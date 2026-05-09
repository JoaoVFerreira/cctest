package cctest

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	peer "github.com/hyperledger/fabric-protos-go/peer"
)

// Context is created fresh for each It block.
type Context struct {
	t                  *testing.T
	name               string
	config             Config
	stub               *MockStub
	transactionContext contractapi.TransactionContextInterface
	identity           ClientIdentity
	txID               string
	timestamp          time.Time
	events             *EventLog
}

func newContext(t *testing.T, config Config) *Context {
	t.Helper()

	ctx := &Context{
		t:         t,
		name:      t.Name(),
		config:    config,
		identity:  normalizeIdentity(config),
		txID:      "cctest-tx",
		timestamp: time.Now().UTC(),
	}
	ctx.events = newEventLog(t)
	ctx.stub = newMockStub(t.Name(), config.chaincode, ctx.events)
	ctx.stub.ChannelID = config.channelID
	ctx.applyIdentity()

	txCtx := &contractapi.TransactionContext{}
	txCtx.SetStub(ctx.stub)
	txCtx.SetClientIdentity(newMockClientIdentity(ctx.identity))
	ctx.transactionContext = txCtx

	for _, setup := range config.contextSetups {
		if err := setup(ctx); err != nil {
			t.Fatalf("context setup: %v", err)
		}
	}

	if config.initArgs != nil {
		res := ctx.MockInit(config.initArgs...)
		if res.Status >= 400 {
			t.Fatalf("chaincode init failed: status=%d message=%s", res.Status, res.Message)
		}
	}

	ctx.seedState(config.initialState)
	ctx.seedJSONState(config.initialJSONState)

	return ctx
}

// T returns the underlying Go test handle.
func (c *Context) T() *testing.T {
	c.t.Helper()
	return c.t
}

// Name returns the full Go subtest name for the current test context.
func (c *Context) Name() string {
	return c.name
}

// Expect starts a fatal assertion for the current test context.
func (c *Context) Expect(actual any) *Expectation {
	c.t.Helper()
	return newExpectation(c.t, actual)
}

// Stub returns the Fabric chaincode stub for this test context.
func (c *Context) Stub() shim.ChaincodeStubInterface {
	c.t.Helper()
	return c.stub
}

// MockInvoke invokes the configured chaincode with a function name and raw args.
func (c *Context) MockInvoke(function string, args ...[]byte) peer.Response {
	c.t.Helper()
	return c.MockInvokeWithTransient(function, nil, args...)
}

// MockInvokeWithTransient invokes the configured chaincode with transient data.
func (c *Context) MockInvokeWithTransient(function string, transient map[string][]byte, args ...[]byte) peer.Response {
	c.t.Helper()

	if c.config.chaincode == nil {
		return peer.Response{Status: 500, Message: "cctest: no chaincode configured"}
	}

	invokeArgs := make([][]byte, 0, len(args)+1)
	invokeArgs = append(invokeArgs, []byte(function))
	invokeArgs = append(invokeArgs, args...)

	return c.runChaincode(invokeArgs, transient, c.config.chaincode.Invoke)
}

// MockInit invokes the configured chaincode Init with raw args.
func (c *Context) MockInit(args ...[]byte) peer.Response {
	c.t.Helper()

	if c.config.chaincode == nil {
		return peer.Response{Status: 500, Message: "cctest: no chaincode configured"}
	}

	return c.runChaincode(args, nil, c.config.chaincode.Init)
}

func (c *Context) runChaincode(args [][]byte, transient map[string][]byte, fn func(shim.ChaincodeStubInterface) peer.Response) peer.Response {
	c.stub.setArgs(args)
	c.stub.MockTransactionStart(c.txID)
	c.stub.ChannelID = c.config.channelID
	c.applyIdentity()
	if ts, err := ptypes.TimestampProto(c.timestamp); err == nil {
		c.stub.TxTimestamp = ts
	}

	c.stub.TransientMap = nil
	if transient != nil {
		if err := c.stub.SetTransient(cloneBytesMap(transient)); err != nil {
			c.stub.MockTransactionEnd(c.txID)
			return peer.Response{Status: 500, Message: err.Error()}
		}
	}

	defer func() {
		c.stub.TransientMap = nil
		c.stub.MockTransactionEnd(c.txID)
	}()

	return fn(c.stub)
}

// TransactionContext returns a contractapi-compatible transaction context.
func (c *Context) TransactionContext() contractapi.TransactionContextInterface {
	c.t.Helper()
	return c.transactionContext
}

// PutState writes public ledger state for this context.
func (c *Context) PutState(key string, value []byte) {
	c.t.Helper()
	c.withSetupTransaction(func() {
		if err := c.stub.PutState(key, cloneBytes(value)); err != nil {
			c.t.Fatalf("PutState(%q): %v", key, err)
		}
	})
}

