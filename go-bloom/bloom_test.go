package gobloom

import (
	"fmt"
	"testing"
)

func TestBloom_BasicAddAndContains(t *testing.T) {
	// 10,000 items with 1% false positive rate
	bf, err := New(10000, 0.01)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	keys := []string{"apple", "banana", "cherry", "date", "elderberry"}
	for _, k := range keys {
		bf.AddString(k)
	}

	// All added keys must definitely return true
	for _, k := range keys {
		if !bf.ContainsString(k) {
			t.Fatalf("expected key %s to be present in bloom filter", k)
		}
	}

	// Non-added keys must return false (with 99% probability)
	absent := []string{"fig", "grape", "honeydew", "kiwi", "lemon"}
	for _, k := range absent {
		if bf.ContainsString(k) {
			t.Fatalf("key %s was not added but returned true", k)
		}
	}
}

func TestBloom_FalsePositiveRate(t *testing.T) {
	const count = 5000
	const targetFPR = 0.01 // 1%

	bf, _ := New(count, targetFPR)

	for i := 0; i < count; i++ {
		bf.AddString(fmt.Sprintf("user_id_%d", i))
	}

	// Test 10,000 elements NOT in the filter
	falsePositives := 0
	const testQueries = 10000

	for i := count; i < count+testQueries; i++ {
		if bf.ContainsString(fmt.Sprintf("user_id_%d", i)) {
			falsePositives++
		}
	}

	actualFPR := float64(falsePositives) / float64(testQueries)
	// Allow acceptable statistical tolerance (e.g. up to 2%)
	if actualFPR > 0.02 {
		t.Fatalf("false positive rate exceeded: actual %f, target %f", actualFPR, targetFPR)
	}
}

func TestBloom_SerializeRoundtrip(t *testing.T) {
	bf, _ := New(1000, 0.01)
	bf.AddString("order_101")
	bf.AddString("order_102")

	data := bf.Bytes()
	if len(data) == 0 {
		t.Fatal("serialized data should not be empty")
	}

	restored, err := FromBytes(data)
	if err != nil {
		t.Fatalf("FromBytes failed: %v", err)
	}

	if !restored.ContainsString("order_101") || !restored.ContainsString("order_102") {
		t.Fatal("restored bloom filter failed to retain items")
	}
	if restored.ContainsString("order_999") {
		t.Fatal("restored bloom filter should not contain order_999")
	}
}

func TestBloom_Reset(t *testing.T) {
	bf, _ := New(100, 0.01)
	bf.AddString("test_key")

	if !bf.ContainsString("test_key") {
		t.Fatal("expected test_key present before reset")
	}

	bf.Reset()

	if bf.ContainsString("test_key") {
		t.Fatal("test_key should not be present after reset")
	}
}

func TestCountingFilter_AddAndRemove(t *testing.T) {
	cf, err := NewCounting(1000, 0.01)
	if err != nil {
		t.Fatalf("failed to create CountingFilter: %v", err)
	}

	cf.AddString("user_101")
	cf.AddString("user_102")

	if !cf.ContainsString("user_101") || !cf.ContainsString("user_102") {
		t.Fatal("expected user_101 and user_102 to be present")
	}
	if cf.ContainsString("user_999") {
		t.Fatal("expected user_999 to not be present")
	}

	if cf.Elements() != 2 {
		t.Errorf("expected 2 elements, got %d", cf.Elements())
	}

	// Remove user_101
	removed := cf.RemoveString("user_101")
	if !removed {
		t.Fatal("expected successful removal of user_101")
	}
	if cf.ContainsString("user_101") {
		t.Fatal("user_101 should no longer be present after removal")
	}
	if !cf.ContainsString("user_102") {
		t.Fatal("user_102 should still be present after removing user_101")
	}

	if cf.Elements() != 1 {
		t.Errorf("expected 1 element, got %d", cf.Elements())
	}

	// Reset
	cf.Reset()
	if cf.ContainsString("user_102") {
		t.Fatal("user_102 should not be present after reset")
	}
}


func BenchmarkContains(b *testing.B) {
	bf, _ := New(100000, 0.01)
	for i := 0; i < 100000; i++ {
		bf.AddString(fmt.Sprintf("item_%d", i))
	}

	key := []byte("item_54321")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bf.Contains(key)
	}
}
