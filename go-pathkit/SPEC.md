# SPEC: go-pathkit

## 1. Metadata
- **Module**: `go-pathkit`
- **Package**: `pathkit`
- **Location**: `C:\Library\go-pathkit`
- **Go Version**: `1.23+`
- **Dependencies**: Zero external dependencies (100% Go Standard Library: `os`, `path/filepath`, `crypto/sha256`, `crypto/md5`, `crypto/rand`, `encoding/hex`, `errors`, `fmt`, `io`, `strings`).

---

## 2. Core Capabilities
1. **Path Traversal Defense**:
   - `SafeJoin(baseDir, untrustedPath)` validates relative escapes and guarantees that target files remain strictly jailed inside `baseDir`.
2. **Crash-Proof Atomic File Operations**:
   - `WriteFileAtomic(filename, data, perm)` writes to a temporary file in the destination folder, flushes kernel buffers via `Sync()`, and performs an atomic filesystem replace via `os.Rename`.
3. **Low-Memory Streaming File Hashes**:
   - `SHA256` and `MD5` compute cryptographic checksums via streaming buffers without full file heap allocation.
4. **Filtered Directory Tree Traversal**:
   - `Walk(root, opts, fn)` supports depth limits, directory skips (e.g. `node_modules`, `.git`), and extension filtering.

---

## 3. Public API
```go
func SafeJoin(baseDir, untrustedPath string) (string, error)
func WriteFileAtomic(filename string, data []byte, perm os.FileMode) error
func SHA256(filePath string) (string, error)
func MD5(filePath string) (string, error)
func EnsureDir(dirPath string) error
func Exists(path string) bool

type WalkOptions struct {
    MaxDepth   int
    Extensions []string
    IgnoreDirs []string
    MinSize    int64
    MaxSize    int64
}

func Walk(root string, opts WalkOptions, fn func(path string, info os.FileInfo) error) error
```
