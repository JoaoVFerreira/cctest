# cctest

## What This Is

cctest is an open-source Go package for unit testing Hyperledger Fabric chaincodes built with cc-tools. It provides a Jest-like developer experience through Ginkgo/Gomega patterns while running entirely in memory on top of `cc-tools/mock.MockStub`, with no Docker, Fabric network, or CouchDB dependency.

The initial project is intentionally unit-mode only: it targets business logic, transaction behavior, validation paths, permissions, events, deterministic time, fixtures, snapshots, and result assertions. Query/CouchDB harness work is separate future scope, not part of this v1 roadmap.

## Core Value

Chaincode developers can write fast, deterministic, expressive unit tests for cc-tools business logic using only `go test`.

## Requirements

### Validated

(None yet - ship to validate)

### Active

- [ ] Provide a standalone Go module for cctest unit-mode testing.
- [ ] Expose a Jest-like test authoring experience using Ginkgo v2 and Gomega without wrapping their core primitives unnecessarily.
- [ ] Build `Context` and `Builder` APIs around a process-local, single-spec, single-goroutine `MockStub` test context.
- [ ] Support deterministic `Invoke`, `Query`, `RawInvoke`, transient data, and per-call caller changes.
- [ ] Inject deterministic transaction timestamps after `MockTransactionStart` so `GetTxTimestamp` can be tested reliably.
- [ ] Simulate caller identity atomically by updating both `Creator` and `SignedProposal`.
- [ ] Wrap Fabric responses in a `Result` API with explicit JSON parsing state, raw payload access, object accessors, array accessors, and clear failures for misuse.
- [ ] Support direct state inspection for specific assets and existence checks.
- [ ] Provide fixtures with ordered fail-fast loading, structured errors, alias-to-key mapping, and event capture.
- [ ] Provide snapshot, restore, and reset helpers that deep-copy state, keys, private state, and endorsement policy state.
- [ ] Provide table-driven test helpers for repeated transaction cases.
- [ ] Provide only high-value custom matchers: field assertions and asset containment diagnostics.
- [ ] Include an internal self-test chaincode so cctest tests do not depend on a consumer project.
- [ ] Include examples and README documentation that demonstrate expected consumer usage.
- [ ] Include CI coverage for unit tests, Ginkgo JUnit output, cc-tools compatibility, and consumer vendor mode.
- [ ] Publish a v0.x-compatible package aligned with the supported cc-tools major/minor policy.

### Out of Scope

- Query/CouchDB/index testing - separate companion project or future milestone, not v1 unit-mode scope.
- Real Fabric integration testing - endorsement, MVCC, validation phase, lifecycle, ordering, consensus, and multi-peer behavior require a Fabric test network such as Microfab or fabric-samples test-network.
- Mandatory in-repo prototype phase - user selected direct standalone module work rather than requiring `chaincode/internal/cctesthelpers/` first.
- Custom JUnit reporter - Ginkgo v2 already supports `--junit-report`.
- Generic Ginkgo wrapper DSL - users should use `Describe`, `It`, `BeforeEach`, and related Ginkgo primitives directly.
- Redundant status/error matchers - chainable `Result` helpers and direct Gomega expectations are sufficient unless the matcher adds better diagnostics.
- Rich query support in unit mode - `GetQueryResult` and CouchDB selector behavior belong outside the unit-mode package.

## Context

The source brief is `cc_unit_v1.md`, a revised post-review plan for a cctest package. It incorporates prior objections around MockStub internals, timestamp control, signed proposal consistency, fixture error reporting, snapshot completeness, dependency policy, CI compatibility, and avoiding overpromising Fabric behavior.

The intended users are developers writing cc-tools chaincodes who need fast local tests for business rules. The package should make common chaincode test workflows concise without hiding the fact that this is unit testing over an in-memory mock, not Fabric integration testing.

The design depends on `github.com/hyperledger-labs/cc-tools` `MockStub`, Fabric chaincode interfaces, Fabric protobuf types, Ginkgo v2, and Gomega. The plan specifically avoids Docker, CouchDB, and Fabric network setup for v1.

The current brief contains one module-path inconsistency: the overview mentions `github.com/JoaoVFerreira/cctest`, while examples and phase notes mention `github.com/goledgerdev/cctest`. The final module path should be settled before Phase 1 implementation begins.

## Constraints

- **Runtime model**: Unit mode must run fully in memory with `go test` - this is the product's main value and the boundary that keeps tests fast.
- **Domain honesty**: The package must explicitly document that it does not replace Fabric integration tests - avoiding false confidence is part of the design.
- **Compatibility**: cctest v0.x should support cc-tools v1.0.x, with CI enforcing the supported matrix rather than relying on documentation alone.
- **Concurrency**: `Context` is process-local, single-spec, and single-goroutine - shared global contexts across parallel specs are unsupported.
- **Determinism**: Time must come from controlled transaction timestamps, never wall clock behavior inside chaincode tests.
- **Identity consistency**: Caller simulation must update Fabric identity surfaces atomically so permission tests do not exercise impossible states.
- **Dependency policy**: cctest must not vendor cc-tools; consumer projects that vendor dependencies must be tested through a dedicated consumer-vendor-mode CI job.
- **API scope**: Keep the initial API narrow enough to be stable, favoring explicit helpers with strong diagnostics over broad wrappers.

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Start as standalone module, not mandatory in-repo prototype | User selected direct standalone module work for this project setup | Pending |
| Limit v1 to unit mode | User selected `cc_unit_v1.md` as the initialization source and kept Query/CouchDB mode separate | Pending |
| Use Ginkgo v2 and Gomega directly | Provides mature BDD structure and matchers without maintaining a custom runner | Pending |
| Replicate the needed MockInvoke lifecycle for controlled timestamps | `MockTransactionStart` overwrites timestamps, so timestamp injection must happen after transaction start and before chaincode invocation | Pending |
| Keep caller identity updates atomic | `Creator` and `SignedProposal` must agree for realistic permission tests | Pending |
| Expose explicit JSON parsing state on Result | Silent parsing hides test failures and makes non-object payload behavior ambiguous | Pending |
| Keep custom matchers minimal | Only matchers with richer diagnostics justify maintenance cost | Pending |
| Use Ginkgo native JUnit output | Avoids maintaining duplicate reporter functionality | Pending |
| Enforce cc-tools compatibility in CI | The package depends on mock internals, so compatibility must fail loudly when dependencies change | Pending |
| Resolve module path before implementation | The brief mentions both `github.com/JoaoVFerreira/cctest` and `github.com/goledgerdev/cctest` | Pending |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `$gsd-transition`):
1. Requirements invalidated? -> Move to Out of Scope with reason
2. Requirements validated? -> Move to Validated with phase reference
3. New requirements emerged? -> Add to Active
4. Decisions to log? -> Add to Key Decisions
5. "What This Is" still accurate? -> Update if drifted

**After each milestone** (via `$gsd-complete-milestone`):
1. Full review of all sections
2. Core Value check - still the right priority?
3. Audit Out of Scope - reasons still valid?
4. Update Context with current state

---
*Last updated: 2026-05-09 after initialization*
