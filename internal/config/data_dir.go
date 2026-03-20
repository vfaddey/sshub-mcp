package config

import (
	"os"
	"path/filepath"
	"runtime"
)

// dataDir returns the per-user data directory for sshub-mcp (SQLite, etc.).
// goos is usually runtime.GOOS; exposed for tests.
func dataDir(goos, home, xdgDataHome string) string {
	switch goos {
	case "darwin":
		if home == "" {
			return "sshub-mcp"
		}
		return filepath.Join(home, "Library", "Application Support", "sshub-mcp")
	case "linux", "freebsd", "openbsd", "netbsd", "dragonfly":
		base := xdgDataHome
		if base == "" {
			if home == "" {
				return filepath.Join(".local", "share", "sshub-mcp")
			}
			base = filepath.Join(home, ".local", "share")
		}
		return filepath.Join(base, "sshub-mcp")
	default:
		if home == "" {
			return filepath.Join(".local", "share", "sshub-mcp")
		}
		return filepath.Join(home, ".local", "share", "sshub-mcp")
	}
}

func defaultDataDir() string {
	home, _ := os.UserHomeDir()
	return dataDir(runtime.GOOS, home, os.Getenv("XDG_DATA_HOME"))
}

func defaultDBPath() string {
	return filepath.Join(defaultDataDir(), "sshub-mcp.db")
}
