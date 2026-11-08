package mimekit

import (
	"bytes"
	"io"
	"net/http"
	"strings"
)

// ContentType details the sniffed MIME metadata.
type ContentType struct {
	MIME         string
	Extension    string
	IsImage      bool
	IsDocument   bool
	IsArchive    bool
	IsAudioVideo bool
	IsExecutable bool
}

// Magic byte signatures
var (
	magicPNG   = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	magicJPEG  = []byte{0xFF, 0xD8, 0xFF}
	magicGIF87 = []byte("GIF87a")
	magicGIF89 = []byte("GIF89a")
	magicPDF   = []byte("%PDF-")
	magicZIP   = []byte{0x50, 0x4B, 0x03, 0x04}
	magicGZIP  = []byte{0x1F, 0x8B}
	magicELF   = []byte{0x7F, 0x45, 0x4C, 0x46} // Linux ELF
	magicPE    = []byte{0x4D, 0x5A}             // Windows MZ
	magicMachO = []byte{0xFE, 0xED, 0xFA, 0xCE} // macOS Mach-O
	magicSheb  = []byte("#!")                    // Shell script
)

// Sniff inspects the first 512 bytes of data to determine its true MIME type.
func Sniff(data []byte) ContentType {
	if len(data) == 0 {
		return ContentType{MIME: "application/octet-stream", Extension: "bin"}
	}

	header := data
	if len(header) > 512 {
		header = header[:512]
	}

	// 1. Check dangerous executables first
	if isExecHeader(header) {
		return ContentType{
			MIME:         "application/x-executable",
			Extension:    "bin",
			IsExecutable: true,
		}
	}

	// 2. Specific signature matches
	if bytes.HasPrefix(header, magicPNG) {
		return ContentType{MIME: "image/png", Extension: "png", IsImage: true}
	}
	if bytes.HasPrefix(header, magicJPEG) {
		return ContentType{MIME: "image/jpeg", Extension: "jpg", IsImage: true}
	}
	if bytes.HasPrefix(header, magicGIF87) || bytes.HasPrefix(header, magicGIF89) {
		return ContentType{MIME: "image/gif", Extension: "gif", IsImage: true}
	}
	if len(header) >= 12 && bytes.Equal(header[0:4], []byte("RIFF")) && bytes.Equal(header[8:12], []byte("WEBP")) {
		return ContentType{MIME: "image/webp", Extension: "webp", IsImage: true}
	}
	if bytes.HasPrefix(header, magicPDF) {
		return ContentType{MIME: "application/pdf", Extension: "pdf", IsDocument: true}
	}
	if bytes.HasPrefix(header, magicZIP) {
		return ContentType{MIME: "application/zip", Extension: "zip", IsArchive: true}
	}
	if bytes.HasPrefix(header, magicGZIP) {
		return ContentType{MIME: "application/gzip", Extension: "gz", IsArchive: true}
	}
	if len(header) >= 12 && bytes.Equal(header[0:4], []byte("RIFF")) && bytes.Equal(header[8:12], []byte("WAVE")) {
		return ContentType{MIME: "audio/wav", Extension: "wav", IsAudioVideo: true}
	}
	if len(header) >= 8 && bytes.Equal(header[4:8], []byte("ftyp")) {
		return ContentType{MIME: "video/mp4", Extension: "mp4", IsAudioVideo: true}
	}

	// SVG text check
	trimmed := bytes.TrimSpace(header)
	if bytes.HasPrefix(trimmed, []byte("<?xml")) || bytes.HasPrefix(trimmed, []byte("<svg")) {
		if bytes.Contains(trimmed, []byte("<svg")) {
			return ContentType{MIME: "image/svg+xml", Extension: "svg", IsImage: true}
		}
	}

	// 3. Fallback to Go's standard library detector
	detected := http.DetectContentType(header)
	ext := canonicalExtension(detected)

	return ContentType{
		MIME:         detected,
		Extension:    ext,
		IsImage:      strings.HasPrefix(detected, "image/"),
		IsAudioVideo: strings.HasPrefix(detected, "audio/") || strings.HasPrefix(detected, "video/"),
		IsDocument:   strings.HasPrefix(detected, "text/") || detected == "application/pdf",
	}
}

func isExecHeader(header []byte) bool {
	if bytes.HasPrefix(header, magicELF) ||
		bytes.HasPrefix(header, magicPE) ||
		bytes.HasPrefix(header, magicMachO) ||
		bytes.HasPrefix(header, magicSheb) {
		return true
	}
	return false
}

// SniffReader sniffs the stream, then returns ContentType and a reconstructed io.Reader
// so the caller can read the entire stream from byte 0.
func SniffReader(r io.Reader) (ContentType, io.Reader, error) {
	if r == nil {
		return ContentType{}, nil, ErrEmptyPayload
	}

	buf := make([]byte, 512)
	n, err := io.ReadFull(r, buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return ContentType{}, nil, err
	}

	buf = buf[:n]
	ct := Sniff(buf)

	// Combine sniffed buffer and remaining reader stream
	reconstructed := io.MultiReader(bytes.NewReader(buf), r)
	return ct, reconstructed, nil
}

// IsSafeImage returns true if data matches safe image MIME types and is not executable.
func IsSafeImage(data []byte) bool {
	ct := Sniff(data)
	return ct.IsImage && !ct.IsExecutable
}

// IsExecutable returns true if the content begins with executable magic bytes.
func IsExecutable(data []byte) bool {
	ct := Sniff(data)
	return ct.IsExecutable
}

func canonicalExtension(mime string) string {
	parts := strings.Split(mime, ";")
	clean := strings.TrimSpace(parts[0])

	switch clean {
	case "image/png":
		return "png"
	case "image/jpeg":
		return "jpg"
	case "image/gif":
		return "gif"
	case "image/webp":
		return "webp"
	case "image/svg+xml":
		return "svg"
	case "application/pdf":
		return "pdf"
	case "application/zip":
		return "zip"
	case "application/gzip":
		return "gz"
	case "text/plain":
		return "txt"
	case "text/html":
		return "html"
	case "application/json":
		return "json"
	case "video/mp4":
		return "mp4"
	case "audio/wav":
		return "wav"
	default:
		return "bin"
	}
}
