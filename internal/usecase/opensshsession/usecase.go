package opensshsession

import (
	"context"
	"strings"
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

func (u *UseCase) Execute(ctx context.Context, projectID, hostID string) (domain.SSHSession, error) {
	projectID = strings.TrimSpace(projectID)
	hostID = strings.TrimSpace(hostID)
	if projectID == "" || hostID == "" {
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
