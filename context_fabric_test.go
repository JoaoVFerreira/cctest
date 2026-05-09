package cctest

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/ledger/queryresult"
	peer "github.com/hyperledger/fabric-protos-go/peer"
)

type testChaincode struct{}

func (c *testChaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {
	args := stub.GetStringArgs()
	if len(args) > 0 && args[0] == "init" {
		if err := stub.PutState("initialized", []byte("true")); err != nil {
			return peer.Response{Status: 500, Message: err.Error()}
		}
	}
	return peer.Response{Status: 200}
}

func (c *testChaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	fn, args := stub.GetFunctionAndParameters()
	switch fn {
	case "put":
		if len(args) != 2 {
			return peer.Response{Status: 400, Message: "expected key and value"}
		}
		if err := stub.PutState(args[0], []byte(args[1])); err != nil {
			return peer.Response{Status: 500, Message: err.Error()}
		}
		return peer.Response{Status: 200}
	case "event":
		if err := stub.SetEvent("created", []byte("asset-1")); err != nil {
			return peer.Response{Status: 500, Message: err.Error()}
		}
		return peer.Response{Status: 200}
	case "metadata":
		mspID, err := cid.GetMSPID(stub)
		if err != nil {
			return peer.Response{Status: 500, Message: err.Error()}
		}
		role, found, err := cid.GetAttributeValue(stub, "role")
		if err != nil {
			return peer.Response{Status: 500, Message: err.Error()}
		}
		ts, err := stub.GetTxTimestamp()
		if err != nil {
			return peer.Response{Status: 500, Message: err.Error()}
		}
		payload, _ := json.Marshal(map[string]any{
			"msp":   mspID,
			"role":  role,
			"found": found,
			"txID":  stub.GetTxID(),
			"sec":   ts.Seconds,
		})
		return peer.Response{Status: 200, Payload: payload}
	case "transient":
		transient, err := stub.GetTransient()
		if err != nil {
			return peer.Response{Status: 500, Message: err.Error()}
		}
		return peer.Response{Status: 200, Payload: transient["@request"]}
	default:
		return peer.Response{Status: 404, Message: fn}
	}
}

func TestFabricContextLedgerAndInvoke(t *testing.T) {
	Describe(t, "Fabric", func(s *Suite) {
		s.It("isolates state and invokes chaincode", func(ctx *Context) {
			ctx.PutState("asset-1", []byte("before"))
			ctx.Expect(string(ctx.GetState("asset-1"))).ToEqual("before")
			ctx.Expect(ctx.HasState("asset-1")).ToBeTrue()

			res := ctx.MockInvoke("put", []byte("asset-1"), []byte("after"))

			ctx.Expect(res.Status).ToEqual(int32(200))
			ctx.Expect(string(ctx.GetState("asset-1"))).ToEqual("after")
			ctx.DeleteState("asset-1")
			ctx.Expect(ctx.HasState("asset-1")).ToBeFalse()
		})

		s.It("gets fresh state for each It", func(ctx *Context) {
			ctx.Expect(ctx.HasState("asset-1")).ToBeFalse()
		})
	}, WithChaincode(&testChaincode{}))
}

func TestFabricContextJSONPrivateDataAndEvents(t *testing.T) {
	Describe(t, "Fabric", func(s *Suite) {
		s.It("handles JSON and private data helpers", func(ctx *Context) {
			ctx.PutJSON("asset-json", map[string]any{"id": "asset-json", "count": 2})

			var got map[string]any
			ctx.GetJSON("asset-json", &got)
			ctx.Expect(got["id"]).ToEqual("asset-json")

			ctx.PutPrivateData("collection", "secret", []byte("value"))
			ctx.Expect(string(ctx.GetPrivateData("collection", "secret"))).ToEqual("value")
			ctx.ClearPrivateData("collection")
			ctx.Expect(ctx.GetPrivateData("collection", "secret")).ToBeNil()
		})

		s.It("captures emitted events", func(ctx *Context) {
			res := ctx.MockInvoke("event")

			ctx.Expect(res.Status).ToEqual(int32(200))
			ctx.Events().ExpectEmitted("created")
			ctx.Events().ExpectPayload("created", []byte("asset-1"))
		})
	}, WithChaincode(&testChaincode{}))
}

func TestFabricIdentityMetadataAndInitialState(t *testing.T) {
	ts := time.Date(2025, 2, 3, 4, 5, 6, 0, time.UTC)

	Describe(t, "Fabric", func(s *Suite) {
		s.It("exposes identity and transaction metadata", func(ctx *Context) {
			ctx.SetTxID("tx-123")
			ctx.SetTimestamp(ts)

			res := ctx.MockInvoke("metadata")

			ctx.Expect(res.Status).ToEqual(int32(200))
			ctx.Expect(res.Payload).ToMatchJSON(map[string]any{
				"msp":   "Org2MSP",
				"role":  "auditor",
				"found": true,
				"txID":  "tx-123",
				"sec":   float64(ts.Unix()),
			})

			mspID, err := ctx.TransactionContext().GetClientIdentity().GetMSPID()
			ctx.Expect(err).ToBeNil()
			ctx.Expect(mspID).ToEqual("Org2MSP")

			stubMSP, err := cid.GetMSPID(ctx.Stub())
			ctx.Expect(err).ToBeNil()
			ctx.Expect(stubMSP).ToEqual("Org2MSP")
			ctx.Expect(cid.AssertAttributeValue(ctx.Stub(), "role", "auditor")).ToBeNil()

			cidID, err := cid.GetID(ctx.Stub())
			ctx.Expect(err).ToBeNil()
			ctx.Expect(cidID).Not().ToEqual("")

			cert, err := cid.GetX509Certificate(ctx.Stub())
			ctx.Expect(err).ToBeNil()
			ctx.Expect(cert.Subject.CommonName).ToEqual("alice")
		})

		s.It("seeds initial state per test", func(ctx *Context) {
			ctx.Expect(string(ctx.GetState("seed"))).ToEqual("value")
			var asset map[string]any
			ctx.GetJSON("json-seed", &asset)
			ctx.Expect(asset["@assetType"]).ToEqual("book")
		})
	}, WithChaincode(&testChaincode{}),
		WithInitialState(map[string][]byte{"seed": []byte("value")}),
		WithInitialJSONState(map[string]any{"json-seed": map[string]any{"@assetType": "book"}}),
		WithClientIdentity(ClientIdentity{
			MSPID:      "Org2MSP",
			ID:         "alice",
			Attributes: map[string]string{"role": "auditor"},
		}))
}

func TestFabricQueriesAndCompositeKeys(t *testing.T) {
	Describe(t, "Fabric", func(s *Suite) {
		s.It("supports rich query selector subset", func(ctx *Context) {
			ctx.PutJSON("asset-1", map[string]any{"@assetType": "book", "owner": "joao", "count": 2})
			ctx.PutJSON("asset-2", map[string]any{"@assetType": "book", "owner": "maria", "count": 5})
			ctx.PutJSON("asset-3", map[string]any{"@assetType": "car", "owner": "joao", "count": 7})

			iter, err := ctx.Stub().GetQueryResult(`{"selector":{"@assetType":"book","count":{"$gte":2}}}`)
			ctx.Expect(err).ToBeNil()
			ctx.Expect(collectKeys(ctx, iter)).ToDeepEqual([]string{"asset-1", "asset-2"})

			_, err = ctx.Stub().GetQueryResult(`{"selector":{"count":{"$unknown":"2"}}}`)
			ctx.Expect(err).ToErrorContain("unsupported")
		})

		s.It("supports composite key and private data range helpers", func(ctx *Context) {
			key1, err := ctx.Stub().CreateCompositeKey("asset", []string{"book", "1"})
			ctx.Expect(err).ToBeNil()
			key2, err := ctx.Stub().CreateCompositeKey("asset", []string{"book", "2"})
			ctx.Expect(err).ToBeNil()
			ctx.PutState(key1, []byte(`{"id":"1"}`))
			ctx.PutState(key2, []byte(`{"id":"2"}`))

			iter, err := ctx.Stub().GetStateByPartialCompositeKey("asset", []string{"book"})
			ctx.Expect(err).ToBeNil()
			ctx.Expect(collectKeys(ctx, iter)).ToDeepEqual([]string{key1, key2})

			ctx.PutPrivateData("collection", "a", []byte(`{"id":"a"}`))
			ctx.PutPrivateData("collection", "b", []byte(`{"id":"b"}`))
			privateIter, err := ctx.Stub().GetPrivateDataByRange("collection", "a", "z")
			ctx.Expect(err).ToBeNil()
			ctx.Expect(collectKeys(ctx, privateIter)).ToDeepEqual([]string{"a", "b"})
		})
	})
}

func TestFabricInitTransientHistoryAndAdvancedQueries(t *testing.T) {
	Describe(t, "Fabric", func(s *Suite) {
		s.It("runs chaincode init and transient requests", func(ctx *Context) {
			ctx.Expect(string(ctx.GetState("initialized"))).ToEqual("true")

			res := ctx.MockInvokeWithTransient("transient", map[string][]byte{
				"@request": []byte(`{"secret":"value"}`),
			})

			ctx.Expect(res.Status).ToEqual(int32(200))
			ctx.Expect(res.Payload).ToMatchJSON(map[string]any{"secret": "value"})
		})

		s.It("records public state history", func(ctx *Context) {
			ctx.SetTxID("tx-1")
			ctx.MockInvoke("put", []byte("asset-history"), []byte(`{"version":1}`))
			ctx.SetTxID("tx-2")
			ctx.MockInvoke("put", []byte("asset-history"), []byte(`{"version":2}`))
			ctx.DeleteState("asset-history")

			iter, err := ctx.Stub().GetHistoryForKey("asset-history")
			ctx.Expect(err).ToBeNil()
			defer iter.Close()

			var txIDs []string
			var deletes []bool
			for iter.HasNext() {
				item, err := iter.Next()
				ctx.Expect(err).ToBeNil()
				txIDs = append(txIDs, item.TxId)
				deletes = append(deletes, item.IsDelete)
			}

			ctx.Expect(txIDs).ToDeepEqual([]string{"tx-1", "tx-2", "cctest-setup"})
			ctx.Expect(deletes).ToDeepEqual([]bool{false, false, true})
		})

		s.It("supports dot paths, arrays, sort, limit, and bookmarks", func(ctx *Context) {
			ctx.PutJSON("asset-1", map[string]any{
				"@assetType":  "bem",
				"status":      "ready",
				"setores":     []any{"juridico", "vendas"},
				"organizacao": map[string]any{"@key": "org-1"},
				"ordem":       2,
			})
			ctx.PutJSON("asset-2", map[string]any{
				"@assetType":  "bem",
				"status":      "ready",
				"setores":     []any{"financeiro"},
				"organizacao": map[string]any{"@key": "org-2"},
				"ordem":       1,
			})
			ctx.PutJSON("asset-3", map[string]any{
				"@assetType":  "bem",
				"status":      "blocked",
				"setores":     []any{"juridico"},
				"organizacao": map[string]any{"@key": "org-1"},
				"ordem":       3,
			})

			iter, err := ctx.Stub().GetQueryResult(`{
				"selector":{
					"@assetType":"bem",
					"organizacao.@key":"org-1",
					"setores":{"$elemMatch":{"$eq":"juridico"}},
					"status":{"$in":["ready","blocked"]}
				},
				"sort":[{"ordem":"desc"}],
				"limit":1
			}`)
			ctx.Expect(err).ToBeNil()
			ctx.Expect(collectKeys(ctx, iter)).ToDeepEqual([]string{"asset-3"})

			page, metadata, err := ctx.Stub().GetQueryResultWithPagination(`{
				"selector":{"@assetType":"bem"},
				"sort":[{"ordem":"asc"}]
			}`, 2, "")
			ctx.Expect(err).ToBeNil()
			ctx.Expect(collectKeys(ctx, page)).ToDeepEqual([]string{"asset-2", "asset-1"})
			ctx.Expect(metadata.Bookmark).ToEqual("asset-1")
		})
	}, WithChaincode(&testChaincode{}), WithInitArgs([]byte("init")))
}

func collectKeys(ctx *Context, iter shim.StateQueryIteratorInterface) []string {
	ctx.T().Helper()
	var keys []string
	for iter.HasNext() {
		kv, err := iter.Next()
		ctx.Expect(err).ToBeNil()
		keys = append(keys, kv.Key)
	}
	ctx.Expect(iter.Close()).ToBeNil()
	return keys
}

var _ shim.Chaincode = (*testChaincode)(nil)
var _ = (*queryresult.KV)(nil)
