package adaptivepool

import (
	"sync"
	"sync/atomic"
	"time"
)

// Task represents a unit of work.
type Task func()

// Config configures the adaptive worker pool.
type Config struct {
	MinWorkers  int
	MaxWorkers  int
	IdleTimeout time.Duration
	QueueSize   int
}

// DefaultConfig provides balanced pool defaults.
func DefaultConfig() Config {
	return Config{
		MinWorkers:  2,
		MaxWorkers:  50,
		IdleTimeout: 2 * time.Second,
		QueueSize:   100,
	}
}

// Pool manages dynamic worker scaling and task execution.
type Pool struct {
	cfg         Config
	tasks       chan Task
	activeCount int32
	closed      int32
	stopChan    chan struct{}
	wg          sync.WaitGroup
	mu          sync.Mutex
}

// New creates and starts an adaptive worker pool.
func New(cfg Config) *Pool {
	if cfg.MinWorkers < 1 {
		cfg.MinWorkers = 1
	}
	if cfg.MaxWorkers < cfg.MinWorkers {
		cfg.MaxWorkers = cfg.MinWorkers
	}
	if cfg.IdleTimeout <= 0 {
		cfg.IdleTimeout = 2 * time.Second
	}
	if cfg.QueueSize < 0 {
		cfg.QueueSize = 50
	}

	p := &Pool{
		cfg:      cfg,
		tasks:    make(chan Task, cfg.QueueSize),
		stopChan: make(chan struct{}),
	}

	// Spawn baseline minimum workers
	for i := 0; i < cfg.MinWorkers; i++ {
		p.spawnWorker(true)
	}

	return p
}

func (p *Pool) spawnWorker(isCore bool) {
	atomic.AddInt32(&p.activeCount, 1)
	p.wg.Add(1)

	go func() {
		defer func() {
			atomic.AddInt32(&p.activeCount, -1)
			p.wg.Done()
		}()

		for {
			if isCore {
				select {
				case <-p.stopChan:
					// Non-blocking drain
					for {
						select {
						case task := <-p.tasks:
							safeRun(task)
						default:
							return
						}
					}
				case task := <-p.tasks:
					safeRun(task)
				}
			} else {
				select {
				case <-p.stopChan:
					// Non-blocking drain
					for {
						select {
						case task := <-p.tasks:
							safeRun(task)
						default:
							return
						}
					}
				case task := <-p.tasks:
					safeRun(task)
				case <-time.After(p.cfg.IdleTimeout):
					// Idle timeout exceeded, surplus worker exits
					return
				}
			}
		}
	}()
}

func safeRun(task Task) {
	defer func() {
		_ = recover() // Catch task panic
	}()
	task()
}

func (p *Pool) maybeScale() {
	p.mu.Lock()
	defer p.mu.Unlock()

	current := int(atomic.LoadInt32(&p.activeCount))
	if len(p.tasks) > 0 && current < p.cfg.MaxWorkers {
		p.spawnWorker(false)
	}
}

// Submit queues a task for execution, blocking if the queue is full.
func (p *Pool) Submit(task Task) error {
	if task == nil {
		return ErrNilTask
	}
	if atomic.LoadInt32(&p.closed) == 1 {
		return ErrPoolClosed
	}

	select {
	case <-p.stopChan:
		return ErrPoolClosed
	case p.tasks <- task:
		p.maybeScale()
		return nil
	}
}

// TrySubmit enqueues a task without blocking. Returns true if queued, false if queue full.
func (p *Pool) TrySubmit(task Task) bool {
	if task == nil || atomic.LoadInt32(&p.closed) == 1 {
		return false
	}

	select {
	case p.tasks <- task:
		p.maybeScale()
		return true
	default:
		return false
	}
}

// SubmitWithTimeout attempts to submit a task within the given timeout duration.
func (p *Pool) SubmitWithTimeout(task Task, timeout time.Duration) error {
	if task == nil {
		return ErrNilTask
	}
	if atomic.LoadInt32(&p.closed) == 1 {
		return ErrPoolClosed
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-p.stopChan:
		return ErrPoolClosed
	case <-timer.C:
		return ErrTimeout
	case p.tasks <- task:
		p.maybeScale()
		return nil
	}
}

// ActiveWorkers returns the current number of active worker goroutines.
func (p *Pool) ActiveWorkers() int {
	return int(atomic.LoadInt32(&p.activeCount))
}

// PendingTasks returns the number of tasks waiting in the queue.
func (p *Pool) PendingTasks() int {
	return len(p.tasks)
}

// Close gracefully closes the pool, drains remaining tasks, and waits for workers to terminate.
func (p *Pool) Close() error {
	if !atomic.CompareAndSwapInt32(&p.closed, 0, 1) {
		return nil
	}

	close(p.stopChan)
	p.wg.Wait()
	return nil
}
