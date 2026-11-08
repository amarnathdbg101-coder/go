package filekit

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
)

// Default upload configurations.
const (
	DefaultMaxFileSize = 10 * 1024 * 1024 // 10 MB
	DefaultJPEGQuality = 85
)

// Config holds runtime options for the Uploader.
type Config struct {
	MaxFileSize     int64    // Maximum allowed file size in bytes (default: 10MB)
	AllowedMimes    []string // Whitelist of MIME types (default: DefaultImageMimes)
	SubDir          string   // Relative subdirectory (e.g. "products", "avatars", "invoices")
	ResizeMaxWidth  int      // Optional max width for image resizing (0 = no resize)
	ResizeMaxHeight int      // Optional max height for image resizing (0 = no resize)
	JPEGQuality     int      // JPEG compression quality 1-100 (default: 85)
}

// Result contains complete metadata of an uploaded and processed file.
type Result struct {
	Filename     string `json:"filename"`                // Stored unique filename
	URL          string `json:"url"`                     // Publicly accessible URL
	Size         int64  `json:"size"`                    // Final file size in bytes
	MimeType     string `json:"mime_type"`               // Verified MIME content type
	Width        int    `json:"width,omitempty"`         // Image pixel width (if image)
	Height       int    `json:"height,omitempty"`        // Image pixel height (if image)
	OriginalName string `json:"original_name,omitempty"` // Cleaned original filename
}

// Uploader is the high-level orchestration engine for file uploads, security validation,
// image resizing, and multi-storage driver dispatching.
type Uploader struct {
	storage Storage
	cfg     Config
}

// Option configures Uploader parameters.
type Option func(*Config)

// WithMaxFileSize sets custom size limit in bytes.
func WithMaxFileSize(maxBytes int64) Option {
	return func(c *Config) {
		if maxBytes > 0 {
			c.MaxFileSize = maxBytes
		}
	}
}

// WithAllowedMimes sets permitted MIME types whitelist.
func WithAllowedMimes(mimes ...string) Option {
	return func(c *Config) {
		if len(mimes) > 0 {
			c.AllowedMimes = mimes
		}
	}
}

// WithSubDir sets the destination folder partition (e.g. "avatars").
func WithSubDir(subDir string) Option {
	return func(c *Config) {
		c.SubDir = strings.Trim(subDir, "/\\")
	}
}

// WithImageResize enables automatic aspect-ratio constrained downscaling.
func WithImageResize(maxWidth, maxHeight int) Option {
	return func(c *Config) {
		c.ResizeMaxWidth = maxWidth
		c.ResizeMaxHeight = maxHeight
	}
}

// WithJPEGQuality sets JPEG compression factor (1-100).
func WithJPEGQuality(quality int) Option {
	return func(c *Config) {
		if quality > 0 && quality <= 100 {
			c.JPEGQuality = quality
		}
	}
}

