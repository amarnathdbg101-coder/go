package filekit

import (
	"context"
	"io"
)

// Storage defines the persistence contract for file and image storage.
// Local disk, AWS S3, Cloudflare R2, ya MinIO ke liye ye interface implement hota hai.
type Storage interface {
	// Save stores the file stream at relativePath and returns the public access URL.
	Save(ctx context.Context, relativePath string, r io.Reader, size int64, contentType string) (publicURL string, err error)

	// Open retrieves the file stream for reading. Caller is responsible for closing the reader.
	Open(ctx context.Context, relativePath string) (io.ReadCloser, error)

	// Delete removes the file from storage.
	Delete(ctx context.Context, relativePath string) error

	// Exists checks if the file exists in storage.
	Exists(ctx context.Context, relativePath string) (bool, error)
}
