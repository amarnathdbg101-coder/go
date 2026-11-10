# go-eventstore

`go-eventstore` provides a production-grade Event Sourcing storage engine in pure Go (inspired by Greg Young's EventStoreDB and Axon Framework). It persists state as an immutable sequence of domain events grouped into streams with strict optimistic concurrency control, deterministic replays, and point-in-time snapshots.

## Key Features

- **Append-Only Event Streams**: Aggregate Roots are modeled as independent, isolated streams of immutable business facts.
- **Optimistic Concurrency Control (OCC)**: `ExpectedRevision` checks prevent silent overwrites and concurrent conflict corruption across distributed writers.
- **Global Sequential Log (`ReadAll`)**: Global monotonically increasing log position facilitates outbox projectors, read-model denormalization, and CQRS subscribers.
- **State Snapshotting**: Save and load aggregate state snapshots (`SaveSnapshot` / `LoadSnapshot`) to bypass replaying thousands of historical events on initialization.
- **Thread-Safe**: RWMutex protected for concurrent high-throughput append and read operations.
- **Zero Dependencies**: Pure Go standard library.

## Quick Start

```go
package main

import (
	"fmt"
	"go-eventstore"
)

func main() {
	store := eventstore.New()

	// Append initial domain events with NoStream expected revision
	streamID := "bank-account-101"
	_, err := store.AppendToStream(streamID, eventstore.NoStream, []eventstore.EventData{
		{EventType: "AccountOpened", Payload: []byte(`{"initial_balance": 1000}`)},
		{EventType: "FundsDeposited", Payload: []byte(`{"amount": 500}`)},
	})
	if err != nil {
		panic(err)
	}

	// Replay stream to reconstruct aggregate state
	events, _ := store.ReadStream(streamID, 1, 100)
	for _, e := range events {
		fmt.Printf("Replaying Rev %d [%s]: %s\n", e.Revision, e.EventType, string(e.Payload))
	}

	// Append with Optimistic Concurrency Control
	_, err = store.AppendToStream(streamID, 2, []eventstore.EventData{
		{EventType: "FundsWithdrawn", Payload: []byte(`{"amount": 200}`)},
	})
	if err != nil {
		fmt.Println("Concurrent write conflict:", err)
	}
}
```