// NewUploader creates an initialized Uploader instance.
func NewUploader(storage Storage, opts ...Option) (*Uploader, error) {
	if storage == nil {
		return nil, ErrStorageNil
	}

	cfg := Config{
		MaxFileSize:  DefaultMaxFileSize,
		AllowedMimes: DefaultImageMimes,
		JPEGQuality:  DefaultJPEGQuality,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	return &Uploader{
		storage: storage,
		cfg:     cfg,
	}, nil
}

// Upload handles general file validation (magic bytes check, size limit) and saves to storage.
func (u *Uploader) Upload(ctx context.Context, originalFilename string, r io.Reader) (*Result, error) {
	// Read full payload bounded by MaxFileSize + 1 to prevent memory exhaustion
	limitReader := io.LimitReader(r, u.cfg.MaxFileSize+1)
	data, err := io.ReadAll(limitReader)
	if err != nil {
		return nil, fmt.Errorf("filekit: read failed: %w", err)
	}

	if len(data) == 0 {
		return nil, ErrEmptyFile
	}

	if int64(len(data)) > u.cfg.MaxFileSize {
		return nil, ErrFileTooLarge
	}

	// Sniff real magic bytes
	detectedMime := DetectMime(data)
	if !IsAllowedMime(detectedMime, u.cfg.AllowedMimes) {
		return nil, ErrInvalidMimeType
	}

	// Generate collision-proof safe filename
	safeFilename := GenerateSafeFilename(detectedMime)
	relativePath := safeFilename
	if u.cfg.SubDir != "" {
		relativePath = path.Join(u.cfg.SubDir, safeFilename)
	}

	// Save to underlying storage driver
	publicURL, err := u.storage.Save(ctx, relativePath, bytes.NewReader(data), int64(len(data)), detectedMime)
	if err != nil {
		return nil, err
	}

	return &Result{
		Filename:     safeFilename,
		URL:          publicURL,
		Size:         int64(len(data)),
		MimeType:     detectedMime,
		OriginalName: SanitizeOriginalName(originalFilename),
	}, nil
}

// UploadImage decodes, inspects dimensions, optionally resizes, compresses, and saves the image.
func (u *Uploader) UploadImage(ctx context.Context, originalFilename string, r io.Reader) (*Result, error) {
	// 1. Read bounded payload
	limitReader := io.LimitReader(r, u.cfg.MaxFileSize+1)
	data, err := io.ReadAll(limitReader)
	if err != nil {
		return nil, fmt.Errorf("filekit: read failed: %w", err)
	}

	if len(data) == 0 {
		return nil, ErrEmptyFile
	}

	if int64(len(data)) > u.cfg.MaxFileSize {
		return nil, ErrFileTooLarge
	}

	// 2. Validate magic bytes
	detectedMime := DetectMime(data)
	if !IsAllowedMime(detectedMime, u.cfg.AllowedMimes) {
		return nil, ErrInvalidMimeType
	}

	// 3. Decode image to verify it is genuinely valid and not corrupted
	img, format, err := DecodeImage(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	origW := bounds.Dx()
	origH := bounds.Dy()
	finalW, finalH := origW, origH

	var finalBytes []byte

	// 4. Resize if bounds are specified and original exceeds bounds
	if u.cfg.ResizeMaxWidth > 0 && u.cfg.ResizeMaxHeight > 0 {
		targetW, targetH := CalculateAspectFit(origW, origH, u.cfg.ResizeMaxWidth, u.cfg.ResizeMaxHeight)
		if targetW != origW || targetH != origH {
			img = ResizeImage(img, targetW, targetH)
			finalW, finalH = targetW, targetH
		}
	}

	// 5. Compress and re-encode
	finalBytes, err = CompressAndEncode(img, format, u.cfg.JPEGQuality)
	if err != nil {
		// Fallback to original bytes if re-encoding fails
		finalBytes = data
	}

	// 6. Generate safe filename
	safeFilename := GenerateSafeFilename(detectedMime)
	relativePath := safeFilename
	if u.cfg.SubDir != "" {
		relativePath = path.Join(u.cfg.SubDir, safeFilename)
	}

	// 7. Store
	publicURL, err := u.storage.Save(ctx, relativePath, bytes.NewReader(finalBytes), int64(len(finalBytes)), detectedMime)
	if err != nil {
		return nil, err
	}

	return &Result{
		Filename:     safeFilename,
		URL:          publicURL,
		Size:         int64(len(finalBytes)),
		MimeType:     detectedMime,
		Width:        finalW,
		Height:       finalH,
		OriginalName: SanitizeOriginalName(originalFilename),
	}, nil
}

// UploadFromRequest parses an incoming multipart HTTP form file and uploads it.
func (u *Uploader) UploadFromRequest(r *http.Request, formField string) (*Result, error) {
	// Parse multipart body with max memory buffer
	if err := r.ParseMultipartForm(u.cfg.MaxFileSize); err != nil {
		return nil, fmt.Errorf("filekit: failed to parse multipart form: %w", err)
	}

	file, header, err := r.FormFile(formField)
	if err != nil {
		return nil, fmt.Errorf("filekit: missing form file field '%s': %w", formField, err)
	}
	defer file.Close()

	// If allowed MIME types are images, route to UploadImage for compression/resizing
	isImageUploader := false
	for _, m := range u.cfg.AllowedMimes {
		if strings.HasPrefix(m, "image/") {
			isImageUploader = true
			break
		}
	}

	if isImageUploader {
		return u.UploadImage(r.Context(), header.Filename, file)
	}

	return u.Upload(r.Context(), header.Filename, file)
}

// Storage returns the underlying storage driver.
func (u *Uploader) Storage() Storage {
	return u.storage
}
