package mvcc

import (
	"sync"
	"sync/atomic"
)

// Version represents a point-in-time value revision of a key.
type Version struct {
	TxID        uint64
	DeletedTxID uint64
	Value       []byte
	Committed   bool
}

// Engine coordinates multi-version storage, snapshot isolation, and vacuuming.
type Engine struct {
	mu          sync.RWMutex
	cfg         Config
	lastTxID    uint64
	activeTxs   map[uint64]bool
	store       map[string][]*Version
	commitCount int
}

// New creates an initialized MVCC engine.
func New(opts ...Option) *Engine {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	return &Engine{
		cfg:       cfg,
		activeTxs: make(map[uint64]bool),
		store:     make(map[string][]*Version),
	}
}

// Begin starts a new transaction with a point-in-time snapshot of the database.
func (e *Engine) Begin() *Transaction {
	e.mu.Lock()
	defer e.mu.Unlock()

	txID := atomic.AddUint64(&e.lastTxID, 1)
	e.activeTxs[txID] = true

	// Capture currently active transactions at snapshot time
	snapshot := make(map[uint64]bool, len(e.activeTxs))
	for id := range e.activeTxs {
		if id != txID {
			snapshot[id] = true
		}
	}

	return &Transaction{
		id:             txID,
		engine:         e,
		activeSnapshot: snapshot,
		writes:         make(map[string][]byte),
		deletes:        make(map[string]bool),
	}
}

// Transaction represents an isolated atomic unit of read/write operations.
type Transaction struct {
	id             uint64
	engine         *Engine
	activeSnapshot map[uint64]bool
	writes         map[string][]byte
	deletes        map[string]bool
	closed         bool
}

// ID returns the unique transaction sequence number.
func (tx *Transaction) ID() uint64 {
	return tx.id
}

// Get reads the latest visible committed version of key within the transaction snapshot.
func (tx *Transaction) Get(key []byte) ([]byte, error) {
	if len(key) == 0 {
		return nil, ErrEmptyKey
	}
	if tx.closed {
		return nil, ErrTxClosed
	}

	kStr := string(key)

	// 1. Check uncommitted local writes within current transaction
	if tx.deletes[kStr] {
		return nil, ErrKeyNotFound
	}
	if val, ok := tx.writes[kStr]; ok {
		cp := make([]byte, len(val))
		copy(cp, val)
		return cp, nil
	}

	// 2. Scan version chain in storage under read lock
	tx.engine.mu.RLock()
	defer tx.engine.mu.RUnlock()

	chain, exists := tx.engine.store[kStr]
	if !exists || len(chain) == 0 {
		return nil, ErrKeyNotFound
	}

	for _, v := range chain {
		if !v.Committed {
			continue // Uncommitted changes from other transactions are invisible
		}
		if v.TxID > tx.id {
			continue // Created after our transaction started
		}
		if tx.activeSnapshot[v.TxID] {
			continue // Was active/uncommitted when our transaction started
		}

		// Check if deleted before or during our snapshot
		if v.DeletedTxID != 0 {
			if v.DeletedTxID <= tx.id && !tx.activeSnapshot[v.DeletedTxID] {
				return nil, ErrKeyNotFound // Key was deleted before our snapshot
			}
		}

		cp := make([]byte, len(v.Value))
		copy(cp, v.Value)
		return cp, nil
	}

	return nil, ErrKeyNotFound
}

// Set stages a key-value write in the transaction.
func (tx *Transaction) Set(key, val []byte) error {
	if len(key) == 0 {
		return ErrEmptyKey
	}
	if tx.closed {
		return ErrTxClosed
	}

	kStr := string(key)
	delete(tx.deletes, kStr)
	valCopy := make([]byte, len(val))
	copy(valCopy, val)
	tx.writes[kStr] = valCopy
	return nil
}

// Delete stages a deletion tombstone for key in the transaction.
func (tx *Transaction) Delete(key []byte) error {
	if len(key) == 0 {
		return ErrEmptyKey
	}
	if tx.closed {
		return ErrTxClosed
	}

	kStr := string(key)
	delete(tx.writes, kStr)
	tx.deletes[kStr] = true
	return nil
}

// Commit atomically publishes all staged writes and tombstones. Returns ErrWriteConflict on conflict.
func (tx *Transaction) Commit() error {
	if tx.closed {
		return ErrTxClosed
	}

	tx.engine.mu.Lock()
	defer tx.engine.mu.Unlock()

	// 1. Write-Write Conflict Detection (First-Committer-Wins)
	for kStr := range tx.writes {
		if err := tx.checkConflict(kStr); err != nil {
			tx.finish()
			return err
		}
	}
	for kStr := range tx.deletes {
		if err := tx.checkConflict(kStr); err != nil {
			tx.finish()
			return err
		}
	}

	// 2. Apply writes
	for kStr, val := range tx.writes {
		newVer := &Version{
			TxID:      tx.id,
			Value:     val,
			Committed: true,
		}
		// Prepend to version chain (newest first)
		tx.engine.store[kStr] = append([]*Version{newVer}, tx.engine.store[kStr]...)
	}

	// 3. Apply deletes
	for kStr := range tx.deletes {
		if chain, exists := tx.engine.store[kStr]; exists && len(chain) > 0 {
			// Mark latest visible version as deleted by this transaction
			chain[0].DeletedTxID = tx.id
		}
	}

	tx.finish()
	tx.engine.commitCount++
	if tx.engine.cfg.AutoVacuumThreshold > 0 && tx.engine.commitCount >= tx.engine.cfg.AutoVacuumThreshold {
		tx.engine.vacuumInternal()
		tx.engine.commitCount = 0
	}

	return nil
}

func (tx *Transaction) checkConflict(kStr string) error {
	chain, exists := tx.engine.store[kStr]
	if !exists || len(chain) == 0 {
		return nil
	}

	latest := chain[0]
	// If latest version was committed after our snapshot started, or was concurrent in our active snapshot: conflict!
	if latest.Committed && (latest.TxID > tx.id || tx.activeSnapshot[latest.TxID]) {
		return ErrWriteConflict
	}
	// If latest version is currently being modified by another active transaction: conflict!
	if !latest.Committed && latest.TxID != tx.id {
		return ErrWriteConflict
	}
	return nil
}

// Rollback cancels all staged writes.
func (tx *Transaction) Rollback() error {
	if tx.closed {
		return ErrTxClosed
	}
	tx.engine.mu.Lock()
	defer tx.engine.mu.Unlock()

	tx.finish()
	return nil
}

func (tx *Transaction) finish() {
	tx.closed = true
	delete(tx.engine.activeTxs, tx.id)
}

// Vacuum cleans up obsolete versions that are older than any active transaction.
func (e *Engine) Vacuum() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.vacuumInternal()
}

func (e *Engine) vacuumInternal() int {
	// Find oldest active transaction ID
	minActive := e.lastTxID + 1
	for id := range e.activeTxs {
		if id < minActive {
			minActive = id
		}
	}

	pruned := 0
	for kStr, chain := range e.store {
		var newChain []*Version
		foundCommitted := false

		for _, v := range chain {
			if v.Committed && v.TxID < minActive {
				if v.DeletedTxID != 0 && v.DeletedTxID < minActive {
					pruned++
					continue // Both created and deleted before minActive -> dead tuple
				}
				if foundCommitted {
					pruned++
					continue // Superseded by a newer committed version before minActive -> dead tuple
				}
				foundCommitted = true
			}
			newChain = append(newChain, v)
		}

		if len(newChain) == 0 {
			delete(e.store, kStr)
		} else {
			e.store[kStr] = newChain
		}
	}
	return pruned
}
