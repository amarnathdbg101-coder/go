package pgexplain

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"
)

// ExplainFunc represents an execution hook that returns raw JSON EXPLAIN bytes for a query.
type ExplainFunc func(ctx context.Context, query string, args ...any) ([]byte, error)

// Explainer coordinates slow query interception and PostgreSQL EXPLAIN plan auditing.
type Explainer struct {
	cfg      Config
	db       *sql.DB
	executor ExplainFunc
	mu       sync.Mutex
	closed   bool
}

// New creates an Explainer instance with optional SQL DB or custom plan executor.
func New(opts ...Option) *Explainer {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	return &Explainer{
		cfg: cfg,
	}
}

// SetDB connects the explainer directly to a live database/sql.DB instance.
func (e *Explainer) SetDB(db *sql.DB) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.db = db
	e.executor = func(ctx context.Context, query string, args ...any) ([]byte, error) {
		explainQuery := fmt.Sprintf("EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON) %s", query)
		var jsonRaw string
		err := db.QueryRowContext(ctx, explainQuery, args...).Scan(&jsonRaw)
		if err != nil {
			return nil, err
		}
		return []byte(jsonRaw), nil
	}
}

// SetExecutor sets a custom executor (ideal for mock testing and custom drivers).
func (e *Explainer) SetExecutor(fn ExplainFunc) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.executor = fn
}

// Audit evaluates the execution duration of a completed query.
// If duration exceeds SlowThreshold, it captures and parses the PostgreSQL execution plan.
func (e *Explainer) Audit(ctx context.Context, query string, duration time.Duration, args ...any) (*Report, error) {
	e.mu.Lock()
	if e.closed {
		e.mu.Unlock()
		return nil, ErrExplainerClosed
	}
	exec := e.executor
	e.mu.Unlock()

	if duration < e.cfg.SlowThreshold {
		return nil, nil
	}

	if exec == nil {
		return nil, nil
	}

	runAnalysis := func() (*Report, error) {
		explainCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		rawJSON, err := exec(explainCtx, query, args...)
		if err != nil {
			return nil, err
		}

		rep, err := ParsePlanJSON(rawJSON, query)
		if err != nil {
			return nil, err
		}

		if e.cfg.OnSlowQuery != nil {
			e.cfg.OnSlowQuery(rep)
		}
		return rep, nil
	}

	if e.cfg.AsyncExplain {
		go func() {
			_, _ = runAnalysis()
		}()
		return nil, nil
	}

	return runAnalysis()
}

// Close closes the explainer.
func (e *Explainer) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.closed = true
	return nil
}
