package gocrypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

// AESCipher provides high-performance AES-GCM authenticated symmetric encryption.
type AESCipher struct {
	key []byte
}

// NewAES initializes an AES cipher with a 16, 24, or 32 byte key.
func NewAES(key []byte) (*AESCipher, error) {
	keyLen := len(key)
	if keyLen != 16 && keyLen != 24 && keyLen != 32 {
		return nil, fmt.Errorf("%w: got %d bytes", ErrInvalidKeyLength, keyLen)
	}
	keyCopy := make([]byte, keyLen)
	copy(keyCopy, key)
	return &AESCipher{key: keyCopy}, nil
}

// NewAES256 initializes AES strictly requiring a 32-byte (256-bit) key for military-grade security.
func NewAES256(key []byte) (*AESCipher, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("%w: AES-256 requires exactly 32 bytes, got %d", ErrInvalidKeyLength, len(key))
	}
	return NewAES(key)
}

// Encrypt encrypts plaintext using AES-GCM and prepends a cryptographically secure 12-byte nonce.
// Output format: [ 12-byte Nonce | Ciphertext | 16-byte Auth Tag ]
func (c *AESCipher) Encrypt(plaintext []byte) ([]byte, error) {
	return c.EncryptWithAAD(plaintext, nil)
}

// EncryptWithAAD encrypts plaintext with Additional Authenticated Data (AAD).
func (c *AESCipher) EncryptWithAAD(plaintext, aad []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, fmt.Errorf("crypto: create aes block: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("crypto: create gcm: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("crypto: generate nonce: %w", err)
	}

	// Seal appends ciphertext + auth tag to nonce
	ciphertext := gcm.Seal(nonce, nonce, plaintext, aad)
	return ciphertext, nil
}

// Decrypt authenticates and decrypts AES-GCM ciphertext created by Encrypt.
func (c *AESCipher) Decrypt(ciphertextWithNonce []byte) ([]byte, error) {
	return c.DecryptWithAAD(ciphertextWithNonce, nil)
}

// DecryptWithAAD authenticates and decrypts AES-GCM ciphertext using Additional Authenticated Data.
func (c *AESCipher) DecryptWithAAD(ciphertextWithNonce, aad []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, fmt.Errorf("crypto: create aes block: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("crypto: create gcm: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertextWithNonce) < nonceSize {
		return nil, ErrCiphertextTooShort
	}

	nonce, ciphertext := ciphertextWithNonce[:nonceSize], ciphertextWithNonce[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, aad)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	return plaintext, nil
}

// EncryptString encrypts a UTF-8 string and returns a URL-safe Base64 string.
func (c *AESCipher) EncryptString(plaintext string) (string, error) {
	encrypted, err := c.Encrypt([]byte(plaintext))
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(encrypted), nil
}

// DecryptString decodes a URL-safe Base64 string and decrypts it back to plain text.
func (c *AESCipher) DecryptString(base64Ciphertext string) (string, error) {
	data, err := base64.RawURLEncoding.DecodeString(base64Ciphertext)
	if err != nil {
		// Fallback to standard Base64 if not RawURL
		var stdErr error
		data, stdErr = base64.StdEncoding.DecodeString(base64Ciphertext)
		if stdErr != nil {
			return "", fmt.Errorf("crypto: invalid base64 payload: %w", err)
		}
	}

	decrypted, err := c.Decrypt(data)
	if err != nil {
		return "", err
	}
	return string(decrypted), nil
}
