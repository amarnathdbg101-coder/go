package lsmtree_test

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/amarnathdbg101-coder/go/go-lsmtree"
)

func TestLSM_BasicPutGetDelete(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "lsm_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	db, err := lsmtree.Open(lsmtree.WithDir(tempDir))
	if err != nil {
		t.Fatalf("failed to open LSM db: %v", err)
	}
	defer db.Close()

	// Put
	if err := db.Put([]byte("user:101"), []byte("Amar")); err != nil {
		t.Fatalf("put failed: %v", err)
	}

	// Get
	val, err := db.Get([]byte("user:101"))
	if err != nil || string(val) != "Amar" {
		t.Fatalf("get failed: expected Amar, got %s (err: %v)", string(val), err)
	}

	// Overwrite
	_ = db.Put([]byte("user:101"), []byte("Amar Deep"))
	val, _ = db.Get([]byte("user:101"))
	if string(val) != "Amar Deep" {
		t.Fatalf("expected updated value Amar Deep, got %s", string(val))
	}

	// Delete
	if err := db.Delete([]byte("user:101")); err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	_, err = db.Get([]byte("user:101"))
	if err != lsmtree.ErrKeyNotFound {
		t.Fatalf("expected ErrKeyNotFound after deletion, got %v", err)
	}
}

func TestLSM_FlushAndSSTableRead(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "lsm_flush_*")
	defer os.RemoveAll(tempDir)

	// Tiny MemTable capacity (100 bytes) forces frequent flushes
	db, err := lsmtree.Open(
		lsmtree.WithDir(tempDir),
		lsmtree.WithMemTableCapacity(100),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Insert 20 keys, should flush to multiple SSTables
	for i := 0; i < 20; i++ {
		k := fmt.Sprintf("k_%03d", i)
		v := fmt.Sprintf("val_data_%03d", i)
		_ = db.Put([]byte(k), []byte(v))
	}

	// Verify SSTable files were created on disk
	sstFiles, _ := filepath.Glob(filepath.Join(tempDir, "*.sst"))
	if len(sstFiles) < 2 {
		t.Fatalf("expected multiple SSTables on disk, got %d", len(sstFiles))
	}

	// Read all keys from SSTables
	for i := 0; i < 20; i++ {
		k := fmt.Sprintf("k_%03d", i)
		val, err := db.Get([]byte(k))
		if err != nil {
			t.Fatalf("failed to get %s from flushed SSTable: %v", k, err)
		}
		expected := fmt.Sprintf("val_data_%03d", i)
		if string(val) != expected {
			t.Fatalf("value mismatch: expected %s, got %s", expected, string(val))
		}
	}
	_ = db.Close()
}

func TestLSM_WALCrashRecovery(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "lsm_recovery_*")
	defer os.RemoveAll(tempDir)

	// Open DB, insert into MemTable, close without explicit Flush
	db, _ := lsmtree.Open(
		lsmtree.WithDir(tempDir),
		lsmtree.WithMemTableCapacity(1024*1024), // Large capacity, stays in MemTable + WAL
	)
	_ = db.Put([]byte("session:abc"), []byte("active_token"))
	_ = db.Close()

	// Reopen DB simulating recovery after restart
	dbRecovered, err := lsmtree.Open(lsmtree.WithDir(tempDir))
	if err != nil {
		t.Fatalf("recovery open failed: %v", err)
	}
	defer dbRecovered.Close()

	val, err := dbRecovered.Get([]byte("session:abc"))
	if err != nil || string(val) != "active_token" {
		t.Fatalf("WAL recovery failed: expected active_token, got %s (err: %v)", string(val), err)
	}
}

func TestLSM_Compaction(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "lsm_compact_*")
	defer os.RemoveAll(tempDir)

	db, _ := lsmtree.Open(
		lsmtree.WithDir(tempDir),
		lsmtree.WithMemTableCapacity(60),
	)

	// Create SSTable 1
	_ = db.Put([]byte("item:1"), []byte("v1"))
	_ = db.Put([]byte("item:2"), []byte("v1"))
	_ = db.Flush()

	// Create SSTable 2: update item:1 and delete item:2
	_ = db.Put([]byte("item:1"), []byte("v2"))
	_ = db.Delete([]byte("item:2"))
	_ = db.Flush()

	filesBefore, _ := filepath.Glob(filepath.Join(tempDir, "*.sst"))
	if len(filesBefore) != 2 {
		t.Fatalf("expected 2 SSTables before compaction, got %d", len(filesBefore))
	}

	// Compact
	if err := db.Compact(); err != nil {
		t.Fatalf("compaction failed: %v", err)
	}

	filesAfter, _ := filepath.Glob(filepath.Join(tempDir, "*.sst"))
	if len(filesAfter) != 1 {
		t.Fatalf("expected 1 merged SSTable after compaction, got %d", len(filesAfter))
	}

	// Verify item:1 has newest value v2
	val, err := db.Get([]byte("item:1"))
	if err != nil || string(val) != "v2" {
		t.Fatalf("compacted value mismatch: expected v2, got %s (err: %v)", string(val), err)
	}

	// Verify item:2 tombstone was honored and purged
	_, err = db.Get([]byte("item:2"))
	if err != lsmtree.ErrKeyNotFound {
		t.Fatalf("expected ErrKeyNotFound for deleted item:2, got %v", err)
	}
	_ = db.Close()
}

func TestLSM_Scan(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "lsm_scan_*")
	defer os.RemoveAll(tempDir)

	db, _ := lsmtree.Open(lsmtree.WithDir(tempDir))
	defer db.Close()

	for i := 0; i < 10; i++ {
		k := fmt.Sprintf("user_%02d", i)
		v := fmt.Sprintf("val_%02d", i)
		_ = db.Put([]byte(k), []byte(v))
	}

	// Range Scan [user_03 .. user_07] (5 keys)
	var keys []string
	err := db.Scan([]byte("user_03"), []byte("user_07"), func(k, v []byte) error {
		keys = append(keys, string(k))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(keys) != 5 {
		t.Fatalf("expected 5 keys, got %d", len(keys))
	}
	if keys[0] != "user_03" || keys[4] != "user_07" {
		t.Fatalf("unexpected range boundary: %v", keys)
	}
}

func TestLSM_Concurrency(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "lsm_concur_*")
	defer os.RemoveAll(tempDir)

	db, _ := lsmtree.Open(
		lsmtree.WithDir(tempDir),
		lsmtree.WithMemTableCapacity(200),
	)
	defer db.Close()

	var wg sync.WaitGroup
	for w := 0; w < 5; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				k := fmt.Sprintf("w_%d_k_%d", workerID, i)
				v := fmt.Sprintf("v_%d", i)
				_ = db.Put([]byte(k), []byte(v))
				_, _ = db.Get([]byte(k))
			}
		}(w)
	}
	wg.Wait()
}

