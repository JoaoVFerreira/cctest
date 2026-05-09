# Phase 5: cctools Adapter

## Scope

Add optional integration for `github.com/hyperledger-labs/cc-tools`.

## Requirements

- Add `github.com/JoaoVFerreira/cctest/cctools`.
- Support `SetupCC()` style setup.
- Support cctools transaction invocation.
- Support cctools asset helpers.
- Keep core `cctest` independent from cc-tools dependencies.
- Validate behavior against a real cc-tools chaincode sample.

## Deliverables

- `cctools` subpackage.
- Setup, invoke, query, and asset helper APIs.
- Compatibility tests using a representative cc-tools chaincode.

