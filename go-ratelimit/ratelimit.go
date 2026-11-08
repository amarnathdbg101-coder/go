package ratelimit

import (
	"context"
	"net/http"
	"time"
)

// Limiter is the central rate limiting engine.
type Limiter struct {
	store           Store
	rule            Rule
	keyFunc         KeyFunc
	onLimitExceeded ErrorHandler
}

// Option configures Limiter functional options.
type Option func(*Limiter)

// WithKeyFunc overrides default IP key extractor with custom key logic (e.g., API Key, UserID).
func WithKeyFunc(fn KeyFunc) Option {
	return func(l *Limiter) {
		if fn != nil {
			l.keyFunc = fn
		}
	}
}

// WithErrorHandler customizes the HTTP 429 response renderer.
func WithErrorHandler(fn ErrorHandler) Option {
	return func(l *Limiter) {
		if fn != nil {
			l.onLimitExceeded = fn
		}
	}
}

// New creates and initializes a ready-to-use Limiter instance.
// If store is nil, defaults to high-performance in-memory Token Bucket store.
func New(store Store, rule Rule, opts ...Option) (*Limiter, error) {
	if rule.Capacity <= 0 {
		rule.Capacity = 100
	}
	if rule.Window <= 0 {
		rule.Window = 1 * time.Minute
	}

	if store == nil {
		store = NewMemoryStore()
	}

	limiter := &Limiter{
		store:           store,
		rule:            rule,
		keyFunc:         KeyByIP,
		onLimitExceeded: DefaultErrorHandler,
	}

	for _, opt := range opts {
		opt(limiter)
	}

	return limiter, nil
}

// NewIPLimiter creates a convenience rate limiter configured by client IP.
func NewIPLimiter(capacity int, window time.Duration) (*Limiter, error) {
	return New(NewMemoryStore(), Rule{Capacity: capacity, Window: window})
}

// Take evaluates rate limit for an incoming HTTP request.
func (l *Limiter) Take(r *http.Request) (Result, error) {
	key := l.keyFunc(r)
	return l.store.Take(r.Context(), key, l.rule)
}

// TakeKey evaluates rate limit for an explicit string key.
func (l *Limiter) TakeKey(ctx context.Context, key string) (Result, error) {
	return l.store.Take(ctx, key, l.rule)
}

// Store returns the underlying storage implementation.
func (l *Limiter) Store() Store {
	return l.store
}

// Rule returns the active rate limit rule.
func (l *Limiter) Rule() Rule {
	return l.rule
}

// Close gracefully stops storage background cleanup.
func (l *Limiter) Close() error {
	if l.store != nil {
		return l.store.Close()
	}
	return nil
}
