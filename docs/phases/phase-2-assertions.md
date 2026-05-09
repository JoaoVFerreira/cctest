# Phase 2: Assertions

## Scope

Implement readable assertions through `ctx.Expect(...)`.

## Requirements

- Add equality, deep equality, nil, boolean, length, containment, error, and JSON assertions.
- Add negation through `.Not()`.
- Make assertion methods call `testing.T.Helper()`.
- Use fatal failures by default to avoid cascading errors.
- Provide clear failure messages and useful diffs for structured values.
- Normalize JSON before comparison instead of comparing raw strings.

## Deliverables

- `Expectation` API.
- JSON assertion helpers.
- Tests covering success and failure behavior for the initial matcher set.

