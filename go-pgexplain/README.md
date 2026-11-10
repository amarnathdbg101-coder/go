# go-pgexplain 🐘

> **Automated PostgreSQL Slow Query Interceptor, Execution Plan Parser, and Missing Index Auditor (Go 1.23+).**  
> Intercepts queries exceeding duration thresholds, automatically issues `EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON)`, and identifies Sequential Scans (`Seq Scan`) and missing indexes before they degrade production.

---

## 🚀 Key Features

1. **Automatic Slow Query Interception**:
   - Triggers only when a query execution exceeds `SlowThreshold` (e.g. 100ms).
2. **PostgreSQL JSON Plan Analyzer**:
   - Parses recursive execution trees: `Seq Scan`, `Index Scan`, `Bitmap Index Scan`, `Planning Time`, and `Execution Time`.
3. **Missing Index Detection**:
   - Flags dangerous sequential scans on tables with high costs or $> 500$ rows.
4. **Buffer I/O Telemetry**:
   - Extracts `Shared Hit Blocks` (RAM cache hits) vs `Shared Read Blocks` (slow disk I/O) to identify memory pressure.
5. **Non-Blocking Background Analysis**:
   - Runs `EXPLAIN` queries asynchronously by default without holding up caller responses.
6. **Zero Dependencies**:
   - 100% pure standard library (`database/sql`, `encoding/json`, `sync`, `time`).

---

## 📦 Quickstart

```go
package main

import (
	"database/sql"
	"fmt"
	"time"

	"go-pgexplain"
)

func main() {
	rawDB, _ := sql.Open("postgres", "postgres://localhost/mydb")

	explainer := pgexplain.New(
		pgexplain.WithSlowThreshold(50*time.Millisecond),
		pgexplain.WithOnSlowQuery(func(rep *pgexplain.Report) {
			if rep.MissingIndexAlert {
				fmt.Printf("🚨 SLOW QUERY ALERT on %s! %s\n", rep.SeqScans[0].Table, rep.Summary)
			}
		}),
	)
	defer explainer.Close()
	explainer.SetDB(rawDB)
}
```
