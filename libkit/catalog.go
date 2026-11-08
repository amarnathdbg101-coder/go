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
	Description string
	StdLib      string
	Status      string
}

// FindLibraryRoot finds the root C:\Library directory.
func FindLibraryRoot() string {
	// 1. Check current directory and parents
	curr, err := os.Getwd()
	if err == nil {
		for {
			if _, err := os.Stat(filepath.Join(curr, "CATALOG.md")); err == nil {
				return curr
			}
			parent := filepath.Dir(curr)
			if parent == curr {
				break
			}
			curr = parent
		}
	}

	// 2. Fallback to standard C:\Library
	if _, err := os.Stat(`C:\Library\CATALOG.md`); err == nil {
		return `C:\Library`
	}

	return ""
}

// LoadCatalog parses CATALOG.md to extract all cataloged libraries.
func LoadCatalog() ([]LibraryItem, error) {
	rootDir := FindLibraryRoot()
	if rootDir == "" {
		return nil, fmt.Errorf("could not locate C:\\Library\\CATALOG.md")
	}

	catalogFile := filepath.Join(rootDir, "CATALOG.md")
	file, err := os.Open(catalogFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var items []LibraryItem
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
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

		libPath := filepath.Join(rootDir, name)

		items = append(items, LibraryItem{
			ID:          id,
			Name:        name,
			Path:        libPath,
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
