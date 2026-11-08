package jsonpath

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type tokenType int

const (
	tokenField tokenType = iota
	tokenIndex
	tokenWildcard
)

type token struct {
	typ   tokenType
	key   string
	index int
}

// ParsePath parses a JSONPath expression (e.g. "$.store.book[0].title" or "items[*].id") into tokens.
func ParsePath(path string) ([]token, error) {
	path = strings.TrimSpace(path)
	if path == "" || path == "$" {
		return nil, nil
	}

	// Strip leading "$." or "$"
	if strings.HasPrefix(path, "$.") {
		path = path[2:]
	} else if strings.HasPrefix(path, "$") {
		path = path[1:]
	}

	var tokens []token
	i := 0
	n := len(path)

	for i < n {
		if path[i] == '.' {
			i++
			continue
		}

		if path[i] == '[' {
			closeIdx := strings.IndexByte(path[i:], ']')
			if closeIdx == -1 {
				return nil, fmt.Errorf("%w: unclosed '['", ErrInvalidPath)
			}
			inside := path[i+1 : i+closeIdx]
			inside = strings.TrimSpace(inside)

			if inside == "*" {
				tokens = append(tokens, token{typ: tokenWildcard})
			} else {
				idx, err := strconv.Atoi(inside)
				if err != nil {
					return nil, fmt.Errorf("%w: invalid array index '%s'", ErrInvalidPath, inside)
				}
				tokens = append(tokens, token{typ: tokenIndex, index: idx})
			}

			i += closeIdx + 1
			continue
		}

		// Field name until next '.' or '['
		start := i
		for i < n && path[i] != '.' && path[i] != '[' {
			i++
		}
		field := strings.TrimSpace(path[start:i])
		if field != "" {
			tokens = append(tokens, token{typ: tokenField, key: field})
		}
	}

	return tokens, nil
}

// Get extracts a value from raw JSON bytes using a JSONPath expression.
func Get(jsonData []byte, path string) (any, error) {
	var parsed any
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}

	return GetFromAny(parsed, path)
}

// GetFromAny extracts a value from an already unmarshaled data structure (map/slice).
func GetFromAny(data any, path string) (any, error) {
	tokens, err := ParsePath(path)
	if err != nil {
		return nil, err
	}

	return evalTokens(data, tokens)
}

func evalTokens(data any, tokens []token) (any, error) {
	if len(tokens) == 0 {
		return data, nil
	}

	curr := tokens[0]
	rest := tokens[1:]

	switch curr.typ {
	case tokenField:
		m, ok := data.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%w: expected object for key '%s'", ErrTypeMismatch, curr.key)
		}
		val, found := m[curr.key]
		if !found {
			return nil, fmt.Errorf("%w: '%s'", ErrNotFound, curr.key)
		}
		return evalTokens(val, rest)

	case tokenIndex:
		slice, ok := data.([]any)
		if !ok {
			return nil, fmt.Errorf("%w: expected array for index [%d]", ErrTypeMismatch, curr.index)
		}
		if curr.index < 0 || curr.index >= len(slice) {
			return nil, fmt.Errorf("%w: index %d, len %d", ErrIndexOutOfBounds, curr.index, len(slice))
		}
		return evalTokens(slice[curr.index], rest)

	case tokenWildcard:
		slice, ok := data.([]any)
		if !ok {
			return nil, fmt.Errorf("%w: expected array for wildcard [*]", ErrTypeMismatch)
		}
		var results []any
		for _, item := range slice {
			res, err := evalTokens(item, rest)
			if err == nil {
				results = append(results, res)
			}
		}
		return results, nil
	}

	return nil, ErrInvalidPath
}

// GetString extracts a string value from raw JSON bytes.
func GetString(jsonData []byte, path string) (string, error) {
	v, err := Get(jsonData, path)
	if err != nil {
		return "", err
	}
	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("%w: expected string, got %T", ErrTypeMismatch, v)
	}
	return s, nil
}

// GetInt extracts an integer value (converting from JSON float64 or int).
func GetInt(jsonData []byte, path string) (int64, error) {
	v, err := Get(jsonData, path)
	if err != nil {
		return 0, err
	}
	switch num := v.(type) {
	case float64:
		return int64(num), nil
	case int:
		return int64(num), nil
	case int64:
		return num, nil
	default:
		return 0, fmt.Errorf("%w: expected numeric value, got %T", ErrTypeMismatch, v)
	}
}

// GetFloat extracts a float64 value.
func GetFloat(jsonData []byte, path string) (float64, error) {
	v, err := Get(jsonData, path)
	if err != nil {
		return 0, err
	}
	switch num := v.(type) {
	case float64:
		return num, nil
	case int:
		return float64(num), nil
	case int64:
		return float64(num), nil
	default:
		return 0, fmt.Errorf("%w: expected float, got %T", ErrTypeMismatch, v)
	}
}

// GetBool extracts a boolean value.
func GetBool(jsonData []byte, path string) (bool, error) {
	v, err := Get(jsonData, path)
	if err != nil {
		return false, err
	}
	b, ok := v.(bool)
	if !ok {
		return false, fmt.Errorf("%w: expected bool, got %T", ErrTypeMismatch, v)
	}
	return b, nil
}

// GetSlice extracts a slice of values.
func GetSlice(jsonData []byte, path string) ([]any, error) {
	v, err := Get(jsonData, path)
	if err != nil {
		return nil, err
	}
	s, ok := v.([]any)
	if !ok {
		return nil, fmt.Errorf("%w: expected array/slice, got %T", ErrTypeMismatch, v)
	}
	return s, nil
}
