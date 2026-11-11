package eventemitter

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestEventEmitter_Basic(t *testing.T) {
	em := New()

	var count int
	unsub, err := em.On("user.login", func(event string, payload any) {
		count++
	})
	if err != nil {
		t.Fatalf("On failed: %v", err)
	}

	em.Emit("user.login", "user-1")
	em.Emit("user.login", "user-2")

	if count != 2 {
		t.Fatalf("expected count 2, got %d", count)
	}

	// Unsubscribe
	unsub()
	em.Emit("user.login", "user-3")
	if count != 2 {
		t.Fatalf("expected count 2 after unsub, got %d", count)
	}
}

func TestEventEmitter_Once(t *testing.T) {
	em := New()

	var onceCount int
	_, _ = em.Once("app.init", func(event string, payload any) {
		onceCount++
	})

	em.Emit("app.init", nil)
	em.Emit("app.init", nil)
	em.Emit("app.init", nil)

	if onceCount != 1 {
		t.Fatalf("expected Once listener to fire exactly 1 time, got %d", onceCount)
	}

	if em.ListenerCount("app.init") != 0 {
		t.Fatalf("expected listener count to be 0 after Once fired")
	}
}

func TestEventEmitter_Wildcards(t *testing.T) {
	em := New()

	var orderEvents []string
	var allEvents []string

	_, _ = em.On("order.*", func(event string, payload any) {
		orderEvents = append(orderEvents, event)
	})
	_, _ = em.On("*", func(event string, payload any) {
		allEvents = append(allEvents, event)
	})

	em.Emit("order.placed", 1)
	em.Emit("order.shipped", 2)
	em.Emit("user.signup", 3)

	if len(orderEvents) != 2 || orderEvents[0] != "order.placed" || orderEvents[1] != "order.shipped" {
		t.Fatalf("unexpected order events: %v", orderEvents)
	}
	if len(allEvents) != 3 {
		t.Fatalf("expected allEvents to have 3 items, got %d", len(allEvents))
	}
}

func TestEventEmitter_PanicIsolation(t *testing.T) {
	em := New()

	var reachedSecond bool
	_, _ = em.On("test.event", func(event string, payload any) {
		panic("boom") // should be caught safely
	})
	_, _ = em.On("test.event", func(event string, payload any) {
		reachedSecond = true
	})

	em.Emit("test.event", nil)

	if !reachedSecond {
		t.Fatalf("expected second listener to execute despite first listener panic")
	}
}

func TestEventEmitter_EmitAsync(t *testing.T) {
	em := New()
	var asyncHits int32

	_, _ = em.On("event.async", func(event string, payload any) {
		atomic.AddInt32(&asyncHits, 1)
	})

	em.EmitAsync("event.async", nil)
	em.EmitAsync("event.async", nil)

	time.Sleep(50 * time.Millisecond)

	if atomic.LoadInt32(&asyncHits) != 2 {
		t.Fatalf("expected 2 async hits, got %d", asyncHits)
	}
}

func TestEventEmitter_LeakWarning(t *testing.T) {
	em := New()
	em.SetMaxListeners(3)

	var leakWarned bool
	em.SetLeakWarningHandler(func(event string, count int) {
		leakWarned = true
	})

	for i := 0; i < 4; i++ {
		_, _ = em.On("heavy.event", func(event string, payload any) {})
	}

	if !leakWarned {
		t.Fatalf("expected leak warning triggered when count exceeded 3")
	}
}

func TestEventEmitter_Concurrent(t *testing.T) {
	em := New()
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			unsub, _ := em.On("burst", func(event string, payload any) {})
			em.Emit("burst", id)
			unsub()
		}(i)
	}
	wg.Wait()
}
