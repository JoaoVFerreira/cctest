# Phase 1: Core Suite Runner

## Scope

Implement the core test DSL:

- `Describe`
- `Suite`
- `It`
- `BeforeEach`
- `AfterEach`

## Requirements

- Map suites and test cases to standard `testing.T.Run` subtests.
- Support nested `Describe` blocks.
- Run `BeforeEach` hooks from outer suite to inner suite.
- Run `AfterEach` hooks from inner suite to outer suite.
- Ensure `AfterEach` hooks run through deferred cleanup when a test exits early.
- Create a fresh `*Context` for every `It`.

## Deliverables

- Minimal Go module for `github.com/JoaoVFerreira/cctest`.
- Core runner implementation.
- Focused tests for execution order, nested suites, skipped tests, and context isolation.

