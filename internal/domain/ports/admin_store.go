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
	DeleteProject(ctx context.Context, projectID int64) error

	ListHostsByProject(ctx context.Context, projectID int64) ([]domain.Host, error)
	CreateHost(ctx context.Context, projectID int64, in HostCreate) (domain.Host, error)
	DeleteHost(ctx context.Context, projectID, hostID int64) error

	ListTokensAll(ctx context.Context) ([]domain.APIToken, error)
	IssueToken(ctx context.Context, label string, projectIDs []int64) (plainSecret string, t domain.APIToken, err error)
	DeleteToken(ctx context.Context, tokenID int64) error
}

type TokenResolver interface {
	ResolveToken(ctx context.Context, plainSecret string) (domain.AccessScope, error)
}
