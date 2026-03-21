package listsshsessions

import (
	"context"

	"sshub-mcp/internal/domain"
	"sshub-mcp/internal/domain/ports"
	"sshub-mcp/internal/usecase/scopectx"
)

type UseCase struct {
	Gateway ports.SSHGateway
}

func (u *UseCase) Execute(ctx context.Context, projectID int64) ([]domain.SSHSession, error) {
	if projectID <= 0 {
		return nil, domain.ErrValidation
	}
	scope := scopectx.From(ctx)
	if !scope.MayAccessProject(projectID) {
		return nil, domain.ErrForbidden
	}
	return u.Gateway.ListOpenByProject(ctx, projectID)
}
