package saferegex

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestSafeRegex_Basic(t *testing.T) {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	validEmail := "test.user@example.com"
	invalidEmail := "not-an-email"

	matched, err := MatchString(pattern, validEmail, 2*time.Second)
	if err != nil || !matched {
		t.Fatalf("expected valid email match: %v", err)
	}

	matched, err = MatchString(pattern, invalidEmail, 2*time.Second)
	if err != nil || matched {
		t.Fatalf("expected invalid email to not match: %v", err)
	}
}

func TestSafeRegex_FindAndReplace(t *testing.T) {
	pattern := `\d+`
	src := "Order 9821 with 3 items"

	found, err := FindString(pattern, src, 2*time.Second)
	if err != nil || found != "9821" {
		t.Fatalf("expected to find 9821, got '%s', err: %v", found, err)
	}

	replaced, err := ReplaceAllString(pattern, src, "NUM", 2*time.Second)
	if err != nil || replaced != "Order NUM with NUM items" {
		t.Fatalf("unexpected replace result: '%s'", replaced)
	}
}

func TestSafeRegex_InvalidPattern(t *testing.T) {
	_, err := MatchString("[a-z", "abc", 2*time.Second)
	if !errors.Is(err, ErrInvalidPattern) {
		t.Fatalf("expected ErrInvalidPattern, got: %v", err)
	}
}

func TestSafeRegex_LRUEviction(t *testing.T) {
	engine := New(Config{Capacity: 3, DefaultTimeout: 2 * time.Second})

	// Add 3 patterns
	_, _ = engine.getOrCompile(`a+`)
	_, _ = engine.getOrCompile(`b+`)
	_, _ = engine.getOrCompile(`c+`)

	if len(engine.cache) != 3 {
		t.Fatalf("expected 3 items in cache")
	}

	// Access a+ to move it to front
	_, _ = engine.getOrCompile(`a+`)

	// Add 4th pattern: should evict b+ (least recently used)
	_, _ = engine.getOrCompile(`d+`)

	if len(engine.cache) != 3 {
		t.Fatalf("expected cache size 3 after eviction")
	}

	if _, ok := engine.cache[`b+`]; ok {
		t.Fatalf("expected b+ to be evicted")
	}
	if _, ok := engine.cache[`a+`]; !ok {
		t.Fatalf("expected a+ to remain in cache")
	}
}

func TestSafeRegex_ContextCancellation(t *testing.T) {
	engine := New(DefaultConfig())
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // pre-canceled

	_, err := engine.MatchString(ctx, `\d+`, "12345", 100*time.Millisecond)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got: %v", err)
	}
}

func TestSafeRegex_Concurrent(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			pattern := `^user-\d+$`
			matched, err := MatchString(pattern, "user-1234", 2*time.Second)
			if err != nil || !matched {
				t.Errorf("match failed: %v", err)
			}
		}(i)
	}
	wg.Wait()
}
