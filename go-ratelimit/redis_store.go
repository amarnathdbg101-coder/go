package ratelimit

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// RedisClient defines the minimal interface required by go-ratelimit to interact with Redis.
// Compatible with go-redis/v8, go-redis/v9, redigo, or custom Redis clients.
type RedisClient interface {
	Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error)
}

// RedisStore provides atomic, distributed rate limiting across microservice clusters using Redis Lua Script.
type RedisStore struct {
	client RedisClient
	prefix string
}

// NewRedisStore creates a Redis-backed rate limit store.
func NewRedisStore(client RedisClient, prefix string) *RedisStore {
	if prefix == "" {
		prefix = "ratelimit"
	}
	return &RedisStore{
		client: client,
		prefix: prefix,
	}
}

// Lua script executing sliding window rate limit atomically in Redis.
const redisSlidingWindowLua = `
local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local windowSeconds = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local reqID = ARGV[4]
local clearBefore = now - windowSeconds

-- Remove expired entries older than window
redis.call('ZREMRANGEBYSCORE', key, 0, clearBefore)

-- Count active entries in window
local count = redis.call('ZCARD', key)

if count < capacity then
    redis.call('ZADD', key, now, now .. ':' .. reqID)
    redis.call('EXPIRE', key, windowSeconds)
    return {1, capacity - count - 1}
else
    return {0, 0}
end
`

// Take executes the atomic Lua script in Redis.
func (s *RedisStore) Take(ctx context.Context, key string, rule Rule) (Result, error) {
	if s.client == nil {
		return Result{}, fmt.Errorf("ratelimit: redis client is nil")
	}

	redisKey := fmt.Sprintf("%s:%s", s.prefix, key)
	now := time.Now()
	nowUnix := now.Unix()
	reqID := randomHex(4)

	res, err := s.client.Eval(ctx, redisSlidingWindowLua, []string{redisKey}, rule.Capacity, int64(rule.Window.Seconds()), nowUnix, reqID)
	if err != nil {
		return Result{}, fmt.Errorf("ratelimit: redis eval error: %w", err)
	}

	arr, ok := res.([]interface{})
	if !ok || len(arr) < 2 {
		return Result{Allowed: true}, nil // Fallback allow on parsing edge cases
	}

	allowedInt, _ := arr[0].(int64)
	remainingInt, _ := arr[1].(int64)

	allowed := allowedInt == 1
	remaining := int(remainingInt)

	var retryAfter time.Duration
	if !allowed {
		retryAfter = rule.Window
	}

	return Result{
		Allowed:    allowed,
		Remaining:  remaining,
		ResetAt:    now.Add(rule.Window),
		RetryAfter: retryAfter,
	}, nil
}

// Close satisfies Store interface.
func (s *RedisStore) Close() error {
	return nil
}

func randomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
