package gowebhook

import "errors"

var (
	// ErrInvalidSignature indicates the HMAC signature does not match payload.
	ErrInvalidSignature = errors.New("gowebhook: invalid webhook signature")

	// ErrSignatureExpired indicates the signature timestamp is older than allowed tolerance.
	ErrSignatureExpired = errors.New("gowebhook: signature timestamp expired or in future")

	// ErrMalformedHeader indicates the signature header format is invalid.
	ErrMalformedHeader = errors.New("gowebhook: malformed webhook signature header")

	// ErrEmptySecret indicates a secret key was not provided.
	ErrEmptySecret = errors.New("gowebhook: webhook secret cannot be empty")

	// ErrMaxRetriesExceeded indicates all retry attempts failed.
	ErrMaxRetriesExceeded = errors.New("gowebhook: maximum delivery retries exceeded")
)
