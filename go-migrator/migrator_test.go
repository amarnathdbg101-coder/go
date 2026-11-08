package gomigrator

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"testing/fstest"
)

func TestComputeChecksum(t *testing.T) {
	sql := "CREATE TABLE users (id SERIAL PRIMARY KEY, email TEXT);"
	expectedSum := sha256.Sum256([]byte(sql))
	expected := hex.EncodeToString(expectedSum[:])

	got := ComputeChecksum(sql)
	if got != expected {
		t.Fatalf("expected checksum %s, got %s", expected, got)
	}
}

func TestFromFS_ParsingAndSorting(t *testing.T) {
	// Mock filesystem with out-of-order files
	mockFS := fstest.MapFS{
		"migrations/0002_create_orders.up.sql":   &fstest.MapFile{Data: []byte("CREATE TABLE orders (id INT);")},
		"migrations/0002_create_orders.down.sql": &fstest.MapFile{Data: []byte("DROP TABLE orders;")},
		"migrations/0001_create_users.up.sql":    &fstest.MapFile{Data: []byte("CREATE TABLE users (id INT);")},
		"migrations/0001_create_users.down.sql":  &fstest.MapFile{Data: []byte("DROP TABLE users;")},
	}

	migrations, err := FromFS(mockFS, "migrations")
	if err != nil {
		t.Fatalf("FromFS failed: %v", err)
	}

	if len(migrations) != 2 {
		t.Fatalf("expected 2 migrations, got %d", len(migrations))
	}

	// Verify chronological order: 1 then 2
	if migrations[0].Version != 1 || migrations[0].Name != "create_users" {
		t.Fatalf("expected first migration to be 0001_create_users, got %+v", migrations[0])
	}
	if migrations[1].Version != 2 || migrations[1].Name != "create_orders" {
		t.Fatalf("expected second migration to be 0002_create_orders, got %+v", migrations[1])
	}

	if migrations[0].UpSQL != "CREATE TABLE users (id INT);" {
		t.Fatalf("unexpected UpSQL: %s", migrations[0].UpSQL)
	}
	if migrations[0].DownSQL != "DROP TABLE users;" {
		t.Fatalf("unexpected DownSQL: %s", migrations[0].DownSQL)
	}
	if migrations[0].Checksum == "" {
		t.Fatal("checksum should be populated")
	}
}

func TestMigrationError_Unwrap(t *testing.T) {
	origErr := ErrNoMigrations
	migErr := &MigrationError{
		Version: 1,
		Name:    "init",
		Err:     origErr,
	}

	if migErr.Unwrap() != origErr {
		t.Fatal("Unwrap failed to return underlying error")
	}
}
