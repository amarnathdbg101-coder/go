package authkit

import (
	"context"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Config holds runtime configuration options for the go-authkit Engine.
type Config struct {
	// JWTSecret is the secret key used to HMAC-SHA256 sign JWT tokens. (Required)
	JWTSecret string

	// AccessTokenTTL defines how long the access token remains valid. Default: 15 minutes.
	AccessTokenTTL time.Duration

	// RefreshTokenTTL defines how long the refresh token remains valid. Default: 7 days.
	RefreshTokenTTL time.Duration

	// BcryptCost is the hashing complexity factor. Default: 12.
	BcryptCost int

	// OTPTTL defines how long a generated OTP remains valid. Default: 5 minutes.
	OTPTTL time.Duration

	// OTPMaxAttempts is the maximum allowed incorrect OTP attempts before lockout. Default: 3.
	OTPMaxAttempts int

	// Issuer is embedded in JWT RegisteredClaims (iss). Default: "go-authkit".
	Issuer string

	// DefaultRole is assigned to newly registered users when role is omitted. Default: "user".
	DefaultRole string

	// OnSendOTP is an optional hook callback invoked when an OTP is generated (e.g. to send email or SMS).
	// If nil, OTP is generated and stored in-memory (and logged in development).
	OnSendOTP func(ctx context.Context, email, otp string) error
}

// DefaultConfig provides a robust production baseline configuration.
// Caller sirf apna custom JWTSecret specify karega.
func DefaultConfig(jwtSecret string) Config {
	return Config{
		JWTSecret:       jwtSecret,
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		BcryptCost:      bcrypt.DefaultCost + 2, // 12
		OTPTTL:          5 * time.Minute,
		OTPMaxAttempts:  3,
		Issuer:          "go-authkit",
		DefaultRole:     "user",
	}
}

// Validate checks required fields and enforces security standards.
func (c *Config) Validate() error {
	if c.JWTSecret == "" {
		return errors.New("authkit: JWTSecret must not be empty")
	}
	if len(c.JWTSecret) < 16 {
		return errors.New("authkit: JWTSecret must be at least 16 characters for HMAC-SHA256 security")
	}
	if c.AccessTokenTTL <= 0 {
		c.AccessTokenTTL = 15 * time.Minute
	}
	if c.RefreshTokenTTL <= 0 {
		c.RefreshTokenTTL = 7 * 24 * time.Hour
	}
	if c.BcryptCost < bcrypt.MinCost || c.BcryptCost > bcrypt.MaxCost {
		c.BcryptCost = bcrypt.DefaultCost
	}
	if c.OTPTTL <= 0 {
		c.OTPTTL = 5 * time.Minute
	}
	if c.OTPMaxAttempts <= 0 {
		c.OTPMaxAttempts = 3
	}
	if c.Issuer == "" {
		c.Issuer = "go-authkit"
	}
	if c.DefaultRole == "" {
		c.DefaultRole = "user"
	}
	return nil
}