// GetState reads public ledger state for this context.
func (c *Context) GetState(key string) []byte {
	c.t.Helper()
	value, err := c.stub.GetState(key)
	if err != nil {
		c.t.Fatalf("GetState(%q): %v", key, err)
	}
	return cloneBytes(value)
}

// DeleteState deletes public ledger state for this context.
func (c *Context) DeleteState(key string) {
	c.t.Helper()
	c.withSetupTransaction(func() {
		if err := c.stub.DelState(key); err != nil {
			c.t.Fatalf("DeleteState(%q): %v", key, err)
		}
	})
}

// HasState returns true when a key exists in public state.
func (c *Context) HasState(key string) bool {
	c.t.Helper()
	value, err := c.stub.GetState(key)
	if err != nil {
		c.t.Fatalf("HasState(%q): %v", key, err)
	}
	return value != nil
}

// ClearState deletes every public state key.
func (c *Context) ClearState() {
	c.t.Helper()
	c.stub.clearState()
}

// PutJSON marshals and writes public ledger state.
func (c *Context) PutJSON(key string, value any) {
	c.t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		c.t.Fatalf("PutJSON(%q): %v", key, err)
	}
	c.PutState(key, raw)
}

// GetJSON reads public ledger state and unmarshals it into out.
func (c *Context) GetJSON(key string, out any) {
	c.t.Helper()
	raw := c.GetState(key)
	if raw == nil {
		c.t.Fatalf("GetJSON(%q): key not found", key)
	}
	if err := json.Unmarshal(raw, out); err != nil {
		c.t.Fatalf("GetJSON(%q): %v", key, err)
	}
}

// PutPrivateData writes private data state for this context.
func (c *Context) PutPrivateData(collection, key string, value []byte) {
	c.t.Helper()
	if err := c.stub.PutPrivateData(collection, key, cloneBytes(value)); err != nil {
		c.t.Fatalf("PutPrivateData(%q, %q): %v", collection, key, err)
	}
}

// GetPrivateData reads private data state for this context.
func (c *Context) GetPrivateData(collection, key string) []byte {
	c.t.Helper()
	value, err := c.stub.GetPrivateData(collection, key)
	if err != nil {
		c.t.Fatalf("GetPrivateData(%q, %q): %v", collection, key, err)
	}
	return cloneBytes(value)
}

// ClearPrivateData deletes every key in a private data collection.
func (c *Context) ClearPrivateData(collection string) {
	c.t.Helper()
	delete(c.stub.PvtState, collection)
}

// SetClientIdentity updates the identity exposed by Stub and TransactionContext.
func (c *Context) SetClientIdentity(identity ClientIdentity) {
	c.t.Helper()
	c.identity = normalizeIdentity(Config{mspID: c.config.mspID, identity: identity})
	c.applyIdentity()
	if settable, ok := c.transactionContext.(contractapi.SettableTransactionContextInterface); ok {
		settable.SetClientIdentity(newMockClientIdentity(c.identity))
	}
}

// ClientIdentity returns the configured client identity.
func (c *Context) ClientIdentity() ClientIdentity {
	c.t.Helper()
	return c.identity.Clone()
}

// SetTxID updates the transaction ID used by MockInvoke.
func (c *Context) SetTxID(txID string) {
	c.t.Helper()
	c.txID = txID
}

// SetTimestamp updates the transaction timestamp used by MockInvoke.
func (c *Context) SetTimestamp(ts time.Time) {
	c.t.Helper()
	c.timestamp = ts.UTC()
}

// Events returns the event log for this context.
func (c *Context) Events() *EventLog {
	c.t.Helper()
	return c.events
}

func (c *Context) seedState(state map[string][]byte) {
	for key, value := range state {
		c.PutState(key, value)
	}
}

func (c *Context) seedJSONState(state map[string]any) {
	for key, value := range state {
		raw, err := json.Marshal(value)
		if err != nil {
			c.t.Fatalf("initial JSON state %q: %v", key, err)
		}
		c.PutState(key, raw)
	}
}

func (c *Context) withSetupTransaction(fn func()) {
	c.stub.MockTransactionStart("cctest-setup")
	defer c.stub.MockTransactionEnd("cctest-setup")
	fn()
}

func (c *Context) applyIdentity() {
	creator, certPEM, err := serializedIdentity(c.identity)
	if err != nil {
		c.t.Fatalf("client identity: %v", err)
	}
	c.identity.CertPEM = certPEM
	c.stub.Creator = creator
}

func normalizeIdentity(config Config) ClientIdentity {
	identity := config.identity.Clone()
	if identity.MSPID == "" {
		identity.MSPID = config.mspID
	}
	if identity.MSPID == "" {
		identity.MSPID = "Org1MSP"
	}
	if identity.ID == "" {
		identity.ID = "cctest-user"
	}
	if identity.Attributes == nil {
		identity.Attributes = map[string]string{}
	}
	return identity
}

var _ contractapi.TransactionContextInterface = (*contractapi.TransactionContext)(nil)
var _ cid.ClientIdentity = (*mockClientIdentity)(nil)
