package mockdb

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"regexp"
	"sync"
	"sync/atomic"
)

var (
	driverOnce sync.Once
	dsnCounter int64
	registryMu sync.RWMutex
	registry   = make(map[string]*Mock)
)

func initDriver() {
	sql.Register("mockdb", &mockDriver{})
}

type mockDriver struct{}

func (d *mockDriver) Open(name string) (driver.Conn, error) {
	registryMu.RLock()
	mock, ok := registry[name]
	registryMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("mockdb: no mock registered for dsn '%s'", name)
	}
	return &mockConn{mock: mock}, nil
}

// Result implements driver.Result.
type Result struct {
	lastInsertID int64
	rowsAffected int64
}

func NewResult(lastInsertID, rowsAffected int64) *Result {
	return &Result{lastInsertID: lastInsertID, rowsAffected: rowsAffected}
}

func (r *Result) LastInsertId() (int64, error) { return r.lastInsertID, nil }
func (r *Result) RowsAffected() (int64, error) { return r.rowsAffected, nil }

// Rows implements driver.Rows.
type Rows struct {
	cols []string
	rows [][]driver.Value
	pos  int
}

// NewRows initializes mock rows with given column names.
func NewRows(columns []string) *Rows {
	return &Rows{cols: columns}
}

// AddRow appends a row of values.
func (r *Rows) AddRow(values ...any) *Rows {
	dvals := make([]driver.Value, len(values))
	for i, v := range values {
		dvals[i] = driver.Value(v)
	}
	r.rows = append(r.rows, dvals)
	return r
}

func (r *Rows) Columns() []string { return r.cols }
func (r *Rows) Close() error      { return nil }

func (r *Rows) Next(dest []driver.Value) error {
	if r.pos >= len(r.rows) {
		return io.EOF
	}
	row := r.rows[r.pos]
	for i := range dest {
		if i < len(row) {
			dest[i] = row[i]
		}
	}
	r.pos++
	return nil
}

type expectation interface {
	matchQuery(query string, args []driver.NamedValue) (bool, error)
	fulfilled() bool
}

// ExpectedQuery represents an expected SQL SELECT statement.
type ExpectedQuery struct {
	queryRegex *regexp.Regexp
	args       []driver.Value
	rows       *Rows
	err        error
	matched    bool
}

func (eq *ExpectedQuery) matchQuery(query string, args []driver.NamedValue) (bool, error) {
	if !eq.queryRegex.MatchString(query) {
		return false, nil
	}
	if len(eq.args) > 0 {
		if len(eq.args) != len(args) {
			return false, fmt.Errorf("%w: expected %d args, got %d", ErrArgumentMismatch, len(eq.args), len(args))
		}
		for i, exp := range eq.args {
			if fmt.Sprint(exp) != fmt.Sprint(args[i].Value) {
				return false, fmt.Errorf("%w: arg %d: expected %v, got %v", ErrArgumentMismatch, i, exp, args[i].Value)
			}
		}
	}
	eq.matched = true
	return true, nil
}

func (eq *ExpectedQuery) fulfilled() bool { return eq.matched }

// WithArgs specifies the expected arguments.
func (eq *ExpectedQuery) WithArgs(args ...any) *ExpectedQuery {
	for _, a := range args {
		eq.args = append(eq.args, driver.Value(a))
	}
	return eq
}

// WillReturnRows assigns mock rows to return.
func (eq *ExpectedQuery) WillReturnRows(rows *Rows) *ExpectedQuery {
	eq.rows = rows
	return eq
}

// WillReturnError configures an error to return.
func (eq *ExpectedQuery) WillReturnError(err error) *ExpectedQuery {
	eq.err = err
	return eq
}

// ExpectedExec represents an expected SQL INSERT/UPDATE/DELETE statement.
type ExpectedExec struct {
	execRegex *regexp.Regexp
	args      []driver.Value
	result    *Result
	err       error
	matched   bool
}

