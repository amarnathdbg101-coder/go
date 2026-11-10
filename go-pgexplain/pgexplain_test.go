package pgexplain_test

import (
	"context"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/amarnathdbg101-coder/go/go-pgexplain"
)

const sampleSeqScanJSON = `[
  {
    "Plan": {
      "Node Type": "Seq Scan",
      "Relation Name": "orders",
      "Total Cost": 520.50,
      "Plan Rows": 15000,
      "Actual Rows": 14200,
      "Actual Total Time": 125.4,
      "Shared Hit Blocks": 80,
      "Shared Read Blocks": 450,
      "Filter": "(customer_id = 99)"
    },
    "Planning Time": 0.25,
    "Execution Time": 125.8
  }
]`

const sampleIndexScanJSON = `[
  {
    "Plan": {
      "Node Type": "Index Scan",
      "Relation Name": "users",
      "Index Name": "idx_users_email",
      "Total Cost": 8.25,
      "Plan Rows": 1,
      "Actual Rows": 1,
      "Actual Total Time": 0.12,
      "Shared Hit Blocks": 4,
      "Shared Read Blocks": 0
    },
    "Planning Time": 0.10,
    "Execution Time": 0.15
  }
]`

func TestParsePlanJSONWithSeqScan(t *testing.T) {
	rep, err := pgexplain.ParsePlanJSON([]byte(sampleSeqScanJSON), "SELECT * FROM orders WHERE customer_id = 99")
	if err != nil {
		t.Fatalf("failed to parse plan: %v", err)
	}

	if len(rep.SeqScans) != 1 {
		t.Fatalf("expected 1 seq scan, got %d", len(rep.SeqScans))
	}
	if rep.SeqScans[0].Table != "orders" {
		t.Errorf("expected table 'orders', got %s", rep.SeqScans[0].Table)
	}
	if !rep.MissingIndexAlert {
		t.Errorf("expected MissingIndexAlert to be true for 14200 row sequential scan")
	}
	if rep.SharedReadBlocks != 450 {
		t.Errorf("expected 450 shared read blocks, got %d", rep.SharedReadBlocks)
	}
	if !strings.Contains(rep.Summary, "CRITICAL: Sequential Scan") {
		t.Errorf("expected critical warning in summary, got: %s", rep.Summary)
	}
}

func TestParsePlanJSONWithIndexScan(t *testing.T) {
	rep, err := pgexplain.ParsePlanJSON([]byte(sampleIndexScanJSON), "SELECT * FROM users WHERE email = 'a@b.com'")
	if err != nil {
		t.Fatalf("failed to parse index plan: %v", err)
	}

	if len(rep.IndexScans) != 1 {
		t.Fatalf("expected 1 index scan, got %d", len(rep.IndexScans))
	}
	if rep.MissingIndexAlert {
		t.Errorf("should not alert missing index on indexed lookup")
	}
	if !strings.Contains(rep.Summary, "OPTIMAL") {
		t.Errorf("expected optimal status in summary, got: %s", rep.Summary)
	}
}

func TestExplainerAudit(t *testing.T) {
	var alertCount int32
	explainer := pgexplain.New(
		pgexplain.WithSlowThreshold(50*time.Millisecond),
		pgexplain.WithAsyncExplain(false),
		pgexplain.WithOnSlowQuery(func(report *pgexplain.Report) {
			atomic.AddInt32(&alertCount, 1)
		}),
	)
	defer explainer.Close()

	// Mock plan executor
	explainer.SetExecutor(func(ctx context.Context, query string, args ...any) ([]byte, error) {
		return []byte(sampleSeqScanJSON), nil
	})

	ctx := context.Background()

	// 1. Fast query (10ms) -> should NOT trigger audit
	repFast, _ := explainer.Audit(ctx, "SELECT 1", 10*time.Millisecond)
	if repFast != nil {
		t.Errorf("fast query should not produce an audit report")
	}

	// 2. Slow query (150ms) -> should trigger audit
	repSlow, err := explainer.Audit(ctx, "SELECT * FROM orders WHERE customer_id = 99", 150*time.Millisecond)
	if err != nil || repSlow == nil {
		t.Fatalf("slow query failed to produce report: %v", err)
	}

	if atomic.LoadInt32(&alertCount) != 1 {
		t.Errorf("expected alert callback count 1, got %d", atomic.LoadInt32(&alertCount))
	}
}

