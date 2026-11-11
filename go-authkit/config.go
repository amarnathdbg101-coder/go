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

	// OTPCooldown is the minimum time between consecutive OTP generation requests for the same email. Default: 60s.
	OTPCooldown time.Duration

	// MaxBodyBytes limits the incoming HTTP request payload size to defend against memory exhaustion DoS. Default: 64 KB.
	MaxBodyBytes int64

	// MinPasswordLength is the minimum acceptable password length. Default: 8.
	MinPasswordLength int

	// Issuer is embedded in JWT RegisteredClaims (iss). Default: "go-authkit".
	Issuer string

	// DefaultRole is assigned to newly registered users when role is omitted. Default: "user".
	DefaultRole string

	// OnSendOTP is an optional hook callback invoked when an OTP is generated (e.g. to send email or SMS).
	// If nil, OTP is generated and stored in-memory (and logged in development).
	OnSendOTP func(ctx context.Context, email, otp string) error
}

// Option represents a functional option for configuring Config.
type Option func(*Config)

// WithAccessTokenTTL sets the access token validity duration.
func WithAccessTokenTTL(d time.Duration) Option {
	return func(c *Config) { c.AccessTokenTTL = d }
}

// WithRefreshTokenTTL sets the refresh token validity duration.
func WithRefreshTokenTTL(d time.Duration) Option {
	return func(c *Config) { c.RefreshTokenTTL = d }
}

// WithBcryptCost sets the Bcrypt hashing complexity cost (MinCost to MaxCost).
func WithBcryptCost(cost int) Option {
	return func(c *Config) { c.BcryptCost = cost }
}

// WithOTPTTL sets the time before an unconsumed OTP expires.
func WithOTPTTL(d time.Duration) Option {
	return func(c *Config) { c.OTPTTL = d }
}

// WithOTPMaxAttempts sets the maximum invalid OTP attempts before brute force lockout.
func WithOTPMaxAttempts(attempts int) Option {
	return func(c *Config) { c.OTPMaxAttempts = attempts }
}

// WithOTPCooldown sets the throttle duration between consecutive OTP requests for the same identity.
func WithOTPCooldown(d time.Duration) Option {
	return func(c *Config) { c.OTPCooldown = d }
}

// WithMaxBodyBytes sets the maximum HTTP request body read limit in bytes.
func WithMaxBodyBytes(bytes int64) Option {
	return func(c *Config) { c.MaxBodyBytes = bytes }
}

// WithMinPasswordLength sets the minimum allowed password length.
func WithMinPasswordLength(length int) Option {
	return func(c *Config) { c.MinPasswordLength = length }
}

// WithIssuer sets the JWT issuer claim.
func WithIssuer(issuer string) Option {
	return func(c *Config) { c.Issuer = issuer }
}

// WithDefaultRole sets the fallback user role if none is supplied during registration.
func WithDefaultRole(role string) Option {
	return func(c *Config) { c.DefaultRole = role }
}

// WithOnSendOTP configures an external delivery hook for OTP codes (e.g. email, SMS, push).
func WithOnSendOTP(fn func(ctx context.Context, email, otp string) error) Option {
	return func(c *Config) { c.OnSendOTP = fn }
}

// DefaultConfig provides a robust production baseline configuration.
// Caller sirf apna custom JWTSecret specify karega.
func DefaultConfig(jwtSecret string) Config {
	return Config{
		JWTSecret:         jwtSecret,
		AccessTokenTTL:    15 * time.Minute,
		RefreshTokenTTL:   7 * 24 * time.Hour,
		BcryptCost:        bcrypt.DefaultCost + 2, // 12
		OTPTTL:            5 * time.Minute,
		OTPMaxAttempts:    3,
		OTPCooldown:       60 * time.Second,
		MaxBodyBytes:      64 * 1024, // 64 KB
		MinPasswordLength: 8,
		Issuer:            "go-authkit",
		DefaultRole:       "user",
	}
}

// NewConfig initializes Config with production defaults and applies functional options.
func NewConfig(jwtSecret string, opts ...Option) Config {
	cfg := DefaultConfig(jwtSecret)
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
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
	if c.OTPCooldown < 0 {
		c.OTPCooldown = 60 * time.Second
	}
	if c.MaxBodyBytes <= 0 {
		c.MaxBodyBytes = 64 * 1024
	}
	if c.MinPasswordLength <= 0 {
		c.MinPasswordLength = 8
	}
	if c.Issuer == "" {
		c.Issuer = "go-authkit"
	}
	if c.DefaultRole == "" {
		c.DefaultRole = "user"
	}
	return nil
}
