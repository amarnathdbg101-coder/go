package filekit

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

type memoryFile struct {
	data        []byte
	contentType string
}

// MemoryStorage is an in-memory storage driver useful for unit tests, RAM disks, and ephemeral uploads.
type MemoryStorage struct {
	mu      sync.RWMutex
	files   map[string]memoryFile
	baseURL string
}

// NewMemoryStorage creates a thread-safe in-memory storage driver.
func NewMemoryStorage(baseURL string) *MemoryStorage {
	return &MemoryStorage{
		files:   make(map[string]memoryFile),
		baseURL: strings.TrimRight(baseURL, "/"),
	}
}

// Save stores the file buffer in memory.
func (s *MemoryStorage) Save(ctx context.Context, relativePath string, r io.Reader, size int64, contentType string) (string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}

	cleanRel := strings.TrimPrefix(strings.ReplaceAll(relativePath, "\\", "/"), "/")

	s.mu.Lock()
	defer s.mu.Unlock()

	s.files[cleanRel] = memoryFile{
		data:        data,
		contentType: contentType,
	}

	return fmt.Sprintf("%s/%s", s.baseURL, cleanRel), nil
}

// Open retrieves the file stream from memory.
func (s *MemoryStorage) Open(ctx context.Context, relativePath string) (io.ReadCloser, error) {
	cleanRel := strings.TrimPrefix(strings.ReplaceAll(relativePath, "\\", "/"), "/")

	s.mu.RLock()
	defer s.mu.RUnlock()

	f, ok := s.files[cleanRel]
	if !ok {
		return nil, os.ErrNotExist
	}

	return io.NopCloser(bytes.NewReader(f.data)), nil
}

// Delete removes the file from memory.
func (s *MemoryStorage) Delete(ctx context.Context, relativePath string) error {
	cleanRel := strings.TrimPrefix(strings.ReplaceAll(relativePath, "\\", "/"), "/")

	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.files, cleanRel)
	return nil
}

// Exists checks if file exists in memory.
func (s *MemoryStorage) Exists(ctx context.Context, relativePath string) (bool, error) {
	cleanRel := strings.TrimPrefix(strings.ReplaceAll(relativePath, "\\", "/"), "/")

	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok := s.files[cleanRel]
	return ok, nil
}
