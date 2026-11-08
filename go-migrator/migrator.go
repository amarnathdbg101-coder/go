package gomigrator

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// AppliedMigration records a successfully applied migration in the database.
type AppliedMigration struct {
	Version   int64
	Name      string
	Checksum  string
	AppliedAt time.Time
}

// MigrationStatus reports whether a script has been executed.
type MigrationStatus struct {
	Version   int64
	Name      string
	Applied   bool
	AppliedAt *time.Time
}

// Migrator executes schema migrations inside transactional boundaries.
type Migrator struct {
	db         *sql.DB
	migrations []Migration
	tableName  string
}

// New creates a Migrator.
func New(db *sql.DB, migrations []Migration, tableName ...string) *Migrator {
	tbl := "schema_migrations"
	if len(tableName) > 0 && tableName[0] != "" {
		tbl = tableName[0]
	}

	return &Migrator{
		db:         db,
		migrations: migrations,
		tableName:  tbl,
	}
}

func (m *Migrator) initSchemaTable(ctx context.Context) error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			version BIGINT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			checksum VARCHAR(64) NOT NULL,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);`, m.tableName)

	_, err := m.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("migrator: create migrations table: %w", err)
	}
	return nil
}

// GetApplied returns a map of version -> AppliedMigration from the database.
func (m *Migrator) GetApplied(ctx context.Context) (map[int64]AppliedMigration, error) {
	if err := m.initSchemaTable(ctx); err != nil {
		return nil, err
	}

	query := fmt.Sprintf("SELECT version, name, checksum, applied_at FROM %s ORDER BY version ASC", m.tableName)
	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("migrator: query applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[int64]AppliedMigration)
	for rows.Next() {
		var app AppliedMigration
		if err := rows.Scan(&app.Version, &app.Name, &app.Checksum, &app.AppliedAt); err != nil {
			return nil, fmt.Errorf("migrator: scan applied row: %w", err)
		}
		applied[app.Version] = app
	}

	return applied, rows.Err()
}

// Migrate executes all pending migrations in ascending version order.
func (m *Migrator) Migrate(ctx context.Context) (int, error) {
	applied, err := m.GetApplied(ctx)
	if err != nil {
		return 0, err
	}

	appliedCount := 0

	for _, mig := range m.migrations {
		app, isApplied := applied[mig.Version]
		if isApplied {
			// Integrity check: verify checksum matches
			if app.Checksum != mig.Checksum {
				return appliedCount, fmt.Errorf("%w: version %04d (%s) expected checksum %s, got %s",
					ErrChecksumMismatch, mig.Version, mig.Name, app.Checksum, mig.Checksum)
			}
			continue // Already applied
		}

		// Apply migration inside an isolated transaction
		tx, err := m.db.BeginTx(ctx, nil)
		if err != nil {
			return appliedCount, fmt.Errorf("migrator: begin tx: %w", err)
		}

		if mig.UpSQL != "" {
			if _, err := tx.ExecContext(ctx, mig.UpSQL); err != nil {
				_ = tx.Rollback()
				return appliedCount, &MigrationError{
					Version: mig.Version,
					Name:    mig.Name,
					Err:     err,
				}
			}
		}

		// Record migration in schema_migrations
		insertQuery := fmt.Sprintf("INSERT INTO %s (version, name, checksum, applied_at) VALUES ($1, $2, $3, NOW())", m.tableName)
		if _, err := tx.ExecContext(ctx, insertQuery, mig.Version, mig.Name, mig.Checksum); err != nil {
			_ = tx.Rollback()
			return appliedCount, fmt.Errorf("migrator: record migration %04d: %w", mig.Version, err)
		}

		if err := tx.Commit(); err != nil {
			return appliedCount, fmt.Errorf("migrator: commit tx: %w", err)
		}

		appliedCount++
	}

	return appliedCount, nil
}

// Rollback reverses the last N applied migrations.
func (m *Migrator) Rollback(ctx context.Context, steps int) (int, error) {
	if steps <= 0 {
		steps = 1
	}

	applied, err := m.GetApplied(ctx)
	if err != nil {
		return 0, err
	}

	rolledBackCount := 0

	// Iterate in reverse order
	for i := len(m.migrations) - 1; i >= 0; i-- {
		if rolledBackCount >= steps {
			break
		}

		mig := m.migrations[i]
		if _, isApplied := applied[mig.Version]; !isApplied {
			continue // Not applied, cannot rollback
		}

		tx, err := m.db.BeginTx(ctx, nil)
		if err != nil {
			return rolledBackCount, fmt.Errorf("migrator: begin rollback tx: %w", err)
		}

		if mig.DownSQL != "" {
			if _, err := tx.ExecContext(ctx, mig.DownSQL); err != nil {
				_ = tx.Rollback()
				return rolledBackCount, &MigrationError{
					Version: mig.Version,
					Name:    mig.Name,
					Err:     err,
				}
			}
		}

		deleteQuery := fmt.Sprintf("DELETE FROM %s WHERE version = $1", m.tableName)
		if _, err := tx.ExecContext(ctx, deleteQuery, mig.Version); err != nil {
			_ = tx.Rollback()
			return rolledBackCount, fmt.Errorf("migrator: delete migration record %04d: %w", mig.Version, err)
		}

		if err := tx.Commit(); err != nil {
			return rolledBackCount, fmt.Errorf("migrator: commit rollback tx: %w", err)
		}

		rolledBackCount++
	}

	return rolledBackCount, nil
}
