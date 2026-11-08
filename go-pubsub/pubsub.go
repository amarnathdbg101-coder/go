package gopubsub

import (
	"sync"
	"sync/atomic"
)

// OverflowStrategy defines how to handle slow consumers whose channel buffer is full.
type OverflowStrategy int

const (
	// DropOldest drops the oldest pending message to make room for the latest message (ideal for live GPS/telemetry).
	DropOldest OverflowStrategy = iota
	// DropNew discards the new message without blocking other subscribers.
	DropNew
)

type subscriber[T any] struct {
	id     uint64
	ch     chan T
	topic  string
	ps     *PubSub[T]
	closed bool
}

// Subscription represents an active channel subscription to a topic.
type Subscription[T any] struct {
	sub *subscriber[T]
	Ch  <-chan T
}

// Unsubscribe cancels the subscription and removes it from the topic.
func (s *Subscription[T]) Unsubscribe() {
	s.sub.ps.unsubscribe(s.sub.topic, s.sub.id)
}

// PubSub provides thread-safe, non-blocking topic-based pub-sub fanout.
type PubSub[T any] struct {
	mu         sync.RWMutex
	topics     map[string]map[uint64]*subscriber[T]
	nextID     uint64
	bufferSize int
	strategy   OverflowStrategy
	closed     bool
}

// New constructs a PubSub engine with specified subscriber channel buffer size and overflow policy.
func New[T any](bufferSize int, strategy OverflowStrategy) *PubSub[T] {
	if bufferSize <= 0 {
		bufferSize = 64
	}

	return &PubSub[T]{
		topics:     make(map[string]map[uint64]*subscriber[T]),
		bufferSize: bufferSize,
		strategy:   strategy,
	}
}

// Subscribe registers a new subscriber channel on the specified topic.
func (ps *PubSub[T]) Subscribe(topic string) *Subscription[T] {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.closed {
		ch := make(chan T)
		close(ch)
		return &Subscription[T]{Ch: ch}
	}

	subID := atomic.AddUint64(&ps.nextID, 1)
	ch := make(chan T, ps.bufferSize)

	sub := &subscriber[T]{
		id:    subID,
		ch:    ch,
		topic: topic,
		ps:    ps,
	}

	if _, exists := ps.topics[topic]; !exists {
		ps.topics[topic] = make(map[uint64]*subscriber[T])
	}
	ps.topics[topic][subID] = sub

	return &Subscription[T]{
		sub: sub,
		Ch:  ch,
	}
}

func (ps *PubSub[T]) unsubscribe(topic string, id uint64) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	subs, ok := ps.topics[topic]
	if !ok {
		return
	}

	sub, ok := subs[id]
	if !ok {
		return
	}

	if !sub.closed {
		sub.closed = true
		close(sub.ch)
	}

	delete(subs, id)
	if len(subs) == 0 {
		delete(ps.topics, topic)
	}
}

// Publish broadcasts a message to all active subscribers of the topic non-blockingly.
// Returns the number of subscribers receiving the message.
func (ps *PubSub[T]) Publish(topic string, msg T) int {
	ps.mu.RLock()
	if ps.closed {
		ps.mu.RUnlock()
		return 0
	}

	subsMap, ok := ps.topics[topic]
	if !ok || len(subsMap) == 0 {
		ps.mu.RUnlock()
		return 0
	}

	// Snapshot active subscribers to release lock during send
	subs := make([]*subscriber[T], 0, len(subsMap))
	for _, sub := range subsMap {
		subs = append(subs, sub)
	}
	ps.mu.RUnlock()

	delivered := 0
	for _, sub := range subs {
		if ps.deliver(sub, msg) {
			delivered++
		}
	}

	return delivered
}

func (ps *PubSub[T]) deliver(sub *subscriber[T], msg T) bool {
	select {
	case sub.ch <- msg:
		return true

	default:
		// Channel buffer is full: apply overflow strategy
		switch ps.strategy {
		case DropOldest:
			// Discard oldest message
			select {
			case <-sub.ch:
			default:
			}
			// Enqueue latest
			select {
			case sub.ch <- msg:
				return true
			default:
				return false
			}

		case DropNew:
			// Drop this message for this slow consumer
			return false
		}
	}
	return false
}

// SubscribersCount returns the number of active subscribers on a topic.
func (ps *PubSub[T]) SubscribersCount(topic string) int {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if subs, ok := ps.topics[topic]; ok {
		return len(subs)
	}
	return 0
}

// Close closes all subscriber channels and shuts down the PubSub engine.
func (ps *PubSub[T]) Close() {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.closed {
		return
	}
	ps.closed = true

	for _, subs := range ps.topics {
		for _, sub := range subs {
			if !sub.closed {
				sub.closed = true
				close(sub.ch)
			}
		}
	}
	ps.topics = nil
}
