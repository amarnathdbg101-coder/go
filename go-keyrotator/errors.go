package keyrotator

import "errors"

var (
	// ErrNoKeysConfigured indicates that the rotator was initialized with an empty key list.
	ErrNoKeysConfigured = errors.New("keyrotator: no api keys configured in pool")

	// ErrAllKeysExhausted indicates that all keys in the pool are rate-limited or in cooldown.
	ErrAllKeysExhausted = errors.New("keyrotator: all api keys are currently rate-limited or exhausted")

	// ErrKeyNotFound indicates the requested key does not exist in the rotator pool.
	ErrKeyNotFound = errors.New("keyrotator: key not found in pool")
)
