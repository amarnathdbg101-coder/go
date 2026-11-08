package gostopwatch

import (
	"strings"
	"testing"
	"time"
)

func TestStopwatch_LapsAndSplit(t *testing.T) {
	sw := Start("order_pipeline")

	time.Sleep(10 * time.Millisecond)
	d1 := sw.Lap("validate_cart")
	if d1 < 10*time.Millisecond {
		t.Errorf("lap 1 duration too short: %v", d1)
	}

	time.Sleep(20 * time.Millisecond)
	d2 := sw.Lap("charge_card")
	if d2 < 20*time.Millisecond {
		t.Errorf("lap 2 duration too short: %v", d2)
	}

	total := sw.Stop()
	if total < 30*time.Millisecond {
		t.Errorf("total duration too short: %v", total)
	}

	laps := sw.Laps()
	if len(laps) != 2 {
		t.Fatalf("expected 2 laps, got %d", len(laps))
	}

	if laps[0].Name != "validate_cart" || laps[1].Name != "charge_card" {
		t.Errorf("lap names mismatch: %v, %v", laps[0].Name, laps[1].Name)
	}

	// Split 2 should be greater than split 1
	if laps[1].Split <= laps[0].Split {
		t.Errorf("cumulative split order invalid: split1=%v, split2=%v", laps[0].Split, laps[1].Split)
	}
}

func TestStopwatch_SummaryAndToMap(t *testing.T) {
	sw := Start("test_summary")
	time.Sleep(5 * time.Millisecond)
	_ = sw.Lap("phase_a")
	time.Sleep(5 * time.Millisecond)
	_ = sw.Lap("phase_b")
	sw.Stop()

	// Summary format
	sum := sw.Summary()
	if !strings.Contains(sum, "=== Stopwatch: test_summary") || !strings.Contains(sum, "phase_a") || !strings.Contains(sum, "phase_b") {
		t.Errorf("summary output format mismatch:\n%s", sum)
	}

	// Map format
	m := sw.ToMap()
	if m["name"] != "test_summary" {
		t.Errorf("map name mismatch: %v", m["name"])
	}
	if m["laps_count"] != 2 {
		t.Errorf("map laps_count mismatch: %v", m["laps_count"])
	}
}

func TestStopwatch_Reset(t *testing.T) {
	sw := Start("test_reset")
	time.Sleep(5 * time.Millisecond)
	_ = sw.Lap("phase_1")
	sw.Stop()

	sw.Reset()
	if len(sw.Laps()) != 0 {
		t.Errorf("expected 0 laps after reset, got %d", len(sw.Laps()))
	}
	if sw.Elapsed() > 50*time.Millisecond {
		t.Errorf("elapsed should reset to near zero")
	}
}
