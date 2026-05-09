<!-- GSD:project-start source:PROJECT.md -->
## Project

**cctest**

cctest is an open-source Go package for unit testing Hyperledger Fabric chaincodes built with cc-tools. It provides a Jest-like developer experience through Ginkgo/Gomega patterns while running entirely in memory on top of `cc-tools/mock.MockStub`, with no Docker, Fabric network, or CouchDB dependency.

The initial project is intentionally unit-mode only: it targets business logic, transaction behavior, validation paths, permissions, events, deterministic time, fixtures, snapshots, and result assertions. Query/CouchDB harness work is separate future scope, not part of this v1 roadmap.

**Core Value:** Chaincode developers can write fast, deterministic, expressive unit tests for cc-tools business logic using only `go test`.

### Constraints

- **Runtime model**: Unit mode must run fully in memory with `go test` - this is the product's main value and the boundary that keeps tests fast.
- **Domain honesty**: The package must explicitly document that it does not replace Fabric integration tests - avoiding false confidence is part of the design.
- **Compatibility**: cctest v0.x should support cc-tools v1.0.x, with CI enforcing the supported matrix rather than relying on documentation alone.
- **Concurrency**: `Context` is process-local, single-spec, and single-goroutine - shared global contexts across parallel specs are unsupported.
- **Determinism**: Time must come from controlled transaction timestamps, never wall clock behavior inside chaincode tests.
- **Identity consistency**: Caller simulation must update Fabric identity surfaces atomically so permission tests do not exercise impossible states.
- **Dependency policy**: cctest must not vendor cc-tools; consumer projects that vendor dependencies must be tested through a dedicated consumer-vendor-mode CI job.
- **API scope**: Keep the initial API narrow enough to be stable, favoring explicit helpers with strong diagnostics over broad wrappers.
<!-- GSD:project-end -->

<!-- GSD:stack-start source:STACK.md -->
## Technology Stack

Technology stack not yet documented. Will populate after codebase mapping or first phase.
<!-- GSD:stack-end -->

<!-- GSD:conventions-start source:CONVENTIONS.md -->
## Conventions

Conventions not yet established. Will populate as patterns emerge during development.
<!-- GSD:conventions-end -->

<!-- GSD:architecture-start source:ARCHITECTURE.md -->
## Architecture

Architecture not yet mapped. Follow existing patterns found in the codebase.
<!-- GSD:architecture-end -->

<!-- GSD:skills-start source:skills/ -->
## Project Skills

No project skills found. Add skills to any of: `.claude/skills/`, `.agents/skills/`, `.cursor/skills/`, `.github/skills/`, or `.codex/skills/` with a `SKILL.md` index file.
<!-- GSD:skills-end -->

<!-- GSD:workflow-start source:GSD defaults -->
## GSD Workflow Enforcement

Before using Edit, Write, or other file-changing tools, start work through a GSD command so planning artifacts and execution context stay in sync.

Use these entry points:
- `/gsd-quick` for small fixes, doc updates, and ad-hoc tasks
- `/gsd-debug` for investigation and bug fixing
- `/gsd-execute-phase` for planned phase work

Do not make direct repo edits outside a GSD workflow unless the user explicitly asks to bypass it.
<!-- GSD:workflow-end -->



<!-- GSD:profile-start -->
## Developer Profile

> Profile not yet configured. Run `/gsd-profile-user` to generate your developer profile.
> This section is managed by `generate-claude-profile` -- do not edit manually.
<!-- GSD:profile-end -->
