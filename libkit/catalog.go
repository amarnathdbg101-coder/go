package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// LibraryItem represents a cataloged Go library.
type LibraryItem struct {
	ID          int
	Name        string
	Path        string
	Module      string
	ModuleDir   string
	ImportPath  string
	Description string
	StdLib      string
	Status      string
}

// FindLibraryRoot finds the library workspace root.
func FindLibraryRoot() string {
	if root := os.Getenv("LIBKIT_ROOT"); root != "" {
		if hasCatalog(root) {
			return filepath.Clean(root)
		}
		return ""
	}

	curr, err := os.Getwd()
	if err == nil {
		if root := findRootFrom(curr); root != "" {
			return root
		}
	}

	executable, err := os.Executable()
	if err == nil {
		return findRootFrom(filepath.Dir(executable))
	}

	return ""
}

func findRootFrom(start string) string {
	curr := start
	for {
		if hasCatalog(curr) {
			return curr
		}
		parent := filepath.Dir(curr)
		if parent == curr {
			return ""
		}
		curr = parent
	}
}

func hasCatalog(root string) bool {
	info, err := os.Stat(filepath.Join(root, "CATALOG.md"))
	return err == nil && !info.IsDir()
}

// LoadCatalog parses CATALOG.md to extract all cataloged libraries.
func LoadCatalog() ([]LibraryItem, error) {
	rootDir := FindLibraryRoot()
	if rootDir == "" {
		return nil, fmt.Errorf("could not locate CATALOG.md; run from the library workspace or set LIBKIT_ROOT")
	}

	catalogFile := filepath.Join(rootDir, "CATALOG.md")
	file, err := os.Open(catalogFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var items []LibraryItem
	scanner := bufio.NewScanner(file)
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "|") || strings.Contains(line, "---|") {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) < 7 {
			continue
		}

		idStr := strings.TrimSpace(parts[1])
		id, err := strconv.Atoi(idStr)
		if err != nil {
			continue
		}

		nameRaw := strings.TrimSpace(parts[2])
		name := strings.Trim(nameRaw, "*` ")

		desc := strings.TrimSpace(parts[4])
		stdlib := strings.TrimSpace(parts[5])
		status := strings.TrimSpace(parts[6])

		pathLink := strings.TrimSpace(parts[3])
		linkSeparator := strings.Index(pathLink, "](")
		if linkSeparator < 0 || !strings.HasSuffix(pathLink, ")") {
			return nil, fmt.Errorf("catalog line %d has an invalid library path link", lineNumber)
		}
		relativePath := pathLink[linkSeparator+2 : len(pathLink)-1]
		if !strings.HasPrefix(relativePath, "./") {
			return nil, fmt.Errorf("catalog line %d path must be workspace-relative", lineNumber)
		}
		relativePath = strings.TrimPrefix(relativePath, "./")
		cleanRelativePath := filepath.Clean(filepath.FromSlash(relativePath))
		if cleanRelativePath == "." || cleanRelativePath == ".." ||
			strings.HasPrefix(cleanRelativePath, ".."+string(filepath.Separator)) ||
			filepath.IsAbs(cleanRelativePath) {
			return nil, fmt.Errorf("catalog line %d has an invalid library path", lineNumber)
		}

		pathSegments := strings.Split(filepath.ToSlash(cleanRelativePath), "/")
		module := pathSegments[0]
		libPath := filepath.Join(rootDir, cleanRelativePath)
		moduleDir := filepath.Join(rootDir, module)

		items = append(items, LibraryItem{
			ID:          id,
			Name:        name,
			Path:        libPath,
			Module:      module,
			ModuleDir:   moduleDir,
			ImportPath:  "github.com/amarnathdbg101-coder/go/" + filepath.ToSlash(cleanRelativePath),
			Description: desc,
			StdLib:      stdlib,
			Status:      status,
		})
	}

	return items, scanner.Err()
}

// SearchCatalog filters libraries matching query string in name or description.
func SearchCatalog(query string) ([]LibraryItem, error) {
	items, err := LoadCatalog()
	if err != nil {
		return nil, err
	}

	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return items, nil
	}

	var results []LibraryItem
	for _, item := range items {
		if strings.Contains(strings.ToLower(item.Name), query) ||
			strings.Contains(strings.ToLower(item.Description), query) {
			results = append(results, item)
		}
	}

	return results, nil
}

// GetLibrary returns a specific library item by exact name or numeric ID.
func GetLibrary(identifier string) (*LibraryItem, error) {
	items, err := LoadCatalog()
	if err != nil {
		return nil, err
	}

	idNum, _ := strconv.Atoi(identifier)

	for _, item := range items {
		if strings.EqualFold(item.Name, identifier) || (idNum > 0 && item.ID == idNum) {
			return &item, nil
		}
	}

	return nil, fmt.Errorf("library '%s' not found in catalog", identifier)
}

// ReadDoc reads the specified document file (SPEC.md or README.md) from the library folder.
func (item *LibraryItem) ReadDoc(filename string) (string, error) {
	docPath := filepath.Join(item.Path, filename)
	data, err := os.ReadFile(docPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
