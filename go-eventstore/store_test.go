package eventstore_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/amarnathdbg101-coder/go/go-eventstore"
)

func TestEventStore_AppendAndRead(t *testing.T) {
	store := eventstore.New()

	events := []eventstore.EventData{
		{EventType: "AccountOpened", Payload: []byte(`{"id":"ACC-101","holder":"Amar"}`)},
		{EventType: "FundsDeposited", Payload: []byte(`{"amount":5000}`)},
	}

	recorded, err := store.AppendToStream("account-101", eventstore.NoStream, events)
	if err != nil {
		t.Fatalf("append failed: %v", err)
	}

	if len(recorded) != 2 {
		t.Fatalf("expected 2 recorded events, got %d", len(recorded))
	}
	if recorded[0].Revision != 1 || recorded[1].Revision != 2 {
		t.Fatalf("stream revisions unexpected: rev1=%d, rev2=%d", recorded[0].Revision, recorded[1].Revision)
	}

	// Read stream from revision 1
	readEvents, err := store.ReadStream("account-101", 1, 10)
	if err != nil {
		t.Fatalf("read stream failed: %v", err)
	}
	if len(readEvents) != 2 {
		t.Fatalf("expected 2 events read, got %d", len(readEvents))
	}
	if readEvents[0].EventType != "AccountOpened" || readEvents[1].EventType != "FundsDeposited" {
		t.Fatalf("event type mismatch: %v", readEvents)
	}
}

func TestEventStore_OptimisticConcurrency(t *testing.T) {
	store := eventstore.New()

	// Initial event sets stream revision to 1
	_, err := store.AppendToStream("cart-555", eventstore.NoStream, []eventstore.EventData{
		{EventType: "CartCreated", Payload: []byte(`{}`)},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Concurrently, Tx1 tries to append expecting revision 1
	_, err = store.AppendToStream("cart-555", 1, []eventstore.EventData{
		{EventType: "ItemAdded", Payload: []byte(`{"item":"book"}`)},
	})
	if err != nil {
		t.Fatalf("Tx1 append should succeed: %v", err)
	}

	// Tx2 tries to append also expecting revision 1 (stale revision!) -> must fail
	_, err = store.AppendToStream("cart-555", 1, []eventstore.EventData{
		{EventType: "ItemAdded", Payload: []byte(`{"item":"pen"}`)},
	})
	if err != eventstore.ErrWrongExpectedVersion {
		t.Fatalf("expected ErrWrongExpectedVersion for stale append, got: %v", err)
	}
}

func TestEventStore_Snapshot(t *testing.T) {
	store := eventstore.New()

	_ = store.SaveSnapshot("order-999", 10, []byte(`{"status":"fulfilled","total":1500}`))

	snap, err := store.LoadSnapshot("order-999")
	if err != nil {
		t.Fatalf("load snapshot failed: %v", err)
	}
	if snap.Revision != 10 || string(snap.State) != `{"status":"fulfilled","total":1500}` {
		t.Fatalf("unexpected snapshot state: %+v", snap)
	}
}

func TestEventStore_ReadAllGlobalLog(t *testing.T) {
	store := eventstore.New()

	_, _ = store.AppendToStream("stream-A", eventstore.AnyVersion, []eventstore.EventData{{EventType: "A1"}})
	_, _ = store.AppendToStream("stream-B", eventstore.AnyVersion, []eventstore.EventData{{EventType: "B1"}})
	_, _ = store.AppendToStream("stream-A", eventstore.AnyVersion, []eventstore.EventData{{EventType: "A2"}})

	global := store.ReadAll(1, 10)
	if len(global) != 3 {
		t.Fatalf("expected 3 global events, got %d", len(global))
	}

	for i := range global {
		expectedPos := uint64(i + 1)
		if global[i].GlobalPosition != expectedPos {
			t.Fatalf("global position mismatch at index %d: expected %d, got %d", i, expectedPos, global[i].GlobalPosition)
		}
	}
}

func TestEventStore_Concurrency(t *testing.T) {
	store := eventstore.New()
	var wg sync.WaitGroup

	for w := 0; w < 10; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			streamID := fmt.Sprintf("stream-worker-%d", workerID)
			for i := 0; i < 50; i++ {
				_, _ = store.AppendToStream(streamID, eventstore.AnyVersion, []eventstore.EventData{
					{EventType: "WorkerTick", Payload: []byte("data")},
				})
			}
		}(w)
	}

	wg.Wait()

	all := store.ReadAll(1, 1000)
	if len(all) != 500 {
		t.Fatalf("expected 500 events in global log, got %d", len(all))
	}
}

