package gochecksum

import (
	"os"
	"path/filepath"
	"testing"
)

func TestChecksum_KnownVectors(t *testing.T) {
	data := []byte("hello world")

	// SHA256
	expectedSHA256 := "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
	actualSHA256, err := FromBytes(data, SHA256)
	if err != nil || actualSHA256 != expectedSHA256 {
		t.Errorf("SHA256 expected %s, got %s (err %v)", expectedSHA256, actualSHA256, err)
	}

	// MD5
	expectedMD5 := "5eb63bbbe01eeed093cb22bb8f5acdc3"
	actualMD5, err := FromBytes(data, MD5)
	if err != nil || actualMD5 != expectedMD5 {
		t.Errorf("MD5 expected %s, got %s (err %v)", expectedMD5, actualMD5, err)
	}

	// VerifyBytes
	if !VerifyBytes(data, expectedSHA256, SHA256) {
		t.Errorf("VerifyBytes failed for valid SHA256")
	}
	if VerifyBytes(data, "invalidhex", SHA256) {
		t.Errorf("VerifyBytes passed for invalid hex")
	}
}

func TestChecksum_StreamingFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")
	content := []byte("streaming checksum test payload")

	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	expectedSHA256, _ := FromBytes(content, SHA256)

	actualHex, err := FromFile(filePath, SHA256)
	if err != nil {
		t.Fatalf("FromFile failed: %v", err)
	}
	if actualHex != expectedSHA256 {
		t.Errorf("expected %s, got %s", expectedSHA256, actualHex)
	}

	valid, err := VerifyFile(filePath, expectedSHA256, SHA256)
	if err != nil || !valid {
		t.Errorf("VerifyFile failed: valid=%v, err=%v", valid, err)
	}
}

func TestChecksum_Unsupported(t *testing.T) {
	_, err := FromBytes([]byte("test"), "UNKNOWN_ALGO")
	if err != ErrUnsupportedAlgorithm {
		t.Errorf("expected ErrUnsupportedAlgorithm, got %v", err)
	}
}
