package pathkit

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// SafeJoin joins untrusted user paths to a baseDir and guarantees that the resulting path
// is strictly a subpath of baseDir. Protects against ../ traversal attacks.
func SafeJoin(baseDir, untrustedPath string) (string, error) {
	if baseDir == "" || untrustedPath == "" {
		return "", ErrEmptyPath
	}

	absBase, err := filepath.Abs(filepath.Clean(baseDir))
	if err != nil {
		return "", err
	}

	// Clean target path and join
	cleanUntrusted := filepath.Clean(untrustedPath)
	target := filepath.Join(absBase, cleanUntrusted)
	absTarget, err := filepath.Abs(target)
	if err != nil {
		return "", err
	}

	// Ensure absTarget starts with absBase
	rel, err := filepath.Rel(absBase, absTarget)
	if err != nil || strings.HasPrefix(rel, "..") || rel == "." && cleanUntrusted != "." && cleanUntrusted != "" {
		if strings.HasPrefix(rel, "..") {
			return "", fmt.Errorf("%w: path '%s' escapes '%s'", ErrPathTraversal, untrustedPath, baseDir)
		}
	}

	return absTarget, nil
}

// WriteFileAtomic safely writes data to a temporary file in the target directory,
// flushes buffers to physical storage via Sync(), closes the file, and renames it atomically.
// Prevents half-written or corrupted files during crashes or power cuts.
func WriteFileAtomic(filename string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Generate random suffix for temp file in same directory
	var b [8]byte
	_, _ = rand.Read(b[:])
	tmpName := filepath.Join(dir, fmt.Sprintf(".tmp_%s_%s", filepath.Base(filename), hex.EncodeToString(b[:])))

	f, err := os.OpenFile(tmpName, os.O_WRONLY|os.O_CREATE|os.O_EXCL, perm)
	if err != nil {
		return err
	}

	// Clean up temp file if rename doesn't complete
	closed := false
	defer func() {
		if !closed {
			_ = f.Close()
			_ = os.Remove(tmpName)
		}
	}()

	if _, err := f.Write(data); err != nil {
		return err
	}

	// Force write to disk
	if err := f.Sync(); err != nil {
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}
	closed = true

	// Atomic replace
	return os.Rename(tmpName, filename)
}

// SHA256 streams the file content and calculates its SHA-256 hexadecimal hash.
func SHA256(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// MD5 streams the file content and calculates its MD5 hexadecimal hash.
func MD5(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// WalkOptions configures directory walking filters.
type WalkOptions struct {
	MaxDepth   int
	Extensions []string // e.g. [".go", ".json"]
	IgnoreDirs []string // e.g. [".git", "node_modules", "vendor"]
	MinSize    int64
	MaxSize    int64
}

// Walk traverses directory hierarchy starting at root applying the configured filters.
func Walk(root string, opts WalkOptions, fn func(path string, info os.FileInfo) error) error {
	root = filepath.Clean(root)
	rootDepth := strings.Count(root, string(os.PathSeparator))

	extMap := make(map[string]bool)
	for _, ext := range opts.Extensions {
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		extMap[strings.ToLower(ext)] = true
	}

	ignoreMap := make(map[string]bool)
	for _, dir := range opts.IgnoreDirs {
		ignoreMap[dir] = true
	}

	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if path != root && ignoreMap[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		// Depth filter
		if opts.MaxDepth > 0 {
			depth := strings.Count(path, string(os.PathSeparator)) - rootDepth
			if depth > opts.MaxDepth {
				return nil
			}
		}

		// Extension filter
		if len(extMap) > 0 {
			ext := strings.ToLower(filepath.Ext(path))
			if !extMap[ext] {
				return nil
			}
		}

		// Size filter
		if opts.MinSize > 0 && info.Size() < opts.MinSize {
			return nil
		}
		if opts.MaxSize > 0 && info.Size() > opts.MaxSize {
			return nil
		}

		return fn(path, info)
	})
}

// EnsureDir checks if a directory exists, creating it if necessary.
func EnsureDir(dirPath string) error {
	return os.MkdirAll(dirPath, 0755)
}

// Exists checks whether a file or directory exists.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || !os.IsNotExist(err)
}
