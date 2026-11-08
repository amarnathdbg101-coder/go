# SPEC: go-sqlbuilder

## 1. Metadata
- **Module**: `go-sqlbuilder`
- **Package**: `gosqlbuilder`
- **Location**: `C:\Library\go-sqlbuilder`
- **Go Version**: `1.23+`
- **Dependencies**: Zero external dependencies (100% Go Standard Library: `strings`, `strconv`, `fmt`, `errors`).

---

## 2. Core Capabilities
1. **Zero-Reflection Parameterized SQL**:
   - Generates PostgreSQL-compliant `$1, $2, $3, ...` indexed placeholders automatically without inlining values or string concatenation.
2. **Dynamic Queries**:
   - `SelectBuilder`: chains `.Where("col = ?", val)`, `.WhereIn("role", "a", "b")`, `.Join()`, `.OrderBy()`, `.Limit()`, `.Offset()`.
   - `InsertBuilder`: single and batch multi-row inserts with `RETURNING`.
   - `UpdateBuilder`: `.Set("col", val)` with parameterized assignments.
   - `DeleteBuilder`: clean parameterized deletions.
3. **SQL Injection Defense**:
   - Values are never concatenated into the SQL statement; passed strictly via `[]any` to `database/sql`.

---

## 3. Public API
```go
func Select(columns ...string) *SelectBuilder
func (b *SelectBuilder) From(table string) *SelectBuilder
func (b *SelectBuilder) Join(joinType, table, on string) *SelectBuilder
func (b *SelectBuilder) Where(condition string, args ...any) *SelectBuilder
func (b *SelectBuilder) WhereIn(column string, args ...any) *SelectBuilder
func (b *SelectBuilder) OrderBy(column, direction string) *SelectBuilder
func (b *SelectBuilder) Limit(limit int) *SelectBuilder
func (b *SelectBuilder) Offset(offset int) *SelectBuilder
func (b *SelectBuilder) Build() (string, []any, error)

func InsertInto(table string) *InsertBuilder
func Update(table string) *UpdateBuilder
func DeleteFrom(table string) *DeleteBuilder
```
