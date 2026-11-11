package hotconfig

import (
	"crypto/sha256"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// HotConfig provides lock-free, zero-downtime hot reloading for application configurations.
type HotConfig[T any] struct {
	path         string
	opts         Options[T]
	current      atomic.Pointer[T]
	lastChecksum [32]byte
	lastModTime  time.Time

	mu       sync.Mutex
	stopChan chan struct{}
	wg       sync.WaitGroup
	closed   bool
}

// Watch loads the configuration from path, verifies its checksum, and starts background file watching.
func Watch[T any](path string, options ...Option[T]) (*HotConfig[T], error) {
	opts := DefaultOptions[T]()
	for _, opt := range options {
		opt(&opts)
	}

	h := &HotConfig[T]{
		path:     path,
		opts:     opts,
		stopChan: make(chan struct{}),
	}

	// Initial load must succeed
	if err := h.Reload(); err != nil {
		return nil, err
	}

	// Start background watcher loop
	h.wg.Add(1)
	go h.watchLoop()

	return h, nil
}

// Get returns a pointer to the active configuration.
// It executes in sub-nanosecond time via a lock-free atomic pointer read without mutex contention.
func (h *HotConfig[T]) Get() *T {
	return h.current.Load()
}

// Reload forces an immediate read, SHA-256 validation, and atomic swap.
func (h *HotConfig[T]) Reload() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.closed {
		return ErrWatcherClosed
	}

	data, err := os.ReadFile(h.path)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrConfigNotFound, err)
	}

	// Checksum verification
	sum := sha256.Sum256(data)
	if sum == h.lastChecksum && h.current.Load() != nil {
		// Content unchanged, skip re-parsing
		return nil
	}

	// Unmarshal
	newCfg, err := h.opts.Unmarshaler(data)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrCorruptConfig, err)
	}

	// Validation schema check
	if h.opts.Validator != nil {
		if err := h.opts.Validator(newCfg); err != nil {
			return fmt.Errorf("%w: %v", ErrValidationFailed, err)
		}
	}

	// Atomic Pointer Swap (Zero downtime!)
	oldCfg := h.current.Swap(newCfg)
	h.lastChecksum = sum

	if stat, err := os.Stat(h.path); err == nil {
		h.lastModTime = stat.ModTime()
	}

	if oldCfg != nil && h.opts.OnChange != nil {
		// Invoke callback outside lock in separate goroutine
		go h.opts.OnChange(oldCfg, newCfg)
	}

	return nil
}

func (h *HotConfig[T]) watchLoop() {
	defer h.wg.Done()
	ticker := time.NewTicker(h.opts.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-h.stopChan:
			return
		case <-ticker.C:
			stat, err := os.Stat(h.path)
			if err != nil {
				continue
			}

			// Check if modified
			if stat.ModTime().After(h.lastModTime) {
				_ = h.Reload()
			}
		}
	}
}

// Close gracefully stops the background file watcher.
func (h *HotConfig[T]) Close() error {
	h.mu.Lock()
	if h.closed {
		h.mu.Unlock()
		return nil
	}
	h.closed = true
	close(h.stopChan)
	h.mu.Unlock()

	h.wg.Wait()
	return nil
}
