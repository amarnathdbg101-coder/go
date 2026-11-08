package goevents

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
)

type subscription struct {
	id      uint64
	pattern string
	handler HandlerFunc
	bus     *Bus
}

// Unsubscribe removes the subscription from the bus.
func (s *subscription) Unsubscribe() {
	s.bus.unsubscribe(s.id)
}

type asyncItem struct {
	ctx   context.Context
	event Event
}

// Config tunes the EventBus workers and queue.
type Config struct {
	WorkerCount int
	QueueSize   int
	OnPanic     func(topic string, recovered any)
	OnDLQ       func(event Event, err error)
}

// Bus provides thread-safe synchronous and asynchronous event pub-sub.
type Bus struct {
	mu          sync.RWMutex
	subs        map[uint64]*subscription
	nextSubID   uint64
	middlewares []MiddlewareFunc
	asyncChan   chan asyncItem
	wg          sync.WaitGroup
	closed      bool
	config      Config
}

// New constructs an EventBus with worker pool for async events.
func New(cfg ...Config) *Bus {
	c := Config{
		WorkerCount: 4,
		QueueSize:   1024,
		OnPanic: func(topic string, r any) {
			// Default no-op log or ignore
		},
		OnDLQ: nil,
	}

	if len(cfg) > 0 {
		if cfg[0].WorkerCount > 0 {
			c.WorkerCount = cfg[0].WorkerCount
		}
		if cfg[0].QueueSize > 0 {
			c.QueueSize = cfg[0].QueueSize
		}
		if cfg[0].OnPanic != nil {
			c.OnPanic = cfg[0].OnPanic
		}
		if cfg[0].OnDLQ != nil {
			c.OnDLQ = cfg[0].OnDLQ
		}
	}

	b := &Bus{
		subs:      make(map[uint64]*subscription),
		asyncChan: make(chan asyncItem, c.QueueSize),
		config:    c,
	}

	// Start worker pool
	for i := 0; i < c.WorkerCount; i++ {
		b.wg.Add(1)
		go b.worker()
	}

	return b
}

// Use appends one or more middlewares to the bus pipeline.
func (b *Bus) Use(middlewares ...MiddlewareFunc) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.middlewares = append(b.middlewares, middlewares...)
}

// Subscribe registers a listener for topics matching the pattern.
func (b *Bus) Subscribe(pattern string, handler HandlerFunc) (*subscription, error) {
	if pattern == "" {
		return nil, ErrInvalidPattern
	}
	if handler == nil {
		return nil, fmt.Errorf("events: handler cannot be nil")
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil, ErrBusClosed
	}

	subID := atomic.AddUint64(&b.nextSubID, 1)
	sub := &subscription{
		id:      subID,
		pattern: pattern,
		handler: handler,
		bus:     b,
	}
	b.subs[subID] = sub

	return sub, nil
}

func (b *Bus) unsubscribe(id uint64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.subs, id)
}

// Publish dispatches an event synchronously to all matching subscribers.
// Any panic in an individual subscriber is caught so other listeners complete.
func (b *Bus) Publish(ctx context.Context, event Event) error {
	b.mu.RLock()
	if b.closed {
		b.mu.RUnlock()
		return ErrBusClosed
	}

	// Collect matching subscriptions
	var matchedHandlers []HandlerFunc
	for _, sub := range b.subs {
		if MatchTopic(sub.pattern, event.Topic) {
			matchedHandlers = append(matchedHandlers, sub.handler)
		}
	}
	middlewares := b.middlewares
	b.mu.RUnlock()

	if len(matchedHandlers) == 0 {
		if b.config.OnDLQ != nil {
			b.config.OnDLQ(event, fmt.Errorf("no subscriber matched topic: %s", event.Topic))
		}
		return nil
	}

	var firstErr error
	for _, handler := range matchedHandlers {
		h := handler
		// Apply middlewares
		for i := len(middlewares) - 1; i >= 0; i-- {
			h = middlewares[i](h)
		}

		if err := b.safeExecute(ctx, h, event); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			if b.config.OnDLQ != nil {
				b.config.OnDLQ(event, err)
			}
		}
	}

	return firstErr
}

// PublishAsync queues the event to be processed concurrently by the background worker pool.
func (b *Bus) PublishAsync(event Event) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return ErrBusClosed
	}

	item := asyncItem{
		ctx:   context.Background(),
		event: event,
	}

	select {
	case b.asyncChan <- item:
		return nil
	default:
		return ErrQueueFull
	}
}

func (b *Bus) worker() {
	defer b.wg.Done()

	for item := range b.asyncChan {
		_ = b.Publish(item.ctx, item.event)
	}
}

func (b *Bus) safeExecute(ctx context.Context, h HandlerFunc, e Event) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if b.config.OnPanic != nil {
				b.config.OnPanic(e.Topic, r)
			}
			err = fmt.Errorf("events: handler panic on topic '%s': %v", e.Topic, r)
		}
	}()

	return h(ctx, e)
}

// Close shuts down the bus gracefully, draining all pending asynchronous events.
func (b *Bus) Close() error {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return nil
	}
	b.closed = true
	close(b.asyncChan)
	b.mu.Unlock()

	// Wait for worker pool to finish draining queue
	b.wg.Wait()
	return nil
}

// SubscriptionsCount returns active subscriber count.
func (b *Bus) SubscriptionsCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.subs)
}
