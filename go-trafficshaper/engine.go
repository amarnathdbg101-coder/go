package trafficshaper

import (
	"container/heap"
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type ticket struct {
	id       uint64
	priority int
	index    int
	ready    chan struct{}
	canceled bool
}

type priorityQueue []*ticket

func (pq priorityQueue) Len() int { return len(pq) }

func (pq priorityQueue) Less(i, j int) bool {
	// Higher priority first
	if pq[i].priority != pq[j].priority {
		return pq[i].priority > pq[j].priority
	}
	// For same priority, FIFO (lower id first)
	return pq[i].id < pq[j].id
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *priorityQueue) Push(x any) {
	n := len(*pq)
	item := x.(*ticket)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *priorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[0 : n-1]
	return item
}

// Stats provides real-time telemetry on the shaper.
type Stats struct {
	Admitted uint64
	Queued   uint64
	Dropped  uint64
	InQueue  int
}

// Shaper coordinates token allocation and pacing queues.
type Shaper struct {
	cfg        Config
	mu         sync.Mutex
	pq         priorityQueue
	nextID     uint64
	tokens     float64
	lastRefill time.Time

	closed     bool
	closeChan  chan struct{}
	notifyChan chan struct{}

	admitted uint64
	queued   uint64
	dropped  uint64
}

// New creates and starts a new traffic shaper.
func New(opts ...Option) (*Shaper, error) {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	if cfg.Rate <= 0 {
		return nil, ErrInvalidConfig
	}
	if cfg.Burst < 1 {
		cfg.Burst = 1
	}
	if cfg.Algorithm == AlgorithmLeakyBucket {
		cfg.Burst = 1 // Leaky bucket allows 0 burst beyond 1 unit
	}

	s := &Shaper{
		cfg:        cfg,
		tokens:     float64(cfg.Burst),
		lastRefill: time.Now(),
		closeChan:  make(chan struct{}),
		notifyChan: make(chan struct{}, 1),
	}
	heap.Init(&s.pq)

	go s.dispatchLoop()
	return s, nil
}

// Acquire requests admission with default priority 0.
func (s *Shaper) Acquire(ctx context.Context) error {
	return s.AcquireWithPriority(ctx, 0)
}

// AcquireWithPriority requests admission with an explicit priority integer (higher = prioritized).
func (s *Shaper) AcquireWithPriority(ctx context.Context, priority int) error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return ErrShaperClosed
	}

	s.refillTokens()

	// If tokens are available and queue is empty, consume immediately
	if s.pq.Len() == 0 && s.tokens >= 1.0 {
		s.tokens -= 1.0
		atomic.AddUint64(&s.admitted, 1)
		s.mu.Unlock()
		return nil
	}

	// Check queue capacity
	if s.cfg.QueueCapacity > 0 && s.pq.Len() >= s.cfg.QueueCapacity {
		atomic.AddUint64(&s.dropped, 1)
		s.mu.Unlock()
		return ErrQueueFull
	}

	// Enqueue ticket
	s.nextID++
	t := &ticket{
		id:       s.nextID,
		priority: priority,
		ready:    make(chan struct{}),
	}
	heap.Push(&s.pq, t)
	atomic.AddUint64(&s.queued, 1)

	// Trigger dispatch loop
	select {
	case s.notifyChan <- struct{}{}:
	default:
	}

	s.mu.Unlock()

	// Wait for admission or context cancellation
	select {
	case <-t.ready:
		return nil
	case <-ctx.Done():
		s.mu.Lock()
		t.canceled = true
		s.mu.Unlock()
		return ctx.Err()
	case <-s.closeChan:
		return ErrShaperClosed
	}
}

func (s *Shaper) refillTokens() {
	now := time.Now()
	elapsed := now.Sub(s.lastRefill).Seconds()
	s.lastRefill = now

	s.tokens += elapsed * s.cfg.Rate
	maxTokens := float64(s.cfg.Burst)
	if s.tokens > maxTokens {
		s.tokens = maxTokens
	}
}

func (s *Shaper) dispatchLoop() {
	// Interval between token refills
	interval := time.Duration(float64(time.Second) / s.cfg.Rate)
	if interval <= 0 {
		interval = time.Millisecond
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-s.closeChan:
			return
		case <-s.notifyChan:
			s.flushQueue()
		case <-ticker.C:
			s.flushQueue()
		}
	}
}

func (s *Shaper) flushQueue() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return
	}

	s.refillTokens()

	for s.pq.Len() > 0 && s.tokens >= 1.0 {
		t := heap.Pop(&s.pq).(*ticket)
		if t.canceled {
			continue // Skip tickets that timed out while in queue
		}

		s.tokens -= 1.0
		atomic.AddUint64(&s.admitted, 1)
		close(t.ready) // Wake up caller
	}
}

// Stats returns a snapshot of runtime metrics.
func (s *Shaper) Stats() Stats {
	s.mu.Lock()
	defer s.mu.Unlock()

	return Stats{
		Admitted: atomic.LoadUint64(&s.admitted),
		Queued:   atomic.LoadUint64(&s.queued),
		Dropped:  atomic.LoadUint64(&s.dropped),
		InQueue:  s.pq.Len(),
	}
}

// Close gracefully stops the traffic shaper and cancels pending tickets.
func (s *Shaper) Close() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	close(s.closeChan)

	// Drain remaining tickets
	for s.pq.Len() > 0 {
		t := heap.Pop(&s.pq).(*ticket)
		t.canceled = true
	}
	s.mu.Unlock()
	return nil
}
