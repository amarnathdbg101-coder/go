package gocrypto

import "errors"

var (
	// ErrInvalidKeyLength occurs when an AES key is not 16, 24, or 32 bytes.
	ErrInvalidKeyLength = errors.New("crypto: invalid key size (must be 16, 24, or 32 bytes for AES)")

	// ErrCiphertextTooShort occurs when encrypted data is shorter than nonce size.
	ErrCiphertextTooShort = errors.New("crypto: ciphertext too short")

	// ErrDecryptionFailed occurs when authentication tag mismatch or payload corrupted.
	ErrDecryptionFailed = errors.New("crypto: decryption failed or data tampered")

	// ErrInvalidHMAC occurs when HMAC signature verification fails.
	ErrInvalidHMAC = errors.New("crypto: invalid hmac signature")

	// ErrInvalidIterations occurs when KDF iteration count is too low.
	ErrInvalidIterations = errors.New("crypto: kdf iterations must be at least 1000")
)
