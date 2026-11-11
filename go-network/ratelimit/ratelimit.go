package ratelimit

import (
	"context"
	"net/http"
	"time"
)

type Limiter struct {
	store           Store
	rule            Rule
	keyFunc         KeyFunc
	onLimitExceeded ErrorHandler
}

type Option func(*Limiter)

func WithKeyFunc(fn KeyFunc) Option {
	return func(l *Limiter) {
		if fn != nil {
			l.keyFunc = fn
		}
	}
}

func WithErrorHandler(fn ErrorHandler) Option {
	return func(l *Limiter) {
		if fn != nil {
			l.onLimitExceeded = fn
		}
	}
}

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

func NewIPLimiter(capacity int, window time.Duration) (*Limiter, error) {
	return New(NewMemoryStore(), Rule{Capacity: capacity, Window: window})
}

func (l *Limiter) Take(r *http.Request) (Result, error) {
	key := l.keyFunc(r)
	return l.store.Take(r.Context(), key, l.rule)
}

func (l *Limiter) TakeKey(ctx context.Context, key string) (Result, error) {
	return l.store.Take(ctx, key, l.rule)
}

func (l *Limiter) Store() Store {
	return l.store
}

func (l *Limiter) Rule() Rule {
	return l.rule
}

func (l *Limiter) Close() error {
	if l.store != nil {
		return l.store.Close()
	}
	return nil
}
