package set

import (
	"encoding/json"
	"sync"
	"testing"
)

func TestSet_BasicOps(t *testing.T) {
	s := New("apple", "banana", "cherry")

	if s.Size() != 3 {
		t.Errorf("expected size 3, got %d", s.Size())
	}
	if !s.Contains("apple") || !s.Contains("banana") {
		t.Errorf("expected set to contain apple and banana")
	}
	if s.Contains("mango") {
		t.Errorf("expected set not to contain mango")
	}

	s.Remove("banana")
	if s.Size() != 2 || s.Contains("banana") {
		t.Errorf("expected banana to be removed")
	}

	if !s.ContainsAll("apple", "cherry") {
		t.Errorf("expected ContainsAll to be true")
	}
	if !s.ContainsAny("mango", "cherry") {
		t.Errorf("expected ContainsAny to be true")
	}

	s.Clear()
	if !s.IsEmpty() || s.Size() != 0 {
		t.Errorf("expected set to be empty after clear")
	}
}

func TestSet_Algebra(t *testing.T) {
	s1 := New(1, 2, 3, 4)
	s2 := New(3, 4, 5, 6)

	// Union: {1, 2, 3, 4, 5, 6}
	union := s1.Union(s2)
	if union.Size() != 6 || !union.ContainsAll(1, 2, 3, 4, 5, 6) {
		t.Errorf("union mismatch: %+v", union.ToSlice())
	}

	// Intersect: {3, 4}
	intersect := s1.Intersect(s2)
	if intersect.Size() != 2 || !intersect.ContainsAll(3, 4) {
		t.Errorf("intersect mismatch: %+v", intersect.ToSlice())
	}

	// Difference (s1 - s2): {1, 2}
	diff := s1.Difference(s2)
	if diff.Size() != 2 || !diff.ContainsAll(1, 2) {
		t.Errorf("difference mismatch: %+v", diff.ToSlice())
	}

	// Symmetric Difference: {1, 2, 5, 6}
	symDiff := s1.SymmetricDifference(s2)
	if symDiff.Size() != 4 || !symDiff.ContainsAll(1, 2, 5, 6) {
		t.Errorf("symmetric difference mismatch: %+v", symDiff.ToSlice())
	}

	// Subsets
	sub := New(1, 2)
	if !sub.IsSubset(s1) {
		t.Errorf("expected {1, 2} to be subset of s1")
	}
	if !s1.IsSuperset(sub) {
		t.Errorf("expected s1 to be superset of {1, 2}")
	}
}

func TestSet_FilterAndEach(t *testing.T) {
	s := New(1, 2, 3, 4, 5, 6, 7, 8)
	evens := s.Filter(func(n int) bool {
		return n%2 == 0
	})

	if evens.Size() != 4 || !evens.ContainsAll(2, 4, 6, 8) {
		t.Errorf("filter mismatch: %+v", evens.ToSlice())
	}

	sum := 0
	evens.Each(func(n int) {
		sum += n
	})
	if sum != 20 {
		t.Errorf("expected sum 20, got %d", sum)
	}
}

func TestSet_JSON(t *testing.T) {
	s := New("red", "green", "blue")
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("failed to marshal set: %v", err)
	}

	var restored Set[string]
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("failed to unmarshal set: %v", err)
	}

	if !s.Equal(&restored) {
		t.Errorf("expected restored set to equal original set")
	}
}

func TestSet_ThreadSafety(t *testing.T) {
	ts := NewThreadSafe[int]()
	var wg sync.WaitGroup
	const routines = 30
	const iterations = 50

	for i := 0; i < routines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				val := id*1000 + j
				ts.Add(val)
				if !ts.Contains(val) {
					t.Errorf("concurrent contains check failed for %d", val)
				}
				if j%2 == 0 {
					ts.Remove(val)
				}
			}
		}(i)
	}

	wg.Wait()
}
