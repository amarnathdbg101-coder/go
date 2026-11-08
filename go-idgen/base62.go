package goidgen

import (
	"strings"
)

const base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// EncodeBase62 encodes a 64-bit integer into a compact Base62 alphanumeric string.
func EncodeBase62(n uint64) string {
	if n == 0 {
		return "0"
	}

	var buf [11]byte
	i := len(buf)

	for n > 0 {
		i--
		buf[i] = base62Chars[n%62]
		n /= 62
	}

	return string(buf[i:])
}

// DecodeBase62 decodes a Base62 string back into an unsigned 64-bit integer.
func DecodeBase62(s string) (uint64, error) {
	if s == "" {
		return 0, ErrInvalidBase62
	}

	var n uint64
	for i := 0; i < len(s); i++ {
		c := s[i]
		idx := strings.IndexByte(base62Chars, c)
		if idx == -1 {
			return 0, ErrInvalidBase62
		}
		n = n*62 + uint64(idx)
	}

	return n, nil
}
