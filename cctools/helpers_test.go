package cctools

import (
	"encoding/json"
	"testing"

	"github.com/JoaoVFerreira/cctest"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	peer "github.com/hyperledger/fabric-protos-go/peer"
)

type adapterChaincode struct{}

func (c *adapterChaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return peer.Response{Status: 200}
}

func (c *adapterChaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	fn, args := stub.GetFunctionAndParameters()
	if fn == "secret" {
		transient, err := stub.GetTransient()
		if err != nil {
			return peer.Response{Status: 500, Message: err.Error()}
		}
		return peer.Response{Status: 200, Payload: transient["@request"]}
	}
	if fn != "createAsset" {
		return peer.Response{Status: 404, Message: fn}
	}
	if len(args) != 1 {
		return peer.Response{Status: 400, Message: "expected request"}
	}

	var request map[string]any
	if err := json.Unmarshal([]byte(args[0]), &request); err != nil {
		return peer.Response{Status: 400, Message: err.Error()}
	}
	raw, _ := json.Marshal(request)
	if err := stub.PutState("asset-1", raw); err != nil {
		return peer.Response{Status: 500, Message: err.Error()}
	}
	return peer.Response{Status: 200, Payload: raw}
}

func TestCCToolsHelpers(t *testing.T) {
	setupCalls := 0

	cctest.Describe(t, "cctools", func(s *cctest.Suite) {
		s.It("runs setup and invokes JSON requests", func(ctx *cctest.Context) {
			res := From(ctx).Invoke("createAsset", map[string]any{"id": "asset-1"})

			ctx.Expect(setupCalls).ToEqual(1)
			ctx.Expect(res.Status).ToEqual(int32(200))
			ctx.Expect(res.Payload).ToMatchJSON(map[string]any{"id": "asset-1"})
		})

		s.It("supports direct asset helpers", func(ctx *cctest.Context) {
			helpers := From(ctx)
			helpers.PutAsset("book", "book-1", map[string]any{"title": "Domain-Driven Design"})

			asset := helpers.GetAsset("book", "book-1")
			ctx.Expect(asset["title"]).ToEqual("Domain-Driven Design")

			matches := helpers.SearchAssets(map[string]any{"@assetType": "book"})
			ctx.Expect(matches).ToHaveLen(1)
		})

		s.It("passes transient cc-tools request payloads", func(ctx *cctest.Context) {
			res := From(ctx).InvokeWithTransient("secret", nil, map[string]any{"password": "hidden"})

			ctx.Expect(res.Status).ToEqual(int32(200))
			ctx.Expect(res.Payload).ToMatchJSON(map[string]any{"password": "hidden"})
		})
	}, WithSetup(func() error {
		setupCalls++
		return nil
	}), WithChaincode(&adapterChaincode{}))
}

var _ shim.Chaincode = (*adapterChaincode)(nil)
