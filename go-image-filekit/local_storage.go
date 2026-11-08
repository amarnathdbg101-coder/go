package filekit

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// LocalStorage stores files on the local filesystem.
// Production-ready path traversal protection aur auto-directory creation ke saath built hai.
type LocalStorage struct {
	baseDir string
	baseURL string
}

// NewLocalStorage initializes local disk storage.
// baseDir: local directory path (e.g. "./uploads" or "/var/data/uploads")
// baseURL: public URL prefix (e.g. "/uploads" or "https://api.myapp.com/uploads")
func NewLocalStorage(baseDir, baseURL string) (*LocalStorage, error) {
	cleanBase := filepath.Clean(baseDir)
	if err := os.MkdirAll(cleanBase, 0755); err != nil {
		return nil, fmt.Errorf("filekit: failed to create base storage directory: %w", err)
	}

	return &LocalStorage{
		baseDir: cleanBase,
		baseURL: strings.TrimRight(baseURL, "/"),
	}, nil
}

// Save writes incoming stream to disk safely.
func (s *LocalStorage) Save(ctx context.Context, relativePath string, r io.Reader, size int64, contentType string) (string, error) {
	fullPath, err := s.resolveSafePath(relativePath)
	if err != nil {
		return "", err
	}

	// Ensure parent directories exist
	parentDir := filepath.Dir(fullPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return "", fmt.Errorf("filekit: failed to create parent directory: %w", err)
	}

	// Create destination file
	dst, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return "", fmt.Errorf("filekit: failed to create file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, r); err != nil {
		return "", fmt.Errorf("filekit: failed to write file data: %w", err)
	}

	// Construct public URL
	cleanRel := strings.TrimPrefix(filepath.ToSlash(relativePath), "/")
	publicURL := fmt.Sprintf("%s/%s", s.baseURL, cleanRel)
	return publicURL, nil
}

// Open retrieves a readable stream for the requested file.
func (s *LocalStorage) Open(ctx context.Context, relativePath string) (io.ReadCloser, error) {
	fullPath, err := s.resolveSafePath(relativePath)
	if err != nil {
		return nil, err
	}
	return os.Open(fullPath)
}

// Delete removes the file from disk.
func (s *LocalStorage) Delete(ctx context.Context, relativePath string) error {
	fullPath, err := s.resolveSafePath(relativePath)
	if err != nil {
		return err
	}
	err = os.Remove(fullPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// Exists checks if file is present on disk.
func (s *LocalStorage) Exists(ctx context.Context, relativePath string) (bool, error) {
	fullPath, err := s.resolveSafePath(relativePath)
	if err != nil {
		return false, err
	}
	_, err = os.Stat(fullPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// resolveSafePath verifies and prevents directory traversal attacks (e.g. "../../etc/passwd").
func (s *LocalStorage) resolveSafePath(relativePath string) (string, error) {
	cleanRel := filepath.Clean(relativePath)

	// Check for path traversal attempts
	if strings.HasPrefix(cleanRel, "..") || filepath.IsAbs(cleanRel) {
		return "", ErrPathTraversal
	}

	fullPath := filepath.Join(s.baseDir, cleanRel)

	// Final security boundary check
	relCheck, err := filepath.Rel(s.baseDir, fullPath)
	if err != nil || strings.HasPrefix(relCheck, "..") {
		return "", ErrPathTraversal
	}

	return fullPath, nil
}
