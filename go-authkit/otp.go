package authkit

import (
	"crypto/rand"
	"crypto/subtle"
	"fmt"
	"math/big"
	"sync"
	"time"
)

// otpEntry stores state for an active one-time password.
type otpEntry struct {
	code      string
	expiresAt time.Time
	attempts  int
}

// OTPManager manages in-memory, concurrency-safe OTP storage, attempt limits, and automated expiry cleanup.
type OTPManager struct {
	mu          sync.RWMutex
	store       map[string]*otpEntry
	ttl         time.Duration
	maxAttempts int
	stopChan    chan struct{}
	stopOnce    sync.Once
}

// NewOTPManager creates an OTPManager and starts the background cleanup goroutine.
func NewOTPManager(ttl time.Duration, maxAttempts int) *OTPManager {
	m := &OTPManager{
		store:       make(map[string]*otpEntry),
		ttl:         ttl,
		maxAttempts: maxAttempts,
		stopChan:    make(chan struct{}),
	}

	// Concurrency pattern: Start background cleaner goroutine to prevent memory leaks from expired OTPs.
	go m.startCleanupLoop()

	return m
}

// GenerateCode creates a cryptographically secure 6-digit numeric OTP (100000 - 999999).
// crypto/rand use kiya gaya hai taaki predictable pseudorandom attacks prevent ho sake.
func GenerateCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return "", fmt.Errorf("authkit: failed to generate crypto random OTP: %w", err)
	}
	return fmt.Sprintf("%06d", n.Int64()+100000), nil
}

// Create generates and stores a fresh OTP for the given email with configured TTL.
func (m *OTPManager) Create(email string) (string, error) {
	code, err := GenerateCode()
	if err != nil {
		return "", err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.store[email] = &otpEntry{
		code:      code,
		expiresAt: time.Now().Add(m.ttl),
		attempts:  0,
	}

	return code, nil
}

// Verify validates the user provided OTP against stored entry in constant time.
// Handles brute force attack protection: max 3 attempts ke baad lockout ho jata hai.
func (m *OTPManager) Verify(email, code string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry, exists := m.store[email]
	if !exists {
		return ErrInvalidOTP
	}

	// Expiry check
	if time.Now().After(entry.expiresAt) {
		delete(m.store, email)
		return ErrOTPExpired
	}

	// Brute force threshold check
	if entry.attempts >= m.maxAttempts {
		delete(m.store, email)
		return ErrMaxAttemptsExceeded
	}

	// Constant-time comparison to prevent timing side-channel attacks
	match := subtle.ConstantTimeCompare([]byte(entry.code), []byte(code)) == 1

	if !match {
		entry.attempts++
		if entry.attempts >= m.maxAttempts {
			delete(m.store, email)
			return ErrMaxAttemptsExceeded
		}
		return ErrInvalidOTP
	}

	// Success: Consume OTP immediately (prevent replay attacks)
	delete(m.store, email)
	return nil
}

// startCleanupLoop runs periodically to delete expired records from memory.
func (m *OTPManager) startCleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.cleanupExpired()
		case <-m.stopChan:
			return // Goroutine cleanly exits when OTPManager is closed
		}
	}
}

// cleanupExpired removes stale OTPs.
func (m *OTPManager) cleanupExpired() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for email, entry := range m.store {
		if now.After(entry.expiresAt) {
			delete(m.store, email)
		}
	}
}

// Close stops the background cleanup loop and frees memory.
func (m *OTPManager) Close() {
	m.stopOnce.Do(func() {
		close(m.stopChan)
	})
}
