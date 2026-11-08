package resilience

import (
	"context"
	"crypto/rand"
	"math/big"
	"sync"
	"time"
)

type circuitState int

const (
	stateClosed circuitState = iota
	stateOpen
	stateHalfOpen
)

// RetryConfig configures the retry and exponential backoff parameters.
type RetryConfig struct {
	MaxAttempts       int
	InitialInterval   time.Duration
	BackoffMultiplier float64
	MaxInterval       time.Duration
	Jitter            bool
}

// CircuitBreakerConfig configures failure thresholds and reset intervals for the circuit breaker.
type CircuitBreakerConfig struct {
	FailureThreshold int
	ResetTimeout     time.Duration
}

// Policy coordinates execution through timeout, retry, circuit breaker, bulkhead, and fallback layers.
type Policy[T any] struct {
	timeout time.Duration

	retry *RetryConfig

	cbMu         sync.Mutex
	cbConfig     *CircuitBreakerConfig
	cbState      circuitState
	failureCount int
	lastTripTime time.Time

	bulkheadSem chan struct{}

	fallback func(ctx context.Context, err error) (T, error)
}

// NewPolicy creates a new, unconfigured resilience Policy pipeline for type T.
func NewPolicy[T any]() *Policy[T] {
	return &Policy[T]{}
}

// WithTimeout sets a per-attempt execution deadline timeout.
func (p *Policy[T]) WithTimeout(d time.Duration) *Policy[T] {
	p.timeout = d
	return p
}

// WithRetry configures retry policies with backoff and jitter.
func (p *Policy[T]) WithRetry(cfg RetryConfig) *Policy[T] {
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 3
	}
	if cfg.InitialInterval <= 0 {
		cfg.InitialInterval = 50 * time.Millisecond
	}
	if cfg.BackoffMultiplier <= 1.0 {
		cfg.BackoffMultiplier = 2.0
	}
	if cfg.MaxInterval <= 0 {
		cfg.MaxInterval = 2 * time.Second
	}
	p.retry = &cfg
	return p
}

// WithCircuitBreaker enables automatic fault isolation via circuit breaking.
func (p *Policy[T]) WithCircuitBreaker(cfg CircuitBreakerConfig) *Policy[T] {
	if cfg.FailureThreshold <= 0 {
		cfg.FailureThreshold = 5
	}
	if cfg.ResetTimeout <= 0 {
		cfg.ResetTimeout = 10 * time.Second
	}
	p.cbConfig = &cfg
	p.cbState = stateClosed
	return p
}

// WithBulkhead limits max concurrent in-flight executions.
func (p *Policy[T]) WithBulkhead(maxConcurrent int) *Policy[T] {
	if maxConcurrent > 0 {
		p.bulkheadSem = make(chan struct{}, maxConcurrent)
	}
	return p
}

// WithFallback specifies a degradation fallback function invoked when execution fails.
func (p *Policy[T]) WithFallback(fn func(ctx context.Context, err error) (T, error)) *Policy[T] {
	p.fallback = fn
	return p
}

// Execute coordinates the execution of fn through all configured policy stages.
func (p *Policy[T]) Execute(ctx context.Context, fn func(ctx context.Context) (T, error)) (T, error) {
	// 1. Bulkhead Concurrency Guard
	if p.bulkheadSem != nil {
		select {
		case p.bulkheadSem <- struct{}{}:
			defer func() { <-p.bulkheadSem }()
		default:
			return p.handleFallback(ctx, ErrBulkheadFull)
		}
	}

	// 2. Circuit Breaker State Check
	if p.cbConfig != nil {
		if !p.allowCircuitRequest() {
			return p.handleFallback(ctx, ErrCircuitOpen)
		}
	}

	// 3. Execution with Retries and Timeouts
	var lastErr error
	maxAttempts := 1
	if p.retry != nil {
		maxAttempts = p.retry.MaxAttempts
	}

	interval := 50 * time.Millisecond
	if p.retry != nil {
		interval = p.retry.InitialInterval
	}

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// Attempt Timeout
		attemptCtx := ctx
		var cancel context.CancelFunc
		if p.timeout > 0 {
			attemptCtx, cancel = context.WithTimeout(ctx, p.timeout)
		}

		res, err := fn(attemptCtx)
		if cancel != nil {
			cancel()
		}

		if err == nil {
			p.reportCircuitSuccess()
			return res, nil
		}

		lastErr = err

		// Check if context was cancelled by caller
		if ctx.Err() != nil {
			return p.handleFallback(ctx, ctx.Err())
		}

		// Backoff sleep if not last attempt
		if attempt < maxAttempts {
			sleepDuration := interval
			if p.retry != nil && p.retry.Jitter {
				sleepDuration = applyJitter(sleepDuration)
			}

			select {
			case <-time.After(sleepDuration):
			case <-ctx.Done():
				return p.handleFallback(ctx, ctx.Err())
			}

			if p.retry != nil {
				interval = time.Duration(float64(interval) * p.retry.BackoffMultiplier)
				if interval > p.retry.MaxInterval {
					interval = p.retry.MaxInterval
				}
			}
		}
	}

	p.reportCircuitFailure()
	return p.handleFallback(ctx, lastErr)
}

func (p *Policy[T]) allowCircuitRequest() bool {
	p.cbMu.Lock()
	defer p.cbMu.Unlock()

	now := time.Now()
	switch p.cbState {
	case stateClosed:
		return true
	case stateOpen:
		if now.Sub(p.lastTripTime) >= p.cbConfig.ResetTimeout {
			p.cbState = stateHalfOpen
			return true // Canary trial probe
		}
		return false
	case stateHalfOpen:
		return true
	}
	return true
}

func (p *Policy[T]) reportCircuitSuccess() {
	if p.cbConfig == nil {
		return
	}
	p.cbMu.Lock()
	defer p.cbMu.Unlock()

	p.failureCount = 0
	p.cbState = stateClosed
}

func (p *Policy[T]) reportCircuitFailure() {
	if p.cbConfig == nil {
		return
	}
	p.cbMu.Lock()
	defer p.cbMu.Unlock()

	p.failureCount++
	if p.failureCount >= p.cbConfig.FailureThreshold || p.cbState == stateHalfOpen {
		p.cbState = stateOpen
		p.lastTripTime = time.Now()
	}
}

func (p *Policy[T]) handleFallback(ctx context.Context, err error) (T, error) {
	if p.fallback != nil {
		return p.fallback(ctx, err)
	}
	var zero T
	return zero, err
}

func applyJitter(d time.Duration) time.Duration {
	if d <= 0 {
		return d
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(d/2)))
	if err != nil {
		return d
	}
	return d/2 + time.Duration(n.Int64())
}
