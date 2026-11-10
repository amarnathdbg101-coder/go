package pgexplain

import (
	"encoding/json"
	"fmt"
)

// PlanNode represents a single operation in a recursive PostgreSQL execution plan.
type PlanNode struct {
	NodeType         string     `json:"Node Type"`
	RelationName     string     `json:"Relation Name,omitempty"`
	IndexName        string     `json:"Index Name,omitempty"`
	TotalCost        float64    `json:"Total Cost"`
	PlanRows         float64    `json:"Plan Rows"`
	ActualRows       float64    `json:"Actual Rows"`
	ActualTotalTime  float64    `json:"Actual Total Time"`
	SharedHitBlocks  int        `json:"Shared Hit Blocks"`
	SharedReadBlocks int        `json:"Shared Read Blocks"`
	Filter           string     `json:"Filter,omitempty"`
	Plans            []PlanNode `json:"Plans,omitempty"`
}

type explainRoot struct {
	Plan          PlanNode `json:"Plan"`
	PlanningTime  float64  `json:"Planning Time"`
	ExecutionTime float64  `json:"Execution Time"`
}

// SeqScanWarning details a detected sequential table scan that could benefit from an index.
type SeqScanWarning struct {
	Table      string  `json:"table"`
	Cost       float64 `json:"cost"`
	ActualRows float64 `json:"actual_rows"`
	Filter     string  `json:"filter,omitempty"`
}

// Report aggregates diagnostic findings from a PostgreSQL query execution plan.
type Report struct {
	Query             string           `json:"query"`
	ExecutionTimeMs   float64          `json:"execution_time_ms"`
	PlanningTimeMs    float64          `json:"planning_time_ms"`
	TotalCost         float64          `json:"total_cost"`
	SeqScans          []SeqScanWarning `json:"seq_scans"`
	IndexScans        []string         `json:"index_scans"`
	SharedHitBlocks   int              `json:"shared_hit_blocks"`
	SharedReadBlocks  int              `json:"shared_read_blocks"`
	MissingIndexAlert bool             `json:"missing_index_alert"`
	Summary           string           `json:"summary"`
}

// ParsePlanJSON parses raw PostgreSQL EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON) output.
func ParsePlanJSON(jsonBytes []byte, originalQuery string) (*Report, error) {
	var roots []explainRoot
	if err := json.Unmarshal(jsonBytes, &roots); err != nil {
		return nil, ErrInvalidPlanFormat
	}
	if len(roots) == 0 {
		return nil, ErrNoPlanFound
	}

	root := roots[0]
	rep := &Report{
		Query:           originalQuery,
		ExecutionTimeMs: root.ExecutionTime,
		PlanningTimeMs:  root.PlanningTime,
		TotalCost:       root.Plan.TotalCost,
		SeqScans:        make([]SeqScanWarning, 0),
		IndexScans:      make([]string, 0),
	}

	walkNode(&root.Plan, rep)

	// Evaluate Missing Index heuristic
	if len(rep.SeqScans) > 0 {
		for _, s := range rep.SeqScans {
			if s.ActualRows > 500 || s.Cost > 100 {
				rep.MissingIndexAlert = true
				break
			}
		}
	}

	// Build summary
	if rep.MissingIndexAlert {
		rep.Summary = fmt.Sprintf("CRITICAL: Sequential Scan on %d table(s) detected with high row count or cost! Consider adding an index on table '%s'.", len(rep.SeqScans), rep.SeqScans[0].Table)
	} else if len(rep.IndexScans) > 0 {
		rep.Summary = fmt.Sprintf("OPTIMAL: Query utilizes index scan(s): %v", rep.IndexScans)
	} else {
		rep.Summary = "HEALTHY: Query executed within acceptable parameters."
	}

	return rep, nil
}

func walkNode(node *PlanNode, rep *Report) {
	if node == nil {
		return
	}

	rep.SharedHitBlocks += node.SharedHitBlocks
	rep.SharedReadBlocks += node.SharedReadBlocks

	switch node.NodeType {
	case "Seq Scan":
		rep.SeqScans = append(rep.SeqScans, SeqScanWarning{
			Table:      node.RelationName,
			Cost:       node.TotalCost,
			ActualRows: node.ActualRows,
			Filter:     node.Filter,
		})
	case "Index Scan", "Index Only Scan", "Bitmap Index Scan":
		if node.IndexName != "" {
			rep.IndexScans = append(rep.IndexScans, fmt.Sprintf("%s (%s)", node.IndexName, node.RelationName))
		}
	}

	for i := range node.Plans {
		walkNode(&node.Plans[i], rep)
	}
}
