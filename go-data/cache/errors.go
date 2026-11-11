package cache

import "errors"

var (
	ErrCacheMiss  = errors.New("cache: cache miss")
	ErrKeyExpired = errors.New("cache: key has expired")
	ErrInvalidTTL = errors.New("cache: TTL must be greater than zero")
	ErrStoreNil   = errors.New("cache: store cannot be nil")
)
