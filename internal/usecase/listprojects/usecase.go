package listprojects

import (
	"context"

	"sshub-mcp/internal/domain"
	"sshub-mcp/internal/domain/ports"
	"sshub-mcp/internal/usecase/scopectx"
)

type UseCase struct {
	Catalog ports.Catalog
}

func (u *UseCase) Execute(ctx context.Context) ([]domain.Project, error) {
	return u.Catalog.ListProjects(ctx, scopectx.From(ctx))
}
