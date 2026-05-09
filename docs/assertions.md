# Assertions

Start assertions with `ctx.Expect(actual)`.

Supported matchers:

- `ToEqual(expected)`
- `ToDeepEqual(expected)`
- `ToBeNil()`
- `ToNotBeNil()`
- `ToBeTrue()`
- `ToBeFalse()`
- `ToContain(expected)`
- `ToHaveLen(length)`
- `ToMatchJSON(expected)`
- `ToError()`
- `ToErrorContain(substr)`
- `Not()`

Assertions are fatal and call `testing.T.Helper()` so failures point to the test line.

`ToEqual` uses Go equality for comparable values and falls back to deep comparison for maps, slices, and other non-comparable values.

Structured failures include a first-difference path plus pretty-printed actual and expected values.

`ToMatchJSON` accepts strings, `[]byte`, or Go values. It decodes both sides and compares normalized structures, so object key order does not matter.
