# cctools Adapter

Import the optional adapter:

```go
import cctools "github.com/JoaoVFerreira/cctest/cctools"
```

Available helpers:

- `cctools.WithSetup(func() error)`
- `cctools.WithChaincode(shim.Chaincode)`
- `cctools.From(ctx).StubWrapper()`
- `cctools.From(ctx).Invoke(txName, request)`
- `cctools.From(ctx).InvokeWithTransient(txName, request, transientRequest)`
- `cctools.From(ctx).Query(txName, request)`
- `cctools.From(ctx).QueryWithTransient(txName, request, transientRequest)`
- `cctools.From(ctx).PutAsset(assetType, id, asset)`
- `cctools.From(ctx).GetAsset(assetType, id)`
- `cctools.From(ctx).SearchAssets(selector)`

`Invoke` and `Query` marshal the request as JSON and invoke the configured chaincode with Fabric args shaped as:

```text
[txName, jsonRequest]
```

This matches the common cc-tools test pattern.

`InvokeWithTransient` and `QueryWithTransient` encode `transientRequest` into transient key `@request`, which is the format consumed by `transactions.GetArgs` for private arguments.

The adapter is a nested Go module. Run its tests from the subdirectory:

```bash
cd cctools
go test ./...
```
