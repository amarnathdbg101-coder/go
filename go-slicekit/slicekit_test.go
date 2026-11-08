package slicekit

import (
	"reflect"
	"strings"
	"testing"
)

func TestChunk(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5, 6, 7}
	chunks := Chunk(nums, 3)

	if len(chunks) != 3 {
		t.Fatalf("expected 3 chunks, got %d", len(chunks))
	}
	if !reflect.DeepEqual(chunks[0], []int{1, 2, 3}) ||
		!reflect.DeepEqual(chunks[1], []int{4, 5, 6}) ||
		!reflect.DeepEqual(chunks[2], []int{7}) {
		t.Errorf("chunk mismatch: %+v", chunks)
	}

	// Invalid chunk size
	if Chunk(nums, 0) != nil || Chunk(nums, -1) != nil {
		t.Errorf("expected nil for invalid chunk size")
	}
}

func TestFlatten(t *testing.T) {
	nested := [][]string{
		{"a", "b"},
		{"c"},
		{"d", "e", "f"},
	}

	flat := Flatten(nested)
	expected := []string{"a", "b", "c", "d", "e", "f"}
	if !reflect.DeepEqual(flat, expected) {
		t.Errorf("expected %+v, got %+v", expected, flat)
	}
}

func TestUniqueAndUniqueBy(t *testing.T) {
	nums := []int{1, 2, 2, 3, 1, 4, 5, 4}
	uniq := Unique(nums)
	expected := []int{1, 2, 3, 4, 5}
	if !reflect.DeepEqual(uniq, expected) {
		t.Errorf("unique mismatch: %+v", uniq)
	}

	// UniqueBy
	words := []string{"apple", "Banana", "APPLE", "cherry", "banana"}
	uniqWords := UniqueBy(words, func(s string) string {
		return strings.ToLower(s)
	})
	expectedWords := []string{"apple", "Banana", "cherry"}
	if !reflect.DeepEqual(uniqWords, expectedWords) {
		t.Errorf("uniqueBy mismatch: %+v", uniqWords)
	}
}

func TestGroupBy(t *testing.T) {
	type User struct {
		Name string
		Role string
	}

	users := []User{
		{"Amit", "admin"},
		{"Priya", "editor"},
		{"Rahul", "admin"},
		{"Sneha", "viewer"},
	}

	groups := GroupBy(users, func(u User) string {
		return u.Role
	})

	if len(groups["admin"]) != 2 || len(groups["editor"]) != 1 || len(groups["viewer"]) != 1 {
		t.Errorf("group by mismatch: %+v", groups)
	}
}

func TestPartition(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5, 6, 7, 8}
	evens, odds := Partition(nums, func(n int) bool {
		return n%2 == 0
	})

	if !reflect.DeepEqual(evens, []int{2, 4, 6, 8}) || !reflect.DeepEqual(odds, []int{1, 3, 5, 7}) {
		t.Errorf("partition mismatch: evens=%+v, odds=%+v", evens, odds)
	}
}

func TestReverseAndCompact(t *testing.T) {
	nums := []int{1, 2, 3, 4}
	rev := Reverse(nums)
	if !reflect.DeepEqual(rev, []int{4, 3, 2, 1}) {
		t.Errorf("reverse mismatch: %+v", rev)
	}
	// Verify input wasn't mutated
	if nums[0] != 1 {
		t.Errorf("original slice was mutated")
	}

	ReverseInPlace(nums)
	if !reflect.DeepEqual(nums, []int{4, 3, 2, 1}) {
		t.Errorf("reverse in place mismatch: %+v", nums)
	}

	// Compact
	withZeros := []int{0, 1, 0, 2, 3, 0, 4}
	compacted := Compact(withZeros)
	if !reflect.DeepEqual(compacted, []int{1, 2, 3, 4}) {
		t.Errorf("compact mismatch: %+v", compacted)
	}
}

func TestMapFilterReduce(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5}

	// Map
	doubled := Map(nums, func(n int) int {
		return n * 2
	})
	if !reflect.DeepEqual(doubled, []int{2, 4, 6, 8, 10}) {
		t.Errorf("map mismatch: %+v", doubled)
	}

	// Filter
	gt2 := Filter(nums, func(n int) bool {
		return n > 2
	})
	if !reflect.DeepEqual(gt2, []int{3, 4, 5}) {
		t.Errorf("filter mismatch: %+v", gt2)
	}

	// Reduce
	sum := Reduce(nums, 0, func(acc, n int) int {
		return acc + n
	})
	if sum != 15 {
		t.Errorf("expected sum 15, got %d", sum)
	}
}

func TestFindAndContainsBy(t *testing.T) {
	names := []string{"Bob", "Charlie", "Alice"}

	found, ok := Find(names, func(s string) bool {
		return strings.HasPrefix(s, "A")
	})
	if !ok || found != "Alice" {
		t.Errorf("find failed: found=%s, ok=%v", found, ok)
	}

	idx := FindIndex(names, func(s string) bool {
		return s == "Charlie"
	})
	if idx != 1 {
		t.Errorf("expected index 1, got %d", idx)
	}

	if !ContainsBy(names, func(s string) bool { return len(s) == 7 }) {
		t.Errorf("expected ContainsBy true for length 7 (Charlie)")
	}
}

func TestShuffle(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	Shuffle(nums)
	if len(nums) != 10 {
		t.Errorf("shuffle changed length")
	}
}
