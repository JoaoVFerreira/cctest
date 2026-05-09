# Fabric Mocks

Every `It` receives a fresh `*cctest.Context`.

Available helpers:

- `ctx.Stub()`
- `ctx.MockInit(args...)`
- `ctx.MockInvoke(function, args...)`
- `ctx.MockInvokeWithTransient(function, transient, args...)`
- `ctx.TransactionContext()`
- `ctx.PutState`, `ctx.GetState`, `ctx.DeleteState`, `ctx.HasState`, `ctx.ClearState`
- `ctx.PutJSON`, `ctx.GetJSON`
- `ctx.PutPrivateData`, `ctx.GetPrivateData`, `ctx.ClearPrivateData`
- `ctx.SetClientIdentity`, `ctx.ClientIdentity`
- `ctx.SetTxID`, `ctx.SetTimestamp`
- `ctx.Events()`

The mock is based on Fabric `shimtest.MockStub` with cctest-owned overrides for invocation args, transient data, events, private data deletion, state history, and conservative rich queries.

Supported rich query selector subset:

- field equality: `{"selector":{"owner":"joao"}}`
- dot-path field equality: `{"selector":{"organizacao.@key":"org-1"}}`
- `$eq`, `$ne`
- `$gt`, `$gte`, `$lt`, `$lte` for numbers and strings
- `$exists`
- `$in`, `$nin`, `$all`
- `$elemMatch`
- `$regex`
- top-level `$and` and `$or`
- `sort`, `limit`, and `bookmark`

Unsupported selector operators return an explicit error.
