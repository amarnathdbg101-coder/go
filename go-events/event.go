package goevents

import (
	"context"
	"time"
)

// Event represents an emitted message carrying structured payload and metadata.
type Event struct {
	Topic     string            `json:"topic"`
	Payload   any               `json:"payload"`
	Timestamp time.Time         `json:"timestamp"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// NewEvent constructs a new event stamped with the current UTC timestamp.
func NewEvent(topic string, payload any) Event {
	return Event{
		Topic:     topic,
		Payload:   payload,
		Timestamp: time.Now().UTC(),
		Metadata:  make(map[string]string),
	}
}

// WithMetadata adds a key-value pair to event metadata (e.g. correlation_id).
func (e Event) WithMetadata(key, value string) Event {
	if e.Metadata == nil {
		e.Metadata = make(map[string]string)
	}
	e.Metadata[key] = value
	return e
}

// HandlerFunc defines the callback signature for event listeners.
type HandlerFunc func(ctx context.Context, e Event) error

// MiddlewareFunc intercepts event dispatching for logging, tracing, or metrics.
type MiddlewareFunc func(next HandlerFunc) HandlerFunc
