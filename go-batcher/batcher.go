package gobatcher

import (
	"fmt"
	"sync"
	"time"
)

// FlushFunc defines the handler invoked when a batch is ready to be flushed.
type FlushFunc[T any] func(batch []T) error

// Config controls batching thresholds and operational handlers.
type Config[T any] struct {
	MaxBatchSize  int
	FlushInterval time.Duration
	QueueCapacity int
	FlushFn       FlushFunc[T]
	OnError       func(batch []T, err error)
	OnPanic       func(r any)
}

// Batcher provides high-throughput concurrent micro-batching for arbitrary types T.
type Batcher[T any] struct {
	cfg       Config[T]
	itemChan  chan T
	flushChan chan chan struct{}
	stopChan  chan struct{}
	stopped   bool
	mu        sync.RWMutex
	wg        sync.WaitGroup
}

// New initializes and launches background worker for Batcher.
func New[T any](cfg Config[T]) (*Batcher[T], error) {
	if cfg.MaxBatchSize <= 0 {
		cfg.MaxBatchSize = 100
	}
	if cfg.FlushInterval <= 0 {
		cfg.FlushInterval = 50 * time.Millisecond
	}
	if cfg.QueueCapacity <= 0 {
		cfg.QueueCapacity = cfg.MaxBatchSize * 10
	}
	if cfg.FlushFn == nil {
		return nil, fmt.Errorf("%w: FlushFn must not be nil", ErrInvalidConfig)
	}

	b := &Batcher[T]{
		cfg:       cfg,
		itemChan:  make(chan T, cfg.QueueCapacity),
		flushChan: make(chan chan struct{}, 10),
		stopChan:  make(chan struct{}),
	}

	b.wg.Add(1)
	go b.run()

	return b, nil
}

func (b *Batcher[T]) run() {
	defer b.wg.Done()

	ticker := time.NewTicker(b.cfg.FlushInterval)
	defer ticker.Stop()

	batch := make([]T, 0, b.cfg.MaxBatchSize)

	flushCurrent := func() {
		if len(batch) == 0 {
			return
		}
		itemsToFlush := batch
		batch = make([]T, 0, b.cfg.MaxBatchSize)

		b.safeFlush(itemsToFlush)
	}

	for {
		select {
		case item, ok := <-b.itemChan:
			if !ok {
				// Channel closed, flush remaining and exit
				flushCurrent()
				return
			}
			batch = append(batch, item)
			if len(batch) >= b.cfg.MaxBatchSize {
				flushCurrent()
			}

		case <-ticker.C:
			flushCurrent()

		case doneChan := <-b.flushChan:
			// Drain any pending items sitting in itemChan first before flushing
			for {
				select {
				case item, ok := <-b.itemChan:
					if ok {
						batch = append(batch, item)
						if len(batch) >= b.cfg.MaxBatchSize {
							flushCurrent()
						}
					}
				default:
					goto drainDone
				}
			}
		drainDone:
			flushCurrent()
			close(doneChan)

		case <-b.stopChan:
			// Drain remaining in itemChan
			for {
				select {
				case item := <-b.itemChan:
					batch = append(batch, item)
					if len(batch) >= b.cfg.MaxBatchSize {
						flushCurrent()
					}
				default:
					flushCurrent()
					return
				}
			}
		}
	}
}

func (b *Batcher[T]) safeFlush(items []T) {
	defer func() {
		if r := recover(); r != nil {
			if b.cfg.OnPanic != nil {
				b.cfg.OnPanic(r)
			}
		}
	}()

	if err := b.cfg.FlushFn(items); err != nil {
		if b.cfg.OnError != nil {
			b.cfg.OnError(items, err)
		}
	}
}

// Push appends an item to the queue, blocking if queue capacity is saturated.
func (b *Batcher[T]) Push(item T) error {
	b.mu.RLock()
	if b.stopped {
		b.mu.RUnlock()
		return ErrBatcherStopped
	}
	b.mu.RUnlock()

	b.itemChan <- item
	return nil
}

// TryPush attempts to add an item to queue, returning ErrQueueFull immediately if saturated.
func (b *Batcher[T]) TryPush(item T) error {
	b.mu.RLock()
	if b.stopped {
		b.mu.RUnlock()
		return ErrBatcherStopped
	}
	b.mu.RUnlock()

	select {
	case b.itemChan <- item:
		return nil
	default:
		return ErrQueueFull
	}
}

// Flush triggers an immediate synchronous flush of all accumulated items.
func (b *Batcher[T]) Flush() {
	b.mu.RLock()
	if b.stopped {
		b.mu.RUnlock()
		return
	}
	b.mu.RUnlock()

	done := make(chan struct{})
	b.flushChan <- done
	<-done
}

// Stop initiates graceful shutdown, flushing all buffered items before returning.
func (b *Batcher[T]) Stop() {
	b.mu.Lock()
	if b.stopped {
		b.mu.Unlock()
		return
	}
	b.stopped = true
	b.mu.Unlock()

	close(b.stopChan)
	b.wg.Wait()
}
