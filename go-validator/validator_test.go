package validator_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/shopsilo/go-validator"
)

type SignupPayload struct {
	Name            string `json:"name" validate:"required,min=3"`
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required,strong_password"`
	ConfirmPassword string `json:"confirm_password" validate:"required"`
}

// Custom business validation implementing validator.Validatable interface
func (p SignupPayload) Validate(v *validator.Validator) {
	v.Check(p.Password == p.ConfirmPassword, "confirm_password", "passwords do not match")
}

// -------------------------------------------------------------------------
// 1. Fluent Validation Tests
// -------------------------------------------------------------------------

func TestValidator_Fluent(t *testing.T) {
	v := validator.New()

	email := "invalid-email"
	password := "short"

	v.Required(email, "email", "email is required")
	v.Email(email, "email", "invalid email address")
	v.Min(password, 8, "password", "password must be at least 8 characters")

	if v.Valid() {
		t.Fatalf("expected validation failures")
	}

	if len(v.Errors) != 2 {
		t.Errorf("expected 2 errors, got %d", len(v.Errors))
	}

	if v.Errors["email"] != "invalid email address" {
		t.Errorf("expected 'invalid email address', got '%s'", v.Errors["email"])
	}

	if v.Errors["password"] != "password must be at least 8 characters" {
		t.Errorf("expected password error, got '%s'", v.Errors["password"])
	}

	// Error() returns non-nil error
	if v.Error() == nil {
		t.Errorf("expected non-nil error from v.Error()")
	}
}

// -------------------------------------------------------------------------
// 2. Core Rule Unit Tests
// -------------------------------------------------------------------------

func TestRules_Email(t *testing.T) {
	valid := []string{"test@example.com", "user.name+tag@sub.domain.co.in", "admin@server.org"}
	for _, e := range valid {
		if !validator.IsEmail(e) {
			t.Errorf("expected '%s' to be valid email", e)
		}
	}

	invalid := []string{"plainaddress", "@missinguser.com", "user@", "user@.com", "user@domain..com", ""}
	for _, e := range invalid {
		if validator.IsEmail(e) {
			t.Errorf("expected '%s' to be invalid email", e)
		}
	}
}

func TestRules_Phone(t *testing.T) {
	valid := []string{"+919876543210", "9876543210", "+14155552671", "01123456789"}
	for _, p := range valid {
		if !validator.IsPhone(p) {
			t.Errorf("expected '%s' to be valid phone", p)
		}
	}

	invalid := []string{"123", "abc", "+", "98765abcdef"}
	for _, p := range invalid {
		if validator.IsPhone(p) {
			t.Errorf("expected '%s' to be invalid phone", p)
		}
	}
}

func TestRules_StrongPassword(t *testing.T) {
	if !validator.IsStrongPassword("P@ssw0rd123!") {
		t.Errorf("expected strong password to pass")
	}

	// Weak cases
	if validator.IsStrongPassword("short1!") { // < 8 chars
		t.Errorf("expected short password to fail")
	}
	if validator.IsStrongPassword("alllowercase123!") { // no uppercase
		t.Errorf("expected missing uppercase to fail")
	}
	if validator.IsStrongPassword("ALLUPPERCASE123!") { // no lowercase
		t.Errorf("expected missing lowercase to fail")
	}
	if validator.IsStrongPassword("NoSpecialChars123") { // no special
		t.Errorf("expected missing special char to fail")
	}
}

func TestRules_UUID_And_Slug(t *testing.T) {
	if !validator.IsUUID("c08a491f-0b44-4f24-9b8e-734d2847a9e1") {
		t.Errorf("expected valid UUID")
	}
	if validator.IsUUID("not-a-uuid") {
		t.Errorf("expected invalid UUID")
	}

	if !validator.IsSlug("wireless-ergonomic-mouse-101") {
		t.Errorf("expected valid slug")
	}
	if validator.IsSlug("Invalid Slug With Spaces!") {
		t.Errorf("expected invalid slug")
	}
}

// -------------------------------------------------------------------------
// 3. Struct Tag & Custom Validation Tests
// -------------------------------------------------------------------------

func TestStructTag_Validation(t *testing.T) {
	// Case A: Missing fields
	badPayload := SignupPayload{
		Name:            "Al", // < 3 chars
		Email:           "bad-email",
		Password:        "weak",
		ConfirmPassword: "other",
	}

	errs := validator.ValidateStruct(badPayload)
	if errs == nil {
		t.Fatalf("expected validation errors")
	}

	if errs["name"] == "" || errs["email"] == "" || errs["password"] == "" || errs["confirm_password"] == "" {
		t.Errorf("expected errors for all invalid fields, got: %+v", errs)
	}

	// Case B: Valid payload
	goodPayload := SignupPayload{
		Name:            "Alex Sharma",
		Email:           "alex@example.com",
		Password:        "SecureP@ss123!",
		ConfirmPassword: "SecureP@ss123!",
	}

	goodErrs := validator.ValidateStruct(goodPayload)
	if goodErrs != nil {
		t.Fatalf("expected 0 errors for valid payload, got: %+v", goodErrs)
	}
}

// -------------------------------------------------------------------------
// 4. HTTP Binding & Response Tests
// -------------------------------------------------------------------------

func TestBindAndValidate_HTTP(t *testing.T) {
	jsonBody := `{"name": "John Doe", "email": "john@test.com", "password": "SecretPass123!", "confirm_password": "SecretPass123!"}`
	req := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	var payload SignupPayload
	err := validator.BindAndValidate(req, &payload)
	if err != nil {
		t.Fatalf("bind and validate failed: %v", err)
	}

	if payload.Name != "John Doe" || payload.Email != "john@test.com" {
		t.Errorf("payload fields mismatch: %+v", payload)
	}
}

func TestWriteValidationError(t *testing.T) {
	errs := validator.ValidationErrors{
		"email": "invalid email format",
	}

	w := httptest.NewRecorder()
	validator.WriteValidationError(w, errs)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected status 422, got %d", w.Code)
	}

	respBody := w.Body.String()
	if !strings.Contains(respBody, `"email":"invalid email format"`) {
		t.Errorf("expected json error response, got: %s", respBody)
	}
}
