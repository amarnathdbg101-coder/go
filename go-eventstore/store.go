package eventstore

import (
	"sync"
	"time"
)

// Version defines the release version of go-eventstore.
const Version = "1.0.0"

// Store provides an in-memory event sourcing engine with stream isolation and global event ordering.
type Store struct {
	mu        sync.RWMutex
	streams   map[string][]RecordedEvent
	globalLog []RecordedEvent
	snapshots map[string]Snapshot
}

// New creates an initialized Event Store instance.
func New() *Store {
	return &Store{
		streams:   make(map[string][]RecordedEvent),
		globalLog: make([]RecordedEvent, 0, 128),
		snapshots: make(map[string]Snapshot),
	}
}

// AppendToStream writes one or more domain events to a stream with optimistic concurrency checks.
func (s *Store) AppendToStream(streamID string, expectedRevision uint64, events []EventData) ([]RecordedEvent, error) {
	if len(events) == 0 {
		return nil, ErrEmptyEvents
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	currentEvents := s.streams[streamID]
	currentRevision := uint64(len(currentEvents))

	// Optimistic concurrency verification
	if expectedRevision != AnyVersion {
		if expectedRevision == NoStream {
			if currentRevision != 0 {
				return nil, ErrWrongExpectedVersion
			}
		} else if currentRevision != expectedRevision {
			return nil, ErrWrongExpectedVersion
		}
	}

	now := time.Now()
	recorded := make([]RecordedEvent, len(events))

	for i, e := range events {
		rev := currentRevision + uint64(i) + 1
		gPos := uint64(len(s.globalLog)) + 1

		pCopy := make([]byte, len(e.Payload))
		copy(pCopy, e.Payload)
		mCopy := make([]byte, len(e.Metadata))
		copy(mCopy, e.Metadata)

		rec := RecordedEvent{
			StreamID:       streamID,
			Revision:       rev,
			GlobalPosition: gPos,
			EventType:      e.EventType,
			Payload:        pCopy,
			Metadata:       mCopy,
			Timestamp:      now,
		}

		recorded[i] = rec
		s.globalLog = append(s.globalLog, rec)
	}

	s.streams[streamID] = append(s.streams[streamID], recorded...)
	return recorded, nil
}

// ReadStream reads up to maxCount events from a stream starting at fromRevision (1-based).
func (s *Store) ReadStream(streamID string, fromRevision uint64, maxCount int) ([]RecordedEvent, error) {
	if fromRevision == 0 {
		fromRevision = 1
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	stream, exists := s.streams[streamID]
	if !exists {
		return nil, ErrStreamNotFound
	}

	startIdx := int(fromRevision) - 1
	if startIdx >= len(stream) {
		return nil, nil // Past end of stream
	}

	endIdx := len(stream)
	if maxCount > 0 && startIdx+maxCount < endIdx {
		endIdx = startIdx + maxCount
	}

	res := make([]RecordedEvent, endIdx-startIdx)
	copy(res, stream[startIdx:endIdx])
	return res, nil
}

// ReadAll reads events across all streams starting from a global log position (1-based).
func (s *Store) ReadAll(fromGlobalPosition uint64, maxCount int) []RecordedEvent {
	if fromGlobalPosition == 0 {
		fromGlobalPosition = 1
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	startIdx := int(fromGlobalPosition) - 1
	if startIdx >= len(s.globalLog) {
		return nil
	}

	endIdx := len(s.globalLog)
	if maxCount > 0 && startIdx+maxCount < endIdx {
		endIdx = startIdx + maxCount
	}

	res := make([]RecordedEvent, endIdx-startIdx)
	copy(res, s.globalLog[startIdx:endIdx])
	return res
}

// SaveSnapshot writes a point-in-time snapshot of the aggregate state.
func (s *Store) SaveSnapshot(streamID string, revision uint64, state []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	stCopy := make([]byte, len(state))
	copy(stCopy, state)

	s.snapshots[streamID] = Snapshot{
		StreamID: streamID,
		Revision: revision,
		State:    stCopy,
		SavedAt:  time.Now(),
	}
	return nil
}

// LoadSnapshot retrieves the latest snapshot for a stream.
func (s *Store) LoadSnapshot(streamID string) (Snapshot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	snap, exists := s.snapshots[streamID]
	if !exists {
		return Snapshot{}, ErrSnapshotNotFound
	}
	return snap, nil
}

// StreamRevision returns the current version number of the stream (0 if non-existent).
func (s *Store) StreamRevision(streamID string) uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return uint64(len(s.streams[streamID]))
}
