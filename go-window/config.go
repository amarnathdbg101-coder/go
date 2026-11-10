package window

import "time"

// Config specifies the windowing parameters and callbacks.
type Config[T any] struct {
	Type            WindowType
	Size            time.Duration
	Slide           time.Duration
	InactivityGap   time.Duration
	AllowedLateness time.Duration
	OnTrigger       TriggerFunc[T]
	Reducer         ReducerFunc[T]
}

// Option is a functional option for configuring the window engine.
type Option[T any] func(*Config[T])

// TumblingConfig creates a configuration for fixed tumbling windows.
func TumblingConfig[T any](size time.Duration, trigger TriggerFunc[T]) Config[T] {
	return Config[T]{
		Type:      WindowTumbling,
		Size:      size,
		OnTrigger: trigger,
	}
}

// SlidingConfig creates a configuration for sliding windows.
func SlidingConfig[T any](size, slide time.Duration, trigger TriggerFunc[T]) Config[T] {
	return Config[T]{
		Type:      WindowSliding,
		Size:      size,
		Slide:     slide,
		OnTrigger: trigger,
	}
}

// SessionConfig creates a configuration for inactivity-based session windows.
func SessionConfig[T any](gap time.Duration, trigger TriggerFunc[T]) Config[T] {
	return Config[T]{
		Type:          WindowSession,
		InactivityGap: gap,
		OnTrigger:     trigger,
	}
}

// WithAllowedLateness sets the watermark tolerance for out-of-order events.
func WithAllowedLateness[T any](d time.Duration) Option[T] {
	return func(c *Config[T]) {
		c.AllowedLateness = d
	}
}

// WithReducer sets an incremental aggregation function.
func WithReducer[T any](fn ReducerFunc[T]) Option[T] {
	return func(c *Config[T]) {
		c.Reducer = fn
	}
}
