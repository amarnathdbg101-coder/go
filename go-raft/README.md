# go-raft

`go-raft` provides a pure Go standard library implementation of the Raft Distributed Consensus Protocol (Ongaro & Ousterhout, Stanford). It coordinates distributed state machine replication, leader election, and log consensus across independent cluster nodes, powering distributed databases, key-value stores, and metadata coordinators.

## Key Features

- **Core Raft State Machine**: Complete `Follower`, `Candidate`, and `Leader` role transitions.
- **Randomized Election Timers**: Prevents split votes via randomized election timeouts.
- **Quorum-Based Log Replication**: Leader commits log entries only after majority cluster acknowledgement.
- **Transport Agnostic**: Clean `Transport` interface plugs directly into TCP, gRPC, or in-memory testing harnesses.
- **Deterministic State Machine Delivery**: Committed commands are safely emitted over Go channels (`ApplyCh`).
- **Zero Dependencies**: Pure Go standard library (`sync`, `time`, `math/rand`).

## Quick Start

```go
package main

import (
	"fmt"
	"go-raft"
)

func main() {
	applyCh := make(chan raft.ApplyMsg, 100)

	// Initialize Raft consensus node
	node, err := raft.New(
		"node-1",
		[]string{"node-2", "node-3"},
		myNetworkTransport,
		applyCh,
	)
	if err != nil {
		panic(err)
	}
	defer node.Stop()

	// Propose mutation to cluster
	index, term, err := node.Propose([]byte("user:101:credit=500"))
	if err == nil {
		fmt.Printf("Proposed at index %d in term %d\n", index, term)
	}

	// Consume committed entries
	go func() {
		for msg := range applyCh {
			fmt.Printf("Committed command: %s (Index: %d)\n", string(msg.Command), msg.CommandIndex)
		}
	}()
}
```
