package gopubsub

import (
	"sync"
	"testing"
	"time"
)

type ChatMessage struct {
	Sender string
	Text   string
}

func TestPubSub_Broadcast(t *testing.T) {
	ps := New[ChatMessage](10, DropOldest)
	defer ps.Close()

	topic := "room_101"

	sub1 := ps.Subscribe(topic)
	defer sub1.Unsubscribe()

	sub2 := ps.Subscribe(topic)
	defer sub2.Unsubscribe()

	if ps.SubscribersCount(topic) != 2 {
		t.Fatalf("expected 2 subscribers, got %d", ps.SubscribersCount(topic))
	}

	msg := ChatMessage{Sender: "Alice", Text: "Hello Room!"}
	delivered := ps.Publish(topic, msg)
	if delivered != 2 {
		t.Fatalf("expected 2 deliveries, got %d", delivered)
	}

	select {
	case m1 := <-sub1.Ch:
		if m1.Text != "Hello Room!" {
			t.Fatalf("sub1 message mismatch: %+v", m1)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("sub1 timeout")
	}

	select {
	case m2 := <-sub2.Ch:
		if m2.Text != "Hello Room!" {
			t.Fatalf("sub2 message mismatch: %+v", m2)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("sub2 timeout")
	}
}

func TestPubSub_Unsubscribe(t *testing.T) {
	ps := New[string](5, DropOldest)
	defer ps.Close()

	sub := ps.Subscribe("live_updates")
	if ps.SubscribersCount("live_updates") != 1 {
		t.Fatal("expected 1 subscriber")
	}

	sub.Unsubscribe()

	if ps.SubscribersCount("live_updates") != 0 {
		t.Fatal("expected 0 subscribers after unsubscribe")
	}

	// Channel should be closed
	_, ok := <-sub.Ch
	if ok {
		t.Fatal("expected channel to be closed after unsubscribe")
	}
}

func TestPubSub_SlowConsumer_DropOldest(t *testing.T) {
	// Tiny buffer of 2 items
	ps := New[int](2, DropOldest)
	defer ps.Close()

	sub := ps.Subscribe("sensor_telemetry")
	defer sub.Unsubscribe()

	// Publish 4 items without reading from sub.Ch
	ps.Publish("sensor_telemetry", 1) // in buffer [1]
	ps.Publish("sensor_telemetry", 2) // in buffer [1, 2]
	ps.Publish("sensor_telemetry", 3) // buffer full: drops 1 -> [2, 3]
	ps.Publish("sensor_telemetry", 4) // buffer full: drops 2 -> [3, 4]

	// Read remaining 2 items: must be 3 and 4!
	item1 := <-sub.Ch
	item2 := <-sub.Ch

	if item1 != 3 || item2 != 4 {
		t.Fatalf("expected [3, 4] after dropping oldest, got [%d, %d]", item1, item2)
	}
}

func TestPubSub_SlowConsumer_DropNew(t *testing.T) {
	// Buffer of 2 items
	ps := New[int](2, DropNew)
	defer ps.Close()

	sub := ps.Subscribe("stock_ticker")
	defer sub.Unsubscribe()

	ps.Publish("stock_ticker", 100) // [100]
	ps.Publish("stock_ticker", 200) // [100, 200]
	// 3rd publish must be dropped because buffer is full
	delivered := ps.Publish("stock_ticker", 300)
	if delivered != 0 {
		t.Fatalf("expected 0 delivered for full buffer with DropNew, got %d", delivered)
	}

	item1 := <-sub.Ch
	item2 := <-sub.Ch
	if item1 != 100 || item2 != 200 {
		t.Fatalf("expected [100, 200], got [%d, %d]", item1, item2)
	}
}

func TestPubSub_ConcurrentPublish(t *testing.T) {
	ps := New[int](100, DropOldest)
	defer ps.Close()

	sub := ps.Subscribe("metrics")
	defer sub.Unsubscribe()

	const routines = 10
	const perRoutine = 50
	var wg sync.WaitGroup
	wg.Add(routines)

	for i := 0; i < routines; i++ {
		go func(val int) {
			defer wg.Done()
			for j := 0; j < perRoutine; j++ {
				ps.Publish("metrics", val)
			}
		}(i)
	}

	wg.Wait()

	// Drain messages
	received := 0
	for {
		select {
		case <-sub.Ch:
			received++
		default:
			goto DRAIN_DONE
		}
	}
DRAIN_DONE:
	if received == 0 {
		t.Fatal("expected messages to be received")
	}
}
