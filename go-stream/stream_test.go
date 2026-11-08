package gostream

import (
	"strconv"
	"strings"
	"testing"
	"time"
)

type Product struct {
	ID    int
	Name  string
	Price float64
}

func TestStream_FilterAndMap(t *testing.T) {
	products := []Product{
		{ID: 1, Name: "Keyboard", Price: 1200},
		{ID: 2, Name: "Mouse", Price: 600},
		{ID: 3, Name: "Monitor", Price: 15000},
		{ID: 4, Name: "USB Cable", Price: 250},
	}

	// Filter expensive products (> 1000) and map to uppercase names
	names := Map(
		FromSlice(products).Filter(func(p Product) bool {
			return p.Price >= 1000
		}),
		func(p Product) string {
			return strings.ToUpper(p.Name)
		},
	).ToSlice()

	if len(names) != 2 || names[0] != "KEYBOARD" || names[1] != "MONITOR" {
		t.Fatalf("unexpected filter+map output: %+v", names)
	}
}

func TestStream_Distinct(t *testing.T) {
	numbers := []int{1, 2, 2, 3, 4, 4, 4, 5, 1}
	distinct := FromSlice(numbers).Distinct(func(n int) any {
		return n
	}).ToSlice()

	if len(distinct) != 5 {
		t.Fatalf("expected 5 distinct elements, got %d: %+v", len(distinct), distinct)
	}
}

func TestStream_Sort(t *testing.T) {
	numbers := []int{5, 2, 8, 1, 9}
	sorted := FromSlice(numbers).Sort(func(a, b int) bool {
		return a < b
	}).ToSlice()

	for i := 0; i < len(sorted)-1; i++ {
		if sorted[i] > sorted[i+1] {
			t.Fatalf("slice not sorted properly: %+v", sorted)
		}
	}
}

func TestStream_Chunk(t *testing.T) {
	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	chunks := Chunk(FromSlice(items), 3).ToSlice()

	if len(chunks) != 4 {
		t.Fatalf("expected 4 chunks, got %d", len(chunks))
	}
	if len(chunks[0]) != 3 || len(chunks[3]) != 1 {
		t.Fatalf("unexpected chunk sizes: %+v", chunks)
	}
}

func TestStream_FlatMap(t *testing.T) {
	sentences := []string{"hello world", "go generics"}
	words := FlatMap(FromSlice(sentences), func(s string) []string {
		return strings.Fields(s)
	}).ToSlice()

	if len(words) != 4 || words[0] != "hello" || words[3] != "generics" {
		t.Fatalf("unexpected flatmap output: %+v", words)
	}
}

func TestStream_Reduce(t *testing.T) {
	numbers := []int{10, 20, 30, 40}
	sum := FromSlice(numbers).Reduce(0, func(acc, item int) int {
		return acc + item
	})

	if sum != 100 {
		t.Fatalf("expected sum 100, got %d", sum)
	}
}

func TestStream_ParallelMap_PreservesOrder(t *testing.T) {
	const count = 50
	items := make([]int, count)
	for i := 0; i < count; i++ {
		items[i] = i
	}

	start := time.Now()
	results := ParallelMap(FromSlice(items), 8, func(n int) string {
		time.Sleep(2 * time.Millisecond) // Simulate light I/O or work
		return "VAL_" + strconv.Itoa(n*2)
	}).ToSlice()

	elapsed := time.Since(start)

	// 50 items * 2ms = 100ms sequentially. With 8 workers, should take < 35ms!
	if elapsed >= 80*time.Millisecond {
		t.Logf("Parallel execution took %v", elapsed)
	}

	if len(results) != count {
		t.Fatalf("expected %d results, got %d", count, len(results))
	}

	// Verify order was preserved strictly!
	for i := 0; i < count; i++ {
		expected := "VAL_" + strconv.Itoa(i*2)
		if results[i] != expected {
			t.Fatalf("order mismatch at index %d: expected %s, got %s", i, expected, results[i])
		}
	}
}
