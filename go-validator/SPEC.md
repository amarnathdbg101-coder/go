# SPEC: go-validator

## 1. Scope & Purpose
Ultra-lightweight, high-speed input and struct payload validation engine for Go backends.
- In Scope: 100% reflection-free fluent validation, optional struct tag validation (`validate:"required,email,min=X,max=X,phone,url,uuid,strong_password"`), clean JSON error serialization (`map[string]string`), and HTTP payload binding (`BindAndValidate`).
- Out of Scope: Database foreign-key checks (handled at the repository/database layer).

## 2. Public API Contract
- `New(opts ...Option) *Validator`
- `Validator.Check(ok bool, field, message string)`
- `Validator.Required(val, field, msg string)`
- `Validator.Email(val, field, msg string)`
- `Validator.Min(val, min int, field, msg string)`
- `Validator.Max(val, max int, field, msg string)`
- `Validator.Phone(val, field, msg string)`
- `Validator.StrongPassword(val, field, msg string)`
- `Validator.Valid() bool`
- `Validator.Error() error`
- `ValidateStruct(s any) ValidationErrors`
- `BindAndValidate(r *http.Request, dst any) error`
- `WriteValidationError(w http.ResponseWriter, errs ValidationErrors)`

## 3. Concurrency & Memory Guarantees
- Memory Footprint: Pure Go standard library, zero reflection overhead when using fluent `Check/Required/Email` methods.
- Thread Safety: Individual `Validator` instances are intended per-request (stack allocated, zero global state).

## 4. How to Extend
- Add new rule functions to `rules.go` as pure boolean functions (`IsX(val string) bool`).
- Wire custom struct tag rules inside `struct_tag.go`.
