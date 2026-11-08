package gomigrator

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// Migration represents a single database schema migration step.
type Migration struct {
	Version  int64
	Name     string
	UpSQL    string
	DownSQL  string
	Checksum string
}

// ComputeChecksum generates a SHA-256 hex string for the UpSQL body.
func ComputeChecksum(sql string) string {
	h := sha256.Sum256([]byte(strings.TrimSpace(sql)))
	return hex.EncodeToString(h[:])
}

// NewMigration constructs a Migration struct with calculated checksum.
func NewMigration(version int64, name, upSQL, downSQL string) Migration {
	return Migration{
		Version:  version,
		Name:     name,
		UpSQL:    strings.TrimSpace(upSQL),
		DownSQL:  strings.TrimSpace(downSQL),
		Checksum: ComputeChecksum(upSQL),
	}
}

// FromFS scans an io/fs.FS (such as embed.FS or os.DirFS) for migration scripts.
// Supports files named:
//
//	0001_create_users.up.sql
//	0001_create_users.down.sql
func FromFS(fsys fs.FS, rootDir string) ([]Migration, error) {
	if rootDir == "" {
		rootDir = "."
	}

	entries, err := fs.ReadDir(fsys, rootDir)
	if err != nil {
		return nil, fmt.Errorf("migrator: read fs dir: %w", err)
	}

	// Map of version -> *Migration
	migMap := make(map[int64]*Migration)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		if !strings.HasSuffix(filename, ".sql") {
			continue
		}

		isUp := strings.HasSuffix(filename, ".up.sql")
		isDown := strings.HasSuffix(filename, ".down.sql")
		if !isUp && !isDown {
			continue
		}

		// Strip .up.sql or .down.sql
		baseName := filename
		if isUp {
			baseName = strings.TrimSuffix(filename, ".up.sql")
		} else {
			baseName = strings.TrimSuffix(filename, ".down.sql")
		}

		// Expected format: <version>_<name> e.g. 0001_create_users
		parts := strings.SplitN(baseName, "_", 2)
		if len(parts) < 2 {
			continue
		}

		ver, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			continue
		}
		name := parts[1]

		content, err := fs.ReadFile(fsys, filepath.ToSlash(filepath.Join(rootDir, filename)))
		if err != nil {
			return nil, fmt.Errorf("migrator: read file %s: %w", filename, err)
		}

		mig, exists := migMap[ver]
		if !exists {
			mig = &Migration{
				Version: ver,
				Name:    name,
			}
			migMap[ver] = mig
		}

		if isUp {
			mig.UpSQL = strings.TrimSpace(string(content))
			mig.Checksum = ComputeChecksum(mig.UpSQL)
		} else {
			mig.DownSQL = strings.TrimSpace(string(content))
		}
	}

	migrations := make([]Migration, 0, len(migMap))
	for _, m := range migMap {
		migrations = append(migrations, *m)
	}

	// Sort chronologically by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}
