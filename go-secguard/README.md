# go-secguard 🛡️

Zero-Dependency OWASP Security Headers, Sub-Microsecond WAF, and IP Quarantine Jail for Go.

## 🚀 Key Features
- **Zero External Dependencies**: Pure Go standard library (`net/http`, `regexp`, `sync`, `database/sql`).
- **PostgreSQL Distributed Jail**: Built-in `PostgresJailStorage` allows multiple API servers to share quarantined attacker IPs in real time.
- **OWASP Security Headers**: Injects HSTS (with preload), `X-Content-Type-Options: nosniff`, `X-Frame-Options: SAMEORIGIN`, and Permissions-Policy.
- **Sub-Microsecond WAF**: 100ns fast-path pattern filtering for SQL Injection, Path Traversal, and XSS.
- **Automated IP Quarantine**: Offending clients probing `/.env`, `/.git`, or running `sqlmap` are quarantined into a temporary jail with `Retry-After` headers.
- **Observability Hooks**: Real-time security incident callbacks for SIEM, Slack alerts, or structured logging (`slog`).

---

## 🐘 PostgreSQL Distributed Jail Setup

```go
package main

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/shopsilo/go-secguard"
)

func main() {
	// PostgreSQL connection string
	dsn := "postgres://postgres:secret@localhost:5432/ecommerce?sslmode=disable"
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Initialize shared PostgreSQL jail
	pgJail := secguard.NewPostgresJailStorage(db)
	if err := pgJail.InitSchema(context.Background()); err != nil {
		panic(err)
	}

	guard := secguard.Middleware(
		secguard.WithJailStore(pgJail),
		secguard.WithJailDuration(30 * time.Minute),
	)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/items", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[{"id": 1}]`))
	})

	http.ListenAndServe(":8080", guard(mux))
}
```
