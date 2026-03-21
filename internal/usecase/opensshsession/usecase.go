package opensshsession

import (
	"context"
	"time"

	"sshub-mcp/internal/domain"
	"sshub-mcp/internal/domain/ports"
	"sshub-mcp/internal/usecase/scopectx"
)

type UseCase struct {
	Catalog    ports.Catalog
	Gateway    ports.SSHGateway
	SessionTTL time.Duration
}

func (u *UseCase) Execute(ctx context.Context, projectID, hostID int64) (domain.SSHSession, error) {
	if projectID <= 0 || hostID <= 0 {
		return domain.SSHSession{}, domain.ErrValidation
	}
	scope := scopectx.From(ctx)
	if !scope.MayAccessProject(projectID) {
		return domain.SSHSession{}, domain.ErrForbidden
	}
	host, err := u.Catalog.GetHostForSSH(ctx, scope, projectID, hostID)
	if err != nil {
		return domain.SSHSession{}, err
	}
	return u.Gateway.OpenSession(ctx, host, projectID, u.SessionTTL)
}
