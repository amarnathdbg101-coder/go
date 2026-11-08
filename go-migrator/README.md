# go-migrator 🗄️

Zero-dependency transactional PostgreSQL migration runner with embedded SQL files support (`embed.FS`), atomic rollback, and SHA-256 checksum tampering defense. Adheres strictly to the [Master Library Blueprint](../BLUEPRINT.md).

---

## Features
- 🛡️ **Atomic Transactions**: Har migration script apne isolated transaction (`BEGIN ... COMMIT`) me execute hota hai; failure par turant rollback ho jata hai.
- 🔒 **SHA-256 Checksum Protection**: Agar kisi developer ne purana applied migration modify kiya, to `ErrChecksumMismatch` throw karke dirty state rokk deta hai.
- 📦 **Embedded SQL (`embed.FS`)**: SQL files Go binary ke andar embed ho jati hain (no external files needed in Docker containers).
- ⏪ **Clean Rollback**: Step-wise rollbacks (`Rollback(1)`) using corresponding `.down.sql` files.
- ⚡ **Zero External Dependencies**: 100% pure Go standard library.

---

## Quick Start

### 1. Migrations Directory (`migrations/`)
```
migrations/
├── 0001_create_users.up.sql
├── 0001_create_users.down.sql
├── 0002_create_orders.up.sql
└── 0002_create_orders.down.sql
```

### 2. Running Migrations on App Boot
```go
package main

import (
	"context"
	"database/sql"
	"embed"
	"log"
	gomigrator "go-migrator"
	_ "github.com/lib/pq"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

func main() {
	db, _ := sql.Open("postgres", "postgres://user:pass@localhost:5432/shopsilo?sslmode=disable")

	// Parse embedded SQL files
	migrations, err := gomigrator.FromFS(migrationFS, "migrations")
	if err != nil {
		log.Fatal(err)
	}

	migrator := gomigrator.New(db, migrations)

	// Apply pending migrations atomically
	applied, err := migrator.Migrate(context.Background())
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Printf("Successfully applied %d migrations!\n", applied)
}
```

---

## Running Tests
```bash
go test -v -count=1 ./...
```
