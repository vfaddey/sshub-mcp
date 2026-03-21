package ports

import (
	"context"

	"sshub-mcp/internal/domain"
)

type Catalog interface {
	ListProjects(ctx context.Context, scope domain.AccessScope) ([]domain.Project, error)
	ListHosts(ctx context.Context, scope domain.AccessScope, projectID int64) ([]domain.Host, error)
	GetHostForSSH(ctx context.Context, scope domain.AccessScope, projectID int64, hostID int64) (domain.Host, error)
}
