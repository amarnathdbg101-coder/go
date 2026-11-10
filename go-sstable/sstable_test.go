package sstable_test

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/amarnathdbg101-coder/go/go-sstable"
)

func TestSSTable_BasicWriteAndRead(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "sstable_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	filePath := filepath.Join(tempDir, "00001.sst")

	w, err := sstable.NewWriter(filePath, sstable.WithBlockSize(128))
	if err != nil {
		t.Fatalf("failed to create writer: %v", err)
	}

	keys := []string{
		"apple",
		"banana",
		"cherry",
		"date",
		"elderberry",
		"fig",
		"grape",
	}

	for _, k := range keys {
		val := fmt.Sprintf("val_%s", k)
		if err := w.Append([]byte(k), []byte(val)); err != nil {
			t.Fatalf("failed to append %s: %v", k, err)
		}
	}

	if err := w.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	// Open Reader
	r, err := sstable.Open(filePath)
	if err != nil {
		t.Fatalf("failed to open sstable reader: %v", err)
	}
	defer r.Close()

	if r.Count() != uint64(len(keys)) {
		t.Fatalf("expected %d keys, got %d", len(keys), r.Count())
	}

	// Test point lookups
	for _, k := range keys {
		val, err := r.Get([]byte(k))
		if err != nil {
			t.Fatalf("failed to get %s: %v", k, err)
		}
		expected := fmt.Sprintf("val_%s", k)
		if string(val) != expected {
			t.Fatalf("value mismatch for %s: expected %s, got %s", k, expected, string(val))
		}
	}

	// Test missing key
	_, err = r.Get([]byte("mango"))
	if err != sstable.ErrKeyNotFound {
		t.Fatalf("expected ErrKeyNotFound for mango, got %v", err)
	}
}

func TestSSTable_UnsortedKeysRejection(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "sstable_unsorted_*")
	defer os.RemoveAll(tempDir)

	filePath := filepath.Join(tempDir, "unsorted.sst")
	w, _ := sstable.NewWriter(filePath)

	_ = w.Append([]byte("zebra"), []byte("1"))
	err := w.Append([]byte("apple"), []byte("2"))
	if err != sstable.ErrUnsortedKeys {
		t.Fatalf("expected ErrUnsortedKeys, got %v", err)
	}
	_ = w.Close()
}

func TestSSTable_MultiBlockScan(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "sstable_scan_*")
	defer os.RemoveAll(tempDir)

	filePath := filepath.Join(tempDir, "scan.sst")
	// 64-byte block size forces frequent block flushes
	w, err := sstable.NewWriter(filePath, sstable.WithBlockSize(64))
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("user_%04d", i)
		val := fmt.Sprintf("payload_%04d", i)
		_ = w.Append([]byte(key), []byte(val))
	}
	_ = w.Close()

	r, err := sstable.Open(filePath)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	// Range scan from user_0020 to user_0045 (26 keys)
	var scannedKeys []string
	err = r.Scan([]byte("user_0020"), []byte("user_0045"), func(k, v []byte) error {
		scannedKeys = append(scannedKeys, string(k))
		return nil
	})
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	if len(scannedKeys) != 26 {
		t.Fatalf("expected 26 scanned keys, got %d", len(scannedKeys))
	}
	if scannedKeys[0] != "user_0020" || scannedKeys[25] != "user_0045" {
		t.Fatalf("scan bounds mismatch: first=%s, last=%s", scannedKeys[0], scannedKeys[25])
	}
}

func TestSSTable_Concurrency(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "sstable_concur_*")
	defer os.RemoveAll(tempDir)

	filePath := filepath.Join(tempDir, "concur.sst")
	w, _ := sstable.NewWriter(filePath, sstable.WithBlockSize(128))

	for i := 0; i < 50; i++ {
		key := fmt.Sprintf("key_%03d", i)
		val := fmt.Sprintf("val_%03d", i)
		_ = w.Append([]byte(key), []byte(val))
	}
	_ = w.Close()

	r, _ := sstable.Open(filePath)
	defer r.Close()

	var wg sync.WaitGroup
	for worker := 0; worker < 10; worker++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				k := fmt.Sprintf("key_%03d", i)
				val, err := r.Get([]byte(k))
				if err != nil || string(val) != fmt.Sprintf("val_%03d", i) {
					t.Errorf("worker %d lookup error on %s", id, k)
				}
			}
		}(worker)
	}
	wg.Wait()
}

