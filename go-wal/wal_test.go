package wal_test

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/amarnathdbg101-coder/go/go-wal"
)

func TestWAL_AppendAndReplay(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "wal_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	w, err := wal.Open(wal.WithDir(tempDir))
	if err != nil {
		t.Fatalf("failed to open WAL: %v", err)
	}

	messages := [][]byte{
		[]byte("order_created:id=101"),
		[]byte("payment_captured:order=101:amount=499"),
		[]byte("order_fulfilled:id=101"),
	}

	for _, msg := range messages {
		lsn, err := w.Append(msg)
		if err != nil {
			t.Fatalf("append failed: %v", err)
		}
		if lsn == 0 {
			t.Fatal("expected lsn > 0")
		}
	}

	if w.LastLSN() != 3 {
		t.Fatalf("expected last LSN 3, got %d", w.LastLSN())
	}

	// Replay from LSN 1
	var replayed [][]byte
	err = w.Replay(1, func(rec wal.Record) error {
		replayed = append(replayed, rec.Data)
		return nil
	})
	if err != nil {
		t.Fatalf("replay failed: %v", err)
	}

	if len(replayed) != len(messages) {
		t.Fatalf("expected %d replayed records, got %d", len(messages), len(replayed))
	}

	for i, msg := range messages {
		if string(replayed[i]) != string(msg) {
			t.Fatalf("record %d mismatch: expected %s, got %s", i, string(msg), string(replayed[i]))
		}
	}

	_ = w.Close()
}

func TestWAL_SegmentRollingAndRecovery(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "wal_roll_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// MaxSegmentSize 60 bytes forces segment roll every 1-2 records
	w, err := wal.Open(
		wal.WithDir(tempDir),
		wal.WithMaxSegmentSize(60),
	)
	if err != nil {
		t.Fatalf("failed to open WAL: %v", err)
	}

	for i := 0; i < 5; i++ {
		_, err := w.Append([]byte("sample_payload_data"))
		if err != nil {
			t.Fatalf("append failed at %d: %v", i, err)
		}
	}

	_ = w.Close()

	// Verify multiple segment files exist
	files, err := filepath.Glob(filepath.Join(tempDir, "*.wal"))
	if err != nil || len(files) < 2 {
		t.Fatalf("expected at least 2 segment files, got %d", len(files))
	}

	// Re-open existing directory to simulate recovery after crash
	w2, err := wal.Open(
		wal.WithDir(tempDir),
		wal.WithMaxSegmentSize(60),
	)
	if err != nil {
		t.Fatalf("failed to recover WAL: %v", err)
	}
	defer w2.Close()

	if w2.LastLSN() != 5 {
		t.Fatalf("expected recovered last LSN 5, got %d", w2.LastLSN())
	}

	// Replay must retrieve all 5 records
	count := 0
	err = w2.Replay(1, func(rec wal.Record) error {
		count++
		return nil
	})
	if err != nil || count != 5 {
		t.Fatalf("expected 5 records after recovery, got %d (err: %v)", count, err)
	}
}

func TestWAL_CorruptionDetection(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "wal_corrupt_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	w, err := wal.Open(wal.WithDir(tempDir))
	if err != nil {
		t.Fatalf("failed to open: %v", err)
	}

	_, _ = w.Append([]byte("valid_record_payload"))
	_ = w.Close()

	// Tamper with segment file
	files, _ := filepath.Glob(filepath.Join(tempDir, "*.wal"))
	f, err := os.OpenFile(files[0], os.O_RDWR, 0644)
	if err != nil {
		t.Fatal(err)
	}
	// Overwrite payload byte
	_, _ = f.WriteAt([]byte("X"), 25)
	_ = f.Close()

	w2, err := wal.Open(wal.WithDir(tempDir))
	if err != nil {
		t.Fatalf("open failed: %v", err)
	}
	defer w2.Close()

	err = w2.Replay(1, func(rec wal.Record) error {
		return nil
	})
	if err != wal.ErrCorruptRecord {
		t.Fatalf("expected ErrCorruptRecord on tampered file, got %v", err)
	}
}

func TestWAL_Concurrency(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "wal_concurrent_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	w, err := wal.Open(wal.WithDir(tempDir))
	if err != nil {
		t.Fatalf("failed to open: %v", err)
	}
	defer w.Close()

	var wg sync.WaitGroup
	workers := 10
	perWorker := 20

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < perWorker; j++ {
				_, err := w.Append([]byte("concurrent_data"))
				if err != nil {
					t.Errorf("concurrent append failed: %v", err)
				}
			}
		}(i)
	}

	wg.Wait()

	expectedTotal := uint64(workers * perWorker)
	if w.LastLSN() != expectedTotal {
		t.Fatalf("expected last LSN %d, got %d", expectedTotal, w.LastLSN())
	}
}

