package validator

import "strings"

// Validator accumulates validation errors across fields.
// It is 100% reflection-free for ultra-high performance in microservices.
type Validator struct {
	Errors ValidationErrors
	cfg    Config
}

// New creates an initialized Validator instance.
func New(opts ...Option) *Validator {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	return &Validator{
		Errors: make(ValidationErrors),
		cfg:    cfg,
	}
}

// Valid returns true if no validation errors have been recorded.
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// Error returns ValidationErrors if any errors exist, otherwise nil.
// Allows clean one-line checks:
//
//	if err := v.Error(); err != nil { return err }
func (v *Validator) Error() error {
	if !v.Valid() {
		return v.Errors
	}
	return nil
}

// AddError appends an error message for a given field if not already recorded.
// (Preserves the primary root error for that field).
func (v *Validator) AddError(field, message string) {
	if v.Errors == nil {
		v.Errors = make(ValidationErrors)
	}
	if _, exists := v.Errors[field]; !exists {
		v.Errors[field] = message
	}
}

// Check evaluates a boolean condition. If false, records an error message for the field.
func (v *Validator) Check(ok bool, field, message string) {
	if v.cfg.StopOnFirstError && !v.Valid() {
		return
	}
	if !ok {
		v.AddError(field, message)
	}
}

// Required ensures a string field is not empty or pure whitespace.
func (v *Validator) Required(val, field, message string) {
	if strings.TrimSpace(val) == "" {
		v.AddError(field, message)
	}
}

// Email validates email syntax. If empty, it skips (use Required to enforce presence).
func (v *Validator) Email(val, field, message string) {
	if strings.TrimSpace(val) == "" {
		return
	}
	if !IsEmail(val) {
		v.AddError(field, message)
	}
}

// Min ensures string length is at least min characters.
func (v *Validator) Min(val string, min int, field, message string) {
	if strings.TrimSpace(val) == "" {
		return
	}
	if !MinLength(val, min) {
		v.AddError(field, message)
	}
}

// Max ensures string length does not exceed max characters.
func (v *Validator) Max(val string, max int, field, message string) {
	if strings.TrimSpace(val) == "" {
		return
	}
	if !MaxLength(val, max) {
		v.AddError(field, message)
	}
}

// Phone validates telephone number format.
func (v *Validator) Phone(val, field, message string) {
	if strings.TrimSpace(val) == "" {
		return
	}
	if !IsPhone(val) {
		v.AddError(field, message)
	}
}

// StrongPassword enforces uppercase, lowercase, digit, and symbol constraints.
func (v *Validator) StrongPassword(val, field, message string) {
	if strings.TrimSpace(val) == "" {
		return
	}
	if !IsStrongPassword(val) {
		v.AddError(field, message)
	}
}

// In checks if value is within an allowed set.
func (v *Validator) In(val string, allowed []string, field, message string) {
	if strings.TrimSpace(val) == "" {
		return
	}
	if !In(val, allowed...) {
		v.AddError(field, message)
	}
}
