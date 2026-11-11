package authkit

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"sync"
	"time"
)

var validTableNameRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// ValidateTableName ensures the SQL table identifier contains only safe alphanumeric and underscore characters,
// preventing dynamic SQL injection attacks.
func ValidateTableName(name string) error {
	if !validTableNameRegex.MatchString(name) {
		return fmt.Errorf("%w: '%s'", ErrInvalidTableName, name)
	}
	return nil
}

// UserStore defines the persistence contract required by go-authkit.
// Kisi bhi storage layer (PostgreSQL, MySQL, SQLite, MongoDB, or In-Memory) ke liye
// ye interface implement karke plug and play kiya ja sakta hai.
type UserStore interface {
	// Create persists a new user record. Returns ErrEmailAlreadyExists if email is taken.
	Create(ctx context.Context, user *User) error

	// GetByEmail retrieves a user by exact email address. Returns ErrUserNotFound if missing.
	GetByEmail(ctx context.Context, email string) (*User, error)

	// GetByID retrieves a user by unique identifier. Returns ErrUserNotFound if missing.
	GetByID(ctx context.Context, id string) (*User, error)

	// UpdatePassword updates the user's password hash in persistent storage.
	UpdatePassword(ctx context.Context, id string, newPasswordHash string) error
}

// -------------------------------------------------------------------------
// In-Memory Implementation (Zero-dependency testing & prototyping)
// -------------------------------------------------------------------------

// MemoryUserStore is an in-memory, thread-safe implementation of UserStore.
// Perfect for unit tests and rapid local prototypes without installing databases.
type MemoryUserStore struct {
	mu      sync.RWMutex
	users   map[string]*User  // key: ID
	byEmail map[string]string // key: Email -> value: ID
	counter int64
}

// NewMemoryUserStore initializes an empty in-memory store.
func NewMemoryUserStore() *MemoryUserStore {
	return &MemoryUserStore{
		users:   make(map[string]*User),
		byEmail: make(map[string]string),
	}
}

// Create inserts user into memory.
func (m *MemoryUserStore) Create(ctx context.Context, user *User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.byEmail[user.Email]; exists {
		return ErrEmailAlreadyExists
	}

	if user.ID == "" {
		m.counter++
		user.ID = fmt.Sprintf("usr_%d", m.counter)
	}

	now := time.Now()
	if user.CreatedAt.IsZero() {
		user.CreatedAt = now
	}
	user.UpdatedAt = now

	// Clone user to avoid external mutations
	cloned := *user
	m.users[user.ID] = &cloned
	m.byEmail[user.Email] = user.ID
	return nil
}

// GetByEmail retrieves user by email.
func (m *MemoryUserStore) GetByEmail(ctx context.Context, email string) (*User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	id, exists := m.byEmail[email]
	if !exists {
		return nil, ErrUserNotFound
	}

	user := m.users[id]
	cloned := *user
	return &cloned, nil
}

// GetByID retrieves user by ID.
func (m *MemoryUserStore) GetByID(ctx context.Context, id string) (*User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	user, exists := m.users[id]
	if !exists {
		return nil, ErrUserNotFound
	}

	cloned := *user
	return &cloned, nil
}

// UpdatePassword updates password hash in memory.
func (m *MemoryUserStore) UpdatePassword(ctx context.Context, id string, newPasswordHash string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	user, exists := m.users[id]
	if !exists {
		return ErrUserNotFound
	}

	user.PasswordHash = newPasswordHash
	user.UpdatedAt = time.Now()
	return nil
}

// -------------------------------------------------------------------------
// PostgreSQL Implementation (Production Ready using standard database/sql)
// -------------------------------------------------------------------------

// PostgresUserStore is a production-grade PostgreSQL implementation using database/sql.
type PostgresUserStore struct {
	db        *sql.DB
	tableName string
}

// PostgresOptions allows configuring PostgreSQL store options.
type PostgresOptions struct {
	TableName string // Default: "users"
}

