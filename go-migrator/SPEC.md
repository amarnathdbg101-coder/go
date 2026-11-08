# SPEC: go-migrator

## 1. Metadata
- **Module**: `go-migrator`
- **Package**: `gomigrator`
- **Location**: `C:\Library\go-migrator`
- **Go Version**: `1.23+`
- **Dependencies**: Zero external dependencies (100% Go Standard Library: `database/sql`, `embed`, `io/fs`, `crypto/sha256`, `strings`, `time`, `fmt`, `errors`).

---

## 2. Core Capabilities
1. **Transactional Execution**:
   - Each migration runs inside its own atomic `tx.BeginTx() ... tx.Commit()` boundary, preventing half-applied schema corruptions.
2. **SHA-256 Checksum Tampering Defense**:
   - Detects if an already-applied migration script was edited after execution, throwing `ErrChecksumMismatch`.
3. **Embed.FS Native**:
   - Compiles SQL files directly into the single Go binary using Go standard `embed.FS`.
4. **Up & Down Migrations**:
   - Supports forward upgrades (`Migrate()`) and stepwise rollbacks (`Rollback(steps)`).

---

## 3. Public API
```go
type Migration struct {
    Version  int64
    Name     string
    UpSQL    string
    DownSQL  string
    Checksum string
}

func ComputeChecksum(sql string) string
func NewMigration(version int64, name, upSQL, downSQL string) Migration
func FromFS(fsys fs.FS, rootDir string) ([]Migration, error)

type Migrator struct { ... }
func New(db *sql.DB, migrations []Migration, tableName ...string) *Migrator
func (m *Migrator) Migrate(ctx context.Context) (int, error)
func (m *Migrator) Rollback(ctx context.Context, steps int) (int, error)
func (m *Migrator) GetApplied(ctx context.Context) (map[int64]AppliedMigration, error)
```
