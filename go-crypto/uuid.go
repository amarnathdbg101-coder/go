package gocrypto

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// UUID is a 16-byte Universally Unique Identifier.
type UUID [16]byte

var (
	uuidMu       sync.Mutex
	lastMilli    int64
	uuidSequence uint16
)

// NewUUIDv7 generates an RFC 9562 compliant UUID version 7.
// It is time-ordered with millisecond precision, ensuring optimal B-tree index locality in PostgreSQL.
func NewUUIDv7() (UUID, error) {
	var u UUID

	// Step 1: Read 10 bytes of cryptographically secure random numbers for rand_a (12 bits) and rand_b (62 bits)
	randomBytes := make([]byte, 10)
	if _, err := rand.Read(randomBytes); err != nil {
		return u, fmt.Errorf("crypto: uuid rand failed: %w", err)
	}

	uuidMu.Lock()
	now := time.Now().UnixMilli()
	if now <= lastMilli {
		// Monotonic clock tick within same millisecond
		uuidSequence++
	} else {
		lastMilli = now
		uuidSequence = 0
	}
	seq := uuidSequence
	uuidMu.Unlock()

	// 48-bit millisecond timestamp (bytes 0..5)
	u[0] = byte(now >> 40)
	u[1] = byte(now >> 32)
	u[2] = byte(now >> 24)
	u[3] = byte(now >> 16)
	u[4] = byte(now >> 8)
	u[5] = byte(now)

	// Byte 6: 4-bit version (0111 for v7) + high 4 bits of sequence/rand
	u[6] = 0x70 | byte((seq>>8)&0x0F)
	// Byte 7: low 8 bits of sequence/rand
	u[7] = byte(seq & 0xFF)
	// If sequence is 0, mix with random bytes to prevent predictable entropy
	if seq == 0 {
		u[6] |= (randomBytes[0] & 0x0F)
		u[7] ^= randomBytes[1]
	}

	// Byte 8: 2-bit variant (10xxxxxx) + 6 bits random
	u[8] = 0x80 | (randomBytes[2] & 0x3F)

	// Bytes 9..15: remaining 7 random bytes
	copy(u[9:], randomBytes[3:])

	return u, nil
}

// String returns the canonical 36-character hexadecimal UUID string.
// Example: 018e6931-92be-7b9c-8517-5735b54a7c1b
func (u UUID) String() string {
	buf := make([]byte, 36)
	hex.Encode(buf[0:8], u[0:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], u[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], u[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], u[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:36], u[10:16])
	return string(buf)
}

// ParseUUID parses a canonical 36-character UUID string into a UUID byte array.
func ParseUUID(s string) (UUID, error) {
	var u UUID
	if len(s) != 36 || s[8] != '-' || s[13] != '-' || s[18] != '-' || s[23] != '-' {
		return u, fmt.Errorf("crypto: invalid uuid format")
	}

	clean := s[0:8] + s[9:13] + s[14:18] + s[19:23] + s[24:36]
	b, err := hex.DecodeString(clean)
	if err != nil {
		return u, fmt.Errorf("crypto: invalid hex uuid: %w", err)
	}

	copy(u[:], b)
	return u, nil
}
