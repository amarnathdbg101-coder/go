# go-simhash

`go-simhash` provides a high-performance Locality Sensitive Hashing (LSH) engine implementing Charikar's SimHash algorithm in pure Go. It enables ultra-fast near-duplicate document detection, web scraping deduplication, and plagiarism analysis by mapping high-dimensional text to 64-bit fingerprints where Hamming distance directly correlates with textual similarity.

## Key Features

- **Moses Charikar SimHash Algorithm**: 64-bit fingerprinting with standard bitwise POPCNT distance calculation.
- **Text Preprocessing & Shingling**: Built-in word tokenizer, case normalization, and n-gram shingle generator.
- **Weighted Feature Support**: Weight tokens based on frequency or TF-IDF relevance.
- **In-Memory Near-Duplicate Search Index**: Rapidly scan large document collections to find near-duplicates matching a maximum Hamming distance threshold.
- **Zero Dependencies**: Pure Go standard library (`math/bits`, `hash/fnv`, `unicode`).

## Quick Start

```go
package main

import (
	"fmt"
	"go-simhash"
)

func main() {
	docA := "Golang high concurrency and channels with WaitGroup"
	docB := "Golang high concurrency & channels with WaitGroup" // minor punctuation change
	docC := "Traditional French bakery sourdough bread baking"

	fpA := simhash.ComputeFromText(docA)
	fpB := simhash.ComputeFromText(docB)
	fpC := simhash.ComputeFromText(docC)

	fmt.Printf("Distance(A, B): %d (Similarity: %.2f%%)\n", fpA.Distance(fpB), fpA.Similarity(fpB)*100)
	fmt.Printf("Distance(A, C): %d (Similarity: %.2f%%)\n", fpA.Distance(fpC), fpA.Similarity(fpC)*100)

	// In-memory near-duplicate index
	idx := simhash.NewIndex()
	idx.Insert("article-1", fpA)
	idx.Insert("article-2", fpC)

	// Find near duplicates with Hamming distance <= 3
	matches := idx.FindNearDuplicates(fpB, 3)
	for _, m := range matches {
		fmt.Printf("Near match found: ID=%s, Distance=%d\n", m.ID, m.Distance)
	}
}
```
