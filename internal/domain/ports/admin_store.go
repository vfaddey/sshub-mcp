package ports

import (
	"context"

	"sshub-mcp/internal/domain"
)

type HostCreate struct {
	Name     string
	Address  string
	Port     int
	Username string
	AuthKind domain.HostAuthKind
	Password string
}

type Admin interface {
	ListProjectsAll(ctx context.Context) ([]domain.Project, error)
	CreateProject(ctx context.Context, name string) (domain.Project, error)
	ListHostsByProject(ctx context.Context, projectID string) ([]domain.Host, error)
	CreateHost(ctx context.Context, projectID string, in HostCreate) (domain.Host, error)
	IssueToken(ctx context.Context, label string, projectIDs []string) (plainSecret string, t domain.APIToken, err error)
}

type TokenResolver interface {
	ResolveToken(ctx context.Context, plainSecret string) (domain.AccessScope, error)
}
