package gobitmap

import (
	"testing"
)

func TestBitmap_BasicOps(t *testing.T) {
	bm := New(100)

	// Set bits
	_ = bm.Set(0)
	_ = bm.Set(10)
	_ = bm.Set(63)
	_ = bm.Set(64) // Boundary across word 0 and word 1
	_ = bm.Set(1000) // Dynamic expansion

	if !bm.Test(0) || !bm.Test(10) || !bm.Test(63) || !bm.Test(64) || !bm.Test(1000) {
		t.Errorf("set bits failed verification")
	}

	if bm.Test(1) || bm.Test(50) || bm.Test(999) {
		t.Errorf("unset bits returned true")
	}

	if bm.Count() != 5 {
		t.Errorf("expected count 5, got %d", bm.Count())
	}

	// Toggle
	_ = bm.Toggle(10) // 1 -> 0
	if bm.Test(10) {
		t.Errorf("expected bit 10 to be cleared after toggle")
	}

	_ = bm.Toggle(20) // 0 -> 1
	if !bm.Test(20) {
		t.Errorf("expected bit 20 to be set after toggle")
	}

	// Clear
	_ = bm.Clear(64)
	if bm.Test(64) {
		t.Errorf("expected bit 64 to be cleared")
	}
}

func TestBitmap_SetOperations(t *testing.T) {
	bm1 := New(64)
	bm2 := New(64)

	// bm1 has {1, 2, 3}
	_ = bm1.Set(1)
	_ = bm1.Set(2)
	_ = bm1.Set(3)

	// bm2 has {2, 3, 4}
	_ = bm2.Set(2)
	_ = bm2.Set(3)
	_ = bm2.Set(4)

	// And (Intersection) -> {2, 3}
	andBm := bm1.And(bm2)
	if andBm.Count() != 2 || !andBm.Test(2) || !andBm.Test(3) {
		t.Errorf("And failed: count=%d", andBm.Count())
	}

	// Or (Union) -> {1, 2, 3, 4}
	orBm := bm1.Or(bm2)
	if orBm.Count() != 4 || !orBm.Test(1) || !orBm.Test(2) || !orBm.Test(3) || !orBm.Test(4) {
		t.Errorf("Or failed: count=%d", orBm.Count())
	}

	// AndNot (Difference: bm1 \ bm2) -> {1}
	diffBm := bm1.AndNot(bm2)
	if diffBm.Count() != 1 || !diffBm.Test(1) {
		t.Errorf("AndNot failed: count=%d", diffBm.Count())
	}
}

func TestBitmap_Serialization(t *testing.T) {
	bm := New(200)
	_ = bm.Set(5)
	_ = bm.Set(70)
	_ = bm.Set(150)

	bytes := bm.ToBytes()
	restored := FromBytes(bytes)

	if restored.Count() != 3 {
		t.Fatalf("restored count %d != 3", restored.Count())
	}

	if !restored.Test(5) || !restored.Test(70) || !restored.Test(150) {
		t.Errorf("restored bitmap lost set bits")
	}
}

func BenchmarkBitmap_SetAndTest(b *testing.B) {
	bm := New(1000000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bit := i % 1000000
		_ = bm.Set(bit)
		_ = bm.Test(bit)
	}
}
