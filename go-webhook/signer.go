package gowebhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Sign generates a standard HMAC-SHA256 signature string for webhook payloads.
// Format returned: "t=1700000000,v1=hexsignature"
func Sign(secret []byte, payload []byte, timestamp int64) (string, error) {
	if len(secret) == 0 {
		return "", ErrEmptySecret
	}
	if timestamp == 0 {
		timestamp = time.Now().Unix()
	}

	sig := computeHMAC(secret, payload, timestamp)
	return fmt.Sprintf("t=%d,v1=%s", timestamp, sig), nil
}

// Verify validates a webhook signature header against expected secret and payload.
// If tolerance > 0, checks if the timestamp is within [now - tolerance, now + tolerance].
func Verify(secret []byte, payload []byte, signatureHeader string, tolerance time.Duration) error {
	if len(secret) == 0 {
		return ErrEmptySecret
	}

	timestamp, signatures, err := parseHeader(signatureHeader)
	if err != nil {
		return err
	}

	// Verify timestamp anti-replay
	if tolerance > 0 {
		now := time.Now().Unix()
		diff := now - timestamp
		if diff < 0 {
			diff = -diff
		}
		if time.Duration(diff)*time.Second > tolerance {
			return ErrSignatureExpired
		}
	}

	expectedSig := computeHMAC(secret, payload, timestamp)
	expectedBytes, err := hex.DecodeString(expectedSig)
	if err != nil {
		return ErrInvalidSignature
	}

	for _, sigHex := range signatures {
		candidateBytes, err := hex.DecodeString(sigHex)
		if err != nil {
			continue
		}
		// Constant-time comparison to prevent timing attacks
		if hmac.Equal(expectedBytes, candidateBytes) {
			return nil
		}
	}

	return ErrInvalidSignature
}

func computeHMAC(secret []byte, payload []byte, timestamp int64) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(strconv.FormatInt(timestamp, 10)))
	mac.Write([]byte("."))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

func parseHeader(header string) (int64, []string, error) {
	parts := strings.Split(header, ",")
	var timestamp int64
	var signatures []string

	for _, part := range parts {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := kv[0]
		val := kv[1]

		switch key {
		case "t":
			ts, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return 0, nil, ErrMalformedHeader
			}
			timestamp = ts
		case "v1":
			signatures = append(signatures, val)
		}
	}

	if timestamp == 0 || len(signatures) == 0 {
		return 0, nil, ErrMalformedHeader
	}

	return timestamp, signatures, nil
}
