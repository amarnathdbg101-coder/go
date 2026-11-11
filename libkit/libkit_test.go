package main

import (
	"os"
	"os/exec"
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

func TestLibkit_GetConsolidatedSubpackage(t *testing.T) {
	cases := []struct {
		name       string
		module     string
		importPath string
	}{
		{
			name:       "go-env/envvalidator",
			module:     "go-env",
			importPath: "github.com/amarnathdbg101-coder/go/go-env/envvalidator",
		},
		{
			name:       "go-cachekit/data/cache",
			module:     "go-cachekit",
			importPath: "github.com/amarnathdbg101-coder/go/go-cachekit/data/cache",
		},
	}

	root := FindLibraryRoot()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			item, err := GetLibrary(tc.name)
			if err != nil {
				t.Fatalf("GetLibrary failed: %v", err)
			}
			if item.Module != tc.module || item.ModuleDir != filepath.Join(root, tc.module) {
				t.Fatalf("unexpected module root: module=%q dir=%q", item.Module, item.ModuleDir)
			}
			if item.ImportPath != tc.importPath {
				t.Fatalf("unexpected import path: %q", item.ImportPath)
			}
			if info, err := os.Stat(item.Path); err != nil || !info.IsDir() {
				t.Fatalf("subpackage path is invalid: info=%v err=%v", info, err)
			}
		})
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

	projectDir := filepath.Join(tempDir, projectName)
	cmd := exec.Command("go", "test", "./...")
	cmd.Dir = projectDir
	cmd.Env = append(os.Environ(), "GOWORK=off")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("generated project does not build:\n%s\n%v", output, err)
	}
}

func TestLibkit_ScaffoldDoesNotOverwriteNonEmptyDirectory(t *testing.T) {
	tempDir := t.TempDir()
	projectDir := filepath.Join(tempDir, "existing")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}
	sentinel := filepath.Join(projectDir, "keep.txt")
	if err := os.WriteFile(sentinel, []byte("keep"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := ScaffoldProject(tempDir, "existing"); err == nil {
		t.Fatal("expected an error for a non-empty target directory")
	}
	if data, err := os.ReadFile(sentinel); err != nil || string(data) != "keep" {
		t.Fatalf("existing file was modified: data=%q err=%v", data, err)
	}
}
