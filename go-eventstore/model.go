package eventstore

import "time"

// ExpectedVersion constants for optimistic concurrency control.
const (
	// AnyVersion disables optimistic concurrency revision checks.
	AnyVersion uint64 = ^uint64(0)
	// NoStream requires that the stream does not exist yet (must be revision 0).
	NoStream uint64 = 0
)

// EventData represents an incoming domain event to be appended.
type EventData struct {
	EventType string
	Payload   []byte
	Metadata  []byte
}

// RecordedEvent represents an immutable persisted event with stream and global sequence metadata.
type RecordedEvent struct {
	StreamID       string
	Revision       uint64 // Stream-specific monotonic revision (1, 2, 3...)
	GlobalPosition uint64 // Global monotonic log position across all streams
	EventType      string
	Payload        []byte
	Metadata       []byte
	Timestamp      time.Time
}

// Snapshot represents a saved point-in-time aggregate state.
type Snapshot struct {
	StreamID string
	Revision uint64
	State    []byte
	SavedAt  time.Time
}
