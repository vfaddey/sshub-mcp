package scopectx

import (
	"context"

	"sshub-mcp/internal/domain"
)

type ctxKey struct{}

func With(ctx context.Context, s domain.AccessScope) context.Context {
	return context.WithValue(ctx, ctxKey{}, s)
}

func From(ctx context.Context) domain.AccessScope {
	v, _ := ctx.Value(ctxKey{}).(domain.AccessScope)
	return v
}
