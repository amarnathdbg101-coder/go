package gocrypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
)

// SignHMACSHA256 generates a 32-byte cryptographic HMAC-SHA256 signature.
func SignHMACSHA256(message, secret []byte) []byte {
	h := hmac.New(sha256.New, secret)
	h.Write(message)
	return h.Sum(nil)
}

// SignHMACSHA256Hex generates a lowercase hex-encoded HMAC-SHA256 signature (standard for webhooks like Stripe/Razorpay/GitHub).
func SignHMACSHA256Hex(message, secret []byte) string {
	sig := SignHMACSHA256(message, secret)
	return hex.EncodeToString(sig)
}

// VerifyHMACSHA256 validates a signature using constant-time comparison to prevent timing side-channel attacks.
func VerifyHMACSHA256(message, signature, secret []byte) bool {
	expected := SignHMACSHA256(message, secret)
	return subtle.ConstantTimeCompare(expected, signature) == 1
}

// VerifyHMACSHA256Hex verifies a hex-encoded HMAC signature in constant time.
func VerifyHMACSHA256Hex(message []byte, hexSignature string, secret []byte) bool {
	sigBytes, err := hex.DecodeString(hexSignature)
	if err != nil {
		return false
	}
	return VerifyHMACSHA256(message, sigBytes, secret)
}
