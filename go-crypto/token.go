package gocrypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
)

// RandomBytes generates n cryptographically secure random bytes from crypto/rand.
func RandomBytes(n int) ([]byte, error) {
	if n <= 0 {
		return nil, fmt.Errorf("crypto: byte count must be positive")
	}
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("crypto: rand read failed: %w", err)
	}
	return b, nil
}

// RandomHex generates a secure random hex string of 2*bytesLen characters.
func RandomHex(bytesLen int) (string, error) {
	b, err := RandomBytes(bytesLen)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// RandomBase64URL generates an unpadded URL-safe Base64 random token.
func RandomBase64URL(bytesLen int) (string, error) {
	b, err := RandomBytes(bytesLen)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// RandomNumericOTP generates an unbiased crypto-secure N-digit numeric OTP (e.g. 6 digits: "582914").
func RandomNumericOTP(digits int) (string, error) {
	if digits <= 0 {
		return "", fmt.Errorf("crypto: digits count must be positive")
	}

	// Max value is 10^digits
	max := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(digits)), nil)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", fmt.Errorf("crypto: generate otp failed: %w", err)
	}

	// Format with leading zeros
	format := fmt.Sprintf("%%0%dd", digits)
	return fmt.Sprintf(format, n), nil
}

// HashSHA256 computes a standard SHA-256 hash.
func HashSHA256(data []byte) []byte {
	h := sha256.Sum256(data)
	return h[:]
}

// HashSHA256Hex computes a SHA-256 hash and returns it as a lowercase hex string.
func HashSHA256Hex(data []byte) string {
	return hex.EncodeToString(HashSHA256(data))
}
