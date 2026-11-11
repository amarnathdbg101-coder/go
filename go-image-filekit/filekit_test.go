package filekit_test

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/amarnathdbg101-coder/go/go-image-filekit"
)

// Helper: Generates a valid in-memory PNG image
func createTestPNG(w, h int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x % 255), G: uint8(y % 255), B: 100, A: 255})
		}
	}
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

// -------------------------------------------------------------------------
// 1. LocalStorage & Path Traversal Security Tests
// -------------------------------------------------------------------------

func TestLocalStorage_SaveOpenDelete(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "filekit_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	storage, err := filekit.NewLocalStorage(tempDir, "/uploads")
	if err != nil {
		t.Fatalf("new local storage failed: %v", err)
	}

	ctx := context.Background()
	relPath := "products/test.txt"
	content := []byte("Hello FileKit")

	// 1. Save
	url, err := storage.Save(ctx, relPath, bytes.NewReader(content), int64(len(content)), "text/plain")
	if err != nil {
		t.Fatalf("save failed: %v", err)
	}
	if !strings.HasSuffix(url, "/uploads/products/test.txt") {
		t.Errorf("unexpected url: %s", url)
	}

	// 2. Exists
	exists, err := storage.Exists(ctx, relPath)
	if err != nil || !exists {
		t.Errorf("expected file to exist: %v", err)
	}

	// 3. Open & Read
	reader, err := storage.Open(ctx, relPath)
	if err != nil {
		t.Fatalf("open failed: %v", err)
	}
	readBytes, _ := io.ReadAll(reader)
	_ = reader.Close()

	if string(readBytes) != string(content) {
		t.Errorf("expected '%s', got '%s'", content, readBytes)
	}

	// 4. Delete
	if err := storage.Delete(ctx, relPath); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	existsAfter, _ := storage.Exists(ctx, relPath)
	if existsAfter {
		t.Errorf("file should not exist after delete")
	}
}

func TestLocalStorage_PathTraversalDefense(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "filekit_safe_*")
	defer os.RemoveAll(tempDir)

	storage, _ := filekit.NewLocalStorage(tempDir, "/uploads")
	ctx := context.Background()

	// Attack 1: Classic ../../ traversal
	_, err := storage.Save(ctx, "../../evil.exe", strings.NewReader("bad"), 3, "application/octet-stream")
	if err != filekit.ErrPathTraversal {
		t.Errorf("expected ErrPathTraversal, got: %v", err)
	}

	// Attack 2: Absolute path
	absPath := filepath.Join(tempDir, "abs.txt")
	_, errAbs := storage.Save(ctx, absPath, strings.NewReader("bad"), 3, "text/plain")
	if errAbs != filekit.ErrPathTraversal {
		t.Errorf("expected ErrPathTraversal for absolute path, got: %v", errAbs)
	}
}

// -------------------------------------------------------------------------
// 2. MIME Spoofing Detection Tests
// -------------------------------------------------------------------------

func TestMime_SpoofingBlocked(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "filekit_mime_*")
	defer os.RemoveAll(tempDir)

	storage, _ := filekit.NewLocalStorage(tempDir, "/uploads")
	uploader, _ := filekit.NewUploader(storage, filekit.WithAllowedMimes("image/png", "image/jpeg"))

	// Attacker uploads a shell script renamed to avatar.png
	fakeImagePayload := []byte("#!/bin/bash\necho 'Hacked!'")
	_, err := uploader.Upload(context.Background(), "avatar.png", bytes.NewReader(fakeImagePayload))

	if err != filekit.ErrInvalidMimeType {
		t.Fatalf("expected ErrInvalidMimeType on spoofed image, got: %v", err)
	}
}

// -------------------------------------------------------------------------
// 3. Image Resizing & Aspect-Ratio Fitting Tests
// -------------------------------------------------------------------------

func TestCalculateAspectFit(t *testing.T) {
	// Case 1: Landscape 800x600 fitted to 400x400 -> should be 400x300
	w1, h1 := filekit.CalculateAspectFit(800, 600, 400, 400)
	if w1 != 400 || h1 != 300 {
		t.Errorf("expected 400x300, got %dx%d", w1, h1)
	}

	// Case 2: Portrait 600x800 fitted to 400x400 -> should be 300x400
	w2, h2 := filekit.CalculateAspectFit(600, 800, 400, 400)
	if w2 != 300 || h2 != 400 {
		t.Errorf("expected 300x400, got %dx%d", w2, h2)
	}

	// Case 3: Smaller image 200x150 fitted to 400x400 -> preserves original (no upscaling)
	w3, h3 := filekit.CalculateAspectFit(200, 150, 400, 400)
	if w3 != 200 || h3 != 150 {
		t.Errorf("expected original 200x150, got %dx%d", w3, h3)
	}
}

func TestImage_ResizeAndUpload(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "filekit_resize_*")
	defer os.RemoveAll(tempDir)

	storage, _ := filekit.NewLocalStorage(tempDir, "/uploads")
	uploader, err := filekit.NewUploader(storage,
		filekit.WithSubDir("products"),
		filekit.WithImageResize(400, 300),
		filekit.WithJPEGQuality(80),
	)
	if err != nil {
		t.Fatalf("new uploader error: %v", err)
	}

	// Create 800x600 test image
	originalPNG := createTestPNG(800, 600)

	result, err := uploader.UploadImage(context.Background(), "gadget.png", bytes.NewReader(originalPNG))
	if err != nil {
		t.Fatalf("upload image failed: %v", err)
	}

	// Assertions on resized dimensions
	if result.Width != 400 || result.Height != 300 {
		t.Errorf("expected resized dimensions 400x300, got %dx%d", result.Width, result.Height)
	}
	if result.Size <= 0 {
		t.Errorf("expected positive final size")
	}
	if !strings.Contains(result.URL, "/uploads/products/") {
		t.Errorf("expected URL to contain subdir, got %s", result.URL)
	}

	// Verify file on disk exists and can be opened
	relFile := filepath.Join("products", result.Filename)
	exists, _ := storage.Exists(context.Background(), relFile)
	if !exists {
		t.Errorf("expected stored file to exist at %s", relFile)
	}
}

