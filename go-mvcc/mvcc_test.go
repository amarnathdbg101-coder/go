package mvcc_test

import (
	"sync"
	"testing"

	"github.com/amarnathdbg101-coder/go/go-mvcc"
)

func TestMVCC_SnapshotIsolation(t *testing.T) {
	eng := mvcc.New()

	// Initial data setup
	tx0 := eng.Begin()
	_ = tx0.Set([]byte("balance"), []byte("1000"))
	if err := tx0.Commit(); err != nil {
		t.Fatalf("tx0 commit failed: %v", err)
	}

	// Tx1 starts snapshot
	tx1 := eng.Begin()

	// Tx2 starts, modifies balance, and commits
	tx2 := eng.Begin()
	_ = tx2.Set([]byte("balance"), []byte("2500"))
	if err := tx2.Commit(); err != nil {
		t.Fatalf("tx2 commit failed: %v", err)
	}

	// Tx1 MUST still see original balance (1000) under Snapshot Isolation!
	val1, err := tx1.Get([]byte("balance"))
	if err != nil {
		t.Fatalf("tx1 get failed: %v", err)
	}
	if string(val1) != "1000" {
		t.Fatalf("snapshot isolation violated: expected 1000, got %s", string(val1))
	}
	_ = tx1.Commit()

	// Tx3 starts after Tx2 committed, sees new balance (2500)
	tx3 := eng.Begin()
	val3, err := tx3.Get([]byte("balance"))
	if err != nil || string(val3) != "2500" {
		t.Fatalf("tx3 expected 2500, got %s (err: %v)", string(val3), err)
	}
	_ = tx3.Commit()
}

func TestMVCC_WriteWriteConflict(t *testing.T) {
	eng := mvcc.New()

	// Initial key
	tx0 := eng.Begin()
	_ = tx0.Set([]byte("stock"), []byte("50"))
	_ = tx0.Commit()

	// Concurrent TxA and TxB start simultaneously
	txA := eng.Begin()
	txB := eng.Begin()

	_ = txA.Set([]byte("stock"), []byte("40"))
	_ = txB.Set([]byte("stock"), []byte("30"))

	// TxA commits first
	if err := txA.Commit(); err != nil {
		t.Fatalf("txA commit failed: %v", err)
	}

	// TxB commits second -> MUST fail with ErrWriteConflict (First-Committer-Wins)
	err := txB.Commit()
	if err != mvcc.ErrWriteConflict {
		t.Fatalf("expected ErrWriteConflict for TxB, got: %v", err)
	}
}

func TestMVCC_Rollback(t *testing.T) {
	eng := mvcc.New()

	tx1 := eng.Begin()
	_ = tx1.Set([]byte("order:101"), []byte("pending"))
	_ = tx1.Rollback()

	tx2 := eng.Begin()
	_, err := tx2.Get([]byte("order:101"))
	if err != mvcc.ErrKeyNotFound {
		t.Fatalf("rolled back key should not exist, got err: %v", err)
	}
	_ = tx2.Commit()
}

func TestMVCC_DeleteAndSnapshot(t *testing.T) {
	eng := mvcc.New()

	tx0 := eng.Begin()
	_ = tx0.Set([]byte("profile"), []byte("v1"))
	_ = tx0.Commit()

	// Tx1 snapshot taken BEFORE deletion
	tx1 := eng.Begin()

	// Tx2 deletes key
	tx2 := eng.Begin()
	_ = tx2.Delete([]byte("profile"))
	_ = tx2.Commit()

	// Tx1 can still read key
	val, err := tx1.Get([]byte("profile"))
	if err != nil || string(val) != "v1" {
		t.Fatalf("tx1 should see pre-deletion version, got: %s (err: %v)", string(val), err)
	}
	_ = tx1.Commit()

	// Tx3 started after delete should get ErrKeyNotFound
	tx3 := eng.Begin()
	_, err = tx3.Get([]byte("profile"))
	if err != mvcc.ErrKeyNotFound {
		t.Fatalf("expected ErrKeyNotFound for tx3, got: %v", err)
	}
	_ = tx3.Commit()
}

func TestMVCC_Vacuum(t *testing.T) {
	eng := mvcc.New()

	// Create 10 versions of same key
	for i := 0; i < 10; i++ {
		tx := eng.Begin()
		_ = tx.Set([]byte("counter"), []byte("val"))
		_ = tx.Commit()
	}

	// Vacuum with no active transactions should prune 9 superseded versions
	pruned := eng.Vacuum()
	if pruned != 9 {
		t.Fatalf("expected 9 versions pruned by vacuum, got %d", pruned)
	}

	// Active version remains accessible
	txVerify := eng.Begin()
	val, err := txVerify.Get([]byte("counter"))
	if err != nil || string(val) != "val" {
		t.Fatalf("expected key intact after vacuum, got: %s", string(val))
	}
	_ = txVerify.Commit()
}

func TestMVCC_Concurrency(t *testing.T) {
	eng := mvcc.New()
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				tx := eng.Begin()
				_ = tx.Set([]byte("concur_key"), []byte("worker_val"))
				_ = tx.Commit() // may succeed or conflict, thread safety must hold
			}
		}(i)
	}

	wg.Wait()
}

