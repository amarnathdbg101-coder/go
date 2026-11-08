package gosqlbuilder

import (
	"errors"
	"testing"
)

func TestSelectBuilder_Complex(t *testing.T) {
	query, args, err := Select("id", "name", "email", "created_at").
		From("users").
		Join("LEFT JOIN", "profiles", "profiles.user_id = users.id").
		Where("users.status = ?", "active").
		Where("users.age >= ?", 18).
		WhereIn("users.role", "admin", "manager").
		OrderBy("users.created_at", "DESC").
		Limit(20).
		Offset(40).
		Build()

	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	expectedSQL := "SELECT id, name, email, created_at FROM users LEFT JOIN profiles ON profiles.user_id = users.id WHERE users.status = $1 AND users.age >= $2 AND users.role IN ($3, $4) ORDER BY users.created_at DESC LIMIT $5 OFFSET $6"

	if query != expectedSQL {
		t.Fatalf("SQL mismatch:\nExpected: %s\nGot:      %s", expectedSQL, query)
	}

	if len(args) != 6 {
		t.Fatalf("expected 6 arguments, got %d", len(args))
	}
	if args[0] != "active" || args[1] != 18 || args[2] != "admin" || args[3] != "manager" || args[4] != 20 || args[5] != 40 {
		t.Fatalf("args mismatch: %+v", args)
	}
}

func TestInsertBuilder_SingleAndMultiRow(t *testing.T) {
	// Single row with RETURNING
	query, args, err := InsertInto("products").
		Columns("title", "price", "stock").
		Values("Smartphone", 19999.00, 50).
		Returning("id", "created_at").
		Build()

	if err != nil {
		t.Fatalf("Build single insert failed: %v", err)
	}

	expectedSQL := "INSERT INTO products (title, price, stock) VALUES ($1, $2, $3) RETURNING id, created_at"
	if query != expectedSQL {
		t.Fatalf("expected:\n%s\ngot:\n%s", expectedSQL, query)
	}
	if len(args) != 3 {
		t.Fatalf("expected 3 args, got %d", len(args))
	}

	// Multi-row INSERT
	mQuery, mArgs, err := InsertInto("tags").
		Columns("name").
		Values("electronics").
		Values("gadgets").
		Values("deals").
		Build()

	if err != nil {
		t.Fatalf("Build multi insert failed: %v", err)
	}

	expectedMulti := "INSERT INTO tags (name) VALUES ($1), ($2), ($3)"
	if mQuery != expectedMulti {
		t.Fatalf("expected:\n%s\ngot:\n%s", expectedMulti, mQuery)
	}
	if len(mArgs) != 3 || mArgs[1] != "gadgets" {
		t.Fatalf("unexpected multi args: %+v", mArgs)
	}
}

func TestInsertBuilder_MismatchedValues(t *testing.T) {
	_, _, err := InsertInto("orders").
		Columns("user_id", "total").
		Values(1, 200, "extra_value").
		Build()

	if !errors.Is(err, ErrMismatchedValues) {
		t.Fatalf("expected ErrMismatchedValues, got %v", err)
	}
}

func TestUpdateBuilder(t *testing.T) {
	query, args, err := Update("orders").
		Set("status", "completed").
		Set("notes", "delivered on time").
		Where("id = ?", 101).
		Where("status = ?", "pending").
		Returning("id", "updated_at").
		Build()

	if err != nil {
		t.Fatalf("Build update failed: %v", err)
	}

	expectedSQL := "UPDATE orders SET status = $1, notes = $2 WHERE id = $3 AND status = $4 RETURNING id, updated_at"
	if query != expectedSQL {
		t.Fatalf("expected:\n%s\ngot:\n%s", expectedSQL, query)
	}

	if len(args) != 4 || args[0] != "completed" || args[2] != 101 {
		t.Fatalf("unexpected args: %+v", args)
	}
}

func TestDeleteBuilder(t *testing.T) {
	query, args, err := DeleteFrom("sessions").
		Where("expired_at < ?", "2026-01-01").
		Build()

	if err != nil {
		t.Fatalf("Build delete failed: %v", err)
	}

	expectedSQL := "DELETE FROM sessions WHERE expired_at < $1"
	if query != expectedSQL {
		t.Fatalf("expected:\n%s\ngot:\n%s", expectedSQL, query)
	}
	if len(args) != 1 || args[0] != "2026-01-01" {
		t.Fatalf("unexpected args: %+v", args)
	}
}
