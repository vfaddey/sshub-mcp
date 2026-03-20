package sqlite

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"sshub-mcp/internal/domain"
	"sshub-mcp/internal/domain/ports"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) ListProjects(ctx context.Context, scope domain.AccessScope) ([]domain.Project, error) {
	if len(scope.ProjectIDs) == 0 {
		return nil, nil
	}
	q := `SELECT id, name, created_at, updated_at FROM projects WHERE id IN (` + placeholders(len(scope.ProjectIDs)) + `) ORDER BY name`
	args := make([]any, len(scope.ProjectIDs))
	for i, id := range scope.ProjectIDs {
		args[i] = id
	}
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanProjects(rows)
}

func (s *Store) ListHosts(ctx context.Context, scope domain.AccessScope, projectID string) ([]domain.Host, error) {
	if !scope.MayAccessProject(projectID) {
		return nil, domain.ErrForbidden
	}
	const q = `SELECT id, project_id, name, address, port, username, auth_kind, created_at, updated_at FROM hosts WHERE project_id = ? ORDER BY name`
	rows, err := s.db.QueryContext(ctx, q, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Host
	for rows.Next() {
		h, err := scanHostRowNoPassword(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

func (s *Store) GetHostForSSH(ctx context.Context, scope domain.AccessScope, projectID, hostID string) (domain.Host, error) {
	if !scope.MayAccessProject(projectID) {
		return domain.Host{}, domain.ErrForbidden
	}
	const q = `SELECT id, project_id, name, address, port, username, auth_kind, password, created_at, updated_at FROM hosts WHERE id = ? AND project_id = ?`
	row := s.db.QueryRowContext(ctx, q, hostID, projectID)
	h, err := scanHostRowFull(row)
	if err == sql.ErrNoRows {
		return domain.Host{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Host{}, err
	}
	return h, nil
}

func (s *Store) ListProjectsAll(ctx context.Context) ([]domain.Project, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, created_at, updated_at FROM projects ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanProjects(rows)
}

func (s *Store) CreateProject(ctx context.Context, name string) (domain.Project, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return domain.Project{}, domain.ErrValidation
	}
	now := time.Now().Unix()
	id := uuid.NewString()
	_, err := s.db.ExecContext(ctx, `INSERT INTO projects (id, name, created_at, updated_at) VALUES (?,?,?,?)`, id, name, now, now)
	if err != nil {
		return domain.Project{}, err
	}
	return domain.Project{ID: id, Name: name, CreatedAt: time.Unix(now, 0), UpdatedAt: time.Unix(now, 0)}, nil
}

func (s *Store) ListHostsByProject(ctx context.Context, projectID string) ([]domain.Host, error) {
	const q = `SELECT id, project_id, name, address, port, username, auth_kind, created_at, updated_at FROM hosts WHERE project_id = ? ORDER BY name`
	rows, err := s.db.QueryContext(ctx, q, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Host
	for rows.Next() {
		h, err := scanHostRowNoPassword(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

func (s *Store) CreateHost(ctx context.Context, projectID string, in ports.HostCreate) (domain.Host, error) {
	if strings.TrimSpace(in.Name) == "" || strings.TrimSpace(in.Address) == "" || strings.TrimSpace(in.Username) == "" {
		return domain.Host{}, domain.ErrValidation
	}
	port := in.Port
	if port <= 0 {
		port = 22
	}
	switch in.AuthKind {
	case domain.HostAuthNone, domain.HostAuthAgent:
	case domain.HostAuthPassword:
		if in.Password == "" {
			return domain.Host{}, domain.ErrValidation
		}
	default:
		return domain.Host{}, domain.ErrValidation
	}
	var exists int
	err := s.db.QueryRowContext(ctx, `SELECT 1 FROM projects WHERE id = ?`, projectID).Scan(&exists)
	if err == sql.ErrNoRows {
		return domain.Host{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Host{}, err
	}
	now := time.Now().Unix()
	id := uuid.NewString()
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO hosts (id, project_id, name, address, port, username, auth_kind, password, created_at, updated_at) VALUES (?,?,?,?,?,?,?,?,?,?)`,
		id, projectID, strings.TrimSpace(in.Name), strings.TrimSpace(in.Address), port, strings.TrimSpace(in.Username), string(in.AuthKind), nullIfEmpty(in.Password), now, now,
	)
	if err != nil {
		return domain.Host{}, err
	}
	return domain.Host{
		ID: id, ProjectID: projectID, Name: in.Name, Address: in.Address, Port: port, Username: in.Username,
		AuthKind: in.AuthKind, CreatedAt: time.Unix(now, 0), UpdatedAt: time.Unix(now, 0),
	}, nil
}

func (s *Store) IssueToken(ctx context.Context, label string, projectIDs []string) (string, domain.APIToken, error) {
	label = strings.TrimSpace(label)
	if label == "" {
		return "", domain.APIToken{}, domain.ErrValidation
	}
	for _, pid := range projectIDs {
		var one int
		err := s.db.QueryRowContext(ctx, `SELECT 1 FROM projects WHERE id = ?`, pid).Scan(&one)
		if err == sql.ErrNoRows {
			return "", domain.APIToken{}, domain.ErrValidation
		}
		if err != nil {
			return "", domain.APIToken{}, err
		}
	}
	id := uuid.NewString()
	var rnd [16]byte
	if _, err := rand.Read(rnd[:]); err != nil {
		return "", domain.APIToken{}, err
	}
	plain := id + "." + hex.EncodeToString(rnd[:])
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", domain.APIToken{}, err
	}
	now := time.Now().Unix()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", domain.APIToken{}, err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `INSERT INTO api_tokens (id, label, secret_hash, created_at, updated_at) VALUES (?,?,?,?,?)`, id, label, string(hash), now, now); err != nil {
		return "", domain.APIToken{}, err
	}
	for _, pid := range projectIDs {
		if _, err := tx.ExecContext(ctx, `INSERT INTO token_projects (token_id, project_id) VALUES (?,?)`, id, pid); err != nil {
			return "", domain.APIToken{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return "", domain.APIToken{}, err
	}
	tok := domain.APIToken{ID: id, Label: label, CreatedAt: time.Unix(now, 0), UpdatedAt: time.Unix(now, 0)}
	return plain, tok, nil
}

func (s *Store) ResolveToken(ctx context.Context, plainSecret string) (domain.AccessScope, error) {
	plainSecret = strings.TrimSpace(plainSecret)
	dot := strings.IndexByte(plainSecret, '.')
	if dot <= 0 || dot >= len(plainSecret)-1 {
		return domain.AccessScope{}, domain.ErrValidation
	}
	id := plainSecret[:dot]
	var hash string
	err := s.db.QueryRowContext(ctx, `SELECT secret_hash FROM api_tokens WHERE id = ?`, id).Scan(&hash)
	if err == sql.ErrNoRows {
		return domain.AccessScope{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.AccessScope{}, err
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(plainSecret)) != nil {
		return domain.AccessScope{}, domain.ErrNotFound
	}
	rows, err := s.db.QueryContext(ctx, `SELECT project_id FROM token_projects WHERE token_id = ?`, id)
	if err != nil {
		return domain.AccessScope{}, err
	}
	defer rows.Close()
	var pids []string
	for rows.Next() {
		var pid string
		if err := rows.Scan(&pid); err != nil {
			return domain.AccessScope{}, err
		}
		pids = append(pids, pid)
	}
	if err := rows.Err(); err != nil {
		return domain.AccessScope{}, err
	}
	return domain.AccessScope{TokenID: id, ProjectIDs: pids}, nil
}

func placeholders(n int) string {
	if n == 0 {
		return ""
	}
	b := strings.Builder{}
	for i := range n {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteByte('?')
	}
	return b.String()
}

func scanProjects(rows *sql.Rows) ([]domain.Project, error) {
	var out []domain.Project
	for rows.Next() {
		var p domain.Project
		var ca, ua int64
		if err := rows.Scan(&p.ID, &p.Name, &ca, &ua); err != nil {
			return nil, err
		}
		p.CreatedAt = time.Unix(ca, 0)
		p.UpdatedAt = time.Unix(ua, 0)
		out = append(out, p)
	}
	return out, rows.Err()
}

func scanHostRowNoPassword(rows *sql.Rows) (domain.Host, error) {
	var h domain.Host
	var ca, ua int64
	var ak string
	if err := rows.Scan(&h.ID, &h.ProjectID, &h.Name, &h.Address, &h.Port, &h.Username, &ak, &ca, &ua); err != nil {
		return domain.Host{}, err
	}
	h.AuthKind = domain.HostAuthKind(ak)
	h.CreatedAt = time.Unix(ca, 0)
	h.UpdatedAt = time.Unix(ua, 0)
	return h, nil
}

func scanHostRowFull(sc interface {
	Scan(dest ...any) error
}) (domain.Host, error) {
	var h domain.Host
	var ca, ua int64
	var ak string
	var pwd sql.NullString
	if err := sc.Scan(&h.ID, &h.ProjectID, &h.Name, &h.Address, &h.Port, &h.Username, &ak, &pwd, &ca, &ua); err != nil {
		return domain.Host{}, err
	}
	h.AuthKind = domain.HostAuthKind(ak)
	if pwd.Valid {
		h.Password = pwd.String
	}
	h.CreatedAt = time.Unix(ca, 0)
	h.UpdatedAt = time.Unix(ua, 0)
	return h, nil
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

var (
	_ ports.Catalog       = (*Store)(nil)
	_ ports.Admin         = (*Store)(nil)
	_ ports.TokenResolver = (*Store)(nil)
)
