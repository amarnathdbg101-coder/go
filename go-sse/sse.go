package sse

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Event represents a WHATWG Server-Sent Event message.
type Event struct {
	ID      string
	Event   string
	Data    any
	Retry   time.Duration
	Comment string
}

// Streamer handles writing events to an active HTTP response stream.
type Streamer struct {
	mu      sync.Mutex
	w       http.ResponseWriter
	flusher http.Flusher
	ctx     context.Context
	closed  bool
}

// NewStreamer prepares an HTTP connection for Server-Sent Events by setting headers and flushing.
func NewStreamer(w http.ResponseWriter, r *http.Request) (*Streamer, error) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, ErrStreamingNotSupported
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // Prevent Nginx buffering
	flusher.Flush()

	return &Streamer{
		w:       w,
		flusher: flusher,
		ctx:     r.Context(),
	}, nil
}

// Send formats and writes an Event to the stream and flushes immediately.
func (s *Streamer) Send(e Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrStreamClosed
	}

	select {
	case <-s.ctx.Done():
		s.closed = true
		return ErrStreamClosed
	default:
	}

	var buf bytes.Buffer

	// Keep-alive comment
	if e.Comment != "" {
		buf.WriteString(": ")
		buf.WriteString(e.Comment)
		buf.WriteByte('\n')
	}

	// Event ID
	if e.ID != "" {
		buf.WriteString("id: ")
		buf.WriteString(e.ID)
		buf.WriteByte('\n')
	}

	// Custom Event type
	if e.Event != "" {
		buf.WriteString("event: ")
		buf.WriteString(e.Event)
		buf.WriteByte('\n')
	}

	// Client reconnection retry duration
	if e.Retry > 0 {
		buf.WriteString(fmt.Sprintf("retry: %d\n", e.Retry.Milliseconds()))
	}

	// Data payload
	if e.Data != nil {
		switch v := e.Data.(type) {
		case string:
			lines := strings.Split(v, "\n")
			for _, line := range lines {
				buf.WriteString("data: ")
				buf.WriteString(line)
				buf.WriteByte('\n')
			}
		case []byte:
			lines := bytes.Split(v, []byte("\n"))
			for _, line := range lines {
				buf.WriteString("data: ")
				buf.Write(line)
				buf.WriteByte('\n')
			}
		default:
			jsonData, err := json.Marshal(v)
			if err != nil {
				return err
			}
			buf.WriteString("data: ")
			buf.Write(jsonData)
			buf.WriteByte('\n')
		}
	}

	// SSE messages are delimited by an extra newline
	buf.WriteByte('\n')

	if _, err := s.w.Write(buf.Bytes()); err != nil {
		s.closed = true
		return err
	}

	s.flusher.Flush()
	return nil
}

// SendData formats raw data as a standard unnamed message.
func (s *Streamer) SendData(data any) error {
	return s.Send(Event{Data: data})
}

// Ping sends a comment line to prevent intermediary reverse-proxies from timing out.
func (s *Streamer) Ping() error {
	return s.Send(Event{Comment: "ping"})
}

// Close marks the streamer as closed.
func (s *Streamer) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
}

// Done returns a channel that is closed when the client disconnects.
func (s *Streamer) Done() <-chan struct{} {
	return s.ctx.Done()
}
