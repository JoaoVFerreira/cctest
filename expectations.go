package cctest

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/JoaoVFerreira/cctest/internal/diff"
)

// Expectation is a fatal assertion bound to the current test.
type Expectation struct {
	t      testFatalHelper
	actual any
	negate bool
}

type testFatalHelper interface {
	Helper()
	Fatalf(format string, args ...any)
}

func newExpectation(t testFatalHelper, actual any) *Expectation {
	t.Helper()
	return &Expectation{t: t, actual: actual}
}

// Not returns a negated expectation.
func (e *Expectation) Not() *Expectation {
	e.t.Helper()
	return &Expectation{
		t:      e.t,
		actual: e.actual,
		negate: !e.negate,
	}
}

// ToEqual compares comparable values with Go equality and falls back to deep comparison.
func (e *Expectation) ToEqual(expected any) {
	e.t.Helper()
	e.check(equalValues(e.actual, expected), "expected %s to equal %s\n%s", formatValue(e.actual), formatValue(expected), diff.Format(e.actual, expected))
}

// ToDeepEqual compares values using reflect.DeepEqual.
func (e *Expectation) ToDeepEqual(expected any) {
	e.t.Helper()
	e.check(reflect.DeepEqual(e.actual, expected), "expected %s to deep equal %s\n%s", formatValue(e.actual), formatValue(expected), diff.Format(e.actual, expected))
}

// ToBeNil asserts that the actual value is nil.
func (e *Expectation) ToBeNil() {
	e.t.Helper()
	e.check(isNil(e.actual), "expected %s to be nil", formatValue(e.actual))
}

// ToNotBeNil asserts that the actual value is not nil.
func (e *Expectation) ToNotBeNil() {
	e.t.Helper()
	e.check(!isNil(e.actual), "expected value not to be nil")
}

// ToBeTrue asserts that the actual value is true.
func (e *Expectation) ToBeTrue() {
	e.t.Helper()
	e.check(e.actual == true, "expected %s to be true", formatValue(e.actual))
}

// ToBeFalse asserts that the actual value is false.
func (e *Expectation) ToBeFalse() {
	e.t.Helper()
	e.check(e.actual == false, "expected %s to be false", formatValue(e.actual))
}

// ToContain asserts that strings, slices, arrays, or maps contain the expected value.
func (e *Expectation) ToContain(expected any) {
	e.t.Helper()

	ok, err := contains(e.actual, expected)
	if err != nil {
		e.fail("%v", err)
		return
	}

	e.check(ok, "expected %s to contain %s", formatValue(e.actual), formatValue(expected))
}

// ToHaveLen asserts that the actual value has the expected length.
func (e *Expectation) ToHaveLen(length int) {
	e.t.Helper()

	got, ok := valueLen(e.actual)
	if !ok {
		e.fail("expected %s to have length %d, but value has no length", formatValue(e.actual), length)
		return
	}

	e.check(got == length, "expected %s to have length %d, got %d", formatValue(e.actual), length, got)
}

// ToMatchJSON normalizes actual and expected as JSON and compares the decoded values.
func (e *Expectation) ToMatchJSON(expected any) {
	e.t.Helper()

	actualJSON, err := normalizeJSON(e.actual)
	if err != nil {
		e.fail("actual is not valid JSON: %v", err)
		return
	}

	expectedJSON, err := normalizeJSON(expected)
	if err != nil {
		e.fail("expected is not valid JSON: %v", err)
		return
	}

	e.check(reflect.DeepEqual(actualJSON, expectedJSON), "expected JSON %s to match %s\n%s", formatValue(actualJSON), formatValue(expectedJSON), diff.Format(actualJSON, expectedJSON))
}

// ToError asserts that the actual value is a non-nil error.
func (e *Expectation) ToError() {
	e.t.Helper()

	err, ok := e.actual.(error)
	e.check(ok && err != nil, "expected %s to be a non-nil error", formatValue(e.actual))
}

// ToErrorContain asserts that the actual value is a non-nil error containing substr.
func (e *Expectation) ToErrorContain(substr string) {
	e.t.Helper()

	err, ok := e.actual.(error)
	if !ok || err == nil {
		e.fail("expected %s to be a non-nil error containing %q", formatValue(e.actual), substr)
		return
	}

	e.check(strings.Contains(err.Error(), substr), "expected error %q to contain %q", err.Error(), substr)
}

func (e *Expectation) check(ok bool, format string, args ...any) {
	e.t.Helper()

	if e.negate {
		ok = !ok
	}
	if !ok {
		e.fail(format, args...)
	}
}

func (e *Expectation) fail(format string, args ...any) {
	e.t.Helper()

	if e.negate {
		format = "negated assertion failed: " + format
	}
	e.t.Fatalf(format, args...)
}

func isNil(value any) bool {
	if value == nil {
		return true
	}

	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return rv.IsNil()
	default:
		return false
	}
}

func equalValues(actual, expected any) bool {
	if actual == nil || expected == nil {
		return actual == expected
	}

	actualValue := reflect.ValueOf(actual)
	expectedValue := reflect.ValueOf(expected)
	if actualValue.IsValid() &&
		expectedValue.IsValid() &&
		actualValue.Type() == expectedValue.Type() &&
		actualValue.Type().Comparable() {
		return actualValue.Equal(expectedValue)
	}

	return reflect.DeepEqual(actual, expected)
}

func contains(actual any, expected any) (bool, error) {
	if actual == nil {
		return false, errors.New("expected nil to contain a value")
	}

	if text, ok := actual.(string); ok {
		expectedText, ok := expected.(string)
		if !ok {
			return false, fmt.Errorf("expected string containment value, got %T", expected)
		}
		return strings.Contains(text, expectedText), nil
	}

	rv := reflect.ValueOf(actual)
	switch rv.Kind() {
	case reflect.Array, reflect.Slice:
		for i := 0; i < rv.Len(); i++ {
			if reflect.DeepEqual(rv.Index(i).Interface(), expected) {
				return true, nil
			}
		}
		return false, nil
	case reflect.Map:
		expectedValue := reflect.ValueOf(expected)
		if expectedValue.IsValid() && expectedValue.Type().AssignableTo(rv.Type().Key()) {
			return rv.MapIndex(expectedValue).IsValid(), nil
		}

		for _, key := range rv.MapKeys() {
			if reflect.DeepEqual(rv.MapIndex(key).Interface(), expected) {
				return true, nil
			}
		}
		return false, nil
	default:
		return false, fmt.Errorf("expected %s to support containment", formatValue(actual))
	}
}

func valueLen(value any) (int, bool) {
	if value == nil {
		return 0, false
	}

	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		return rv.Len(), true
	default:
		return 0, false
	}
}

func normalizeJSON(value any) (any, error) {
	var raw []byte

	switch v := value.(type) {
	case string:
		raw = []byte(v)
	case []byte:
		raw = v
	default:
		var err error
		raw, err = json.Marshal(value)
		if err != nil {
			return nil, err
		}
	}

	var decoded any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return nil, err
	}

	return decoded, nil
}

func formatValue(value any) string {
	if value == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%#v (%T)", value, value)
}
