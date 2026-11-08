package gocrypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

const (
	// DefaultPBKDF2Iterations recommended for general password-to-key derivation.
	DefaultPBKDF2Iterations = 100000
)

// DeriveKeyPBKDF2 derives a cryptographic key of specified length using PBKDF2-HMAC-SHA256 (RFC 8018).
// Zero external dependencies, pure standard library.
func DeriveKeyPBKDF2(passphrase string, salt []byte, iterations, keyLen int) ([]byte, error) {
	if iterations < 1000 {
		return nil, ErrInvalidIterations
	}
	if keyLen <= 0 {
		return nil, fmt.Errorf("crypto: key length must be positive")
	}

	prf := hmac.New(sha256.New, []byte(passphrase))
	hashLen := prf.Size()
	numBlocks := (keyLen + hashLen - 1) / hashLen

	var (
		derivedKey = make([]byte, 0, numBlocks*hashLen)
		u          = make([]byte, hashLen)
		f          = make([]byte, hashLen)
		blockIndex = make([]byte, 4)
	)

	for block := 1; block <= numBlocks; block++ {
		prf.Reset()
		prf.Write(salt)
		binary.BigEndian.PutUint32(blockIndex, uint32(block))
		prf.Write(blockIndex)
		copy(u, prf.Sum(nil))
		copy(f, u)

		for iter := 1; iter < iterations; iter++ {
			prf.Reset()
			prf.Write(u)
			copy(u, prf.Sum(nil))
			for i := 0; i < hashLen; i++ {
				f[i] ^= u[i]
			}
		}

		derivedKey = append(derivedKey, f...)
	}

	return derivedKey[:keyLen], nil
}
