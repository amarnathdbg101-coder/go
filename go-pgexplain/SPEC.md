# SPEC: go-pgexplain

## 1. Scope & Purpose
`go-pgexplain` audits PostgreSQL queries by capturing and analyzing execution plans for operations that breach latency thresholds. It diagnoses performance anomalies (sequential scans, high disk buffer reads, unindexed filters) in production.

It does NOT modify database indexes or rewrite queries automatically.

## 2. Public API Contract
```go
type PlanNode struct { ... }
type Report struct { ... }
type Explainer struct { ... }

func New(opts ...Option) *Explainer
func ParsePlanJSON(jsonBytes []byte, query string) (*Report, error)
func (e *Explainer) SetDB(db *sql.DB)
func (e *Explainer) SetExecutor(fn ExplainFunc)
func (e *Explainer) Audit(ctx context.Context, query string, duration time.Duration, args ...any) (*Report, error)
func (e *Explainer) Close() error
```

## 3. Concurrency & Memory Guarantees
- Explainer configuration is thread-safe.
- Parsing allocates AST nodes proportionally to the plan depth.
- Asynchronous explain calls execute in isolated goroutines with bounded timeouts.

## 4. How to Extend
- Add automated Slack/Discord webhook alerts via `go-webhook`.
- Add DDL suggestion generators (e.g. `CREATE INDEX CONCURRENTLY ...`).
