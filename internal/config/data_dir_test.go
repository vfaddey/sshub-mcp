package config

import (
	"path/filepath"
	"testing"
)

func TestDataDir(t *testing.T) {
	t.Parallel()
	if g := dataDir("darwin", "/Users/x", ""); g != filepath.Join("/Users/x", "Library", "Application Support", "sshub-mcp") {
		t.Errorf("darwin: got %q", g)
	}
	if g := dataDir("darwin", "", ""); g != "sshub-mcp" {
		t.Errorf("darwin no home: got %q", g)
	}
	if g := dataDir("linux", "/home/x", "/xdg"); g != filepath.Join("/xdg", "sshub-mcp") {
		t.Errorf("linux+XDG_DATA_HOME: got %q", g)
	}
	if g := dataDir("linux", "/home/x", ""); g != filepath.Join("/home/x", ".local", "share", "sshub-mcp") {
		t.Errorf("linux default: got %q", g)
	}
	if g := dataDir("freebsd", "/home/x", ""); g != filepath.Join("/home/x", ".local", "share", "sshub-mcp") {
		t.Errorf("freebsd: got %q", g)
	}
	if g := dataDir("plan9", "/u", ""); g != filepath.Join("/u", ".local", "share", "sshub-mcp") {
		t.Errorf("fallback OS: got %q", g)
	}
}
