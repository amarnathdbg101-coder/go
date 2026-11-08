package gosecureid

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/bits"
	"strings"
)

const (
	nanoIDAlphabet    = "_-0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	crockfordAlphabet = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"
)

// Token generates a cryptographically secure URL-safe unpadded Base64 token.
func Token(byteLength int) (string, error) {
	if byteLength <= 0 {
		return "", ErrInvalidLength
	}

	buf := make([]byte, byteLength)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// APIKey generates a prefixed API credential (e.g. "sk_live_...", "pk_test_...").
func APIKey(prefix string, byteLength int) (string, error) {
	tok, err := Token(byteLength)
	if err != nil {
		return "", err
	}

	if prefix == "" {
		return tok, nil
	}

	prefix = strings.TrimSuffix(prefix, "_")
	return fmt.Sprintf("%s_%s", prefix, tok), nil
}

// NanoID generates a collision-resistant URL-safe identifier of specified length (default 21).
func NanoID(length int) (string, error) {
	if length <= 0 {
		length = 21
	}
	return Custom(nanoIDAlphabet, length)
}

// InviteCode generates a human-readable, unambiguous uppercase code using Crockford Base32.
func InviteCode(length int) (string, error) {
	if length <= 0 {
		length = 8
	}
	return Custom(crockfordAlphabet, length)
}

// Custom generates a secure identifier using an arbitrary alphabet with mask-based modulo bias prevention.
func Custom(alphabet string, length int) (string, error) {
	if length <= 0 {
		return "", ErrInvalidLength
	}
	if len(alphabet) < 2 {
		return "", ErrEmptyAlphabet
	}

	alphabetLen := len(alphabet)
	// Calculate bit mask to avoid modulo bias
	mask := (2 << (31 - bits.LeadingZeros32(uint32(alphabetLen-1)))) - 1
	step := (length * mask * 8) / (alphabetLen * 5)
	if step < 16 {
		step = 16
	}

	result := make([]byte, length)
	count := 0

	for {
		randomBytes := make([]byte, step)
		if _, err := rand.Read(randomBytes); err != nil {
			return "", err
		}

		for _, b := range randomBytes {
			idx := int(b) & mask
			if idx < alphabetLen {
				result[count] = alphabet[idx]
				count++
				if count == length {
					return string(result), nil
				}
			}
		}
	}
}
