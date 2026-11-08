package sysinfo

import (
	"strings"
	"testing"
)

func TestCollect(t *testing.T) {
	snap := Collect()

	if snap.PID <= 0 {
		t.Errorf("expected positive PID, got %d", snap.PID)
	}
	if snap.NumCPU <= 0 {
		t.Errorf("expected positive NumCPU, got %d", snap.NumCPU)
	}
	if snap.Goroutines <= 0 {
		t.Errorf("expected at least 1 goroutine, got %d", snap.Goroutines)
	}
	if snap.OS == "" || snap.Arch == "" {
		t.Errorf("OS or Arch is empty: os=%s, arch=%s", snap.OS, snap.Arch)
	}
	if snap.Memory.Alloc == 0 || snap.Memory.Sys == 0 {
		t.Errorf("expected non-zero memory stats: %+v", snap.Memory)
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    uint64
		expected string
	}{
		{500, "500 B"},
		{1024, "1.0 KiB"},
		{1024 * 1024 * 5, "5.0 MiB"},
		{1024 * 1024 * 1024 * 2, "2.0 GiB"},
	}

	for _, tt := range tests {
		res := FormatBytes(tt.bytes)
		if res != tt.expected {
			t.Errorf("FormatBytes(%d) = %s; expected %s", tt.bytes, res, tt.expected)
		}
	}
}

func TestHumanMemoryAndToMap(t *testing.T) {
	snap := Collect()

	human := snap.HumanMemory()
	if !strings.Contains(human["alloc"], "B") {
		t.Errorf("expected human memory to contain B unit, got %s", human["alloc"])
	}

	m := snap.ToMap()
	if m["pid"].(int) != snap.PID {
		t.Errorf("expected map pid %d, got %v", snap.PID, m["pid"])
	}
	if m["os"].(string) != snap.OS {
		t.Errorf("expected map os %s, got %v", snap.OS, m["os"])
	}
}
