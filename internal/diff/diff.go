package diff

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
)

// Format returns a compact structural diff and pretty-printed values.
func Format(actual, expected any) string {
	normalizedActual, actualOK := normalize(actual)
	normalizedExpected, expectedOK := normalize(expected)
	if actualOK && expectedOK {
		return fmt.Sprintf("%s\nactual:\n%s\nexpected:\n%s", firstDifference("$", normalizedActual, normalizedExpected), pretty(actual), pretty(expected))
	}
	return fmt.Sprintf("%s\nactual:\n%s\nexpected:\n%s", firstDifference("$", actual, expected), pretty(actual), pretty(expected))
}

func firstDifference(path string, actual, expected any) string {
	if reflect.DeepEqual(actual, expected) {
		return "no structural diff"
	}

	actualMap, actualMapOK := actual.(map[string]any)
	expectedMap, expectedMapOK := expected.(map[string]any)
	if actualMapOK && expectedMapOK {
		keys := mapKeys(actualMap, expectedMap)
		for _, key := range keys {
			actualValue, actualOK := actualMap[key]
			expectedValue, expectedOK := expectedMap[key]
			keyPath := path + "." + key
			switch {
			case !actualOK:
				return fmt.Sprintf("missing actual value at %s: expected %#v", keyPath, expectedValue)
			case !expectedOK:
				return fmt.Sprintf("unexpected actual value at %s: %#v", keyPath, actualValue)
			default:
				if !reflect.DeepEqual(actualValue, expectedValue) {
					return firstDifference(keyPath, actualValue, expectedValue)
				}
			}
		}
	}

	actualSlice, actualSliceOK := actual.([]any)
	expectedSlice, expectedSliceOK := expected.([]any)
	if actualSliceOK && expectedSliceOK {
		if len(actualSlice) != len(expectedSlice) {
			return fmt.Sprintf("length mismatch at %s: actual %d, expected %d", path, len(actualSlice), len(expectedSlice))
		}
		for i := range actualSlice {
			if !reflect.DeepEqual(actualSlice[i], expectedSlice[i]) {
				return firstDifference(fmt.Sprintf("%s[%d]", path, i), actualSlice[i], expectedSlice[i])
			}
		}
	}

	return fmt.Sprintf("value mismatch at %s: actual %#v (%T), expected %#v (%T)", path, actual, actual, expected, expected)
}

func mapKeys(left, right map[string]any) []string {
	seen := make(map[string]struct{}, len(left)+len(right))
	for key := range left {
		seen[key] = struct{}{}
	}
	for key := range right {
		seen[key] = struct{}{}
	}

	keys := make([]string, 0, len(seen))
	for key := range seen {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func pretty(value any) string {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err == nil {
		return string(raw)
	}
	return fmt.Sprintf("%#v", value)
}

func normalize(value any) (any, bool) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, false
	}
	var out any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, false
	}
	return out, true
}
