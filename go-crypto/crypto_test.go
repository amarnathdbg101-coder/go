package gocrypto

import (
	"strings"
	"testing"
	"time"
)

func TestAES256GCM_EncryptDecrypt(t *testing.T) {
	key, err := RandomBytes(32)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	cipher, err := NewAES256(key)
	if err != nil {
		t.Fatalf("NewAES256 failed: %v", err)
	}

	original := "ShopSilo-Secret-Payment-Payload-₹4999.00"
	encrypted, err := cipher.EncryptString(original)
	if err != nil {
		t.Fatalf("EncryptString failed: %v", err)
	}

	if encrypted == original {
		t.Fatal("ciphertext should not match plaintext")
	}

	decrypted, err := cipher.DecryptString(encrypted)
	if err != nil {
		t.Fatalf("DecryptString failed: %v", err)
	}

	if decrypted != original {
		t.Fatalf("expected '%s', got '%s'", original, decrypted)
	}
}

func TestAES256GCM_TamperingDetection(t *testing.T) {
	key, _ := RandomBytes(32)
	cipher, _ := NewAES256(key)

	original := []byte("Sensitive customer KYC data")
	ciphertext, err := cipher.Encrypt(original)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	// Tamper single byte in ciphertext
	ciphertext[len(ciphertext)-1] ^= 0xFF

	_, err = cipher.Decrypt(ciphertext)
	if err != ErrDecryptionFailed {
		t.Fatalf("expected ErrDecryptionFailed on tampered ciphertext, got %v", err)
	}
}

func TestAES256GCM_AAD(t *testing.T) {
	key, _ := RandomBytes(32)
	cipher, _ := NewAES256(key)

	plaintext := []byte("Transfer 10,000 INR")
	aad := []byte("User-ID:1094")

	ciphertext, err := cipher.EncryptWithAAD(plaintext, aad)
	if err != nil {
		t.Fatalf("EncryptWithAAD failed: %v", err)
	}

	// Decrypt with correct AAD
	decrypted, err := cipher.DecryptWithAAD(ciphertext, aad)
	if err != nil {
		t.Fatalf("DecryptWithAAD failed: %v", err)
	}
	if string(decrypted) != string(plaintext) {
		t.Fatalf("expected %s, got %s", plaintext, decrypted)
	}

	// Decrypt with wrong AAD should fail
	wrongAAD := []byte("User-ID:9999")
	_, err = cipher.DecryptWithAAD(ciphertext, wrongAAD)
	if err != ErrDecryptionFailed {
		t.Fatalf("expected ErrDecryptionFailed with mismatched AAD, got %v", err)
	}
}

func TestAES_InvalidKey(t *testing.T) {
	invalidKey := []byte("short-key")
	_, err := NewAES(invalidKey)
	if err == nil {
		t.Fatal("expected error for invalid key length, got nil")
	}

	_, err = NewAES256(make([]byte, 16))
	if err == nil {
		t.Fatal("expected error for non-32 byte key in NewAES256, got nil")
	}
}

func TestHMACSHA256(t *testing.T) {
	secret := []byte("webhook_secret_key_prod_994")
	payload := []byte(`{"event":"payment_success","amount":5000}`)

	sigHex := SignHMACSHA256Hex(payload, secret)
	if sigHex == "" {
		t.Fatal("HMAC signature should not be empty")
	}

	// Verify valid signature
	if !VerifyHMACSHA256Hex(payload, sigHex, secret) {
		t.Fatal("valid HMAC signature verification failed")
	}

	// Verify tampered payload fails
	tampered := []byte(`{"event":"payment_success","amount":500}`)
	if VerifyHMACSHA256Hex(tampered, sigHex, secret) {
		t.Fatal("tampered payload should fail HMAC verification")
	}

	// Verify wrong secret fails
	wrongSecret := []byte("another_secret")
	if VerifyHMACSHA256Hex(payload, sigHex, wrongSecret) {
		t.Fatal("wrong secret should fail HMAC verification")
	}
}

func TestRandomTokens(t *testing.T) {
	hexToken, err := RandomHex(16)
	if err != nil {
		t.Fatalf("RandomHex failed: %v", err)
	}
	if len(hexToken) != 32 {
		t.Fatalf("expected 32 hex chars for 16 bytes, got %d", len(hexToken))
	}

	b64Token, err := RandomBase64URL(32)
	if err != nil {
		t.Fatalf("RandomBase64URL failed: %v", err)
	}
	if len(b64Token) == 0 || strings.Contains(b64Token, "/") {
		t.Fatalf("RandomBase64URL returned invalid characters: %s", b64Token)
	}

	// Test 6-digit OTP
	otp, err := RandomNumericOTP(6)
	if err != nil {
		t.Fatalf("RandomNumericOTP failed: %v", err)
	}
	if len(otp) != 6 {
		t.Fatalf("expected 6-digit OTP, got '%s'", otp)
	}
}

func TestUUIDv7(t *testing.T) {
	id1, err := NewUUIDv7()
	if err != nil {
		t.Fatalf("NewUUIDv7 failed: %v", err)
	}

	time.Sleep(2 * time.Millisecond)

	id2, err := NewUUIDv7()
	if err != nil {
		t.Fatalf("NewUUIDv7 failed: %v", err)
	}

	// UUID string format check: 8-4-4-4-12 = 36 chars
	s1 := id1.String()
	s2 := id2.String()
	if len(s1) != 36 || len(s2) != 36 {
		t.Fatalf("invalid UUID string length: s1=%s, s2=%s", s1, s2)
	}

	// Version 7 check: 13th character (index 14) is '7'
	if s1[14] != '7' || s2[14] != '7' {
		t.Fatalf("expected version '7', got %c and %c", s1[14], s2[14])
	}

	// Time ordering check: s1 < s2 lexicographically
	if s1 >= s2 {
		t.Fatalf("UUIDv7 should be time-ordered: s1=%s, s2=%s", s1, s2)
	}

	// Parse check
	parsed, err := ParseUUID(s1)
	if err != nil {
		t.Fatalf("ParseUUID failed: %v", err)
	}
	if parsed.String() != s1 {
		t.Fatalf("parsed uuid string mismatch: expected %s, got %s", s1, parsed.String())
	}
}

func TestPBKDF2(t *testing.T) {
	passphrase := "SuperSecretPassword#2026"
	salt := []byte("unique_random_salt_99")

	key1, err := DeriveKeyPBKDF2(passphrase, salt, 1000, 32)
	if err != nil {
		t.Fatalf("DeriveKeyPBKDF2 failed: %v", err)
	}
	if len(key1) != 32 {
		t.Fatalf("expected 32 byte key, got %d", len(key1))
	}

	// Consistency test
	key2, err := DeriveKeyPBKDF2(passphrase, salt, 1000, 32)
	if err != nil {
		t.Fatalf("DeriveKeyPBKDF2 second run failed: %v", err)
	}

	for i := range key1 {
		if key1[i] != key2[i] {
			t.Fatalf("PBKDF2 must be deterministic, mismatch at byte %d", i)
		}
	}
}
