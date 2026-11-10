package vectorindex

import (
	"container/heap"
	"sort"
	"sync"
)

// SearchResult represents a nearest neighbor match.
type SearchResult[T any] struct {
	ID      string
	Score   float32
	Payload T
}

type item[T any] struct {
	id      string
	vector  []float32
	payload T
}

// Index represents an in-memory high-dimensional vector search index.
type Index[T any] struct {
	mu    sync.RWMutex
	cfg   Config
	items map[string]*item[T]
}

// New creates an initialized vector index.
func New[T any](dimensions int, opts ...Option) (*Index[T], error) {
	if dimensions <= 0 {
		return nil, ErrDimensionMismatch
	}

	cfg := DefaultConfig(dimensions)
	for _, opt := range opts {
		opt(&cfg)
	}

	return &Index[T]{
		cfg:   cfg,
		items: make(map[string]*item[T]),
	}, nil
}

// Insert adds or updates a vector with associated metadata payload.
func (idx *Index[T]) Insert(id string, vector []float32, payload T) error {
	if len(vector) != idx.cfg.Dimensions {
		return ErrDimensionMismatch
	}

	vCopy := make([]float32, len(vector))
	copy(vCopy, vector)

	idx.mu.Lock()
	defer idx.mu.Unlock()

	idx.items[id] = &item[T]{
		id:      id,
		vector:  vCopy,
		payload: payload,
	}
	return nil
}

// Delete removes an item from the index by ID.
func (idx *Index[T]) Delete(id string) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	if _, exists := idx.items[id]; !exists {
		return ErrItemNotFound
	}
	delete(idx.items, id)
	return nil
}

// Len returns the count of indexed vectors.
func (idx *Index[T]) Len() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return len(idx.items)
}

// Get retrieves an item by ID.
func (idx *Index[T]) Get(id string) (SearchResult[T], error) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	it, exists := idx.items[id]
	if !exists {
		var zero SearchResult[T]
		return zero, ErrItemNotFound
	}

	return SearchResult[T]{
		ID:      it.id,
		Payload: it.payload,
	}, nil
}

// Search finds the Top-K nearest neighbors to the query vector.
func (idx *Index[T]) Search(query []float32, k int) ([]SearchResult[T], error) {
	if len(query) != idx.cfg.Dimensions {
		return nil, ErrDimensionMismatch
	}
	if k <= 0 {
		return nil, ErrInvalidK
	}

	idx.mu.RLock()
	defer idx.mu.RUnlock()

	if len(idx.items) == 0 {
		return nil, nil
	}

	if idx.cfg.Metric == MetricEuclidean {
		// For Euclidean distance, lower is better. Use Max-Heap of size K.
		h := &maxHeap[T]{}
		heap.Init(h)

		for _, it := range idx.items {
			dist := EuclideanDistance(query, it.vector)
			if h.Len() < k {
				heap.Push(h, SearchResult[T]{ID: it.id, Score: dist, Payload: it.payload})
			} else if dist < (*h)[0].Score {
				(*h)[0] = SearchResult[T]{ID: it.id, Score: dist, Payload: it.payload}
				heap.Fix(h, 0)
			}
		}

		res := make([]SearchResult[T], h.Len())
		for i := len(res) - 1; i >= 0; i-- {
			res[i] = heap.Pop(h).(SearchResult[T])
		}
		return res, nil
	}

	// For Cosine / DotProduct, higher score is better. Use Min-Heap of size K.
	h := &minHeap[T]{}
	heap.Init(h)

	for _, it := range idx.items {
		var score float32
		if idx.cfg.Metric == MetricCosine {
			s, err := CosineSimilarity(query, it.vector)
			if err != nil {
				continue
			}
			score = s
		} else {
			score = DotProduct(query, it.vector)
		}

		if h.Len() < k {
			heap.Push(h, SearchResult[T]{ID: it.id, Score: score, Payload: it.payload})
		} else if score > (*h)[0].Score {
			(*h)[0] = SearchResult[T]{ID: it.id, Score: score, Payload: it.payload}
			heap.Fix(h, 0)
		}
	}

	res := make([]SearchResult[T], h.Len())
	for i := len(res) - 1; i >= 0; i-- {
		res[i] = heap.Pop(h).(SearchResult[T])
	}
	// Sort descending (highest similarity first)
	sort.Slice(res, func(i, j int) bool {
		return res[i].Score > res[j].Score
	})
	return res, nil
}

// -------------------------------------------------------------
// Priority Queues for Top-K Streaming
// -------------------------------------------------------------

type minHeap[T any] []SearchResult[T]

func (h minHeap[T]) Len() int           { return len(h) }
func (h minHeap[T]) Less(i, j int) bool { return h[i].Score < h[j].Score }
func (h minHeap[T]) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *minHeap[T]) Push(x any)        { *h = append(*h, x.(SearchResult[T])) }
func (h *minHeap[T]) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

type maxHeap[T any] []SearchResult[T]

func (h maxHeap[T]) Len() int           { return len(h) }
func (h maxHeap[T]) Less(i, j int) bool { return h[i].Score > h[j].Score }
func (h maxHeap[T]) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *maxHeap[T]) Push(x any)        { *h = append(*h, x.(SearchResult[T])) }
func (h *maxHeap[T]) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
