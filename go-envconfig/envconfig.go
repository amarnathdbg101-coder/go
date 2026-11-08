package envconfig

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// LookupFunc defines a function signature for looking up environment variables.
type LookupFunc func(key string) (string, bool)

// Process populates the target struct pointer using environment variables from os.LookupEnv.
func Process(prefix string, target any) error {
	return ProcessWithLookup(prefix, target, os.LookupEnv)
}

// ProcessWithLookup populates the target struct pointer using a custom lookup function.
func ProcessWithLookup(prefix string, target any, lookup LookupFunc) error {
	if lookup == nil {
		lookup = os.LookupEnv
	}

	val := reflect.ValueOf(target)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return ErrNotPointerToStruct
	}

	elem := val.Elem()
	if elem.Kind() != reflect.Struct {
		return ErrNotPointerToStruct
	}

	return processStruct(prefix, elem, lookup)
}

func processStruct(prefix string, structVal reflect.Value, lookup LookupFunc) error {
	structType := structVal.Type()

	for i := 0; i < structVal.NumField(); i++ {
		fieldVal := structVal.Field(i)
		fieldType := structType.Field(i)

		// Skip unexported fields
		if !fieldVal.CanSet() {
			continue
		}

		// Handle embedded or nested struct
		if fieldVal.Kind() == reflect.Struct && fieldType.Type != reflect.TypeOf(time.Time{}) {
			nestedPrefix := prefix
			if envTag := fieldType.Tag.Get("env"); envTag != "" {
				if nestedPrefix != "" {
					nestedPrefix = nestedPrefix + "_" + envTag
				} else {
					nestedPrefix = envTag
				}
			}
			if err := processStruct(nestedPrefix, fieldVal, lookup); err != nil {
				return err
			}
			continue
		}

		envTag := fieldType.Tag.Get("env")
		if envTag == "" {
			continue
		}

		var fullKey string
		if prefix != "" {
			fullKey = prefix + "_" + envTag
		} else {
			fullKey = envTag
		}

		rawVal, found := lookup(fullKey)
		defaultVal := fieldType.Tag.Get("default")
		isRequired := strings.EqualFold(fieldType.Tag.Get("required"), "true")

		if !found || rawVal == "" {
			if defaultVal != "" {
				rawVal = defaultVal
			} else if isRequired {
				return fmt.Errorf("%w: %s", ErrMissingRequired, fullKey)
			} else {
				continue
			}
		}

		splitSep := fieldType.Tag.Get("split")
		if splitSep == "" {
			splitSep = ","
		}

		if err := setFieldValue(fieldVal, rawVal, splitSep); err != nil {
			return fmt.Errorf("field %s: %w", fieldType.Name, err)
		}
	}

	return nil
}

func setFieldValue(field reflect.Value, raw string, splitSep string) error {
	// Handle time.Duration specifically
	if field.Type() == reflect.TypeOf(time.Duration(0)) {
		d, err := time.ParseDuration(raw)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidValue, err)
		}
		field.SetInt(int64(d))
		return nil
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(raw)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidValue, err)
		}
		field.SetInt(v)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidValue, err)
		}
		field.SetUint(v)

	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidValue, err)
		}
		field.SetFloat(v)

	case reflect.Bool:
		v, err := parseBool(raw)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidValue, err)
		}
		field.SetBool(v)

	case reflect.Slice:
		parts := strings.Split(raw, splitSep)
		sliceLen := len(parts)
		newSlice := reflect.MakeSlice(field.Type(), sliceLen, sliceLen)

		for i, part := range parts {
			part = strings.TrimSpace(part)
			targetElem := newSlice.Index(i)
			if err := setFieldValue(targetElem, part, splitSep); err != nil {
				return err
			}
		}
		field.Set(newSlice)

	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedType, field.Kind().String())
	}

	return nil
}

func parseBool(val string) (bool, error) {
	val = strings.ToLower(strings.TrimSpace(val))
	switch val {
	case "1", "t", "true", "yes", "y", "on":
		return true, nil
	case "0", "f", "false", "no", "n", "off":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean string: %q", val)
	}
}

// DumpMasked returns a map of environment variables mapped from the struct,
// masking fields with mask:"true" or secret:"true".
func DumpMasked(target any) map[string]string {
	result := make(map[string]string)
	val := reflect.ValueOf(target)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return result
	}

	dumpStruct("", val, result)
	return result
}

func dumpStruct(prefix string, structVal reflect.Value, out map[string]string) {
	structType := structVal.Type()

	for i := 0; i < structVal.NumField(); i++ {
		fieldVal := structVal.Field(i)
		fieldType := structType.Field(i)

		if fieldVal.Kind() == reflect.Struct && fieldType.Type != reflect.TypeOf(time.Time{}) {
			nestedPrefix := prefix
			if envTag := fieldType.Tag.Get("env"); envTag != "" {
				if nestedPrefix != "" {
					nestedPrefix = nestedPrefix + "_" + envTag
				} else {
					nestedPrefix = envTag
				}
			}
			dumpStruct(nestedPrefix, fieldVal, out)
			continue
		}

		envTag := fieldType.Tag.Get("env")
		if envTag == "" {
			continue
		}

		var fullKey string
		if prefix != "" {
			fullKey = prefix + "_" + envTag
		} else {
			fullKey = envTag
		}

		isMasked := strings.EqualFold(fieldType.Tag.Get("mask"), "true") ||
			strings.EqualFold(fieldType.Tag.Get("secret"), "true")

		if isMasked {
			out[fullKey] = "******"
		} else {
			out[fullKey] = fmt.Sprintf("%v", fieldVal.Interface())
		}
	}
}
