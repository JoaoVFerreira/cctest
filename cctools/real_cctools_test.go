package cctools

import (
	"testing"

	"github.com/JoaoVFerreira/cctest"
	"github.com/hyperledger-labs/cc-tools/assets"
	"github.com/hyperledger-labs/cc-tools/transactions"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	peer "github.com/hyperledger/fabric-protos-go/peer"
)

var bookAssetType = assets.AssetType{
	Tag:         "book",
	Label:       "Book",
	Description: "Book asset used by cctest's cc-tools adapter test",
	Props: []assets.AssetProp{
		{
			Tag:      "id",
			Label:    "ID",
			DataType: "string",
			IsKey:    true,
		},
		{
			Tag:      "title",
			Label:    "Title",
			DataType: "string",
			Required: true,
		},
	},
}

type realCCToolsChaincode struct{}

func (c *realCCToolsChaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success(nil)
}

func (c *realCCToolsChaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	result, err := transactions.Run(stub)
	if err != nil {
		return err.GetErrorResponse()
	}
	return shim.Success(result)
}

func TestRealCCToolsCreateAsset(t *testing.T) {
	cctest.Describe(t, "cc-tools", func(s *cctest.Suite) {
		s.It("runs the real createAsset transaction", func(ctx *cctest.Context) {
			res := From(ctx).Invoke("createAsset", map[string]any{
				"asset": []map[string]any{
					{
						"@assetType": "book",
						"id":         "book-1",
						"title":      "Domain-Driven Design",
					},
				},
			})

			ctx.Expect(res.Status).ToEqual(int32(200))

			matches := From(ctx).SearchAssets(map[string]any{"@assetType": "book"})
			ctx.Expect(matches).ToHaveLen(1)
			ctx.Expect(matches[0]["title"]).ToEqual("Domain-Driven Design")
		})
	}, WithSetup(func() error {
		assets.InitDynamicAssetTypeConfig(assets.DynamicAssetType{})
		assets.InitAssetList([]assets.AssetType{bookAssetType})
		transactions.InitTxList([]transactions.Transaction{transactions.CreateAsset})
		return nil
	}), WithChaincode(&realCCToolsChaincode{}))
}

var _ shim.Chaincode = (*realCCToolsChaincode)(nil)
