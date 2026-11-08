package cachekit

import (
	"context"
	"fmt"
	"time"
)

// RedisClient defines the minimal interface required by go-cachekit for Redis storage.
// Compatible with go-redis/v8, go-redis/v9, redigo, or custom Redis clients.
type RedisClient interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Del(ctx context.Context, keys ...string) error
	Keys(ctx context.Context, pattern string) ([]string, error)
	FlushDB(ctx context.Context) error
}

// RedisStore provides a production-grade L2 remote cache backed by Redis.
type RedisStore struct {
	client RedisClient
	prefix string
}

// NewRedisStore initializes a Redis cache store with optional key prefix.
func NewRedisStore(client RedisClient, prefix string) *RedisStore {
	return &RedisStore{
		client: client,
		prefix: prefix,
	}
}

func (s *RedisStore) formatKey(key string) string {
	if s.prefix == "" {
		return key
	}
	return fmt.Sprintf("%s:%s", s.prefix, key)
}

// Get retrieves raw bytes from Redis.
func (s *RedisStore) Get(ctx context.Context, key string) ([]byte, error) {
	if s.client == nil {
		return nil, ErrStoreNil
	}
	val, err := s.client.Get(ctx, s.formatKey(key))
	if err != nil {
		return nil, ErrCacheMiss
	}
	return val, nil
}

// Set stores raw bytes with specified TTL in Redis.
func (s *RedisStore) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if s.client == nil {
		return ErrStoreNil
	}
	if ttl <= 0 {
		return ErrInvalidTTL
	}
	return s.client.Set(ctx, s.formatKey(key), value, ttl)
}

// Delete evicts a key from Redis.
func (s *RedisStore) Delete(ctx context.Context, key string) error {
	if s.client == nil {
		return ErrStoreNil
	}
	return s.client.Del(ctx, s.formatKey(key))
}

// DeletePrefix evicts all keys matching prefix from Redis.
func (s *RedisStore) DeletePrefix(ctx context.Context, prefix string) error {
	if s.client == nil {
		return ErrStoreNil
	}
	pattern := s.formatKey(prefix + "*")
	keys, err := s.client.Keys(ctx, pattern)
	if err != nil || len(keys) == 0 {
		return nil
	}
	return s.client.Del(ctx, keys...)
}

// Clear flushes Redis database.
func (s *RedisStore) Clear(ctx context.Context) error {
	if s.client == nil {
		return ErrStoreNil
	}
	return s.client.FlushDB(ctx)
}

// Close satisfies Store interface.
func (s *RedisStore) Close() error {
	return nil
}
