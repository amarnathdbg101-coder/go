package window_test

import (
	"errors"
	"testing"
	"time"

	"github.com/amarnathdbg101-coder/go/go-window"
)

func TestTumblingWindowSum(t *testing.T) {
	var triggered []window.WindowResult[int]

	cfg := window.TumblingConfig[int](100*time.Millisecond, func(res window.WindowResult[int]) {
		triggered = append(triggered, res)
	})

	engine, err := window.New(cfg, window.WithReducer[int](func(a, b int) int {
		return a + b
	}))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}
	defer engine.Close()

	baseTime := time.Date(2026, 11, 10, 10, 0, 0, 0, time.UTC)

	// Add 3 events in same tumbling window
	_ = engine.Add(window.Event[int]{Key: "user_1", Timestamp: baseTime.Add(10 * time.Millisecond), Value: 10})
	_ = engine.Add(window.Event[int]{Key: "user_1", Timestamp: baseTime.Add(30 * time.Millisecond), Value: 25})
	_ = engine.Add(window.Event[int]{Key: "user_1", Timestamp: baseTime.Add(80 * time.Millisecond), Value: 15})

	// Add event in next window to advance time past 100ms
	_ = engine.Add(window.Event[int]{Key: "user_1", Timestamp: baseTime.Add(150 * time.Millisecond), Value: 5})

	if len(triggered) != 1 {
		t.Fatalf("expected 1 completed tumbling window, got %d", len(triggered))
	}

	res := triggered[0]
	if res.Count != 3 {
		t.Errorf("expected count 3, got %d", res.Count)
	}
	if res.Result != 50 { // 10 + 25 + 15 = 50
		t.Errorf("expected sum 50, got %d", res.Result)
	}
}

func TestSessionWindow(t *testing.T) {
	var sessions []window.WindowResult[string]

	cfg := window.SessionConfig[string](50*time.Millisecond, func(res window.WindowResult[string]) {
		sessions = append(sessions, res)
	})

	engine, err := window.New(cfg)
	if err != nil {
		t.Fatalf("failed to create session engine: %v", err)
	}
	defer engine.Close()

	start := time.Now()

	// Session 1: events arriving within 50ms gap
	_ = engine.Add(window.Event[string]{Key: "session_user", Timestamp: start, Value: "click_home"})
	_ = engine.Add(window.Event[string]{Key: "session_user", Timestamp: start.Add(20 * time.Millisecond), Value: "view_product"})

	// Gap of 80ms (exceeds 50ms gap!) -> Closes Session 1
	_ = engine.Add(window.Event[string]{Key: "session_user", Timestamp: start.Add(100 * time.Millisecond), Value: "checkout"})

	if len(sessions) != 1 {
		t.Fatalf("expected 1 closed session, got %d", len(sessions))
	}

	s1 := sessions[0]
	if s1.Count != 2 {
		t.Errorf("expected session 1 count 2, got %d", s1.Count)
	}
}

func TestWatermarkLateEventDrop(t *testing.T) {
	cfg := window.TumblingConfig[int](1*time.Minute, func(res window.WindowResult[int]) {})
	engine, _ := window.New(cfg, window.WithAllowedLateness[int](10*time.Second))
	defer engine.Close()

	now := time.Now()

	// Advance watermark to now
	_ = engine.Add(window.Event[int]{Key: "k", Timestamp: now, Value: 1})

	// Event 20 seconds in the past (exceeds 10s allowed lateness) -> must be dropped!
	err := engine.Add(window.Event[int]{Key: "k", Timestamp: now.Add(-20 * time.Second), Value: 2})
	if !errors.Is(err, window.ErrLateEventDropped) {
		t.Fatalf("expected ErrLateEventDropped, got: %v", err)
	}
}

