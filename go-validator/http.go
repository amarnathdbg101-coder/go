package validator

import (
	"encoding/json"
	"io"
	"net/http"
)

// DefaultMaxBodyBytes limits incoming JSON bodies to 1 MB to prevent memory exhaustion attacks.
const DefaultMaxBodyBytes = 1024 * 1024 // 1 MB

// BindAndValidate parses JSON request body into dst and executes both struct tag and custom validations.
// Works seamlessly in any Chi, standard net/http, or Gorilla Mux handler.
func BindAndValidate(r *http.Request, dst any) error {
	// 1. Bound request reader
	r.Body = http.MaxBytesReader(nil, r.Body, DefaultMaxBodyBytes)

	// 2. Decode JSON
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields() // Strict JSON security

	if err := dec.Decode(dst); err != nil {
		if err == io.EOF {
			return ErrInvalidPayload
		}
		return err
	}

	// 3. Execute validations (struct tags + custom Validate() method)
	if errs := ValidateStruct(dst); errs != nil {
		return errs
	}

	return nil
}

// WriteValidationError formats and writes a standard HTTP 422 Unprocessable Entity response.
func WriteValidationError(w http.ResponseWriter, errs ValidationErrors) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusUnprocessableEntity)

	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": false,
		"message": "Validation failed",
		"errors":  errs,
	})
}
