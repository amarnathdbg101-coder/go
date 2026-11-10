package window

import "time"

// WindowType identifies the mathematical windowing strategy.
type WindowType string

const (
	// WindowTumbling specifies fixed-size, non-overlapping time intervals.
	WindowTumbling WindowType = "TUMBLING"

	// WindowSliding specifies overlapping intervals that advance at regular slide intervals.
	WindowSliding WindowType = "SLIDING"

	// WindowSession specifies dynamic windows demarcated by periods of inactivity gap.
	WindowSession WindowType = "SESSION"
)

// Event encapsulates a timestamped payload associated with an optional grouping key.
type Event[T any] struct {
	Key       string    `json:"key"`
	Timestamp time.Time `json:"timestamp"`
	Value     T         `json:"value"`
}

// WindowResult contains the aggregated output of an expired/triggered window.
type WindowResult[T any] struct {
	Key       string    `json:"key"`
	Start     time.Time `json:"start"`
	End       time.Time `json:"end"`
	Count     int       `json:"count"`
	Result    T         `json:"result"`
	Values    []T       `json:"values,omitempty"`
}

// ReducerFunc aggregates an incoming value into an accumulated state.
type ReducerFunc[T any] func(acc, val T) T

// TriggerFunc is invoked whenever a window reaches its boundary.
type TriggerFunc[T any] func(res WindowResult[T])