// NewPostgresUserStore creates a new PostgreSQL store backed by standard *sql.DB.
// Validates TableName against a strict identifier whitelist to prevent SQL injection.
func NewPostgresUserStore(db *sql.DB, opts *PostgresOptions) *PostgresUserStore {
	tableName := "users"
	if opts != nil && opts.TableName != "" {
		if err := ValidateTableName(opts.TableName); err == nil {
			tableName = opts.TableName
		}
	}
	return &PostgresUserStore{
		db:        db,
		tableName: tableName,
	}
}

// TableName returns the active database table name.
func (s *PostgresUserStore) TableName() string {
	return s.tableName
}

// InitSchema automatically provisions the users table and indexes in PostgreSQL if not already present.
func (s *PostgresUserStore) InitSchema(ctx context.Context) error {
	if s.db == nil {
		return errors.New("authkit: postgres db connection is nil")
	}

	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id VARCHAR(64) PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			role VARCHAR(64) NOT NULL,
			created_at TIMESTAMPTZ NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_%s_email ON %s(email);
	`, s.tableName, s.tableName, s.tableName)

	_, err := s.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("authkit: postgres init schema failed: %w", err)
	}
	return nil
}

// Create inserts a user into PostgreSQL.
func (s *PostgresUserStore) Create(ctx context.Context, user *User) error {
	now := time.Now()
	if user.CreatedAt.IsZero() {
		user.CreatedAt = now
	}
	user.UpdatedAt = now

	if user.ID == "" {
		user.ID = generateUniqueID()
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (id, email, password_hash, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, s.tableName)

	_, err := s.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.Role,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		// Postgres duplicate key error code 23505 check (or general unique violation)
		if isUniqueViolation(err) {
			return ErrEmailAlreadyExists
		}
		return fmt.Errorf("authkit: postgres insert failed: %w", err)
	}
	return nil
}

// GetByEmail retrieves user by email from PostgreSQL.
func (s *PostgresUserStore) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := fmt.Sprintf(`
		SELECT id, email, password_hash, role, created_at, updated_at
		FROM %s
		WHERE email = $1
		LIMIT 1
	`, s.tableName)

	row := s.db.QueryRowContext(ctx, query, email)

	var u User
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("authkit: postgres query failed: %w", err)
	}
	return &u, nil
}

// GetByID retrieves user by ID from PostgreSQL.
func (s *PostgresUserStore) GetByID(ctx context.Context, id string) (*User, error) {
	query := fmt.Sprintf(`
		SELECT id, email, password_hash, role, created_at, updated_at
		FROM %s
		WHERE id = $1
		LIMIT 1
	`, s.tableName)

	row := s.db.QueryRowContext(ctx, query, id)

	var u User
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("authkit: postgres query failed: %w", err)
	}
	return &u, nil
}

// UpdatePassword updates the password_hash in PostgreSQL.
func (s *PostgresUserStore) UpdatePassword(ctx context.Context, id string, newPasswordHash string) error {
	query := fmt.Sprintf(`
		UPDATE %s
		SET password_hash = $1, updated_at = $2
		WHERE id = $3
	`, s.tableName)

	result, err := s.db.ExecContext(ctx, query, newPasswordHash, time.Now(), id)
	if err != nil {
		return fmt.Errorf("authkit: postgres update password failed: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrUserNotFound
	}
	return nil
}

// isUniqueViolation checks whether the error is a PostgreSQL unique constraint violation.
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	// PostgreSQL unique_violation error code is 23505
	return contains(err.Error(), "23505") || contains(err.Error(), "unique constraint") || contains(err.Error(), "duplicate key")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > 0 && searchSubstr(s, substr)))
}

func searchSubstr(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// generateUniqueID creates a crypto-secure, timestamp-ordered, collision-free user ID.
func generateUniqueID() string {
	var b [8]byte
	_, err := rand.Read(b[:])
	if err != nil {
		return fmt.Sprintf("usr_%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("usr_%d_%s", time.Now().UnixNano(), hex.EncodeToString(b[:]))
}
