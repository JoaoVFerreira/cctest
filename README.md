# cctest

`cctest` is a Go test helper for Hyperledger Fabric chaincode. It keeps tests compatible with `go test` while adding a small suite DSL, fatal expectations, Fabric mock helpers, and an optional cc-tools adapter.

```go
func TestAssetContract(t *testing.T) {
    cctest.Describe(t, "AssetContract", func(s *cctest.Suite) {
        s.It("reads seeded state", func(ctx *cctest.Context) {
            ctx.PutJSON("asset-1", map[string]any{"id": "asset-1", "owner": "joao"})

            var asset map[string]any
            ctx.GetJSON("asset-1", &asset)

            ctx.Expect(asset["owner"]).ToEqual("joao")
        })
    })
}
```

## Packages

- `github.com/JoaoVFerreira/cctest`: core suite, assertions, Fabric mock context, ledger helpers, events, identity, and query helpers.
- `github.com/JoaoVFerreira/cctest/cctools`: optional helpers for cc-tools style JSON transaction requests and `stubwrapper.StubWrapper`.

## Core API

```go
cctest.Describe(t, "suite", func(s *cctest.Suite) {
    s.BeforeEach(func(ctx *cctest.Context) {})
    s.AfterEach(func(ctx *cctest.Context) {})

    s.It("case", func(ctx *cctest.Context) {
        ctx.Expect(true).ToBeTrue()
    })
})
```

Supported options:

- `WithChaincode(shim.Chaincode)`
- `WithInitArgs(args ...[]byte)`
- `WithMSP(string)`
- `WithChannelID(string)`
- `WithClientIdentity(cctest.ClientIdentity)`
- `WithInitialState(map[string][]byte)`
- `WithInitialJSONState(map[string]any)`
- `WithPrettyOutput(bool)`
- `WithColorOutput(bool)`

## Terminal Output

When tests run with `go test -v`, `cctest` prints a Jest-like reporter with colored `PASS`, `FAIL`, and `SKIP` lines plus a short summary. The native Go `=== RUN` and `--- PASS` lines still appear because suites are real `testing.T` subtests.

Disable the reporter with `CCTEST_PRETTY=0` or per suite with `cctest.WithPrettyOutput(false)`. Disable ANSI colors with `NO_COLOR=1`, `CCTEST_COLOR=0`, or `cctest.WithColorOutput(false)`.

For clean `go test -v` output without Go's duplicated `=== RUN` and `--- PASS` subtest lines, opt in with a package-level `TestMain`:

```go
func TestMain(m *testing.M) {
    cctest.Main(m)
}
```

Set `CCTEST_CLEAN=0` to keep Go's raw verbose output.

## Verification

```bash
go test ./...
(cd cctools && go test ./...)
```

## Compatibility

The core module currently targets:

- `github.com/hyperledger/fabric-chaincode-go v0.0.0-20210603161043-af0e3898842a`
- `github.com/hyperledger/fabric-contract-api-go v1.1.1`
- `github.com/hyperledger/fabric-protos-go v0.0.0-20210528200356-82833ecdac31`

The optional `cctools` adapter is a nested module that targets `github.com/hyperledger-labs/cc-tools v1.0.3`.

## Local Consumption

Until this repository is published, consumers can use local replaces:

```go
require (
    github.com/JoaoVFerreira/cctest v0.0.0
    github.com/JoaoVFerreira/cctest/cctools v0.0.0
)

replace github.com/JoaoVFerreira/cctest => /home/joaov/cctest
replace github.com/JoaoVFerreira/cctest/cctools => /home/joaov/cctest/cctools
```
