package window

import (
	"sync"
	"time"
)

type activeWindow[T any] struct {
	key       string
	start     time.Time
	end       time.Time
	values    []T
	accum     T
	hasAccum  bool
	count     int
	lastEvent time.Time
}

// Engine processes live event streams into bounded time windows.
type Engine[T any] struct {
	cfg       Config[T]
	mu        sync.Mutex
	windows   map[string][]*activeWindow[T]
	watermark time.Time
	closed    bool
}

// New creates a new stream windowing engine.
func New[T any](cfg Config[T], opts ...Option[T]) (*Engine[T], error) {
	if cfg.Type == WindowTumbling && cfg.Size <= 0 {
		return nil, ErrInvalidDuration
	}
	if cfg.Type == WindowSliding && (cfg.Size <= 0 || cfg.Slide <= 0) {
		return nil, ErrInvalidDuration
	}
	if cfg.Type == WindowSession && cfg.InactivityGap <= 0 {
		return nil, ErrInvalidDuration
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	return &Engine[T]{
		cfg:     cfg,
		windows: make(map[string][]*activeWindow[T]),
	}, nil
}

// Add routes an incoming event into its corresponding time windows.
func (e *Engine[T]) Add(event Event[T]) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.closed {
		return ErrWindowClosed
	}

	// Watermark checking
	if !e.watermark.IsZero() && e.cfg.AllowedLateness > 0 {
		if event.Timestamp.Before(e.watermark.Add(-e.cfg.AllowedLateness)) {
			return ErrLateEventDropped
		}
	}
	if event.Timestamp.After(e.watermark) {
		e.watermark = event.Timestamp
	}

	switch e.cfg.Type {
	case WindowTumbling:
		e.addTumbling(event)
	case WindowSliding:
		e.addSliding(event)
	case WindowSession:
		e.addSession(event)
	}

	// Trigger expired windows
	e.triggerExpiredWindows(event.Timestamp)

	return nil
}

func (e *Engine[T]) addTumbling(event Event[T]) {
	size := e.cfg.Size
	// Calculate window boundaries aligned to Unix epoch
	windowStartUnix := event.Timestamp.UnixNano() - (event.Timestamp.UnixNano() % size.Nanoseconds())
	start := time.Unix(0, windowStartUnix)
	end := start.Add(size)

	wList := e.windows[event.Key]
	var target *activeWindow[T]
	for _, w := range wList {
		if w.start.Equal(start) {
			target = w
			break
		}
	}

	if target == nil {
		target = &activeWindow[T]{
			key:   event.Key,
			start: start,
			end:   end,
		}
		e.windows[event.Key] = append(e.windows[event.Key], target)
	}

	e.accumulate(target, event.Value)
	target.lastEvent = event.Timestamp
}

func (e *Engine[T]) addSliding(event Event[T]) {
	size := e.cfg.Size
	slide := e.cfg.Slide

	// Find the earliest window start that still covers this event
	lastStartUnix := event.Timestamp.UnixNano() - (event.Timestamp.UnixNano() % slide.Nanoseconds())

	for startUnix := lastStartUnix; startUnix+size.Nanoseconds() > event.Timestamp.UnixNano(); startUnix -= slide.Nanoseconds() {
		start := time.Unix(0, startUnix)
		end := start.Add(size)

		if event.Timestamp.Before(start) || !event.Timestamp.Before(end) {
			continue
		}

		wList := e.windows[event.Key]
		var target *activeWindow[T]
		for _, w := range wList {
			if w.start.Equal(start) {
				target = w
				break
			}
		}

		if target == nil {
			target = &activeWindow[T]{
				key:   event.Key,
				start: start,
				end:   end,
			}
			e.windows[event.Key] = append(e.windows[event.Key], target)
		}

		e.accumulate(target, event.Value)
		target.lastEvent = event.Timestamp
	}
}

func (e *Engine[T]) addSession(event Event[T]) {
	wList := e.windows[event.Key]
	gap := e.cfg.InactivityGap

	if len(wList) == 0 {
		target := &activeWindow[T]{
			key:       event.Key,
			start:     event.Timestamp,
			end:       event.Timestamp.Add(gap),
			lastEvent: event.Timestamp,
		}
		e.accumulate(target, event.Value)
		e.windows[event.Key] = append(e.windows[event.Key], target)
		return
	}

	current := wList[0]
	if event.Timestamp.Sub(current.lastEvent) <= gap {
		// Extend session window
		current.lastEvent = event.Timestamp
		current.end = event.Timestamp.Add(gap)
		e.accumulate(current, event.Value)
	} else {
		// Inactivity gap exceeded: close old session and trigger!
		e.triggerWindow(current)
		target := &activeWindow[T]{
			key:       event.Key,
			start:     event.Timestamp,
			end:       event.Timestamp.Add(gap),
			lastEvent: event.Timestamp,
		}
		e.accumulate(target, event.Value)
		e.windows[event.Key] = []*activeWindow[T]{target}
	}
}

func (e *Engine[T]) accumulate(w *activeWindow[T], val T) {
	w.count++
	w.values = append(w.values, val)
	if e.cfg.Reducer != nil {
		if !w.hasAccum {
			w.accum = val
			w.hasAccum = true
		} else {
			w.accum = e.cfg.Reducer(w.accum, val)
		}
	}
}

func (e *Engine[T]) triggerExpiredWindows(now time.Time) {
	for key, wList := range e.windows {
		var remaining []*activeWindow[T]
		for _, w := range wList {
			if now.After(w.end) || now.Equal(w.end) {
				e.triggerWindow(w)
			} else {
				remaining = append(remaining, w)
			}
		}
		if len(remaining) == 0 {
			delete(e.windows, key)
		} else {
			e.windows[key] = remaining
		}
	}
}

func (e *Engine[T]) triggerWindow(w *activeWindow[T]) {
	if e.cfg.OnTrigger == nil {
		return
	}

	res := WindowResult[T]{
		Key:    w.key,
		Start:  w.start,
		End:    w.end,
		Count:  w.count,
		Result: w.accum,
		Values: w.values,
	}
	e.cfg.OnTrigger(res)
}

// Flush closes and triggers all currently pending windows immediately.
func (e *Engine[T]) Flush() {
	e.mu.Lock()
	defer e.mu.Unlock()

	for key, wList := range e.windows {
		for _, w := range wList {
			e.triggerWindow(w)
		}
		delete(e.windows, key)
	}
}

// Close flushes and shuts down the engine.
func (e *Engine[T]) Close() error {
	e.Flush()
	e.mu.Lock()
	defer e.mu.Unlock()
	e.closed = true
	return nil
}
