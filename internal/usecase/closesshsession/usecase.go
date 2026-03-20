package closesshsession

import (
	"context"
	"strings"

	"sshub-mcp/internal/domain"
	"sshub-mcp/internal/domain/ports"
	"sshub-mcp/internal/usecase/scopectx"
)

type UseCase struct {
	Gateway ports.SSHGateway
}

func (u *UseCase) Execute(ctx context.Context, sessionID string) error {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return domain.ErrValidation
	}
	info, ok := u.Gateway.SessionInfo(sessionID)
	if !ok {
		return domain.ErrNotFound
	}
	scope := scopectx.From(ctx)
	if !scope.MayAccessProject(info.ProjectID) {
		return domain.ErrForbidden
	}
	return u.Gateway.Close(ctx, sessionID)
}
