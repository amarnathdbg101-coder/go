package validator

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Validatable is an interface that structs can implement for custom validation logic.
type Validatable interface {
	Validate(v *Validator)
}

// ValidateStruct inspects struct tags (`validate:"..."`) and maps failures to JSON keys (`json:"..."`).
// Supports: required, email, phone, url, uuid, min=X, max=X, strong_password, alphanumeric.
func ValidateStruct(s any) ValidationErrors {
	v := New()

	val := reflect.ValueOf(s)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil
	}

	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		tag := field.Tag.Get("validate")
		if tag == "" {
			continue
		}

		// Determine JSON field name (falls back to lowercase struct field name)
		fieldName := field.Tag.Get("json")
		if fieldName == "" || fieldName == "-" {
			fieldName = strings.ToLower(field.Name)
		} else {
			if idx := strings.Index(fieldName, ","); idx != -1 {
				fieldName = fieldName[:idx]
			}
		}

		fieldVal := val.Field(i)
		strVal := ""

		switch fieldVal.Kind() {
		case reflect.String:
			strVal = fieldVal.String()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			strVal = strconv.FormatInt(fieldVal.Int(), 10)
		case reflect.Float32, reflect.Float64:
			strVal = strconv.FormatFloat(fieldVal.Float(), 'f', -1, 64)
		}

		rules := strings.Split(tag, ",")
		for _, rule := range rules {
			rule = strings.TrimSpace(rule)
			if rule == "" {
				continue
			}

			switch {
			case rule == "required":
				if strings.TrimSpace(strVal) == "" {
					v.AddError(fieldName, fmt.Sprintf("%s is required", fieldName))
				}
			case rule == "email":
				if strVal != "" && !IsEmail(strVal) {
					v.AddError(fieldName, fmt.Sprintf("%s must be a valid email address", fieldName))
				}
			case rule == "phone":
				if strVal != "" && !IsPhone(strVal) {
					v.AddError(fieldName, fmt.Sprintf("%s must be a valid phone number", fieldName))
				}
			case rule == "url":
				if strVal != "" && !IsURL(strVal) {
					v.AddError(fieldName, fmt.Sprintf("%s must be a valid URL", fieldName))
				}
			case rule == "uuid":
				if strVal != "" && !IsUUID(strVal) {
					v.AddError(fieldName, fmt.Sprintf("%s must be a valid UUID", fieldName))
				}
			case rule == "alphanumeric":
				if strVal != "" && !IsAlphanumeric(strVal) {
					v.AddError(fieldName, fmt.Sprintf("%s must contain only letters and numbers", fieldName))
				}
			case rule == "strong_password":
				if strVal != "" && !IsStrongPassword(strVal) {
					v.AddError(fieldName, fmt.Sprintf("%s must be at least 8 characters long and include uppercase, lowercase, digit, and special symbol", fieldName))
				}
			case strings.HasPrefix(rule, "min="):
				minStr := strings.TrimPrefix(rule, "min=")
				if min, err := strconv.Atoi(minStr); err == nil {
					if strVal != "" && !MinLength(strVal, min) {
						v.AddError(fieldName, fmt.Sprintf("%s must be at least %d characters", fieldName, min))
					}
				}
			case strings.HasPrefix(rule, "max="):
				maxStr := strings.TrimPrefix(rule, "max=")
				if max, err := strconv.Atoi(maxStr); err == nil {
					if strVal != "" && !MaxLength(strVal, max) {
						v.AddError(fieldName, fmt.Sprintf("%s must not exceed %d characters", fieldName, max))
					}
				}
			}
		}
	}

	// If struct implements Validatable, run custom validation rules
	if validatable, ok := s.(Validatable); ok {
		validatable.Validate(v)
	}

	if !v.Valid() {
		return v.Errors
	}
	return nil
}
