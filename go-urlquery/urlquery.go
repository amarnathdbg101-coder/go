package urlquery

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Decode parses url.Values into a struct pointer.
func Decode(values url.Values, target any) error {
	val := reflect.ValueOf(target)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return ErrNotPointerToStruct
	}

	elem := val.Elem()
	if elem.Kind() != reflect.Struct {
		return ErrNotPointerToStruct
	}

	return decodeStruct(values, elem)
}

// DecodeURL extracts query parameters from a raw URL string and decodes them into target.
func DecodeURL(rawURL string, target any) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	return Decode(u.Query(), target)
}

func decodeStruct(values url.Values, structVal reflect.Value) error {
	structType := structVal.Type()

	for i := 0; i < structVal.NumField(); i++ {
		fieldVal := structVal.Field(i)
		fieldType := structType.Field(i)

		if !fieldVal.CanSet() {
			continue
		}

		tag := fieldType.Tag.Get("query")
		if tag == "-" {
			continue
		}

		paramName, defaultVal, isComma := parseTag(tag, fieldType.Name)
		rawList, found := values[paramName]

		if !found || len(rawList) == 0 || (len(rawList) == 1 && rawList[0] == "") {
			if defaultVal != "" {
				rawList = []string{defaultVal}
			} else {
				continue
			}
		}

		if err := setField(fieldVal, rawList, isComma); err != nil {
			return fmt.Errorf("field %s: %w", fieldType.Name, err)
		}
	}

	return nil
}

func parseTag(tag string, defaultName string) (name string, defaultVal string, isComma bool) {
	if tag == "" {
		return strings.ToLower(defaultName), "", false
	}

	parts := strings.Split(tag, ",")
	name = strings.TrimSpace(parts[0])
	if name == "" {
		name = strings.ToLower(defaultName)
	}

	for _, opt := range parts[1:] {
		opt = strings.TrimSpace(opt)
		if strings.HasPrefix(opt, "default=") {
			defaultVal = strings.TrimPrefix(opt, "default=")
		} else if opt == "comma" {
			isComma = true
		}
	}

	return name, defaultVal, isComma
}

func setField(field reflect.Value, rawList []string, isComma bool) error {
	if len(rawList) == 0 {
		return nil
	}

	raw := rawList[0]

	// Handle time.Duration
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
		v, err := strconv.ParseBool(raw)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidValue, err)
		}
		field.SetBool(v)

	case reflect.Slice:
		var items []string
		if isComma {
			// Split all raw strings by comma
			for _, r := range rawList {
				for _, sub := range strings.Split(r, ",") {
					sub = strings.TrimSpace(sub)
					if sub != "" {
						items = append(items, sub)
					}
				}
			}
		} else {
			items = rawList
		}

		newSlice := reflect.MakeSlice(field.Type(), len(items), len(items))
		for i, item := range items {
			elem := newSlice.Index(i)
			if err := setField(elem, []string{item}, false); err != nil {
				return err
			}
		}
		field.Set(newSlice)

	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedType, field.Kind().String())
	}

	return nil
}

// Encode serializes a struct into url.Values.
func Encode(source any) (url.Values, error) {
	val := reflect.ValueOf(source)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil, ErrNotStruct
	}

	values := make(url.Values)
	structType := val.Type()

	for i := 0; i < val.NumField(); i++ {
		fieldVal := val.Field(i)
		fieldType := structType.Field(i)

		tag := fieldType.Tag.Get("query")
		if tag == "-" {
			continue
		}

		paramName, _, isComma := parseTag(tag, fieldType.Name)

		if fieldVal.Kind() == reflect.Slice {
			sliceLen := fieldVal.Len()
			if sliceLen == 0 {
				continue
			}

			if isComma {
				var parts []string
				for j := 0; j < sliceLen; j++ {
					parts = append(parts, fmt.Sprintf("%v", fieldVal.Index(j).Interface()))
				}
				values.Set(paramName, strings.Join(parts, ","))
			} else {
				for j := 0; j < sliceLen; j++ {
					values.Add(paramName, fmt.Sprintf("%v", fieldVal.Index(j).Interface()))
				}
			}
			continue
		}

		strVal := fmt.Sprintf("%v", fieldVal.Interface())
		values.Set(paramName, strVal)
	}

	return values, nil
}

// EncodeString serializes a struct into a URL query parameter string (e.g. "limit=20&page=1").
func EncodeString(source any) (string, error) {
	vals, err := Encode(source)
	if err != nil {
		return "", err
	}
	return vals.Encode(), nil
}
