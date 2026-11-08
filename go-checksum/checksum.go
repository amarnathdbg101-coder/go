package gochecksum

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"hash"
	"hash/adler32"
	"hash/crc32"
	"io"
	"os"
	"strings"
	"sync"
)

// Algorithm defines supported cryptographic and non-cryptographic checksum algorithms.
type Algorithm string

const (
	SHA256  Algorithm = "SHA256"
	SHA512  Algorithm = "SHA512"
	MD5     Algorithm = "MD5"
	SHA1    Algorithm = "SHA1"
	CRC32   Algorithm = "CRC32"
	Adler32 Algorithm = "ADLER32"
)

var bufPool = sync.Pool{
	New: func() any {
		buf := make([]byte, 32*1024) // 32 KB chunks
		return &buf
	},
}

// NewHasher returns a new hash.Hash for the specified algorithm.
func NewHasher(algo Algorithm) (hash.Hash, error) {
	switch strings.ToUpper(string(algo)) {
	case "SHA256":
		return sha256.New(), nil
	case "SHA512":
		return sha512.New(), nil
	case "MD5":
		return md5.New(), nil
	case "SHA1":
		return sha1.New(), nil
	case "CRC32":
		return crc32.NewIEEE(), nil
	case "ADLER32":
		return adler32.New(), nil
	default:
		return nil, ErrUnsupportedAlgorithm
	}
}

// FromReader calculates hex-encoded checksum from any io.Reader using recycled buffer.
func FromReader(r io.Reader, algo Algorithm) (string, error) {
	hasher, err := NewHasher(algo)
	if err != nil {
		return "", err
	}

	bufPtr := bufPool.Get().(*[]byte)
	defer bufPool.Put(bufPtr)

	if _, err := io.CopyBuffer(hasher, r, *bufPtr); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// FromBytes calculates hex-encoded checksum for an in-memory byte slice.
func FromBytes(data []byte, algo Algorithm) (string, error) {
	hasher, err := NewHasher(algo)
	if err != nil {
		return "", err
	}
	hasher.Write(data)
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// FromFile streams a file from disk and calculates its checksum with constant low memory footprint.
func FromFile(filePath string, algo Algorithm) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	return FromReader(f, algo)
}

// VerifyFile verifies that file matches expected hex checksum using constant-time comparison.
func VerifyFile(filePath string, expectedHex string, algo Algorithm) (bool, error) {
	actualHex, err := FromFile(filePath, algo)
	if err != nil {
		return false, err
	}
	return constantTimeCompare(actualHex, expectedHex), nil
}

// VerifyBytes verifies that byte slice matches expected hex checksum.
func VerifyBytes(data []byte, expectedHex string, algo Algorithm) bool {
	actualHex, err := FromBytes(data, algo)
	if err != nil {
		return false
	}
	return constantTimeCompare(actualHex, expectedHex)
}

func constantTimeCompare(a, b string) bool {
	a = strings.ToLower(strings.TrimSpace(a))
	b = strings.ToLower(strings.TrimSpace(b))

	bytesA, errA := hex.DecodeString(a)
	bytesB, errB := hex.DecodeString(b)
	if errA != nil || errB != nil {
		return false
	}

	return hmac.Equal(bytesA, bytesB)
}
