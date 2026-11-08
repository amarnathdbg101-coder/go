package sse

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type mockNonFlusher struct {
	http.ResponseWriter
}

func TestNewStreamer_Headers(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/events", nil)

	streamer, err := NewStreamer(rec, req)
	if err != nil {
		t.Fatalf("unexpected error creating streamer: %v", err)
	}

	h := rec.Header()
	if h.Get("Content-Type") != "text/event-stream" {
		t.Errorf("expected Content-Type text/event-stream, got %s", h.Get("Content-Type"))
	}
	if h.Get("Cache-Control") != "no-cache" {
		t.Errorf("expected Cache-Control no-cache, got %s", h.Get("Cache-Control"))
	}
	if h.Get("X-Accel-Buffering") != "no" {
		t.Errorf("expected X-Accel-Buffering no, got %s", h.Get("X-Accel-Buffering"))
	}

	if streamer == nil {
		t.Fatalf("streamer is nil")
	}
}

func TestNewStreamer_NonFlusher(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/events", nil)
	nonFlusher := &mockNonFlusher{ResponseWriter: rec}

	_, err := NewStreamer(nonFlusher, req)
	if !errors.Is(err, ErrStreamingNotSupported) {
		t.Errorf("expected ErrStreamingNotSupported, got: %v", err)
	}
}

func TestStreamer_SendEvent(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/events", nil)
	streamer, _ := NewStreamer(rec, req)

	err := streamer.Send(Event{
		ID:    "msg-1",
		Event: "order_created",
		Retry: 5 * time.Second,
		Data:  "Line 1\nLine 2",
	})
	if err != nil {
		t.Fatalf("send failed: %v", err)
	}

	body := rec.Body.String()
	expectedLines := []string{
		"id: msg-1",
		"event: order_created",
		"retry: 5000",
		"data: Line 1",
		"data: Line 2",
		"",
	}

	for _, line := range expectedLines {
		if !strings.Contains(body, line) {
			t.Errorf("expected body to contain %q, got: \n%s", line, body)
		}
	}
}

func TestStreamer_JSONData(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/events", nil)
	streamer, _ := NewStreamer(rec, req)

	payload := map[string]any{"status": "delivered", "code": 200}
	err := streamer.SendData(payload)
	if err != nil {
		t.Fatalf("sendData failed: %v", err)
	}

	body := rec.Body.String()
	if !strings.Contains(body, `data: {"code":200,"status":"delivered"}`) {
		t.Errorf("JSON data mismatch: %s", body)
	}
}

func TestStreamer_Ping(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/events", nil)
	streamer, _ := NewStreamer(rec, req)

	err := streamer.Ping()
	if err != nil {
		t.Fatalf("ping failed: %v", err)
	}

	if !strings.Contains(rec.Body.String(), ": ping\n\n") {
		t.Errorf("expected ping comment, got: %s", rec.Body.String())
	}
}

func TestStreamer_ContextCancellation(t *testing.T) {
	rec := httptest.NewRecorder()
	ctx, cancel := context.WithCancel(context.Background())
	req := httptest.NewRequest(http.MethodGet, "/events", nil).WithContext(ctx)

	streamer, _ := NewStreamer(rec, req)

	// Cancel context
	cancel()

	err := streamer.SendData("test message")
	if !errors.Is(err, ErrStreamClosed) {
		t.Errorf("expected ErrStreamClosed after context cancellation, got: %v", err)
	}
}
