package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	mcpadapter "sshub-mcp/internal/adapter/mcp"
	"sshub-mcp/internal/adapter/persistence/sqlite"
	sshgateway "sshub-mcp/internal/adapter/sshgateway"
	"sshub-mcp/internal/adapter/web"
	"sshub-mcp/internal/config"
	"sshub-mcp/internal/usecase/closesshsession"
	"sshub-mcp/internal/usecase/execssh"
	"sshub-mcp/internal/usecase/listhosts"
	"sshub-mcp/internal/usecase/listprojects"
	"sshub-mcp/internal/usecase/listsshsessions"
	"sshub-mcp/internal/usecase/opensshsession"
)

type App struct {
	cfg   config.Config
	log   *slog.Logger
	db    *sql.DB
	store *sqlite.Store
	sshGW *sshgateway.Gateway
}

func New(cfg config.Config) (*App, error) {
	if err := config.EnsureParentDir(cfg.DBPath); err != nil {
		return nil, err
	}
	db, err := sqlite.Open(cfg.DBPath)
	if err != nil {
		return nil, fmt.Errorf("sqlite: %w", err)
	}
	if err := sqlite.Migrate(context.Background(), db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	st := sqlite.NewStore(db)
	gw := sshgateway.New()
	return &App{cfg: cfg, log: slog.Default(), db: db, store: st, sshGW: gw}, nil
}

func (a *App) Run() error {
	defer func() { _ = a.db.Close() }()
	tools := &mcpadapter.ToolSet{
		ListProjects: &listprojects.UseCase{Catalog: a.store},
		ListHosts:    &listhosts.UseCase{Catalog: a.store},
		OpenSSHSession: &opensshsession.UseCase{
			Catalog: a.store, Gateway: a.sshGW, SessionTTL: a.cfg.SessionTTL,
		},
		ExecSSH:         &execssh.UseCase{Gateway: a.sshGW},
		CloseSSHSession: &closesshsession.UseCase{Gateway: a.sshGW},
		ListSSHSessions: &listsshsessions.UseCase{Gateway: a.sshGW},
	}
	mcpH := mcpadapter.NewHandler(tools, a.store)
	admin := web.NewRouter(a.store)
	adminMounted := http.StripPrefix(strings.TrimSuffix(a.cfg.AdminPrefix, "/"), admin)
	mux := http.NewServeMux()
	mcpStrip := strings.TrimSuffix(a.cfg.MCPPath, "/")
	mcpMount := mountMCPPath(mcpStrip, mcpH)
	mux.Handle(mcpStrip+"/", mcpMount)
	mux.Handle(mcpStrip, mcpMount)
	mux.Handle(a.cfg.AdminPrefix+"/", adminMounted)
	mux.Handle(a.cfg.AdminPrefix, http.RedirectHandler(a.cfg.AdminPrefix+"/", http.StatusSeeOther))
	a.log.Info("listen", "addr", a.cfg.HTTPAddr, "mcp", a.cfg.MCPPath, "admin", a.cfg.AdminPrefix)
	return http.ListenAndServe(a.cfg.HTTPAddr, mux)
}

func mountMCPPath(prefix string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, prefix) {
			http.NotFound(w, r)
			return
		}
		suffix := strings.TrimPrefix(r.URL.Path, prefix)
		if suffix == "" {
			suffix = "/"
		}
		r2 := r.Clone(r.Context())
		r2.URL = cloneURLPath(r.URL, suffix)
		h.ServeHTTP(w, r2)
	})
}

func cloneURLPath(u *url.URL, path string) *url.URL {
	u2 := *u
	u2.Path = path
	u2.RawPath = ""
	return &u2
}
