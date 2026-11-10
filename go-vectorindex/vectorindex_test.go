package vectorindex_test

import (
	"fmt"
	"math"
	"sync"
	"testing"

	"github.com/amarnathdbg101-coder/go/go-vectorindex"
)

func TestVectorIndex_CosineSimilarity(t *testing.T) {
	idx, err := vectorindex.New[string](4, vectorindex.WithMetric(vectorindex.MetricCosine))
	if err != nil {
		t.Fatalf("failed to create vector index: %v", err)
	}

	// Insert vectors
	_ = idx.Insert("doc_apple", []float32{1.0, 0.0, 0.0, 0.0}, "Apple Fruit")
	_ = idx.Insert("doc_pear", []float32{0.9, 0.1, 0.0, 0.0}, "Pear Fruit")
	_ = idx.Insert("doc_car", []float32{0.0, 1.0, 0.0, 0.0}, "Automobile Vehicle")

	// Query closest to apple
	results, err := idx.Search([]float32{1.0, 0.0, 0.0, 0.0}, 2)
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	if results[0].ID != "doc_apple" {
		t.Fatalf("expected top match to be doc_apple, got %s", results[0].ID)
	}
	if math.Abs(float64(results[0].Score)-1.0) > 0.001 {
		t.Fatalf("expected perfect score 1.0, got %f", results[0].Score)
	}

	if results[1].ID != "doc_pear" {
		t.Fatalf("expected 2nd match to be doc_pear, got %s", results[1].ID)
	}
}

func TestVectorIndex_EuclideanDistance(t *testing.T) {
	idx, err := vectorindex.New[string](3, vectorindex.WithMetric(vectorindex.MetricEuclidean))
	if err != nil {
		t.Fatal(err)
	}

	_ = idx.Insert("p1", []float32{1.0, 1.0, 1.0}, "Point 1")
	_ = idx.Insert("p2", []float32{2.0, 2.0, 2.0}, "Point 2")
	_ = idx.Insert("p3", []float32{10.0, 10.0, 10.0}, "Point 3")

	results, err := idx.Search([]float32{1.1, 1.1, 1.1}, 2)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// Closest should be p1 (distance ~ 0.17)
	if results[0].ID != "p1" {
		t.Fatalf("expected p1 closest, got %s", results[0].ID)
	}
	if results[1].ID != "p2" {
		t.Fatalf("expected p2 second closest, got %s", results[1].ID)
	}
}

func TestVectorIndex_TopKStreaming(t *testing.T) {
	idx, _ := vectorindex.New[int](2, vectorindex.WithMetric(vectorindex.MetricCosine))

	// Insert 100 vectors along different angles
	for i := 0; i < 100; i++ {
		angle := float64(i) * 0.01
		v := []float32{float32(math.Cos(angle)), float32(math.Sin(angle))}
		_ = idx.Insert(fmt.Sprintf("vec_%d", i), v, i)
	}

	// Search Top-5
	results, err := idx.Search([]float32{1.0, 0.0}, 5)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 5 {
		t.Fatalf("expected 5 results, got %d", len(results))
	}

	// Verify descending score order
	for i := 0; i < len(results)-1; i++ {
		if results[i].Score < results[i+1].Score {
			t.Fatalf("results not in descending score order: %f < %f", results[i].Score, results[i+1].Score)
		}
	}
}

func TestVectorIndex_DimensionMismatch(t *testing.T) {
	idx, _ := vectorindex.New[string](3)

	err := idx.Insert("bad_dim", []float32{1.0, 2.0}, "data")
	if err != vectorindex.ErrDimensionMismatch {
		t.Fatalf("expected ErrDimensionMismatch on insert, got %v", err)
	}

	_, err = idx.Search([]float32{1.0}, 5)
	if err != vectorindex.ErrDimensionMismatch {
		t.Fatalf("expected ErrDimensionMismatch on search, got %v", err)
	}
}

func TestVectorIndex_Concurrency(t *testing.T) {
	idx, _ := vectorindex.New[int](4)
	var wg sync.WaitGroup

	for w := 0; w < 5; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				id := fmt.Sprintf("w_%d_i_%d", workerID, i)
				v := []float32{float32(workerID), float32(i), 1.0, 2.0}
				_ = idx.Insert(id, v, i)
				_, _ = idx.Search([]float32{1.0, 2.0, 3.0, 4.0}, 3)
			}
		}(w)
	}

	wg.Wait()
}

