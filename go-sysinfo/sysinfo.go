package sysinfo

import (
	"fmt"
	"os"
	"runtime"
	"time"
)

var processStartTime = time.Now()

// MemorySnapshot holds detailed Go runtime memory allocator telemetry.
type MemorySnapshot struct {
	Alloc         uint64 `json:"alloc_bytes"`
	TotalAlloc    uint64 `json:"total_alloc_bytes"`
	Sys           uint64 `json:"sys_bytes"`
	HeapInuse     uint64 `json:"heap_inuse_bytes"`
	HeapIdle      uint64 `json:"heap_idle_bytes"`
	HeapReleased  uint64 `json:"heap_released_bytes"`
	NumGC         uint32 `json:"num_gc"`
	LastGCPauseNs uint64 `json:"last_gc_pause_ns"`
}

// Snapshot contains a full point-in-time diagnostics view of the process and host.
type Snapshot struct {
	Hostname      string         `json:"hostname"`
	PID           int            `json:"pid"`
	OS            string         `json:"os"`
	Arch          string         `json:"arch"`
	NumCPU        int            `json:"num_cpu"`
	GoVersion     string         `json:"go_version"`
	Goroutines    int            `json:"goroutines"`
	UptimeSeconds float64        `json:"uptime_seconds"`
	Memory        MemorySnapshot `json:"memory"`
	Timestamp     time.Time      `json:"timestamp"`
}

// Collect takes an instantaneous telemetry snapshot of the host, process, and Go runtime.
func Collect() Snapshot {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}

	var lastPause uint64
	if m.NumGC > 0 {
		lastPause = m.PauseNs[(m.NumGC+255)%256]
	}

	return Snapshot{
		Hostname:      hostname,
		PID:           os.Getpid(),
		OS:            runtime.GOOS,
		Arch:          runtime.GOARCH,
		NumCPU:        runtime.NumCPU(),
		GoVersion:     runtime.Version(),
		Goroutines:    runtime.NumGoroutine(),
		UptimeSeconds: time.Since(processStartTime).Seconds(),
		Memory: MemorySnapshot{
			Alloc:         m.Alloc,
			TotalAlloc:    m.TotalAlloc,
			Sys:           m.Sys,
			HeapInuse:     m.HeapInuse,
			HeapIdle:      m.HeapIdle,
			HeapReleased:  m.HeapReleased,
			NumGC:         m.NumGC,
			LastGCPauseNs: lastPause,
		},
		Timestamp: time.Now().UTC(),
	}
}

// FormatBytes formats byte counts into human-readable SI/IEC strings (e.g. "45.2 MB").
func FormatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	prefix := "KMGTPE"[exp]
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), prefix)
}

// HumanMemory returns a map of memory statistics formatted into human-readable strings.
func (s Snapshot) HumanMemory() map[string]string {
	return map[string]string{
		"alloc":         FormatBytes(s.Memory.Alloc),
		"total_alloc":   FormatBytes(s.Memory.TotalAlloc),
		"sys":           FormatBytes(s.Memory.Sys),
		"heap_inuse":    FormatBytes(s.Memory.HeapInuse),
		"heap_idle":     FormatBytes(s.Memory.HeapIdle),
		"heap_released": FormatBytes(s.Memory.HeapReleased),
	}
}

// ToMap converts the snapshot into a map suitable for JSON API responses.
func (s Snapshot) ToMap() map[string]any {
	return map[string]any{
		"hostname":       s.Hostname,
		"pid":            s.PID,
		"os":             s.OS,
		"arch":           s.Arch,
		"num_cpu":        s.NumCPU,
		"go_version":     s.GoVersion,
		"goroutines":     s.Goroutines,
		"uptime_seconds": s.UptimeSeconds,
		"memory":         s.HumanMemory(),
		"timestamp":      s.Timestamp.Format(time.RFC3339),
	}
}
