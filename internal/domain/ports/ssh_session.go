package ports

import (
	"context"
	"time"

	"sshub-mcp/internal/domain"
)

type SSHGateway interface {
	OpenSession(ctx context.Context, host domain.Host, projectID int64, ttl time.Duration) (domain.SSHSession, error)
	SessionInfo(sessionID string) (domain.SSHSession, bool)
	Exec(ctx context.Context, sessionID, command string) (string, error)
	Close(ctx context.Context, sessionID string) error
	ListOpenByProject(ctx context.Context, projectID int64) ([]domain.SSHSession, error)
}
