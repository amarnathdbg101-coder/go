package mockdb

import (
	"errors"
	"testing"
)

func TestMockDB_Query(t *testing.T) {
	db, mock, err := New()
	if err != nil {
		t.Fatalf("failed to init mockdb: %v", err)
	}
	defer db.Close()

	rows := NewRows([]string{"id", "name", "balance"}).
		AddRow(1, "Arjun", 1500.50).
		AddRow(2, "Vikram", 2400.00)

	mock.ExpectQuery(`SELECT id, name, balance FROM users WHERE active = \$1`).
		WithArgs(true).
		WillReturnRows(rows)

	query := `SELECT id, name, balance FROM users WHERE active = $1`
	resRows, err := db.Query(query, true)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	defer resRows.Close()

	var count int
	for resRows.Next() {
		var id int
		var name string
		var balance float64
		if err := resRows.Scan(&id, &name, &balance); err != nil {
			t.Fatalf("scan failed: %v", err)
		}
		count++
	}

	if count != 2 {
		t.Fatalf("expected 2 scanned rows, got %d", count)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations unmet: %v", err)
	}
}

func TestMockDB_Exec(t *testing.T) {
	db, mock, err := New()
	if err != nil {
		t.Fatalf("failed to init mockdb: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(`UPDATE users SET status = \$1 WHERE id = \$2`).
		WithArgs("suspended", 42).
		WillReturnResult(0, 1)

	res, err := db.Exec(`UPDATE users SET status = $1 WHERE id = $2`, "suspended", 42)
	if err != nil {
		t.Fatalf("exec failed: %v", err)
	}

	affected, _ := res.RowsAffected()
	if affected != 1 {
		t.Fatalf("expected 1 row affected, got %d", affected)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations unmet: %v", err)
	}
}

func TestMockDB_Transaction(t *testing.T) {
	db, mock, err := New()
	if err != nil {
		t.Fatalf("failed to init mockdb: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO logs`).WillReturnResult(101, 1)
	mock.ExpectCommit()

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin failed: %v", err)
	}

	_, err = tx.Exec(`INSERT INTO logs VALUES ('login')`)
	if err != nil {
		t.Fatalf("tx exec failed: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		t.Fatalf("commit failed: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations unmet: %v", err)
	}
}

func TestMockDB_UnmetExpectations(t *testing.T) {
	db, mock, err := New()
	if err != nil {
		t.Fatalf("failed to init mockdb: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(`SELECT 1`)

	// We did NOT execute the query!
	err = mock.ExpectationsWereMet()
	if !errors.Is(err, ErrUnmetExpectations) {
		t.Fatalf("expected ErrUnmetExpectations, got %v", err)
	}
}

func TestMockDB_UnexpectedQuery(t *testing.T) {
	db, _, err := New()
	if err != nil {
		t.Fatalf("failed to init mockdb: %v", err)
	}
	defer db.Close()

	_, err = db.Query(`SELECT unexpected FROM table`)
	if !errors.Is(err, ErrUnexpectedQuery) {
		t.Fatalf("expected ErrUnexpectedQuery, got %v", err)
	}
}

func TestMockDB_ArgMismatch(t *testing.T) {
	db, mock, err := New()
	if err != nil {
		t.Fatalf("failed to init mockdb: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(`SELECT \* FROM users WHERE id = \$1`).WithArgs(99)

	_, err = db.Query(`SELECT * FROM users WHERE id = $1`, 42) // wrong arg
	if !errors.Is(err, ErrArgumentMismatch) {
		t.Fatalf("expected ErrArgumentMismatch, got %v", err)
	}
}
