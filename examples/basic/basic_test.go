package basic_test

import (
	"encoding/json"
	"testing"

	"github.com/JoaoVFerreira/cctest"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	peer "github.com/hyperledger/fabric-protos-go/peer"
)

type assetChaincode struct{}

func (c *assetChaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return peer.Response{Status: 200}
}

func (c *assetChaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	fn, args := stub.GetFunctionAndParameters()
	switch fn {
	case "createAsset":
		asset := map[string]any{"id": args[0], "owner": args[1]}
		raw, _ := json.Marshal(asset)
		if err := stub.PutState(args[0], raw); err != nil {
			return peer.Response{Status: 500, Message: err.Error()}
		}
		_ = stub.SetEvent("AssetCreated", []byte(args[0]))
		return peer.Response{Status: 200, Payload: raw}
	default:
		return peer.Response{Status: 404, Message: fn}
	}
}

func TestAssetChaincode(t *testing.T) {
	cctest.Describe(t, "AssetChaincode", func(s *cctest.Suite) {
		s.It("creates an asset", func(ctx *cctest.Context) {
			res := ctx.MockInvoke("createAsset", []byte("asset-1"), []byte("joao"))

			ctx.Expect(res.Status).ToEqual(int32(200))
			ctx.Expect(res.Payload).ToMatchJSON(map[string]any{"id": "asset-1", "owner": "joao"})
			ctx.Events().ExpectEmitted("AssetCreated")
		})
	}, cctest.WithChaincode(&assetChaincode{}))
}

var _ shim.Chaincode = (*assetChaincode)(nil)
