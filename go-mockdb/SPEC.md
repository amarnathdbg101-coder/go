# SPEC: go-mockdb

## 1. Metadata
- **Module**: `go-mockdb`
- **Package**: `mockdb`
- **Location**: `C:\Library\go-mockdb`
- **Go Version**: `1.23+`
- **Dependencies**: Zero external dependencies (100% Go Standard Library: `database/sql`, `database/sql/driver`, `sync`, `regexp`, `fmt`, `errors`, `io`).

---

## 2. Core Capabilities
1. **Standard `database/sql/driver` Mocking**:
   - Registers a virtual `mockdb` SQL driver, allowing real `*sql.DB` instances to be used in unit tests without external databases (PostgreSQL/MySQL/SQLite).
2. **Deterministic Query & Exec Expectations**:
   - Regex-based SQL pattern matching for `ExpectQuery(...)` and `ExpectExec(...)`.
   - Strict argument matching (`WithArgs(...)`) and mock result generation (`WillReturnRows(...)`, `WillReturnResult(...)`, `WillReturnError(...)`).
3. **Transaction Lifecycle Validation**:
   - Explicitly records `ExpectBegin()`, `ExpectCommit()`, and `ExpectRollback()`.
4. **Unmet Expectations Verification**:
   - `ExpectationsWereMet()` ensures zero unexecuted or missed queries during test completion.

---

## 3. Public API
```go
type Result struct
func NewResult(lastInsertID, rowsAffected int64) *Result

type Rows struct
func NewRows(columns []string) *Rows
func (r *Rows) AddRow(values ...any) *Rows

type ExpectedQuery struct
func (eq *ExpectedQuery) WithArgs(args ...any) *ExpectedQuery
func (eq *ExpectedQuery) WillReturnRows(rows *Rows) *ExpectedQuery
func (eq *ExpectedQuery) WillReturnError(err error) *ExpectedQuery

type ExpectedExec struct
func (ee *ExpectedExec) WithArgs(args ...any) *ExpectedExec
func (ee *ExpectedExec) WillReturnResult(lastInsertID, rowsAffected int64) *ExpectedExec
func (ee *ExpectedExec) WillReturnError(err error) *ExpectedExec

type Mock struct
func New() (*sql.DB, *Mock, error)
func (m *Mock) ExpectQuery(pattern string) *ExpectedQuery
func (m *Mock) ExpectExec(pattern string) *ExpectedExec
func (m *Mock) ExpectBegin()
func (m *Mock) ExpectCommit()
func (m *Mock) ExpectRollback()
func (m *Mock) ExpectationsWereMet() error
```
