package pathkit

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPathKit_SafeJoin(t *testing.T) {
	tempDir := t.TempDir()

	// Valid subpaths
	safePath, err := SafeJoin(tempDir, "uploads/avatar.png")
	if err != nil {
		t.Fatalf("unexpected SafeJoin error: %v", err)
	}
	if !strings.HasPrefix(safePath, tempDir) {
		t.Fatalf("expected safe path to be inside tempDir")
	}

	// Path traversal attack: ../../etc/passwd
	_, err = SafeJoin(tempDir, "../../etc/passwd")
	if !errors.Is(err, ErrPathTraversal) {
		t.Fatalf("expected ErrPathTraversal, got: %v", err)
	}

	// Absolute path attempt to break out
	_, err = SafeJoin(tempDir, "/etc/shadow")
	if err == nil {
		// On windows /etc/shadow might become root drive \etc\shadow which escapes tempDir
		// Should fail
		if !strings.HasPrefix(safePath, tempDir) {
			t.Fatalf("expected failure on absolute escape")
		}
	}
}

func TestPathKit_WriteFileAtomic(t *testing.T) {
	tempDir := t.TempDir()
	targetFile := filepath.Join(tempDir, "sub", "config.json")

	content := []byte(`{"version": 1}`)
	err := WriteFileAtomic(targetFile, content, 0644)
	if err != nil {
		t.Fatalf("WriteFileAtomic failed: %v", err)
	}

	// Verify content
	readBack, err := os.ReadFile(targetFile)
	if err != nil || string(readBack) != string(content) {
		t.Fatalf("read back content mismatch: %s", string(readBack))
	}

	// Atomic overwrite
	updatedContent := []byte(`{"version": 2}`)
	err = WriteFileAtomic(targetFile, updatedContent, 0644)
	if err != nil {
		t.Fatalf("second WriteFileAtomic failed: %v", err)
	}

	readBack, _ = os.ReadFile(targetFile)
	if string(readBack) != string(updatedContent) {
		t.Fatalf("expected updated content")
	}
}

func TestPathKit_Hashing(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "sample.txt")
	_ = os.WriteFile(filePath, []byte("Hello Antigravity 2026"), 0644)

	sha, err := SHA256(filePath)
	if err != nil || len(sha) != 64 {
		t.Fatalf("invalid SHA256 hash: %s, err: %v", sha, err)
	}

	md5Hash, err := MD5(filePath)
	if err != nil || len(md5Hash) != 32 {
		t.Fatalf("invalid MD5 hash: %s, err: %v", md5Hash, err)
	}
}

func TestPathKit_Walk(t *testing.T) {
	tempDir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(tempDir, "node_modules"), 0755)
	_ = os.WriteFile(filepath.Join(tempDir, "node_modules", "skip.go"), []byte("package skip"), 0644)

	_ = os.MkdirAll(filepath.Join(tempDir, "pkg"), 0755)
	_ = os.WriteFile(filepath.Join(tempDir, "pkg", "main.go"), []byte("package main"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "pkg", "readme.md"), []byte("# Readme"), 0644)

	var collected []string
	opts := WalkOptions{
		Extensions: []string{".go"},
		IgnoreDirs: []string{"node_modules"},
	}

	err := Walk(tempDir, opts, func(path string, info os.FileInfo) error {
		collected = append(collected, filepath.Base(path))
		return nil
	})

	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	if len(collected) != 1 || collected[0] != "main.go" {
		t.Fatalf("expected only main.go collected, got: %v", collected)
	}
}

func TestPathKit_ExistsAndEnsureDir(t *testing.T) {
	tempDir := t.TempDir()
	subDir := filepath.Join(tempDir, "new_dir")

	if Exists(subDir) {
		t.Fatalf("did not expect subDir to exist")
	}

	if err := EnsureDir(subDir); err != nil {
		t.Fatalf("EnsureDir failed: %v", err)
	}

	if !Exists(subDir) {
		t.Fatalf("expected subDir to exist after EnsureDir")
	}
}
