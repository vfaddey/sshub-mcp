package listhosts

import (
	"context"

	"sshub-mcp/internal/domain"
	"sshub-mcp/internal/domain/ports"
	"sshub-mcp/internal/usecase/scopectx"
)

type UseCase struct {
	Catalog ports.Catalog
}

func (u *UseCase) Execute(ctx context.Context, projectID int64) ([]domain.Host, error) {
	if projectID <= 0 {
		return nil, domain.ErrValidation
	}
	scope := scopectx.From(ctx)
	if !scope.MayAccessProject(projectID) {
		return nil, domain.ErrForbidden
	}
	return u.Catalog.ListHosts(ctx, scope, projectID)
}
