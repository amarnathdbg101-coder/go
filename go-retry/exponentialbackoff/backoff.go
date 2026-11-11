package goexponentialbackoff

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"math"
	"sync"
	"time"
)

// Algorithm specifies the jitter strategy.
type Algorithm int

const (
	FullJitter Algorithm = iota
	EqualJitter
	DecorrelatedJitter
	NoJitter
)

// Config specifies backoff limits and parameters.
type Config struct {
	Base      time.Duration
	Max       time.Duration
	Factor    float64
	Algorithm Algorithm
}

// DefaultConfig provides recommended production retry settings.
func DefaultConfig() Config {
	return Config{
		Base:      100 * time.Millisecond,
		Max:       10 * time.Second,
		Factor:    2.0,
		Algorithm: FullJitter,
	}
}

// Backoff calculates delays between retries using exponential backoff and jitter.
type Backoff struct {
	cfg       Config
	prevSleep time.Duration
	mu        sync.Mutex
}

// New initializes Backoff with config.
func New(cfg Config) *Backoff {
	if cfg.Base <= 0 {
		cfg.Base = 100 * time.Millisecond
	}
	if cfg.Max <= 0 {
		cfg.Max = 10 * time.Second
	}
	if cfg.Factor <= 1.0 {
		cfg.Factor = 2.0
	}

	return &Backoff{
		cfg:       cfg,
		prevSleep: cfg.Base,
	}
}

// Next calculates the duration to sleep before attempt number (0-indexed or 1-indexed).
func (b *Backoff) Next(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	base := float64(b.cfg.Base)
	max := float64(b.cfg.Max)

	// Ceiling calculation: min(max, base * factor^attempt)
	ceiling := base * math.Pow(b.cfg.Factor, float64(attempt))
	if ceiling > max || ceiling < 0 { // overflow check
		ceiling = max
	}

	var sleep time.Duration

	switch b.cfg.Algorithm {
	case FullJitter:
		// sleep = rand(0, ceiling)
		sleep = time.Duration(randomFloat() * ceiling)

	case EqualJitter:
		// sleep = ceiling/2 + rand(0, ceiling/2)
		half := ceiling / 2.0
		sleep = time.Duration(half + (randomFloat() * half))

	case DecorrelatedJitter:
		// sleep = min(max, rand(base, prevSleep * 3))
		rangeVal := float64(b.prevSleep*3) - base
		if rangeVal < 0 {
			rangeVal = 0
		}
		candidate := base + (randomFloat() * rangeVal)
		if candidate > max {
			candidate = max
		}
		sleep = time.Duration(candidate)

	case NoJitter:
		sleep = time.Duration(ceiling)
	}

	if sleep < b.cfg.Base && b.cfg.Algorithm != FullJitter {
		sleep = b.cfg.Base
	}
	if sleep > b.cfg.Max {
		sleep = b.cfg.Max
	}

	b.prevSleep = sleep
	return sleep
}

// Sleep delays the calling goroutine by Next(attempt) or terminates early if context cancels.
func (b *Backoff) Sleep(ctx context.Context, attempt int) error {
	delay := b.Next(attempt)
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// Reset restores internal state.
func (b *Backoff) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.prevSleep = b.cfg.Base
}

func randomFloat() float64 {
	var b [8]byte
	_, _ = rand.Read(b[:])
	val := binary.LittleEndian.Uint64(b[:])
	return float64(val&(1<<53-1)) / float64(1<<53)
}
