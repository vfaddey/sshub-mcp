package config

import (
	"os"
	"time"
)

const defaultHTTPAddr = "127.0.0.1:8787"

type Config struct {
	HTTPAddr    string
	DBPath      string
	SessionTTL  time.Duration
	AdminPrefix string
	MCPPath     string
}

// envOrLegacy reads primary (SSHUB_MCP_*) first, then legacy (SSH_MCP_*) for backward compatibility.
func envOrLegacy(primary, legacy string) string {
	if v := os.Getenv(primary); v != "" {
		return v
	}
	return os.Getenv(legacy)
}

func Load() Config {
	cfg := Config{
		HTTPAddr:    defaultHTTPAddr,
		DBPath:      defaultDBPathResolved(),
		SessionTTL:  10 * time.Minute,
		AdminPrefix: "/admin",
		MCPPath:     "/mcp",
	}
	if v := envOrLegacy("SSHUB_MCP_HTTP_ADDR", "SSH_MCP_HTTP_ADDR"); v != "" {
		cfg.HTTPAddr = v
	}
	if v := envOrLegacy("SSHUB_MCP_SESSION_TTL", "SSH_MCP_SESSION_TTL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.SessionTTL = d
		}
	}
	return cfg
}

func defaultDBPathResolved() string {
	if v := envOrLegacy("SSHUB_MCP_DB", "SSH_MCP_DB"); v != "" {
		return v
	}
	return defaultDBPath()
}