func (ee *ExpectedExec) matchQuery(query string, args []driver.NamedValue) (bool, error) {
	if !ee.execRegex.MatchString(query) {
		return false, nil
	}
	if len(ee.args) > 0 {
		if len(ee.args) != len(args) {
			return false, fmt.Errorf("%w: expected %d args, got %d", ErrArgumentMismatch, len(ee.args), len(args))
		}
		for i, exp := range ee.args {
			if fmt.Sprint(exp) != fmt.Sprint(args[i].Value) {
				return false, fmt.Errorf("%w: arg %d: expected %v, got %v", ErrArgumentMismatch, i, exp, args[i].Value)
			}
		}
	}
	ee.matched = true
	return true, nil
}

func (ee *ExpectedExec) fulfilled() bool { return ee.matched }

// WithArgs specifies expected arguments.
func (ee *ExpectedExec) WithArgs(args ...any) *ExpectedExec {
	for _, a := range args {
		ee.args = append(ee.args, driver.Value(a))
	}
	return ee
}

// WillReturnResult sets the execution result.
func (ee *ExpectedExec) WillReturnResult(lastInsertID, rowsAffected int64) *ExpectedExec {
	ee.result = NewResult(lastInsertID, rowsAffected)
	return ee
}

// WillReturnError sets the error to return.
func (ee *ExpectedExec) WillReturnError(err error) *ExpectedExec {
	ee.err = err
	return ee
}

// ExpectedTx tracks transaction expectations.
type ExpectedTx struct {
	kind    string // "begin", "commit", "rollback"
	matched bool
}

func (et *ExpectedTx) matchQuery(query string, args []driver.NamedValue) (bool, error) { return false, nil }
func (et *ExpectedTx) fulfilled() bool                                                 { return et.matched }

// Mock manages database expectations and verification.
type Mock struct {
	mu           sync.Mutex
	dsn          string
	expectations []expectation
}

// New creates a new *sql.DB and corresponding *Mock controller.
func New() (*sql.DB, *Mock, error) {
	driverOnce.Do(initDriver)

	dsnID := atomic.AddInt64(&dsnCounter, 1)
	dsn := fmt.Sprintf("mockdb_instance_%d", dsnID)

	mock := &Mock{dsn: dsn}

	registryMu.Lock()
	registry[dsn] = mock
	registryMu.Unlock()

	db, err := sql.Open("mockdb", dsn)
	if err != nil {
		return nil, nil, err
	}

	return db, mock, nil
}

// ExpectQuery registers an expected SELECT query with regex pattern matching.
func (m *Mock) ExpectQuery(pattern string) *ExpectedQuery {
	m.mu.Lock()
	defer m.mu.Unlock()

	eq := &ExpectedQuery{
		queryRegex: regexp.MustCompile(pattern),
	}
	m.expectations = append(m.expectations, eq)
	return eq
}

// ExpectExec registers an expected INSERT/UPDATE/DELETE query with regex pattern matching.
func (m *Mock) ExpectExec(pattern string) *ExpectedExec {
	m.mu.Lock()
	defer m.mu.Unlock()

	ee := &ExpectedExec{
		execRegex: regexp.MustCompile(pattern),
	}
	m.expectations = append(m.expectations, ee)
	return ee
}

// ExpectBegin expects a transaction start.
func (m *Mock) ExpectBegin() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.expectations = append(m.expectations, &ExpectedTx{kind: "begin"})
}

// ExpectCommit expects a transaction commit.
func (m *Mock) ExpectCommit() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.expectations = append(m.expectations, &ExpectedTx{kind: "commit"})
}

// ExpectRollback expects a transaction rollback.
func (m *Mock) ExpectRollback() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.expectations = append(m.expectations, &ExpectedTx{kind: "rollback"})
}

