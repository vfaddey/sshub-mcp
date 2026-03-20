package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// EnsureParentDir creates the parent directory of dbPath if needed (for SQLite file).
// No-op if dbPath has no directory component (e.g. "sshub-mcp.db" in cwd).
func EnsureParentDir(dbPath string) error {
	dir := filepath.Dir(dbPath)
	if dir == "." || dir == "" || dir == string(filepath.Separator) {
		return nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create data dir %q: %w", dir, err)
	}
	return nil
}
