package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLibkit_CatalogLoading(t *testing.T) {
	root := FindLibraryRoot()
	if root == "" {
		t.Fatalf("expected to find library root")
	}

	items, err := LoadCatalog()
	if err != nil {
		t.Fatalf("LoadCatalog failed: %v", err)
	}

	if len(items) < 125 {
		t.Fatalf("expected at least 125 catalog items, got %d", len(items))
	}
}

func TestLibkit_SearchCatalog(t *testing.T) {
	results, err := SearchCatalog("jwt")
	if err != nil {
		t.Fatalf("SearchCatalog failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatalf("expected search for 'jwt' to find results")
	}

	found := false
	for _, r := range results {
		if r.Name == "go-jwt" || r.Name == "go-authkit" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected go-jwt or go-authkit in search results")
	}
}

func TestLibkit_GetLibrary(t *testing.T) {
	item, err := GetLibrary("go-saga")
	if err != nil {
		t.Fatalf("GetLibrary failed: %v", err)
	}

	if item.Name != "go-saga" {
		t.Fatalf("expected go-saga, got %s", item.Name)
	}

	readme, err := item.ReadDoc("README.md")
	if err != nil || len(readme) == 0 {
		t.Fatalf("failed to read README.md for go-saga: %v", err)
	}
}

func TestLibkit_Scaffold(t *testing.T) {
	tempDir := t.TempDir()
	projectName := "demo-service"

	err := ScaffoldProject(tempDir, projectName)
	if err != nil {
		t.Fatalf("ScaffoldProject failed: %v", err)
	}

	expectedFiles := []string{
		"go.mod",
		"README.md",
		"config/config.go",
		"cmd/server/main.go",
		"internal/handler/handler.go",
	}

	for _, rel := range expectedFiles {
		full := filepath.Join(tempDir, projectName, rel)
		if _, err := os.Stat(full); err != nil {
			t.Errorf("missing expected scaffold file: %s", rel)
		}
	}
}
