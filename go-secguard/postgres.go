package secguard

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// PostgresJailStorage implements JailStorage backed by a PostgreSQL database table.
// It allows multiple load-balanced API servers to share a synchronized IP quarantine jail in real time.
type PostgresJailStorage struct {
	db        *sql.DB
	tableName string
}

// NewPostgresJailStorage creates a PostgreSQL-backed IP quarantine storage.
func NewPostgresJailStorage(db *sql.DB, tableName ...string) *PostgresJailStorage {
	tbl := "secguard_jail"
	if len(tableName) > 0 && tableName[0] != "" {
		tbl = tableName[0]
	}
	return &PostgresJailStorage{
		db:        db,
		tableName: tbl,
	}
}

// InitSchema creates the PostgreSQL quarantine table and index if they do not exist.
func (p *PostgresJailStorage) InitSchema(ctx context.Context) error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			ip VARCHAR(45) PRIMARY KEY,
			reason TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			expires_at TIMESTAMPTZ NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_%s_expires ON %s (expires_at);
	`, p.tableName, p.tableName, p.tableName)

	_, err := p.db.ExecContext(ctx, query)
	return err
}

// Jail places an offending IP address into quarantine with upsert semantics.
func (p *PostgresJailStorage) Jail(ctx context.Context, ip string, duration time.Duration, reason string) error {
	if ip == "" || ip == "127.0.0.1" || ip == "::1" {
		return nil
	}

	expiresAt := time.Now().Add(duration)
	query := fmt.Sprintf(`
		INSERT INTO %s (ip, reason, created_at, expires_at)
		VALUES ($1, $2, NOW(), $3)
		ON CONFLICT (ip) DO UPDATE SET
			reason     = EXCLUDED.reason,
			expires_at = EXCLUDED.expires_at
	`, p.tableName)

	_, err := p.db.ExecContext(ctx, query, ip, reason, expiresAt)
	return err
}

// IsJailed checks whether an IP is quarantined, computing remaining time via database server clock.
func (p *PostgresJailStorage) IsJailed(ctx context.Context, ip string) (bool, string, int64, error) {
	query := fmt.Sprintf(`
		SELECT reason, GREATEST(0, CEIL(EXTRACT(EPOCH FROM (expires_at - NOW())))::BIGINT)
		FROM %s
		WHERE ip = $1 AND expires_at > NOW()
	`, p.tableName)

	var (
		reason    string
		remaining int64
	)

	err := p.db.QueryRowContext(ctx, query, ip).Scan(&reason, &remaining)
	if err == sql.ErrNoRows {
		return false, "", 0, nil
	}
	if err != nil {
		return false, "", 0, err
	}

	return true, reason, remaining, nil
}

// Unjail removes an IP address from quarantine.
func (p *PostgresJailStorage) Unjail(ctx context.Context, ip string) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE ip = $1", p.tableName)
	_, err := p.db.ExecContext(ctx, query, ip)
	return err
}

// PurgeExpired frees storage by deleting expired quarantine records.
func (p *PostgresJailStorage) PurgeExpired(ctx context.Context) (int64, error) {
	query := fmt.Sprintf("DELETE FROM %s WHERE expires_at <= NOW()", p.tableName)
	res, err := p.db.ExecContext(ctx, query)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// Close implements JailStorage (leaves *sql.DB lifecycle to application).
func (p *PostgresJailStorage) Close() error {
	return nil
}
