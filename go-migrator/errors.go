package gomigrator

import (
	"errors"
	"fmt"
)

var (
	// ErrNoMigrations indicates no migration scripts were found.
	ErrNoMigrations = errors.New("migrator: no migrations found")

	// ErrChecksumMismatch occurs when an already-applied migration script was modified after execution.
	ErrChecksumMismatch = errors.New("migrator: migration checksum mismatch (script modified after execution)")

	// ErrDuplicateVersion occurs when two migration scripts share the same version number.
	ErrDuplicateVersion = errors.New("migrator: duplicate migration version")
)

// MigrationError wraps the failed migration script and underlying SQL error.
type MigrationError struct {
	Version int64
	Name    string
	Err     error
}

func (e *MigrationError) Error() string {
	return fmt.Sprintf("migrator: migration %04d_%s failed: %v", e.Version, e.Name, e.Err)
}

func (e *MigrationError) Unwrap() error {
	return e.Err
}
