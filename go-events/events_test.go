package goevents

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
)

func TestMatcher(t *testing.T) {
	tests := []struct {
		pattern string
		topic   string
		match   bool
	}{
		{"order.created", "order.created", true},
		{"order.created", "order.updated", false},
		{"order.*", "order.created", true},
		{"order.*", "order.updated", true},
		{"order.*", "order.item.added", false}, // single wildcard should not cross boundary
		{"order.**", "order.item.added", true},
		{"order.>", "order.item.added", true},
		{"**", "any.random.topic.here", true},
		{"user.*.success", "user.login.success", true},
		{"user.*.success", "user.login.failed", false},
	}

	for _, tt := range tests {
		got := MatchTopic(tt.pattern, tt.topic)
		if got != tt.match {
			t.Errorf("MatchTopic(%q, %q) = %v; expected %v", tt.pattern, tt.topic, got, tt.match)
		}
	}
}

func TestBus_SyncPublish(t *testing.T) {
	bus := New()
	defer bus.Close()

	var received atomic.Int32

	sub, err := bus.Subscribe("user.registered", func(ctx context.Context, e Event) error {
		received.Add(1)
		payload := e.Payload.(map[string]string)
		if payload["email"] != "rahul@example.com" {
			t.Errorf("unexpected payload: %v", payload)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	event := NewEvent("user.registered", map[string]string{"email": "rahul@example.com"})
	err = bus.Publish(context.Background(), event)
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	if received.Load() != 1 {
		t.Fatalf("expected 1 received event, got %d", received.Load())
	}

	// Unsubscribe
	sub.Unsubscribe()

	_ = bus.Publish(context.Background(), event)
	if received.Load() != 1 {
		t.Fatalf("received event after unsubscribe!")
	}
}

func TestBus_AsyncPublish(t *testing.T) {
	bus := New(Config{WorkerCount: 2, QueueSize: 100})
	defer bus.Close()

	var wg sync.WaitGroup
	wg.Add(5)

	var count atomic.Int32

	_, _ = bus.Subscribe("order.*", func(ctx context.Context, e Event) error {
		count.Add(1)
		wg.Done()
		return nil
	})

	for i := 0; i < 5; i++ {
		_ = bus.PublishAsync(NewEvent("order.shipped", i))
	}

	wg.Wait()

	if count.Load() != 5 {
		t.Fatalf("expected 5 async events, got %d", count.Load())
	}
}

func TestBus_PanicIsolation(t *testing.T) {
	var panicReported atomic.Bool

	bus := New(Config{
		OnPanic: func(topic string, r any) {
			panicReported.Store(true)
		},
	})
	defer bus.Close()

	var healthyExecuted atomic.Bool

	// Sub 1: Panics
	_, _ = bus.Subscribe("payment.received", func(ctx context.Context, e Event) error {
		panic("fatal unexpected nil pointer")
	})

	// Sub 2: Healthy
	_, _ = bus.Subscribe("payment.*", func(ctx context.Context, e Event) error {
		healthyExecuted.Store(true)
		return nil
	})

	_ = bus.Publish(context.Background(), NewEvent("payment.received", 1000))

	if !panicReported.Load() {
		t.Fatal("expected panic callback to be invoked")
	}
	if !healthyExecuted.Load() {
		t.Fatal("healthy subscriber should execute even if another subscriber panics")
	}
}

func TestBus_DeadLetterQueue(t *testing.T) {
	var dlqReceived atomic.Bool

	bus := New(Config{
		OnDLQ: func(event Event, err error) {
			dlqReceived.Store(true)
		},
	})
	defer bus.Close()

	// Handler returns an error
	_, _ = bus.Subscribe("inventory.deduct", func(ctx context.Context, e Event) error {
		return errors.New("out of stock")
	})

	_ = bus.Publish(context.Background(), NewEvent("inventory.deduct", "ITEM-10"))

	if !dlqReceived.Load() {
		t.Fatal("expected event to land in Dead Letter Queue on error")
	}
}

func TestBus_Middleware(t *testing.T) {
	bus := New()
	defer bus.Close()

	var order []string

	bus.Use(func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, e Event) error {
			order = append(order, "mw1_start")
			err := next(ctx, e)
			order = append(order, "mw1_end")
			return err
		}
	})

	_, _ = bus.Subscribe("test.event", func(ctx context.Context, e Event) error {
		order = append(order, "handler")
		return nil
	})

	_ = bus.Publish(context.Background(), NewEvent("test.event", nil))

	if len(order) != 3 || order[0] != "mw1_start" || order[1] != "handler" || order[2] != "mw1_end" {
		t.Fatalf("unexpected middleware execution order: %v", order)
	}
}
