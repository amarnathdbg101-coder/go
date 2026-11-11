package mimekit

import (
	"bytes"
	"io"
	"testing"
)

func TestMimeKit_Images(t *testing.T) {
	// PNG
	pngData := append([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, []byte("imagedata")...)
	ct := Sniff(pngData)
	if ct.MIME != "image/png" || ct.Extension != "png" || !ct.IsImage {
		t.Fatalf("expected PNG, got %+v", ct)
	}
	if !IsSafeImage(pngData) {
		t.Fatalf("expected IsSafeImage true for PNG")
	}

	// JPEG
	jpgData := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46}
	ct = Sniff(jpgData)
	if ct.MIME != "image/jpeg" || ct.Extension != "jpg" || !ct.IsImage {
		t.Fatalf("expected JPEG, got %+v", ct)
	}

	// SVG
	svgData := []byte("<svg xmlns='http://www.w3.org/2000/svg'><circle/></svg>")
	ct = Sniff(svgData)
	if ct.MIME != "image/svg+xml" || ct.Extension != "svg" || !ct.IsImage {
		t.Fatalf("expected SVG, got %+v", ct)
	}
}

func TestMimeKit_DocumentsAndArchives(t *testing.T) {
	// PDF
	pdfData := []byte("%PDF-1.4\n1 0 obj\n<<>>\nendobj")
	ct := Sniff(pdfData)
	if ct.MIME != "application/pdf" || ct.Extension != "pdf" || !ct.IsDocument {
		t.Fatalf("expected PDF, got %+v", ct)
	}

	// ZIP
	zipData := []byte{0x50, 0x4B, 0x03, 0x04, 0x14, 0x00}
	ct = Sniff(zipData)
	if ct.MIME != "application/zip" || ct.Extension != "zip" || !ct.IsArchive {
		t.Fatalf("expected ZIP, got %+v", ct)
	}
}

func TestMimeKit_ExecutablePolyglotDefense(t *testing.T) {
	// Linux ELF
	elfData := []byte{0x7F, 0x45, 0x4C, 0x46, 0x02, 0x01, 0x01, 0x00}
	ct := Sniff(elfData)
	if !ct.IsExecutable {
		t.Fatalf("expected ELF to be identified as executable")
	}
	if IsSafeImage(elfData) {
		t.Fatalf("ELF binary must never be flagged as safe image")
	}

	// Windows PE (MZ)
	peData := []byte{0x4D, 0x5A, 0x90, 0x00, 0x03, 0x00}
	if !IsExecutable(peData) {
		t.Fatalf("expected PE binary to be flagged executable")
	}

	// Shell script
	shData := []byte("#!/bin/bash\nrm -rf /")
	if !IsExecutable(shData) {
		t.Fatalf("expected shell script to be flagged executable")
	}
}

func TestMimeKit_SniffReaderRewind(t *testing.T) {
	content := append([]byte("%PDF-1.7\n"), bytes.Repeat([]byte("A"), 1000)...)
	reader := bytes.NewReader(content)

	ct, reconstructed, err := SniffReader(reader)
	if err != nil {
		t.Fatalf("SniffReader failed: %v", err)
	}
	if ct.MIME != "application/pdf" {
		t.Fatalf("expected PDF, got %s", ct.MIME)
	}

	// Reconstructed reader must yield the entire 1009 bytes intact
	drained, err := io.ReadAll(reconstructed)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}
	if len(drained) != len(content) || !bytes.Equal(drained, content) {
		t.Fatalf("stream content corrupted after SniffReader")
	}
}
