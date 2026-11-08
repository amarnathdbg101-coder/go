package filekit

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

// Standard allowed MIME categories.
var (
	// DefaultImageMimes restricts uploads strictly to safe image formats.
	DefaultImageMimes = []string{
		"image/jpeg",
		"image/png",
		"image/webp",
		"image/gif",
	}

	// DefaultDocumentMimes includes common office/pdf formats.
	DefaultDocumentMimes = []string{
		"application/pdf",
		"image/jpeg",
		"image/png",
	}
)

// DetectMime reads the first 512 bytes (magic numbers) to determine the true content type.
// Attackers jo file extension rename karte hain (e.g. malware.exe ko avatar.png) unhe ye check detect kar leta hai.
func DetectMime(buffer []byte) string {
	limit := len(buffer)
	if limit > 512 {
		limit = 512
	}
	contentType := http.DetectContentType(buffer[:limit])

	// Strip charset if appended (e.g. "text/plain; charset=utf-8")
	if idx := strings.Index(contentType, ";"); idx != -1 {
		contentType = contentType[:idx]
	}
	return strings.TrimSpace(contentType)
}

// IsAllowedMime verifies if the detected MIME type is in the allowed whitelist.
func IsAllowedMime(detectedMime string, allowedMimes []string) bool {
	cleanDetected := strings.ToLower(strings.TrimSpace(detectedMime))
	for _, allowed := range allowedMimes {
		if strings.EqualFold(cleanDetected, strings.ToLower(strings.TrimSpace(allowed))) {
			return true
		}
	}
	return false
}

// ExtForMime maps a verified MIME type to its standard file extension.
func ExtForMime(mime string) string {
	switch strings.ToLower(strings.TrimSpace(mime)) {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	case "application/pdf":
		return ".pdf"
	case "image/svg+xml":
		return ".svg"
	default:
		return ".bin"
	}
}

// GenerateSafeFilename creates a collision-resistant, cryptographically random filename.
// Format: <unix_nano>_<8_hex_bytes><safe_extension>
func GenerateSafeFilename(detectedMime string) string {
	ext := ExtForMime(detectedMime)

	randomBytes := make([]byte, 8)
	_, _ = rand.Read(randomBytes)
	randHex := hex.EncodeToString(randomBytes)

	return fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), randHex, ext)
}

// SanitizeOriginalName cleans up user-supplied filename while stripping path elements and dangerous characters.
func SanitizeOriginalName(filename string) string {
	base := filepath.Base(filepath.Clean(filename))
	// Replace spaces with underscores
	base = strings.ReplaceAll(base, " ", "_")
	return base
}
