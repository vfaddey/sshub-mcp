package mcpadapter

import (
	"net/http"
	"strings"

	"sshub-mcp/internal/domain/ports"
	"sshub-mcp/internal/usecase/scopectx"
)

type AuthConfig struct {
	Resolver ports.TokenResolver
}

func authMiddleware(next http.Handler, cfg AuthConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cfg.Resolver == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		raw := bearerToken(r.Header.Get("Authorization"))
		if raw == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		scope, err := cfg.Resolver.ResolveToken(r.Context(), raw)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if scope.TokenID == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r.WithContext(scopectx.With(r.Context(), scope)))
	})
}

func bearerToken(h string) string {
	const p = "Bearer "
	h = strings.TrimSpace(h)
	if !strings.HasPrefix(h, p) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(h, p))
}
