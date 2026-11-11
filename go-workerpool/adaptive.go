package goworkerpool

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// AdaptivePool coordinates an elastic set of worker goroutines that automatically scale
// up during traffic bursts and contract during idle periods to conserve OS threads and memory.
type AdaptivePool struct {
	mu          sync.Mutex
	minWorkers  int
	maxWorkers  int
	idleTimeout time.Duration
	tasks       chan Job
	activeCount int32
	closed      int32
	stopChan    chan struct{}
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
	onPanic     func(recovered any)
}

// NewAdaptive creates and starts an elastic worker pool.
func NewAdaptive(minWorkers, maxWorkers, queueSize int, idleTimeout time.Duration, opts ...Option) *AdaptivePool {
	if minWorkers < 1 {
		minWorkers = 2
	}
	if maxWorkers < minWorkers {
		maxWorkers = minWorkers
	}
	if idleTimeout <= 0 {
		idleTimeout = 2 * time.Second
	}
	if queueSize < 0 {
		queueSize = 100
	}

	ctx, cancel := context.WithCancel(context.Background())

	p := &AdaptivePool{
		minWorkers:  minWorkers,
		maxWorkers:  maxWorkers,
		idleTimeout: idleTimeout,
		tasks:       make(chan Job, queueSize),
		stopChan:    make(chan struct{}),
		ctx:         ctx,
		cancel:      cancel,
	}

	// Apply options
	dummyPool := &Pool{}
	for _, opt := range opts {
		opt(dummyPool)
	}
	p.onPanic = dummyPool.onPanic

	// Start baseline minimum workers
	for i := 0; i < minWorkers; i++ {
		p.spawnWorker(false)
	}

	return p
}

func (p *AdaptivePool) spawnWorker(ephemeral bool) {
	atomic.AddInt32(&p.activeCount, 1)
	p.wg.Add(1)

	go func() {
		defer func() {
			atomic.AddInt32(&p.activeCount, -1)
			p.wg.Done()
		}()

		var idleTimer *time.Timer
		if ephemeral {
			idleTimer = time.NewTimer(p.idleTimeout)
			defer idleTimer.Stop()
		}

		for {
			if !ephemeral {
				select {
				case <-p.stopChan:
					return
				case job, ok := <-p.tasks:
					if !ok {
						return
					}
					p.executeJob(job)
				}
			} else {
				if !idleTimer.Stop() {
					select {
					case <-idleTimer.C:
					default:
					}
				}
				idleTimer.Reset(p.idleTimeout)

				select {
				case <-p.stopChan:
					return
				case <-idleTimer.C:
					// Evict idle worker if current workers > minWorkers
					if int(atomic.LoadInt32(&p.activeCount)) > p.minWorkers {
						return
					}
				case job, ok := <-p.tasks:
					if !ok {
						return
					}
					p.executeJob(job)
				}
			}
		}
	}()
}

func (p *AdaptivePool) executeJob(job Job) {
	defer func() {
		if r := recover(); r != nil && p.onPanic != nil {
			p.onPanic(r)
		}
	}()
	job(p.ctx)
}

// Submit enqueues a job for execution. If queue is full and worker count < maxWorkers, spawns an ephemeral worker.
func (p *AdaptivePool) Submit(job Job) error {
	if atomic.LoadInt32(&p.closed) == 1 {
		return ErrPoolClosed
	}

	select {
	case p.tasks <- job:
		return nil
	default:
		// Queue saturated: check if we can expand worker pool
		p.mu.Lock()
		if int(atomic.LoadInt32(&p.activeCount)) < p.maxWorkers {
			p.spawnWorker(true)
		}
		p.mu.Unlock()

		// Try submitting again or block
		select {
		case p.tasks <- job:
			return nil
		case <-p.stopChan:
			return ErrPoolClosed
		}
	}
}

// ActiveWorkers returns the count of currently running worker goroutines.
func (p *AdaptivePool) ActiveWorkers() int {
	return int(atomic.LoadInt32(&p.activeCount))
}

// Stop gracefully closes the pool, drains queued tasks, and waits for workers to finish.
func (p *AdaptivePool) Stop() {
	if atomic.CompareAndSwapInt32(&p.closed, 0, 1) {
		close(p.stopChan)
		close(p.tasks)
		p.cancel()
		p.wg.Wait()
	}
}
