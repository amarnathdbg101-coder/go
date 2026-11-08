package cachekit

import "errors"

// Standard cachekit errors.
var (
	// ErrCacheMiss is returned when a requested key does not exist or has expired.
	ErrCacheMiss = errors.New("cachekit: cache miss")

	// ErrKeyExpired is returned when a key is found but its TTL has elapsed.
	ErrKeyExpired = errors.New("cachekit: key has expired")

	// ErrInvalidTTL is returned when an invalid TTL duration is passed.
	ErrInvalidTTL = errors.New("cachekit: TTL must be greater than zero")

	// ErrStoreNil is returned when engine is initialized with a nil store.
	ErrStoreNil = errors.New("cachekit: store cannot be nil")
)
