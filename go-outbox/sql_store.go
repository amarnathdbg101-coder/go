package gooutbox

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// SQLStore provides production-grade persistence for outbox events backed by PostgreSQL / MySQL 8+.
type SQLStore struct {
	db        *sql.DB
	tableName string
}

// SQLOptions configures SQLStore parameters.
type SQLOptions struct {
	TableName string // Default: "outbox_events"
}

// NewSQLStore creates a new SQLStore backed by *sql.DB.
func NewSQLStore(db *sql.DB, opts *SQLOptions) *SQLStore {
	tableName := "outbox_events"
	if opts != nil && opts.TableName != "" {
		tableName = opts.TableName
	}
	return &SQLStore{
		db:        db,
		tableName: tableName,
	}
}

// InitSchema creates the outbox table and high-performance indexing in PostgreSQL.
func (s *SQLStore) InitSchema(ctx context.Context) error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id VARCHAR(64) PRIMARY KEY,
			topic VARCHAR(128) NOT NULL,
			aggregate_id VARCHAR(128),
			payload BYTEA NOT NULL,
			status VARCHAR(32) NOT NULL DEFAULT 'pending',
			retry_count INT NOT NULL DEFAULT 0,
			max_retries INT NOT NULL DEFAULT 5,
			last_error TEXT,
			next_retry_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			processed_at TIMESTAMP WITH TIME ZONE
		);
		CREATE INDEX IF NOT EXISTS idx_%s_poll ON %s(status, next_retry_at ASC) WHERE status IN ('pending', 'failed');
	`, s.tableName, s.tableName, s.tableName)

	_, err := s.db.ExecContext(ctx, query)
	return err
}

// Save writes an event inside the caller's active transaction (tx) or fallback connection.
func (s *SQLStore) Save(ctx context.Context, tx any, event *Event) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (id, topic, aggregate_id, payload, status, retry_count, max_retries, next_retry_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, s.tableName)

	// Check if caller passed a *sql.Tx
	if sqlTx, ok := tx.(*sql.Tx); ok && sqlTx != nil {
		_, err := sqlTx.ExecContext(ctx, query,
			event.ID, event.Topic, event.AggregateID, event.Payload,
			string(event.Status), event.RetryCount, event.MaxRetries,
			event.NextRetryAt, event.CreatedAt,
		)
		return err
	}

	// Fallback to standard DB exec
	_, err := s.db.ExecContext(ctx, query,
		event.ID, event.Topic, event.AggregateID, event.Payload,
		string(event.Status), event.RetryCount, event.MaxRetries,
		event.NextRetryAt, event.CreatedAt,
	)
	return err
}

// FetchAndLock atomically selects and locks a batch of events using PostgreSQL FOR UPDATE SKIP LOCKED.
func (s *SQLStore) FetchAndLock(ctx context.Context, batchSize int) ([]*Event, error) {
	if batchSize <= 0 {
		batchSize = 50
	}

	query := fmt.Sprintf(`
		UPDATE %s
		SET status = 'processing'
		WHERE id IN (
			SELECT id
			FROM %s
			WHERE status IN ('pending', 'failed')
			  AND next_retry_at <= NOW()
			ORDER BY next_retry_at ASC
			LIMIT $1
			FOR UPDATE SKIP LOCKED
		)
		RETURNING id, topic, aggregate_id, payload, status, retry_count, max_retries, COALESCE(last_error, ''), next_retry_at, created_at
	`, s.tableName, s.tableName)

	rows, err := s.db.QueryContext(ctx, query, batchSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*Event
	for rows.Next() {
		var e Event
		var statusStr string
		var aggID sql.NullString
		var lastErr string

		if err := rows.Scan(
			&e.ID, &e.Topic, &aggID, &e.Payload,
			&statusStr, &e.RetryCount, &e.MaxRetries,
			&lastErr, &e.NextRetryAt, &e.CreatedAt,
		); err != nil {
			return nil, err
		}

		e.Status = Status(statusStr)
		if aggID.Valid {
			e.AggregateID = aggID.String
		}
		e.LastError = lastErr
		events = append(events, &e)
	}

	return events, rows.Err()
}

// MarkCompleted marks the event as completed.
func (s *SQLStore) MarkCompleted(ctx context.Context, id string) error {
	query := fmt.Sprintf(`
		UPDATE %s
		SET status = 'completed', processed_at = NOW(), last_error = NULL
		WHERE id = $1
	`, s.tableName)

	_, err := s.db.ExecContext(ctx, query, id)
	return err
}

// MarkFailed increments retry count and sets next retry schedule.
func (s *SQLStore) MarkFailed(ctx context.Context, id string, err error, nextRetryAfter time.Duration) error {
	query := fmt.Sprintf(`
		UPDATE %s
		SET status = 'failed',
		    retry_count = retry_count + 1,
		    last_error = $2,
		    next_retry_at = NOW() + $3 * INTERVAL '1 second'
		WHERE id = $1
	`, s.tableName)

	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}

	_, qErr := s.db.ExecContext(ctx, query, id, errMsg, int(nextRetryAfter.Seconds()))
	return qErr
}

// MarkDeadLetter moves an event to dead letter status.
func (s *SQLStore) MarkDeadLetter(ctx context.Context, id string, err error) error {
	query := fmt.Sprintf(`
		UPDATE %s
		SET status = 'dead_letter',
		    last_error = $2
		WHERE id = $1
	`, s.tableName)

	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}

	_, qErr := s.db.ExecContext(ctx, query, id, errMsg)
	return qErr
}
