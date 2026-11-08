# Technical Specification: go-checksum

## 1. Domain & Purpose
`go-checksum` is a high-throughput, memory-bounded checksum calculation and file integrity verification engine in pure Go. It supports SHA-256, SHA-512, MD5, SHA-1, CRC32, and Adler-32 with pooled streaming buffer recycling, allowing multi-gigabyte files to be processed in constant <1MB RAM.

## 2. Architecture & Design Principles
- **Recycled Buffer Pooling**: Leverages `sync.Pool` 32KB byte buffers with `io.CopyBuffer` to eliminate dynamic allocations during file reads.
- **Unified Algorithm Interface**: Supports standard cryptographic and polynomial checksums through a single unified `Algorithm` enum.
- **Timing Attack Resistance**: Employs constant-time `hmac.Equal` for hex comparisons in `VerifyFile` and `VerifyBytes`.
- **Zero Heavy Dependencies**: Pure Go standard library (`crypto/sha256`, `crypto/sha512`, `crypto/md5`, `crypto/sha1`, `hash/crc32`, `hash/adler32`, `io`, `os`, `sync`).

## 3. API Surface
- `type Algorithm string` (`SHA256`, `SHA512`, `MD5`, `SHA1`, `CRC32`, `Adler32`)
- `func NewHasher(algo Algorithm) (hash.Hash, error)`
- `func FromReader(r io.Reader, algo Algorithm) (string, error)`
- `func FromBytes(data []byte, algo Algorithm) (string, error)`
- `func FromFile(filePath string, algo Algorithm) (string, error)`
- `func VerifyFile(filePath string, expectedHex string, algo Algorithm) (bool, error)`
- `func VerifyBytes(data []byte, expectedHex string, algo Algorithm) bool`

## 4. Error Handling
- `ErrUnsupportedAlgorithm`: Unknown or unmapped checksum algorithm.
- `ErrChecksumMismatch`: Verification failed.

## 5. Verification & Testing
- Unit tests verify standard IETF/NIST test vectors for SHA-256 and MD5, streaming file chunk calculation, and constant-time integrity validation.
