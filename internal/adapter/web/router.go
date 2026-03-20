package web

import (
	"embed"
	"encoding/json"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"sshub-mcp/internal/domain"
	"sshub-mcp/internal/domain/ports"
)

//go:embed all:static
var staticRoot embed.FS

const maxBody = 1 << 20

func NewRouter(admin ports.Admin) http.Handler {
	staticFS, err := fs.Sub(staticRoot, "static")
	if err != nil {
		panic(err)
	}
	assetsFS, err := fs.Sub(staticFS, "assets")
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	mux.Handle("GET /assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(assetsFS))))
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		serveStaticFile(w, r, staticFS, "index.html")
	})
	a := &api{admin: admin}
	mux.HandleFunc("GET /api/projects", a.listProjects)
	mux.HandleFunc("POST /api/projects", a.createProject)
	mux.HandleFunc("GET /api/projects/{id}/hosts", a.listHosts)
	mux.HandleFunc("POST /api/projects/{id}/hosts", a.createHost)
	mux.HandleFunc("POST /api/tokens", a.issueToken)
	return mux
}

func serveStaticFile(w http.ResponseWriter, r *http.Request, fsys fs.FS, name string) {
	b, err := fs.ReadFile(fsys, name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	ct := mime.TypeByExtension(filepath.Ext(name))
	if ct == "" {
		ct = "application/octet-stream"
	}
	if strings.HasPrefix(ct, "text/") && !strings.Contains(ct, "charset") {
		ct += "; charset=utf-8"
	}
	w.Header().Set("Content-Type", ct)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(b)
}

type api struct {
	admin ports.Admin
}

func (a *api) listProjects(w http.ResponseWriter, r *http.Request) {
	list, err := a.admin.ListProjectsAll(r.Context())
	if err != nil {
		writeErr(w, err)
		return
	}
	if list == nil {
		list = []domain.Project{}
	}
	writeJSON(w, http.StatusOK, list)
}

type createProjectBody struct {
	Name string `json:"name"`
}

func (a *api) createProject(w http.ResponseWriter, r *http.Request) {
	var body createProjectBody
	if !readJSON(w, r, &body) {
		return
	}
	p, err := a.admin.CreateProject(r.Context(), body.Name)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

func (a *api) listHosts(w http.ResponseWriter, r *http.Request) {
	pid := r.PathValue("id")
	list, err := a.admin.ListHostsByProject(r.Context(), pid)
	if err != nil {
		writeErr(w, err)
		return
	}
	if list == nil {
		list = []domain.Host{}
	}
	writeJSON(w, http.StatusOK, list)
}

type createHostBody struct {
	Name     string              `json:"name"`
	Address  string              `json:"address"`
	Port     int                 `json:"port"`
	Username string              `json:"username"`
	AuthKind domain.HostAuthKind `json:"auth_kind"`
	Password string              `json:"password"`
}

func (a *api) createHost(w http.ResponseWriter, r *http.Request) {
	pid := r.PathValue("id")
	var body createHostBody
	if !readJSON(w, r, &body) {
		return
	}
	h, err := a.admin.CreateHost(r.Context(), pid, ports.HostCreate{
		Name: body.Name, Address: body.Address, Port: body.Port, Username: body.Username,
		AuthKind: body.AuthKind, Password: body.Password,
	})
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, h)
}

type issueTokenBody struct {
	Label      string   `json:"label"`
	ProjectIDs []string `json:"project_ids"`
}

type issueTokenResponse struct {
	Token    string          `json:"token"`
	APIToken domain.APIToken `json:"api_token"`
}

func (a *api) issueToken(w http.ResponseWriter, r *http.Request) {
	var body issueTokenBody
	if !readJSON(w, r, &body) {
		return
	}
	plain, tok, err := a.admin.IssueToken(r.Context(), body.Label, body.ProjectIDs)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, issueTokenResponse{Token: plain, APIToken: tok})
}

func readJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxBody)
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		if err == io.EOF {
			http.Error(w, "empty body", http.StatusBadRequest)
			return false
		}
		http.Error(w, "invalid json", http.StatusBadRequest)
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, err error) {
	switch err {
	case domain.ErrNotFound:
		http.Error(w, err.Error(), http.StatusNotFound)
	case domain.ErrForbidden:
		http.Error(w, err.Error(), http.StatusForbidden)
	case domain.ErrValidation:
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
