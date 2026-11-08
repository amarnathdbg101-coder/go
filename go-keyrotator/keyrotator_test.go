package keyrotator

import (
	"sync"
	"testing"
	"time"
)

func TestRotator_Empty(t *testing.T) {
	r := New([]string{})
	if r.HasKeys() {
		t.Fatal("expected HasKeys() to be false")
	}

	_, err := r.GetNextKey()
	if err != ErrNoKeysConfigured {
		t.Fatalf("expected ErrNoKeysConfigured, got %v", err)
	}
}

func TestRotator_RoundRobin(t *testing.T) {
	r := New([]string{"key-1", "key-2", "key-3"})
	if r.KeyCount() != 3 {
		t.Fatalf("expected 3 keys, got %d", r.KeyCount())
	}

	seq := []string{}
	for i := 0; i < 6; i++ {
		k, err := r.GetNextKey()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		seq = append(seq, k)
	}

	expected := []string{"key-1", "key-2", "key-3", "key-1", "key-2", "key-3"}
	for i, exp := range expected {
		if seq[i] != exp {
			t.Fatalf("at index %d: expected %s, got %s", i, exp, seq[i])
		}
	}
}

func TestRotator_RateLimitSkip(t *testing.T) {
	r := New([]string{"key-1", "key-2", "key-3"})

	// Mark key-2 as rate-limited for 1 hour
	r.MarkRateLimited("key-2", 1*time.Hour)

	// Fetch 4 times: key-2 must never appear
	for i := 0; i < 4; i++ {
		k, err := r.GetNextKey()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if k == "key-2" {
			t.Fatalf("got rate-limited key-2!")
		}
	}
}

func TestRotator_AllExhaustedFallback(t *testing.T) {
	r := New([]string{"key-1", "key-2"})

	r.MarkRateLimited("key-1", 10*time.Minute)
	r.MarkRateLimited("key-2", 2*time.Minute) // Earliest recovery

	k, err := r.GetNextKey()
	if err != ErrAllKeysExhausted {
		t.Fatalf("expected ErrAllKeysExhausted, got %v", err)
	}
	if k != "key-2" {
		t.Fatalf("expected earliest recovering key 'key-2', got '%s'", k)
	}
}

func TestRotator_ErrorThresholdTripping(t *testing.T) {
	r := New([]string{"key-1", "key-2"}, WithErrorThreshold(2), WithErrorCooldown(50*time.Millisecond))

	// Trip key-1
	r.MarkError("key-1")
	r.MarkError("key-1")

	h, _ := r.GetHealth("key-1")
	if !h.IsRateLimited {
		t.Fatal("expected key-1 to be rate-limited after 2 errors")
	}

	time.Sleep(60 * time.Millisecond)

	// After cooldown, should recover
	k, err := r.GetNextKey()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if k != "key-1" && k != "key-2" {
		t.Fatalf("unexpected key: %s", k)
	}
}

func TestRotator_ConcurrentStress(t *testing.T) {
	r := New([]string{"key-a", "key-b", "key-c", "key-d"})

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				k, err := r.GetNextKey()
				if err == nil {
					if id%5 == 0 {
						r.MarkError(k)
					} else {
						r.MarkSuccess(k)
					}
				}
			}
		}(i)
	}

	wg.Wait()

	list := r.ListHealth()
	if len(list) != 4 {
		t.Fatalf("expected 4 health records, got %d", len(list))
	}
}

func TestRotator_ObservabilityHooks(t *testing.T) {
	var selectedKey string
	var rateLimitedKey string
	var cooldownDur time.Duration
	var errorStreak int
	var recoveredKey string

	rotator := New([]string{"key-alpha"}, WithHooks(Hooks{
		OnKeySelected: func(key string) {
			selectedKey = key
		},
		OnRateLimited: func(key string, cd time.Duration) {
			rateLimitedKey = key
			cooldownDur = cd
		},
		OnErrorTriggered: func(key string, streak int) {
			errorStreak = streak
		},
		OnCooldownRecovered: func(key string) {
			recoveredKey = key
		},
	}), WithErrorThreshold(2), WithErrorCooldown(50*time.Millisecond))

	k, err := rotator.GetNextKey()
	if err != nil || k != "key-alpha" || selectedKey != "key-alpha" {
		t.Fatalf("OnKeySelected hook failed, got: %s", selectedKey)
	}

	rotator.MarkError("key-alpha")
	if errorStreak != 1 {
		t.Fatalf("expected error streak 1, got %d", errorStreak)
	}

	// Second error should trip rate-limit hook
	rotator.MarkError("key-alpha")
	if rateLimitedKey != "key-alpha" || cooldownDur != 50*time.Millisecond {
		t.Fatalf("OnRateLimited hook failed")
	}

	time.Sleep(60 * time.Millisecond)

	// Fetch again -> should trigger OnCooldownRecovered
	_, _ = rotator.GetNextKey()
	if recoveredKey != "key-alpha" {
		t.Fatalf("OnCooldownRecovered hook failed, got: %s", recoveredKey)
	}
}

