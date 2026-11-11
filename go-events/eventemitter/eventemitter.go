package eventemitter

import (
	"strings"
	"sync"
	"sync/atomic"
)

// Listener is the callback invoked when a subscribed event triggers.
type Listener func(event string, payload any)

type listenerItem struct {
	id   uint64
	fn   Listener
	once bool
}

// Emitter coordinates thread-safe event pub-sub with wildcards and leak alerts.
type Emitter struct {
	mu           sync.RWMutex
	idCounter    uint64
	listeners    map[string][]*listenerItem
	maxListeners int
	onLeakWarn   func(event string, count int)
}

// New creates an initialized Emitter.
func New() *Emitter {
	return &Emitter{
		listeners:    make(map[string][]*listenerItem),
		maxListeners: 30, // Default warning threshold
	}
}

// SetMaxListeners sets the warning threshold before triggering leak alerts.
func (e *Emitter) SetMaxListeners(n int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.maxListeners = n
}

// SetLeakWarningHandler registers a callback invoked when listener count exceeds maxListeners.
func (e *Emitter) SetLeakWarningHandler(fn func(event string, count int)) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.onLeakWarn = fn
}

// On registers a permanent event listener. Returns an unsubscribe function.
func (e *Emitter) On(event string, fn Listener) (func(), error) {
	return e.addListener(event, fn, false)
}

// Once registers a one-shot event listener that automatically unsubscribes after first execution.
func (e *Emitter) Once(event string, fn Listener) (func(), error) {
	return e.addListener(event, fn, true)
}

func (e *Emitter) addListener(event string, fn Listener, once bool) (func(), error) {
	if fn == nil {
		return nil, ErrNilListener
	}
	event = strings.TrimSpace(event)
	if event == "" {
		return nil, ErrEmptyEvent
	}

	id := atomic.AddUint64(&e.idCounter, 1)
	item := &listenerItem{
		id:   id,
		fn:   fn,
		once: once,
	}

	e.mu.Lock()
	e.listeners[event] = append(e.listeners[event], item)
	count := len(e.listeners[event])
	maxL := e.maxListeners
	warnFn := e.onLeakWarn
	e.mu.Unlock()

	if maxL > 0 && count > maxL && warnFn != nil {
		warnFn(event, count)
	}

	unsub := func() {
		e.removeListener(event, id)
	}
	return unsub, nil
}

func (e *Emitter) removeListener(event string, id uint64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	items, ok := e.listeners[event]
	if !ok {
		return
	}

	for i, item := range items {
		if item.id == id {
			e.listeners[event] = append(items[:i], items[i+1:]...)
			if len(e.listeners[event]) == 0 {
				delete(e.listeners, event)
			}
			return
		}
	}
}

// Emit synchronously dispatches an event to all matching listeners (including wildcards).
func (e *Emitter) Emit(event string, payload any) {
	targets, toRemove := e.collectListeners(event)

	// Clean up one-shot listeners
	for _, tr := range toRemove {
		e.removeListener(tr.event, tr.id)
	}

	for _, item := range targets {
		safeInvoke(item.fn, event, payload)
	}
}

// EmitAsync asynchronously dispatches an event, executing each listener in its own goroutine.
func (e *Emitter) EmitAsync(event string, payload any) {
	targets, toRemove := e.collectListeners(event)

	for _, tr := range toRemove {
		e.removeListener(tr.event, tr.id)
	}

	for _, item := range targets {
		go safeInvoke(item.fn, event, payload)
	}
}

type removeTask struct {
	event string
	id    uint64
}

func (e *Emitter) collectListeners(event string) ([]*listenerItem, []removeTask) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var targets []*listenerItem
	var toRemove []removeTask

	for pat, items := range e.listeners {
		if matchEvent(pat, event) {
			for _, item := range items {
				targets = append(targets, item)
				if item.once {
					toRemove = append(toRemove, removeTask{event: pat, id: item.id})
				}
			}
		}
	}

	return targets, toRemove
}

func matchEvent(pattern, event string) bool {
	if pattern == event || pattern == "*" {
		return true
	}
	if strings.HasSuffix(pattern, ".*") {
		prefix := strings.TrimSuffix(pattern, ".*")
		return strings.HasPrefix(event, prefix+".")
	}
	return false
}

func safeInvoke(fn Listener, event string, payload any) {
	defer func() {
		_ = recover() // Catch and isolate listener panic
	}()
	fn(event, payload)
}

// ListenerCount returns the number of listeners currently registered for event.
func (e *Emitter) ListenerCount(event string) int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.listeners[event])
}

// EventNames returns a slice of all unique event names currently subscribed.
func (e *Emitter) EventNames() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	names := make([]string, 0, len(e.listeners))
	for name := range e.listeners {
		names = append(names, name)
	}
	return names
}

// Clear removes all listeners for specified events, or all events if none specified.
func (e *Emitter) Clear(events ...string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if len(events) == 0 {
		e.listeners = make(map[string][]*listenerItem)
		return
	}

	for _, evt := range events {
		delete(e.listeners, evt)
	}
}
