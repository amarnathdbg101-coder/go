package gostopwatch

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// Lap records timing metadata for an individual execution phase.
type Lap struct {
	Name     string        `json:"name"`
	Duration time.Duration `json:"duration"`
	Split    time.Duration `json:"split"` // Cumulative duration since start
}

// Stopwatch provides high-resolution stage profiling and multi-lap execution tracking.
type Stopwatch struct {
	name      string
	startTime time.Time
	stopTime  time.Time
	lastLap   time.Time
	laps      []Lap
	stopped   bool
	mu        sync.Mutex
}

// Start creates and initiates a new Stopwatch.
func Start(name string) *Stopwatch {
	now := time.Now()
	return &Stopwatch{
		name:      name,
		startTime: now,
		lastLap:   now,
		laps:      make([]Lap, 0, 8),
	}
}

// Lap marks the completion of a named stage and records its duration.
func (s *Stopwatch) Lap(name string) time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	dur := now.Sub(s.lastLap)
	split := now.Sub(s.startTime)

	s.laps = append(s.laps, Lap{
		Name:     name,
		Duration: dur,
		Split:    split,
	})

	s.lastLap = now
	return dur
}

// Stop finalizes the stopwatch.
func (s *Stopwatch) Stop() time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.stopped {
		s.stopTime = time.Now()
		s.stopped = true
	}

	return s.stopTime.Sub(s.startTime)
}

// Elapsed returns total time elapsed without necessarily stopping the stopwatch.
func (s *Stopwatch) Elapsed() time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.stopped {
		return s.stopTime.Sub(s.startTime)
	}
	return time.Since(s.startTime)
}

// Laps returns a copy of all recorded laps.
func (s *Stopwatch) Laps() []Lap {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := make([]Lap, len(s.laps))
	copy(result, s.laps)
	return result
}

// Reset clears recorded laps and restarts timer from current time.
func (s *Stopwatch) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	s.startTime = now
	s.lastLap = now
	s.stopTime = time.Time{}
	s.stopped = false
	s.laps = make([]Lap, 0, 8)
}

// Summary returns a formatted profiling breakdown table.
func (s *Stopwatch) Summary() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	total := time.Since(s.startTime)
	if s.stopped {
		total = s.stopTime.Sub(s.startTime)
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("=== Stopwatch: %s (Total: %v) ===\n", s.name, total))

	for i, lap := range s.laps {
		pct := 0.0
		if total > 0 {
			pct = (float64(lap.Duration) / float64(total)) * 100.0
		}
		b.WriteString(fmt.Sprintf("  [%d] %-20s : %10v (%5.1f%%) [split: %v]\n", i+1, lap.Name, lap.Duration, pct, lap.Split))
	}

	return b.String()
}

// ToMap serializes the profile into structured telemetry data.
func (s *Stopwatch) ToMap() map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()

	total := time.Since(s.startTime)
	if s.stopped {
		total = s.stopTime.Sub(s.startTime)
	}

	lapData := make([]map[string]any, len(s.laps))
	for i, l := range s.laps {
		lapData[i] = map[string]any{
			"name":        l.Name,
			"duration_ms": float64(l.Duration.Microseconds()) / 1000.0,
			"split_ms":    float64(l.Split.Microseconds()) / 1000.0,
		}
	}

	return map[string]any{
		"name":        s.name,
		"total_ms":    float64(total.Microseconds()) / 1000.0,
		"laps_count":  len(s.laps),
		"laps":        lapData,
	}
}
