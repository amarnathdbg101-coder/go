package syncchan

import (
	"context"
	"sort"
	"testing"
	"time"
)

func generate(ctx context.Context, start, count int) <-chan int {
	ch := make(chan int)
	go func() {
		defer close(ch)
		for i := 0; i < count; i++ {
			select {
			case <-ctx.Done():
				return
			case ch <- start + i:
			}
		}
	}()
	return ch
}

func TestFanIn(t *testing.T) {
	ctx := context.Background()
	ch1 := generate(ctx, 1, 3)  // 1, 2, 3
	ch2 := generate(ctx, 10, 3) // 10, 11, 12
	ch3 := generate(ctx, 20, 2) // 20, 21

	merged := FanIn(ctx, ch1, ch2, ch3)
	vals := Collect(merged)

	if len(vals) != 8 {
		t.Fatalf("expected 8 merged values, got %d", len(vals))
	}

	sort.Ints(vals)
	expected := []int{1, 2, 3, 10, 11, 12, 20, 21}
	for i, v := range vals {
		if v != expected[i] {
			t.Fatalf("at index %d: expected %d, got %d", i, expected[i], v)
		}
	}
}

func TestFanIn_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ch1 := make(chan int)

	merged := FanIn(ctx, ch1)
	cancel() // abort early

	select {
	case _, ok := <-merged:
		if ok {
			t.Fatalf("expected channel to close immediately on context cancel")
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("timeout waiting for FanIn to abort on context cancel")
	}
}

func TestTee(t *testing.T) {
	ctx := context.Background()
	in := generate(ctx, 1, 5)

	out1, out2 := Tee(ctx, in)

	var vals1, vals2 []int
	var done syncchanDone
	go func() {
		vals1 = Collect(out1)
		done.done1 = true
	}()
	go func() {
		vals2 = Collect(out2)
		done.done2 = true
	}()

	for !done.done1 || !done.done2 {
		time.Sleep(5 * time.Millisecond)
	}

	if len(vals1) != 5 || len(vals2) != 5 {
		t.Fatalf("expected both tee outputs to have 5 items, got %d and %d", len(vals1), len(vals2))
	}
}

type syncchanDone struct {
	done1 bool
	done2 bool
}

func TestTake(t *testing.T) {
	ctx := context.Background()
	in := generate(ctx, 1, 10)

	taken := Take(ctx, in, 4)
	vals := Collect(taken)

	if len(vals) != 4 {
		t.Fatalf("expected 4 items, got %d", len(vals))
	}
	expected := []int{1, 2, 3, 4}
	for i, v := range vals {
		if v != expected[i] {
			t.Fatalf("expected %d at index %d, got %d", expected[i], i, v)
		}
	}
}

func TestDrop(t *testing.T) {
	ctx := context.Background()
	in := generate(ctx, 1, 6) // 1, 2, 3, 4, 5, 6

	dropped := Drop(ctx, in, 3) // skip 1, 2, 3 -> keep 4, 5, 6
	vals := Collect(dropped)

	if len(vals) != 3 {
		t.Fatalf("expected 3 items, got %d", len(vals))
	}
	expected := []int{4, 5, 6}
	for i, v := range vals {
		if v != expected[i] {
			t.Fatalf("expected %d at index %d, got %d", expected[i], i, v)
		}
	}
}

func TestFanOut(t *testing.T) {
	ctx := context.Background()
	in := generate(ctx, 100, 3)

	subs := FanOut(ctx, in, 3)
	if len(subs) != 3 {
		t.Fatalf("expected 3 subscribers")
	}

	var res0, res1, res2 []int
	var wg syncchanDone
	go func() { res0 = Collect(subs[0]); wg.done1 = true }()
	go func() { res1 = Collect(subs[1]); wg.done2 = true }()
	res2 = Collect(subs[2])

	for !wg.done1 || !wg.done2 {
		time.Sleep(5 * time.Millisecond)
	}

	if len(res0) != 3 || len(res1) != 3 || len(res2) != 3 {
		t.Fatalf("fanout mismatch across subscribers")
	}
}

func TestBuffer(t *testing.T) {
	ctx := context.Background()
	in := generate(ctx, 1, 5)

	buffered := Buffer(ctx, in, 10)
	vals := Collect(buffered)

	if len(vals) != 5 {
		t.Fatalf("expected 5 items from buffered chan, got %d", len(vals))
	}
}
