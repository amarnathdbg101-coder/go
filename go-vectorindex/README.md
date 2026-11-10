# go-vectorindex

`go-vectorindex` provides a high-performance in-memory Vector Similarity Search and Nearest Neighbor (KNN) index in pure Go. Built for modern AI applications, Retrieval-Augmented Generation (RAG), and embedding search, it indexes dense floating-point embedding vectors and executes sub-millisecond Top-K queries with loop-unrolled hardware vector arithmetic.

## Key Features

- **Multiple Distance Metrics**:
  - `MetricCosine`: Cosine similarity in $[-1.0, 1.0]$.
  - `MetricEuclidean`: L2 Euclidean distance.
  - `MetricDotProduct`: Raw or normalized inner product.
- **Top-K Streaming Min/Max Heap**: Efficient $O(N \log K)$ neighbor ranking avoids sorting entire collections.
- **Generic Metadata Attachment**: Associate arbitrary Go structs (`Index[T]`) directly with vectors.
- **Loop-Unrolled Vector Arithmetic**: 4-way loop unrolling for pipelined CPU vector dot products.
- **Thread-Safe**: RWMutex protected for concurrent real-time indexing and searching.
- **Zero Dependencies**: Pure Go standard library (`math`, `container/heap`, `sync`).

## Quick Start

```go
package main

import (
	"fmt"
	"go-vectorindex"
)

func main() {
	// Index for 384-dimensional text embeddings (e.g. MiniLM)
	idx, err := vectorindex.New[string](3, vectorindex.WithMetric(vectorindex.MetricCosine))
	if err != nil {
		panic(err)
	}

	// Insert document embeddings
	_ = idx.Insert("doc_1", []float32{0.9, 0.1, 0.0}, "Article on Machine Learning")
	_ = idx.Insert("doc_2", []float32{0.0, 0.1, 0.9}, "Recipe for Sourdough Bread")

	// Search Top-1 most similar
	query := []float32{0.95, 0.05, 0.0}
	results, _ := idx.Search(query, 1)

	for _, r := range results {
		fmt.Printf("Matched: %s (Score: %.4f) -> %s\n", r.ID, r.Score, r.Payload)
	}
}
```
