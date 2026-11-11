package authkit

import "errors"

// Common domain errors returned by go-authkit.
// In errors ko use karke caller application proper HTTP status code or business error handle kar sakti hai.
var (
	ErrUserNotFound         = errors.New("user not found")
	ErrEmailAlreadyExists   = errors.New("email already registered")
	ErrInvalidCredentials   = errors.New("invalid email or password")
	ErrInvalidToken         = errors.New("invalid or malformed token")
	ErrTokenExpired         = errors.New("token has expired")
	ErrInvalidTokenType     = errors.New("invalid token type")
	ErrTokenRevoked         = errors.New("token has been revoked")
	ErrInvalidOTP           = errors.New("invalid OTP")
	ErrOTPExpired           = errors.New("OTP has expired")
	ErrOTPCooldown          = errors.New("please wait before requesting another OTP")
	ErrMaxAttemptsExceeded  = errors.New("maximum OTP verification attempts exceeded, request a new OTP")
	ErrWeakPassword         = errors.New("password does not satisfy minimum length requirements")
	ErrInvalidEmail         = errors.New("invalid email format")
	ErrUnauthorized         = errors.New("unauthorized access")
	ErrForbidden            = errors.New("forbidden: insufficient permissions")
	ErrPayloadTooLarge      = errors.New("request payload exceeds maximum allowed size")
	ErrInvalidTableName     = errors.New("invalid table name identifier")
)
