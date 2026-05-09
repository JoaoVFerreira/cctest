package cctest

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

func matchesSelector(doc map[string]any, selector map[string]any) (bool, error) {
	for field, condition := range selector {
		switch field {
		case "$and":
			parts, ok := condition.([]any)
			if !ok {
				return false, fmt.Errorf("unsupported rich query: $and must be an array")
			}
			for _, part := range parts {
				partSelector, ok := part.(map[string]any)
				if !ok {
					return false, fmt.Errorf("unsupported rich query: $and entries must be selectors")
				}
				matched, err := matchesSelector(doc, partSelector)
				if err != nil || !matched {
					return matched, err
				}
			}
		case "$or":
			parts, ok := condition.([]any)
			if !ok {
				return false, fmt.Errorf("unsupported rich query: $or must be an array")
			}
			anyMatched := false
			for _, part := range parts {
				partSelector, ok := part.(map[string]any)
				if !ok {
					return false, fmt.Errorf("unsupported rich query: $or entries must be selectors")
				}
				matched, err := matchesSelector(doc, partSelector)
				if err != nil {
					return false, err
				}
				if matched {
					anyMatched = true
					break
				}
			}
			if !anyMatched {
				return false, nil
			}
		default:
			values := valuesAtPath(doc, field)
			matched, err := matchesConditionValues(values, condition)
			if err != nil || !matched {
				return matched, err
			}
		}
	}
	return true, nil
}

func matchesCondition(value any, exists bool, condition any) (bool, error) {
	return matchesConditionValues(singleValue(value, exists), condition)
}

func matchesConditionValues(values []any, condition any) (bool, error) {
	exists := len(values) > 0
	ops, ok := condition.(map[string]any)
	if !ok || !isOperatorMap(ops) {
		return anyValueMatches(values, condition), nil
	}

	for op, expected := range ops {
		switch op {
		case "$eq":
			if !anyValueMatches(values, expected) {
				return false, nil
			}
		case "$ne":
			if anyValueMatches(values, expected) {
				return false, nil
			}
		case "$gt", "$gte", "$lt", "$lte":
			if !anyValueCompares(values, expected, op) {
				return false, nil
			}
		case "$exists":
			want, ok := expected.(bool)
			if !ok || exists != want {
				return false, nil
			}
		case "$in":
			expectedValues, ok := anySlice(expected)
			if !ok || !anyIn(values, expectedValues) {
				return false, nil
			}
		case "$nin":
			expectedValues, ok := anySlice(expected)
			if !ok || anyIn(values, expectedValues) {
				return false, nil
			}
		case "$all":
			expectedValues, ok := anySlice(expected)
			if !ok || !allIn(values, expectedValues) {
				return false, nil
			}
		case "$elemMatch":
			if !elemMatch(values, expected) {
				return false, nil
			}
		case "$regex":
			pattern, ok := expected.(string)
			if !ok || !regexMatch(values, pattern) {
				return false, nil
			}
		default:
			return false, fmt.Errorf("unsupported rich query selector operator %s", op)
		}
	}
	return true, nil
}

func valuesAtPath(doc map[string]any, path string) []any {
	parts := strings.Split(path, ".")
	return collectValues(any(doc), parts)
}

func collectValues(value any, parts []string) []any {
	if len(parts) == 0 {
		return flattenValue(value)
	}

	switch typed := value.(type) {
	case map[string]any:
		next, ok := typed[parts[0]]
		if !ok {
			return nil
		}
		return collectValues(next, parts[1:])
	case []any:
		var out []any
		for _, item := range typed {
			out = append(out, collectValues(item, parts)...)
		}
		return out
	default:
		return nil
	}
}

func flattenValue(value any) []any {
	if items, ok := value.([]any); ok {
		return items
	}
	return []any{value}
}

func singleValue(value any, exists bool) []any {
	if !exists {
		return nil
	}
	return flattenValue(value)
}

func isOperatorMap(value map[string]any) bool {
	for key := range value {
		if strings.HasPrefix(key, "$") {
			return true
		}
	}
	return false
}

func anyValueMatches(values []any, expected any) bool {
	for _, value := range values {
		if reflect.DeepEqual(value, expected) {
			return true
		}
		if nested, ok := value.([]any); ok {
			for _, item := range nested {
				if reflect.DeepEqual(item, expected) {
					return true
				}
			}
		}
	}
	return false
}

func anyValueCompares(values []any, expected any, op string) bool {
	for _, value := range values {
		cmp, ok := compareValues(value, expected)
		if !ok {
			continue
		}
		switch op {
		case "$gt":
			if cmp > 0 {
				return true
			}
		case "$gte":
			if cmp >= 0 {
				return true
			}
		case "$lt":
			if cmp < 0 {
				return true
			}
		case "$lte":
			if cmp <= 0 {
				return true
			}
		}
	}
	return false
}

func anySlice(value any) ([]any, bool) {
	if out, ok := value.([]any); ok {
		return out, true
	}
	return nil, false
}

func anyIn(values, expectedValues []any) bool {
	for _, expected := range expectedValues {
		if anyValueMatches(values, expected) {
			return true
		}
	}
	return false
}

func allIn(values, expectedValues []any) bool {
	for _, expected := range expectedValues {
		if !anyValueMatches(values, expected) {
			return false
		}
	}
	return true
}

func elemMatch(values []any, expected any) bool {
	for _, value := range values {
		items, ok := value.([]any)
		if !ok {
			items = []any{value}
		}
		for _, item := range items {
			expectedMap, expectedIsMap := expected.(map[string]any)
			itemMap, itemIsMap := item.(map[string]any)
			if expectedIsMap && itemIsMap && !isOperatorMap(expectedMap) {
				matched, err := matchesSelector(itemMap, expectedMap)
				if err == nil && matched {
					return true
				}
				continue
			}
			matched, err := matchesConditionValues([]any{item}, expected)
			if err == nil && matched {
				return true
			}
		}
	}
	return false
}

func regexMatch(values []any, pattern string) bool {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}
	for _, value := range values {
		text, ok := value.(string)
		if ok && re.MatchString(text) {
			return true
		}
	}
	return false
}

func compareValues(a, b any) (int, bool) {
	if af, ok := toFloat64(a); ok {
		bf, ok := toFloat64(b)
		if !ok {
			return 0, false
		}
		switch {
		case af < bf:
			return -1, true
		case af > bf:
			return 1, true
		default:
			return 0, true
		}
	}

	as, ok := a.(string)
	if !ok {
		return 0, false
	}
	bs, ok := b.(string)
	if !ok {
		return 0, false
	}
	switch {
	case as < bs:
		return -1, true
	case as > bs:
		return 1, true
	default:
		return 0, true
	}
}

func toFloat64(value any) (float64, bool) {
	switch v := value.(type) {
	case int:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case float32:
		return float64(v), true
	case float64:
		return v, true
	default:
		return 0, false
	}
}