// ExpectationsWereMet verifies that all registered expectations were satisfied.
func (m *Mock) ExpectationsWereMet() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, exp := range m.expectations {
		if !exp.fulfilled() {
			return fmt.Errorf("%w: expectation #%d not met", ErrUnmetExpectations, i+1)
		}
	}
	return nil
}

type mockConn struct {
	mock *Mock
}

func (c *mockConn) Prepare(query string) (driver.Stmt, error) {
	return &mockStmt{conn: c, query: query}, nil
}

func (c *mockConn) Close() error { return nil }

func (c *mockConn) Begin() (driver.Tx, error) {
	c.mock.mu.Lock()
	defer c.mock.mu.Unlock()

	for _, exp := range c.mock.expectations {
		if tx, ok := exp.(*ExpectedTx); ok && !tx.matched && tx.kind == "begin" {
			tx.matched = true
			return &mockTx{conn: c}, nil
		}
	}
	return nil, fmt.Errorf("%w: unexpected Begin", ErrUnexpectedQuery)
}

type mockTx struct {
	conn *mockConn
}

func (tx *mockTx) Commit() error {
	tx.conn.mock.mu.Lock()
	defer tx.conn.mock.mu.Unlock()

	for _, exp := range tx.conn.mock.expectations {
		if t, ok := exp.(*ExpectedTx); ok && !t.matched && t.kind == "commit" {
			t.matched = true
			return nil
		}
	}
	return fmt.Errorf("%w: unexpected Commit", ErrUnexpectedQuery)
}

func (tx *mockTx) Rollback() error {
	tx.conn.mock.mu.Lock()
	defer tx.conn.mock.mu.Unlock()

	for _, exp := range tx.conn.mock.expectations {
		if t, ok := exp.(*ExpectedTx); ok && !t.matched && t.kind == "rollback" {
			t.matched = true
			return nil
		}
	}
	return fmt.Errorf("%w: unexpected Rollback", ErrUnexpectedQuery)
}

type mockStmt struct {
	conn  *mockConn
	query string
}

func (s *mockStmt) Close() error { return nil }
func (s *mockStmt) NumInput() int { return -1 }

func (s *mockStmt) Exec(args []driver.Value) (driver.Result, error) {
	named := make([]driver.NamedValue, len(args))
	for i, a := range args {
		named[i] = driver.NamedValue{Ordinal: i + 1, Value: a}
	}
	return s.conn.ExecContext(context.Background(), s.query, named)
}

func (s *mockStmt) Query(args []driver.Value) (driver.Rows, error) {
	named := make([]driver.NamedValue, len(args))
	for i, a := range args {
		named[i] = driver.NamedValue{Ordinal: i + 1, Value: a}
	}
	return s.conn.QueryContext(context.Background(), s.query, named)
}

func (c *mockConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	c.mock.mu.Lock()
	defer c.mock.mu.Unlock()

	for _, exp := range c.mock.expectations {
		if eq, ok := exp.(*ExpectedQuery); ok && !eq.matched {
			matched, err := eq.matchQuery(query, args)
			if err != nil {
				return nil, err
			}
			if matched {
				if eq.err != nil {
					return nil, eq.err
				}
				if eq.rows != nil {
					return eq.rows, nil
				}
				return NewRows([]string{}), nil
			}
		}
	}

	return nil, fmt.Errorf("%w: %s", ErrUnexpectedQuery, query)
}

func (c *mockConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	c.mock.mu.Lock()
	defer c.mock.mu.Unlock()

	for _, exp := range c.mock.expectations {
		if ee, ok := exp.(*ExpectedExec); ok && !ee.matched {
			matched, err := ee.matchQuery(query, args)
			if err != nil {
				return nil, err
			}
			if matched {
				if ee.err != nil {
					return nil, ee.err
				}
				if ee.result != nil {
					return ee.result, nil
				}
				return NewResult(0, 1), nil
			}
		}
	}

	return nil, fmt.Errorf("%w: %s", ErrUnexpectedQuery, query)
}
