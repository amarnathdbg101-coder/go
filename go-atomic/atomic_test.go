package atomic

import (
	"encoding/json"
	"math"
	"sync"
	"testing"
	"time"
)

func TestBool(t *testing.T) {
	b := NewBool(false)
	if b.Load() {
		t.Errorf("expected false")
	}

	b.Store(true)
	if !b.Load() {
		t.Errorf("expected true")
	}

	old := b.Swap(false)
	if !old || b.Load() {
		t.Errorf("expected swap old=true, curr=false")
	}

	if !b.CAS(false, true) || !b.Load() {
		t.Errorf("expected CAS success")
	}

	// Toggle
	newVal := b.Toggle()
	if newVal != false || b.Load() != false {
		t.Errorf("expected toggle to false")
	}

	// JSON
	data, _ := json.Marshal(b)
	var restored Bool
	json.Unmarshal(data, &restored)
	if restored.Load() != false {
		t.Errorf("JSON restore mismatch")
	}
}

func TestString(t *testing.T) {
	s := NewString("initial")
	if s.Load() != "initial" {
		t.Errorf("expected initial")
	}

	s.Store("updated")
	if s.Load() != "updated" {
		t.Errorf("expected updated")
	}

	old := s.Swap("swapped")
	if old != "updated" || s.Load() != "swapped" {
		t.Errorf("swap mismatch")
	}
}

func TestDuration(t *testing.T) {
	d := NewDuration(5 * time.Second)
	if d.Load() != 5*time.Second {
		t.Errorf("expected 5s")
	}

	d.Store(10 * time.Second)
	if d.Load() != 10*time.Second {
		t.Errorf("expected 10s")
	}

	added := d.Add(2 * time.Second)
	if added != 12*time.Second || d.Load() != 12*time.Second {
		t.Errorf("expected 12s after add")
	}
}

func TestTime(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	tm := NewTime(now)
	if !tm.Load().Equal(now) {
		t.Errorf("time mismatch")
	}

	future := now.Add(1 * time.Hour)
	tm.Store(future)
	if !tm.Load().Equal(future) {
		t.Errorf("future time mismatch")
	}
}

func TestFloat64(t *testing.T) {
	f := NewFloat64(10.5)
	if f.Load() != 10.5 {
		t.Errorf("expected 10.5")
	}

	newVal := f.Add(4.5)
	if newVal != 15.0 || f.Load() != 15.0 {
		t.Errorf("expected 15.0 after add")
	}

	if !f.CAS(15.0, 20.0) || f.Load() != 20.0 {
		t.Errorf("expected CAS success to 20.0")
	}
}

func TestValueGeneric(t *testing.T) {
	type Config struct {
		Host string
		Port int
	}

	cfgVal := NewValue(Config{Host: "localhost", Port: 8080})
	curr := cfgVal.Load()
	if curr.Host != "localhost" || curr.Port != 8080 {
		t.Errorf("mismatch config: %+v", curr)
	}

	cfgVal.Store(Config{Host: "remote", Port: 9090})
	if cfgVal.Load().Host != "remote" {
		t.Errorf("store mismatch")
	}
}

func TestConcurrency(t *testing.T) {
	b := NewBool(false)
	f := NewFloat64(0.0)
	d := NewDuration(0)

	var wg sync.WaitGroup
	const routines = 40
	const iterations = 50

	for i := 0; i < routines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				b.Toggle()
				f.Add(1.0)
				d.Add(1 * time.Millisecond)
			}
		}()
	}

	wg.Wait()

	expectedFloat := float64(routines * iterations)
	if math.Abs(f.Load()-expectedFloat) > 1e-9 {
		t.Errorf("concurrency float mismatch: expected %f, got %f", expectedFloat, f.Load())
	}
	expectedDur := time.Duration(routines*iterations) * time.Millisecond
	if d.Load() != expectedDur {
		t.Errorf("concurrency duration mismatch: expected %v, got %v", expectedDur, d.Load())
	}
}
