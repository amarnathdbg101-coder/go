package godebounce

import (
	"sync"
	"time"
)

// Debouncer delays invocation of an action until a specified period of silence has elapsed.
type Debouncer struct {
	wait  time.Duration
	timer *time.Timer
	mu    sync.Mutex
}

// New creates a new Debouncer with the given wait duration.
func New(wait time.Duration) *Debouncer {
	return &Debouncer{
		wait: wait,
	}
}

// Trigger registers an action to be executed after wait duration of inactivity.
// If Trigger is called again before the duration expires, the timer resets.
func (d *Debouncer) Trigger(action func()) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.timer != nil {
		d.timer.Stop()
	}

	d.timer = time.AfterFunc(d.wait, func() {
		action()
	})
}

// Cancel stops any pending scheduled execution.
func (d *Debouncer) Cancel() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.timer != nil {
		d.timer.Stop()
		d.timer = nil
	}
}

// Throttler ensures an action is executed at most once per wait window.
type Throttler struct {
	wait     time.Duration
	lastTime time.Time
	mu       sync.Mutex
}

// NewThrottler creates a new Throttler with the given wait duration.
func NewThrottler(wait time.Duration) *Throttler {
	return &Throttler{
		wait: wait,
	}
}

// Trigger executes action immediately if wait duration has passed since last execution.
// Returns true if executed, false if throttled.
func (t *Throttler) Trigger(action func()) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	if now.Sub(t.lastTime) >= t.wait {
		t.lastTime = now
		action()
		return true
	}

	return false
}

// KeyedDebouncer manages independent debouncers per key (e.g. per user ID or document ID).
type KeyedDebouncer struct {
	wait       time.Duration
	debouncers map[string]*Debouncer
	mu         sync.Mutex
}

// NewKeyed initializes a KeyedDebouncer.
func NewKeyed(wait time.Duration) *KeyedDebouncer {
	return &KeyedDebouncer{
		wait:       wait,
		debouncers: make(map[string]*Debouncer),
	}
}

// Trigger applies debounce logic specifically to the given key.
func (kd *KeyedDebouncer) Trigger(key string, action func()) {
	kd.mu.Lock()
	d, exists := kd.debouncers[key]
	if !exists {
		d = New(kd.wait)
		kd.debouncers[key] = d
	}
	kd.mu.Unlock()

	d.Trigger(func() {
		action()
		kd.mu.Lock()
		delete(kd.debouncers, key)
		kd.mu.Unlock()
	})
}

// Cancel cancels the pending action for a specific key.
func (kd *KeyedDebouncer) Cancel(key string) {
	kd.mu.Lock()
	defer kd.mu.Unlock()

	if d, exists := kd.debouncers[key]; exists {
		d.Cancel()
		delete(kd.debouncers, key)
	}
}

// CancelAll cancels all active pending keyed timers.
func (kd *KeyedDebouncer) CancelAll() {
	kd.mu.Lock()
	defer kd.mu.Unlock()

	for _, d := range kd.debouncers {
		d.Cancel()
	}
	kd.debouncers = make(map[string]*Debouncer)
}