// -------------------------------------------------------------------------
// 4. File Size Limits & Multipart HTTP Form Tests
// -------------------------------------------------------------------------

func TestUploader_MaxFileSizeExceeded(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "filekit_limit_*")
	defer os.RemoveAll(tempDir)

	storage, _ := filekit.NewLocalStorage(tempDir, "/uploads")
	uploader, _ := filekit.NewUploader(storage, filekit.WithMaxFileSize(100)) // 100 bytes limit

	payload := make([]byte, 200) // 200 bytes
	_, err := uploader.Upload(context.Background(), "large.png", bytes.NewReader(payload))

	if err != filekit.ErrFileTooLarge {
		t.Errorf("expected ErrFileTooLarge, got: %v", err)
	}
}

func TestUploader_UploadFromRequest(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "filekit_req_*")
	defer os.RemoveAll(tempDir)

	storage, _ := filekit.NewLocalStorage(tempDir, "/uploads")
	uploader, _ := filekit.NewUploader(storage, filekit.WithSubDir("avatars"))

	// Create synthetic multipart form request
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("avatar", "profile.png")
	if err != nil {
		t.Fatalf("create form file failed: %v", err)
	}
	pngData := createTestPNG(100, 100)
	_, _ = part.Write(pngData)
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	result, err := uploader.UploadFromRequest(req, "avatar")
	if err != nil {
		t.Fatalf("upload from request failed: %v", err)
	}

	if result.OriginalName != "profile.png" {
		t.Errorf("expected original name 'profile.png', got '%s'", result.OriginalName)
	}
	if result.Width != 100 || result.Height != 100 {
		t.Errorf("expected 100x100, got %dx%d", result.Width, result.Height)
	}
}

func TestUploader_DecompressionBombBlocked(t *testing.T) {
	storage := filekit.NewMemoryStorage("/uploads")
	// Set max dimension to 150px
	uploader, err := filekit.NewUploader(storage, filekit.WithMaxDimension(150))
	if err != nil {
		t.Fatalf("NewUploader failed: %v", err)
	}

	// 200x200 image exceeds 150px limit
	bigImgData := createTestPNG(200, 200)
	_, err = uploader.UploadImage(context.Background(), "bomb.png", bytes.NewReader(bigImgData))
	if err != filekit.ErrDecompressionBomb {
		t.Fatalf("expected ErrDecompressionBomb, got: %v", err)
	}
}

func TestMemoryStorage_FullFlow(t *testing.T) {
	storage := filekit.NewMemoryStorage("https://cdn.shopsilo.com/media")
	ctx := context.Background()

	content := []byte("invoice payload data")
	url, err := storage.Save(ctx, "invoices/inv_01.pdf", bytes.NewReader(content), int64(len(content)), "application/pdf")
	if err != nil {
		t.Fatalf("save failed: %v", err)
	}

	if url != "https://cdn.shopsilo.com/media/invoices/inv_01.pdf" {
		t.Errorf("unexpected url: %s", url)
	}

	exists, err := storage.Exists(ctx, "invoices/inv_01.pdf")
	if err != nil || !exists {
		t.Fatalf("expected file to exist in memory storage")
	}

	rc, err := storage.Open(ctx, "invoices/inv_01.pdf")
	if err != nil {
		t.Fatalf("open failed: %v", err)
	}
	readBytes, _ := io.ReadAll(rc)
	_ = rc.Close()

	if string(readBytes) != string(content) {
		t.Fatalf("content mismatch")
	}

	err = storage.Delete(ctx, "invoices/inv_01.pdf")
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	existsAfter, _ := storage.Exists(ctx, "invoices/inv_01.pdf")
	if existsAfter {
		t.Fatalf("file should not exist after delete")
	}
}

func TestCompressAndEncode_GIF_And_PNG(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))

	// GIF
	gifBytes, err := filekit.CompressAndEncode(img, "gif", 80)
	if err != nil || len(gifBytes) == 0 {
		t.Fatalf("gif encoding failed: %v", err)
	}

	// PNG
	pngBytes, err := filekit.CompressAndEncode(img, "png", 80)
	if err != nil || len(pngBytes) == 0 {
		t.Fatalf("png encoding failed: %v", err)
	}

	// Fallback format
	fallbackBytes, err := filekit.CompressAndEncode(img, "unknown_raw", 80)
	if err != nil || len(fallbackBytes) == 0 {
		t.Fatalf("fallback encoding failed: %v", err)
	}
}

func TestExtForMime_All(t *testing.T) {
	cases := map[string]string{
		"image/jpeg":        ".jpg",
		"image/png":         ".png",
		"image/webp":        ".webp",
		"image/gif":         ".gif",
		"application/pdf":   ".pdf",
		"image/svg+xml":     ".svg",
		"application/other": ".bin",
	}

	for mime, expectedExt := range cases {
		got := filekit.ExtForMime(mime)
		if got != expectedExt {
			t.Errorf("for mime %s expected ext %s, got %s", mime, expectedExt, got)
		}
	}
}

func TestUploader_StorageMethod(t *testing.T) {
	storage := filekit.NewMemoryStorage("/uploads")
	uploader, _ := filekit.NewUploader(storage)

	if uploader.Storage() != storage {
		t.Errorf("Storage() should return matching storage reference")
	}
}
