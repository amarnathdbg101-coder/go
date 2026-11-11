package filekit

import "errors"

// Common domain errors for go-image-filekit.
var (
	ErrFileTooLarge      = errors.New("filekit: file size exceeds allowed limit")
	ErrEmptyFile         = errors.New("filekit: uploaded file is empty (0 bytes)")
	ErrInvalidMimeType   = errors.New("filekit: file type not allowed or MIME spoofing detected")
	ErrPathTraversal     = errors.New("filekit: path traversal attempt detected in filename")
	ErrStorageNil        = errors.New("filekit: storage driver must not be nil")
	ErrInvalidImage      = errors.New("filekit: file is not a valid or decodable image")
	ErrDecompressionBomb = errors.New("filekit: image dimensions exceed safe bounds (decompression bomb protection)")
)
