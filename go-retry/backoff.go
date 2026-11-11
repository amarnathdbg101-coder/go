package goretry

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"math"
	"sync"
	"time"
)

// JitterMode specifies how randomness is added to avoid thundering herd problem.
type JitterMode int

const (
	FullJitter JitterMode = iota
	EqualJitter
	DecorrelatedJitter
	NoJitter
)

// Backoff computes delay between retry attempts.
type Backoff interface {
	NextDelay(attempt int) time.Duration
	Sleep(ctx context.Context, attempt int) error
}

type exponentialBackoff struct {
	initial    time.Duration
	max        time.Duration
	multiplier float64
	mu         sync.Mutex
	prevSleep  time.Duration
}

// Exponential creates an exponential backoff strategy (initial * multiplier^attempt).
func Exponential(initial, max time.Duration, multiplier float64) Backoff {
	if initial <= 0 {
		initial = 100 * time.Millisecond
	}
	if max <= 0 {
		max = 5 * time.Second
	}
	if multiplier <= 1.0 {
		multiplier = 2.0
	}
	return &exponentialBackoff{
		initial:    initial,
		max:        max,
		multiplier: multiplier,
		prevSleep:  initial,
	}
}

func (e *exponentialBackoff) NextDelay(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}
	e.mu.Lock()
	defer e.mu.Unlock()

	base := float64(e.initial)
	max := float64(e.max)

	// Guard against float overflow on large attempt numbers
	calculated := base * math.Pow(e.multiplier, float64(attempt))
	if calculated > max || calculated < 0 {
		calculated = max
	}

	d := time.Duration(calculated)
	e.prevSleep = d
	return d
}

func (e *exponentialBackoff) Sleep(ctx context.Context, attempt int) error {
	delay := e.NextDelay(attempt)
	return sleepWithContext(ctx, delay)
}

type decorrelatedBackoff struct {
	base      time.Duration
	max       time.Duration
	prevSleep time.Duration
	mu        sync.Mutex
}

// Decorrelated creates an AWS-recommended decorrelated jitter backoff strategy:
// sleep = min(max, rand(base, prevSleep * 3))
func Decorrelated(base, max time.Duration) Backoff {
	if base <= 0 {
		base = 100 * time.Millisecond
	}
	if max <= 0 {
		max = 10 * time.Second
	}
	return &decorrelatedBackoff{
		base:      base,
		max:       max,
		prevSleep: base,
	}
}

func (d *decorrelatedBackoff) NextDelay(attempt int) time.Duration {
	d.mu.Lock()
	defer d.mu.Unlock()

	base := float64(d.base)
	max := float64(d.max)
	rangeVal := float64(d.prevSleep*3) - base
	if rangeVal < 0 {
		rangeVal = 0
	}

	sleep := time.Duration(base + (randomFloat() * rangeVal))
	if sleep > time.Duration(max) {
		sleep = time.Duration(max)
	}

	d.prevSleep = sleep
	return sleep
}

func (d *decorrelatedBackoff) Sleep(ctx context.Context, attempt int) error {
	delay := d.NextDelay(attempt)
	return sleepWithContext(ctx, delay)
}

type constantBackoff struct {
	interval time.Duration
}

// Constant creates a fixed interval backoff strategy.
func Constant(interval time.Duration) Backoff {
	return &constantBackoff{interval: interval}
}

func (c *constantBackoff) NextDelay(attempt int) time.Duration {
	return c.interval
}

func (c *constantBackoff) Sleep(ctx context.Context, attempt int) error {
	return sleepWithContext(ctx, c.interval)
}

type linearBackoff struct {
	initial time.Duration
	step    time.Duration
	max     time.Duration
}

// Linear creates a linear increasing backoff strategy (initial + attempt*step).
func Linear(initial, step, max time.Duration) Backoff {
	return &linearBackoff{
		initial: initial,
		step:    step,
		max:     max,
	}
}

func (l *linearBackoff) NextDelay(attempt int) time.Duration {
	d := l.initial + time.Duration(attempt)*l.step
	if d > l.max {
		d = l.max
	}
	return d
}

func (l *linearBackoff) Sleep(ctx context.Context, attempt int) error {
	delay := l.NextDelay(attempt)
	return sleepWithContext(ctx, delay)
}

// ApplyJitter applies the specified jitter algorithm to a base duration.
// Uses cryptographically secure random numbers to avoid math/rand global lock contention.
func ApplyJitter(d time.Duration, mode JitterMode) time.Duration {
	if d <= 0 {
		return 0
	}

	switch mode {
	case FullJitter:
		// Random between 0 and d
		return time.Duration(randomFloat() * float64(d))

	case EqualJitter:
		// Half fixed, half random: d/2 + rand(0, d/2)
		half := float64(d) / 2.0
		return time.Duration(half + randomFloat()*half)

	case DecorrelatedJitter:
		// Random between d*0.5 and d*1.5
		return time.Duration(float64(d)*0.5 + randomFloat()*float64(d))

	case NoJitter:
		return d

	default:
		return d
	}
}

// sleepWithContext pauses execution until duration elapses or context is cancelled.
// Leak-proof timer management: Stop() is always called on exit.
func sleepWithContext(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}
	t := time.NewTimer(d)
	defer t.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// randomFloat generates a uniformly distributed float64 in [0.0, 1.0) using crypto/rand.
func randomFloat() float64 {
	var b [8]byte
	_, err := rand.Read(b[:])
	if err != nil {
		return 0.5
	}
	// Extract 53 random bits matching IEEE 754 double precision significand
	return float64(binary.LittleEndian.Uint64(b[:])&(1<<53-1)) / (1 << 53)
}
