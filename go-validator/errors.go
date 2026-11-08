package validator

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrInvalidPayload is returned when JSON unmarshaling fails.
	ErrInvalidPayload = errors.New("validator: malformed or unreadable JSON payload")

	// ErrValidationFailed is returned when one or more validation constraints fail.
	ErrValidationFailed = errors.New("validator: payload failed validation constraints")
)

// ValidationErrors maps failed field names to their specific human-readable error messages.
// Directly serializes to JSON as a clean key-value object for Flutter / React frontends:
//
//	{
//	  "email": "invalid email address",
//	  "password": "password must be at least 8 characters"
//	}
type ValidationErrors map[string]string

// Error implements the standard error interface.
func (v ValidationErrors) Error() string {
	if len(v) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("validation failed: ")
	i := 0
	for field, msg := range v {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("%s: %s", field, msg))
		i++
	}
	return sb.String()
}

// JSON returns serialized JSON bytes of the validation errors.
func (v ValidationErrors) JSON() []byte {
	bytes, _ := json.Marshal(v)
	return bytes
}
