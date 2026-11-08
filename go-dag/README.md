# go-dag 🕸️

> **High-Performance Directed Acyclic Graph Task Runner for Go (Go 1.23+).**  
> Workflows, ETL pipelines, build systems, aur microservice dependency orchestration ke liye concurrent DAG execution engine. Kahn's algorithm ke sath circular dependency (cycle) detect karta hai aur parallel nodes ko safely execute karta hai.

---

## 🚀 Features

- **Kahn's Topological Sort**: Sabhi nodes ka correct execution sequence determine karta hai.
- **Cycle Detection**: Agar graph me circular dependency ho (A -> B -> A), toh execution se pehle hi `ErrCycleDetected` catch karta hai.
- **Concurrency-Bounded**: Limit ke according multiple independent tasks ko ek sath parallel goroutines me execute karta hai.
- **Panic & Failure Isolation**: Agar koi node fail ho jaye ya panic kare, toh downstream dependent nodes cancel ho jaate hain aur error safely return hota hai.
- **Zero External Dependencies**: Pure Go Standard Library!

---

## 📦 Quickstart

```go
package main

import (
	"context"
	"fmt"
	"go-dag"
)

func main() {
	g := dag.New()

	// 1. Nodes register karein
	g.AddNode("fetch_user", func(ctx context.Context) error {
		fmt.Println("1. Fetched User")
		return nil
	})
	g.AddNode("fetch_orders", func(ctx context.Context) error {
		fmt.Println("2. Fetched Orders")
		return nil
	})
	g.AddNode("generate_report", func(ctx context.Context) error {
		fmt.Println("3. Generated Final Report")
		return nil
	})

	// 2. Dependencies define karein (fetch_user aur fetch_orders ke baad hi report banegi)
	g.AddEdge("fetch_user", "generate_report")
	g.AddEdge("fetch_orders", "generate_report")

	// 3. Concurrently run karein
	err := g.Run(context.Background(), 4)
	if err != nil {
		panic(err)
	}
}
```
