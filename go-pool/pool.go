package gopool

import (
	"context"
	"io"
	"sync"
	"time"
)

type pooledItem[T any] struct {
	val      T
	lastUsed time.Time
}

// Config specifies pool parameters.
type Config[T any] struct {
	Capacity    int
	Factory     func(ctx context.Context) (T, error)
	CloseFn     func(item T) error
	HealthCheck func(item T) bool
	IdleTimeout time.Duration
}

// Pool is a bounded, generic resource pool with health validation and idle eviction.
type Pool[T any] struct {
	cfg     Config[T]
	items   chan pooledItem[T]
	sem     chan struct{} // Limits total allocated resources (idle + in-use)
	closed  bool
	closeMu sync.RWMutex
}

// New initializes a bounded Pool.
func New[T any](cfg Config[T]) (*Pool[T], error) {
	if cfg.Capacity <= 0 || cfg.Factory == nil {
		return nil, ErrInvalidConfig
	}

	p := &Pool[T]{
		cfg:   cfg,
		items: make(chan pooledItem[T], cfg.Capacity),
		sem:   make(chan struct{}, cfg.Capacity),
	}

	return p, nil
}

// Acquire retrieves an available resource or instantiates a new one if below capacity.
// Respects context cancellation.
func (p *Pool[T]) Acquire(ctx context.Context) (T, error) {
	var zero T

	p.closeMu.RLock()
	if p.closed {
		p.closeMu.RUnlock()
		return zero, ErrPoolClosed
	}
	p.closeMu.RUnlock()

	for {
		select {
		case <-ctx.Done():
			return zero, ctx.Err()

		case it := <-p.items:
			// Check idle timeout
			if p.cfg.IdleTimeout > 0 && time.Since(it.lastUsed) > p.cfg.IdleTimeout {
				p.destroy(it.val)
				continue
			}

			// Check health
			if p.cfg.HealthCheck != nil && !p.cfg.HealthCheck(it.val) {
				p.destroy(it.val)
				continue
			}

			return it.val, nil

		default:
			// Try acquire a slot to create a new instance
			select {
			case p.sem <- struct{}{}:
				val, err := p.cfg.Factory(ctx)
				if err != nil {
					<-p.sem // release slot on factory error
					return zero, err
				}
				return val, nil

			case it := <-p.items:
				// An item was released while waiting for slot
				if p.cfg.IdleTimeout > 0 && time.Since(it.lastUsed) > p.cfg.IdleTimeout {
					p.destroy(it.val)
					continue
				}
				if p.cfg.HealthCheck != nil && !p.cfg.HealthCheck(it.val) {
					p.destroy(it.val)
					continue
				}
				return it.val, nil

			case <-ctx.Done():
				return zero, ctx.Err()
			}
		}
	}
}

// Release returns a healthy resource back to the pool for reuse.
func (p *Pool[T]) Release(item T) {
	p.closeMu.RLock()
	defer p.closeMu.RUnlock()

	if p.closed {
		p.destroy(item)
		return
	}

	select {
	case p.items <- pooledItem[T]{val: item, lastUsed: time.Now()}:
	default:
		// Queue saturated, destroy
		p.destroy(item)
	}
}

// Discard marks a resource as unhealthy and destroys it without returning to pool.
func (p *Pool[T]) Discard(item T) {
	p.destroy(item)
}

func (p *Pool[T]) destroy(item T) {
	if p.cfg.CloseFn != nil {
		_ = p.cfg.CloseFn(item)
	} else if closer, ok := any(item).(io.Closer); ok {
		_ = closer.Close()
	}

	// Release semaphore slot
	select {
	case <-p.sem:
	default:
	}
}

// Idle returns count of ready idle resources.
func (p *Pool[T]) Idle() int {
	return len(p.items)
}

// Active returns count of resources currently in-use outside the pool.
func (p *Pool[T]) Active() int {
	return len(p.sem) - len(p.items)
}

// Close closes the pool and disposes of all idle resources.
func (p *Pool[T]) Close() error {
	p.closeMu.Lock()
	if p.closed {
		p.closeMu.Unlock()
		return nil
	}
	p.closed = true
	p.closeMu.Unlock()

	close(p.items)
	for it := range p.items {
		p.destroy(it.val)
	}

	return nil
}
