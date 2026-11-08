package gotemplatekit

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

// FilterFunc transforms a string value with an optional argument.
type FilterFunc func(input string, arg string) string

// Engine renders template strings with variables and filter pipelines.
type Engine struct {
	filters map[string]FilterFunc
	strict  bool
	mu      sync.RWMutex
}

// New creates an Engine populated with built-in standard filters.
func New(strict bool) *Engine {
	e := &Engine{
		filters: make(map[string]FilterFunc),
		strict:  strict,
	}

	// Register standard built-in filters
	e.RegisterFilter("upper", func(in, arg string) string {
		return strings.ToUpper(in)
	})
	e.RegisterFilter("lower", func(in, arg string) string {
		return strings.ToLower(in)
	})
	e.RegisterFilter("trim", func(in, arg string) string {
		return strings.TrimSpace(in)
	})
	e.RegisterFilter("default", func(in, arg string) string {
		if in == "" {
			return arg
		}
		return in
	})
	e.RegisterFilter("date", func(in, format string) string {
		if format == "" {
			format = "2006-01-02 15:04:05"
		}
		t, err := time.Parse(time.RFC3339, in)
		if err == nil {
			return t.Format(format)
		}
		return in
	})

	return e
}

// RegisterFilter adds a custom filter function to the engine.
func (e *Engine) RegisterFilter(name string, fn FilterFunc) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.filters[strings.ToLower(name)] = fn
}

// Render processes template string using data context (supports map[string]any or struct).
func (e *Engine) Render(tmpl string, data any) (string, error) {
	var result strings.Builder

	cursor := 0
	for {
		start := strings.Index(tmpl[cursor:], "{{")
		if start == -1 {
			result.WriteString(tmpl[cursor:])
			break
		}
		start += cursor
		result.WriteString(tmpl[cursor:start])

		end := strings.Index(tmpl[start:], "}}")
		if end == -1 {
			// Unclosed tag
			result.WriteString(tmpl[start:])
			break
		}
		end += start

		tagContent := strings.TrimSpace(tmpl[start+2 : end])
		val, err := e.evaluateTag(tagContent, data)
		if err != nil {
			return "", err
		}
		result.WriteString(val)

		cursor = end + 2
	}

	return result.String(), nil
}

func (e *Engine) evaluateTag(tag string, data any) (string, error) {
	parts := strings.Split(tag, "|")
	varName := strings.TrimSpace(parts[0])

	valStr, exists := resolveVariable(varName, data)
	if !exists && e.strict {
		return "", fmt.Errorf("%w: %q", ErrMissingVariable, varName)
	}

	current := valStr

	// Apply filter pipeline
	e.mu.RLock()
	defer e.mu.RUnlock()

	for i := 1; i < len(parts); i++ {
		pipePart := strings.TrimSpace(parts[i])
		if pipePart == "" {
			continue
		}

		filterName := pipePart
		filterArg := ""

		if idx := strings.Index(pipePart, ":"); idx != -1 {
			filterName = strings.TrimSpace(pipePart[:idx])
			filterArg = strings.Trim(strings.TrimSpace(pipePart[idx+1:]), `"'`)
		}

		fn, ok := e.filters[strings.ToLower(filterName)]
		if !ok {
			if e.strict {
				return "", fmt.Errorf("%w: %q", ErrUnknownFilter, filterName)
			}
			continue
		}

		current = fn(current, filterArg)
	}

	return current, nil
}

func resolveVariable(name string, data any) (string, bool) {
	if data == nil {
		return "", false
	}

	// Map lookup
	if m, ok := data.(map[string]any); ok {
		if val, exists := m[name]; exists {
			return formatAny(val), true
		}
		return "", false
	}

	// Struct reflection
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	if v.Kind() == reflect.Struct {
		t := v.Type()
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if strings.EqualFold(field.Name, name) {
				return formatAny(v.Field(i).Interface()), true
			}
			if jsonTag := field.Tag.Get("json"); jsonTag != "" {
				tagName := strings.Split(jsonTag, ",")[0]
				if tagName == name {
					return formatAny(v.Field(i).Interface()), true
				}
			}
		}
	}

	return "", false
}

func formatAny(val any) string {
	if val == nil {
		return ""
	}
	switch v := val.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}
