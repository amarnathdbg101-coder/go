package goenv

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	ErrNotPointerToStruct = errors.New("goenv: target must be a non-nil pointer to a struct")
	ErrRequiredEnvMissing = errors.New("goenv: required environment variable is missing")
)

// Unmarshal populates the target struct pointer using environment variables from os.LookupEnv.
func Unmarshal(target any) error {
	return UnmarshalPrefix("", target)
}

// UnmarshalPrefix populates the target struct pointer prepending prefix to field tags.
func UnmarshalPrefix(prefix string, target any) error {
	val := reflect.ValueOf(target)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return ErrNotPointerToStruct
	}

	elem := val.Elem()
	if elem.Kind() != reflect.Struct {
		return ErrNotPointerToStruct
	}

	return processStruct(prefix, elem)
}

func processStruct(prefix string, structVal reflect.Value) error {
	structType := structVal.Type()

	for i := 0; i < structVal.NumField(); i++ {
		fieldVal := structVal.Field(i)
		fieldType := structType.Field(i)

		if !fieldVal.CanSet() {
			continue
		}

		// Handle embedded or nested struct
		if fieldVal.Kind() == reflect.Struct && fieldType.Type != reflect.TypeOf(time.Time{}) {
			nestedPrefix := prefix
			if envTag := fieldType.Tag.Get("env"); envTag != "" {
				tagParts := strings.Split(envTag, ",")
				if len(tagParts) > 0 && tagParts[0] != "" {
					if nestedPrefix != "" {
						nestedPrefix = nestedPrefix + "_" + tagParts[0]
					} else {
						nestedPrefix = tagParts[0]
					}
				}
			}
			if err := processStruct(nestedPrefix, fieldVal); err != nil {
				return err
			}
			continue
		}

		tag := fieldType.Tag.Get("env")
		if tag == "" {
			continue
		}

		tagParts := strings.Split(tag, ",")
		envKey := strings.TrimSpace(tagParts[0])
		if prefix != "" {
			envKey = prefix + "_" + envKey
		}

		var (
			defaultValue string
			hasDefault   bool
			required     bool
		)

		for _, opt := range tagParts[1:] {
			opt = strings.TrimSpace(opt)
			if opt == "required" {
				required = true
			} else if strings.HasPrefix(opt, "default=") {
				defaultValue = strings.TrimPrefix(opt, "default=")
				hasDefault = true
			}
		}

		envVal, exists := os.LookupEnv(envKey)
		if !exists || envVal == "" {
			if hasDefault {
				envVal = defaultValue
			} else if required {
				return fmt.Errorf("%w: %s (field %s)", ErrRequiredEnvMissing, envKey, fieldType.Name)
			} else {
				continue
			}
		}

		if err := setFieldValue(fieldVal, envVal); err != nil {
			return fmt.Errorf("goenv: failed to set field %s from env %s: %w", fieldType.Name, envKey, err)
		}
	}

	return nil
}

func setFieldValue(val reflect.Value, raw string) error {
	switch val.Kind() {
	case reflect.String:
		val.SetString(raw)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if val.Type() == reflect.TypeOf(time.Duration(0)) {
			d, err := time.ParseDuration(raw)
			if err != nil {
				return err
			}
			val.SetInt(int64(d))
			return nil
		}
		n, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return err
		}
		val.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			return err
		}
		val.SetUint(n)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return err
		}
		val.SetFloat(f)
	case reflect.Bool:
		b, err := strconv.ParseBool(raw)
		if err != nil {
			return err
		}
		val.SetBool(b)
	case reflect.Slice:
		parts := strings.Split(raw, ",")
		slice := reflect.MakeSlice(val.Type(), len(parts), len(parts))
		for i, part := range parts {
			trimmed := strings.TrimSpace(part)
			if err := setFieldValue(slice.Index(i), trimmed); err != nil {
				return err
			}
		}
		val.Set(slice)
	default:
		return fmt.Errorf("unsupported field type: %s", val.Type())
	}
	return nil
}
