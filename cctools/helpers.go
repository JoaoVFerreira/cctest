package cctools

import (
	"encoding/json"

	"github.com/JoaoVFerreira/cctest"
	"github.com/hyperledger-labs/cc-tools/stubwrapper"
	"github.com/hyperledger/fabric-protos-go/ledger/queryresult"
	peer "github.com/hyperledger/fabric-protos-go/peer"
)

// Helpers exposes cc-tools-oriented helpers for a cctest Context.
type Helpers struct {
	ctx *cctest.Context
}

// From creates cc-tools helpers for ctx.
func From(ctx *cctest.Context) *Helpers {
	return &Helpers{ctx: ctx}
}

// StubWrapper returns a cc-tools StubWrapper using the current cctest stub.
func (h *Helpers) StubWrapper() *stubwrapper.StubWrapper {
	return &stubwrapper.StubWrapper{Stub: h.ctx.Stub()}
}

// Invoke marshals request as JSON and invokes txName.
func (h *Helpers) Invoke(txName string, request any) peer.Response {
	return h.invoke(txName, request)
}

// InvokeWithTransient marshals request as public args and transientRequest as transient @request.
func (h *Helpers) InvokeWithTransient(txName string, request any, transientRequest any) peer.Response {
	return h.invokeWithTransient(txName, request, transientRequest)
}

// Query is an alias for Invoke for read-only transactions.
func (h *Helpers) Query(txName string, request any) peer.Response {
	return h.invoke(txName, request)
}

// QueryWithTransient is an alias for InvokeWithTransient for read-only transactions.
func (h *Helpers) QueryWithTransient(txName string, request any, transientRequest any) peer.Response {
	return h.invokeWithTransient(txName, request, transientRequest)
}

// PutAsset writes an asset-shaped JSON object directly into mock state.
func (h *Helpers) PutAsset(assetType, id string, asset map[string]any) {
	copy := make(map[string]any, len(asset)+2)
	for key, value := range asset {
		copy[key] = value
	}
	copy["@assetType"] = assetType
	copy["@key"] = id
	h.ctx.PutJSON(id, copy)
}

// GetAsset reads an asset-shaped JSON object directly from mock state.
func (h *Helpers) GetAsset(assetType, id string) map[string]any {
	var asset map[string]any
	h.ctx.GetJSON(id, &asset)
	if assetType != "" {
		h.ctx.Expect(asset["@assetType"]).ToEqual(assetType)
	}
	return asset
}

// SearchAssets applies the cctest rich-query selector subset and returns JSON objects.
func (h *Helpers) SearchAssets(selector map[string]any) []map[string]any {
	query, err := json.Marshal(map[string]any{"selector": selector})
	h.ctx.Expect(err).ToBeNil()

	iter, err := h.ctx.Stub().GetQueryResult(string(query))
	h.ctx.Expect(err).ToBeNil()
	defer iter.Close()

	var out []map[string]any
	for iter.HasNext() {
		kv, err := iter.Next()
		h.ctx.Expect(err).ToBeNil()
		out = append(out, decodeAsset(h.ctx, kv))
	}
	return out
}

func (h *Helpers) invoke(txName string, request any) peer.Response {
	if request == nil {
		return h.ctx.MockInvoke(txName)
	}
	raw, err := json.Marshal(request)
	h.ctx.Expect(err).ToBeNil()
	return h.ctx.MockInvoke(txName, raw)
}

func (h *Helpers) invokeWithTransient(txName string, request any, transientRequest any) peer.Response {
	var args [][]byte
	if request != nil {
		raw, err := json.Marshal(request)
		h.ctx.Expect(err).ToBeNil()
		args = append(args, raw)
	}

	transient := map[string][]byte{}
	if transientRequest != nil {
		raw, err := json.Marshal(transientRequest)
		h.ctx.Expect(err).ToBeNil()
		transient["@request"] = raw
	}

	return h.ctx.MockInvokeWithTransient(txName, transient, args...)
}

func decodeAsset(ctx *cctest.Context, kv *queryresult.KV) map[string]any {
	ctx.T().Helper()
	var asset map[string]any
	if err := json.Unmarshal(kv.Value, &asset); err != nil {
		ctx.T().Fatalf("decode asset %q: %v", kv.Key, err)
	}
	return asset
}
