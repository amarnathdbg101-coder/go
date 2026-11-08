# SPEC: go-image-filekit

## 1. Scope & Purpose
Production-grade image and file upload engine with Zero-Cgo Pure Go image resizing, real magic byte MIME verification, and path traversal defense.
- In Scope: Magic byte detection (`http.DetectContentType`), collision-proof safe filename generation, pure Go bilinear aspect-ratio resizing, JPEG compression (80-85%), and local disk storage driver.
- Out of Scope: Video/audio transcoding (which requires FFmpeg).

## 2. Public API Contract
- `NewUploader(storage Storage, opts ...Option) (*Uploader, error)`
- `Uploader.Upload(ctx, filename, io.Reader) (*Result, error)`
- `Uploader.UploadImage(ctx, filename, io.Reader) (*Result, error)`
- `Uploader.UploadFromRequest(r *http.Request, formField string) (*Result, error)`
- `NewLocalStorage(baseDir, baseURL string) (*LocalStorage, error)`

## 3. Concurrency & Memory Guarantees
- Bounded payload reading: Max upload size is strictly capped via `io.LimitReader` to prevent memory exhaustion attacks.
- Pure Go image operations require zero external Cgo threads.

## 4. How to Extend
- Implement `Storage` interface (`Save`, `Open`, `Delete`, `Exists`) to add AWS S3, Cloudflare R2, MinIO, or Google Cloud Storage drivers.
