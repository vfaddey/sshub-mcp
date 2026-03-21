package sshgateway

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"

	"sshub-mcp/internal/domain"
	"sshub-mcp/internal/domain/ports"
)

type Gateway struct {
	mu       sync.Mutex
	sessions map[string]*liveSession
}

type liveSession struct {
	meta     domain.SSHSession
	client   *ssh.Client
	stateMu  sync.Mutex
	cwd      string
	envLines []string
}

const mcpPwdMarker = "\n__MCP_PWD__\n"
const mcpPwdMarkerAlt = "__MCP_PWD__"

func New() *Gateway {
	g := &Gateway{sessions: make(map[string]*liveSession)}
	go g.sweepLoop()
	return g
}

func (g *Gateway) OpenSession(ctx context.Context, host domain.Host, projectID int64, ttl time.Duration) (domain.SSHSession, error) {
	addr := fmt.Sprintf("%s:%d", host.Address, host.Port)
	if host.AuthKind == domain.HostAuthAgent && !sshAgentUsable() {
		return domain.SSHSession{}, domain.ErrValidation
	}
	var agentConn net.Conn
	if useSSHAgent(host) {
		sock := os.Getenv("SSH_AUTH_SOCK")
		if sock == "" {
			return domain.SSHSession{}, domain.ErrValidation
		}
		var err error
		agentConn, err = net.Dial("unix", sock)
		if err != nil {
			return domain.SSHSession{}, err
		}
	}
	cfg, err := clientConfig(host, agentConn)
	if err != nil {
		if agentConn != nil {
			_ = agentConn.Close()
		}
		return domain.SSHSession{}, err
	}
	conn, err := dial(ctx, "tcp", addr)
	if err != nil {
		if agentConn != nil {
			_ = agentConn.Close()
		}
		return domain.SSHSession{}, err
	}
	cc, chans, reqs, err := ssh.NewClientConn(conn, addr, cfg)
	if agentConn != nil {
		_ = agentConn.Close()
	}
	if err != nil {
		_ = conn.Close()
		return domain.SSHSession{}, err
	}
	client := ssh.NewClient(cc, chans, reqs)
	now := time.Now()
	meta := domain.SSHSession{
		ID:        uuid.NewString(),
		ProjectID: projectID,
		HostID:    host.ID,
		ExpiresAt: now.Add(ttl),
		CreatedAt: now,
	}
	g.mu.Lock()
	g.sessions[meta.ID] = &liveSession{meta: meta, client: client}
	g.mu.Unlock()
	return meta, nil
}

func (g *Gateway) SessionInfo(sessionID string) (domain.SSHSession, bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	ls, ok := g.sessions[sessionID]
	if !ok || time.Now().After(ls.meta.ExpiresAt) {
		return domain.SSHSession{}, false
	}
	return ls.meta, true
}

func (g *Gateway) Exec(ctx context.Context, sessionID, command string) (string, error) {
	g.mu.Lock()
	ls, ok := g.sessions[sessionID]
	if !ok || time.Now().After(ls.meta.ExpiresAt) {
		g.mu.Unlock()
		return "", domain.ErrNotFound
	}
	client := ls.client
	g.mu.Unlock()

	ls.stateMu.Lock()
	stored := ls.cwd
	envCopy := append([]string(nil), ls.envLines...)
	ls.stateMu.Unlock()

	script := buildStatefulScript(stored, envCopy, command)
	sess, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer sess.Close()

	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer
	sess.Stdout = &stdoutBuf
	sess.Stderr = &stderrBuf
	sess.Stdin = strings.NewReader(script)

	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			_ = sess.Close()
		case <-done:
		}
	}()

	runErr := sess.Run("/bin/bash -s")
	close(done)

	// Preserve user-visible output: return stdout exactly as a terminal would.
	out := stdoutBuf.String()

	// State is emitted on stderr and must not pollute user output.
	_, newPwd, hasPwd := splitMcpPwdOutput(stderrBuf.Bytes())
	if hasPwd {
		ls.stateMu.Lock()
		ls.cwd = newPwd
		ls.stateMu.Unlock()
	}
	appendExportsFromCommand(ls, command)

	if runErr != nil {
		if ctx.Err() != nil {
			return out, ctx.Err()
		}
		errBody, _, _ := splitMcpPwdOutput(stderrBuf.Bytes())
		errBody = strings.TrimSpace(errBody)
		if errBody != "" {
			if out != "" {
				out = strings.TrimRight(out, "\n") + "\n" + errBody + "\n"
			} else {
				out = errBody + "\n"
			}
			return out, fmt.Errorf("%w: %s", runErr, errBody)
		}
		return out, runErr
	}

	return out, nil
}

func (g *Gateway) Close(ctx context.Context, sessionID string) error {
	g.mu.Lock()
	ls, ok := g.sessions[sessionID]
	if !ok {
		g.mu.Unlock()
		return domain.ErrNotFound
	}
	delete(g.sessions, sessionID)
	cl := ls.client
	g.mu.Unlock()
	return cl.Close()
}

func (g *Gateway) ListOpenByProject(ctx context.Context, projectID int64) ([]domain.SSHSession, error) {
	now := time.Now()
	g.mu.Lock()
	defer g.mu.Unlock()
	var out []domain.SSHSession
	for _, ls := range g.sessions {
		if ls.meta.ProjectID != projectID || now.After(ls.meta.ExpiresAt) {
			continue
		}
		out = append(out, ls.meta)
	}
	return out, nil
}

func (g *Gateway) sweepLoop() {
	t := time.NewTicker(30 * time.Second)
	defer t.Stop()
	for range t.C {
		g.sweep()
	}
}

func (g *Gateway) sweep() {
	now := time.Now()
	g.mu.Lock()
	var dead []string
	for id, ls := range g.sessions {
		if now.After(ls.meta.ExpiresAt) {
			dead = append(dead, id)
		}
	}
	for _, id := range dead {
		ls := g.sessions[id]
		delete(g.sessions, id)
		if ls != nil && ls.client != nil {
			_ = ls.client.Close()
		}
	}
	g.mu.Unlock()
}

func dial(ctx context.Context, network, addr string) (net.Conn, error) {
	var d net.Dialer
	return d.DialContext(ctx, network, addr)
}

func bashSingleQuote(s string) string {
	return "'" + strings.ReplaceAll(s, `'`, `'\''`) + "'"
}

func buildStatefulScript(storedCwd string, envLines []string, cmd string) string {
	var b strings.Builder
	b.WriteString("set +e\n")
	for _, e := range envLines {
		t := strings.TrimSpace(e)
		if t != "" {
			b.WriteString(t)
			b.WriteString("\n")
		}
	}
	if storedCwd != "" {
		b.WriteString("cd ")
		b.WriteString(bashSingleQuote(storedCwd))
		b.WriteString("\n")
	}
	b.WriteString(cmd)
	b.WriteString("\n")

	// Emit state marker on stderr so stdout remains purely user-visible output.
	b.WriteString("printf '\\n__MCP_PWD__\\n' 1>&2\n")
	b.WriteString("pwd -P 1>&2\n")
	b.WriteString("printf '\\n' 1>&2\n")

	return b.String()
}

func splitMcpPwdOutput(raw []byte) (body string, pwd string, ok bool) {
	s := string(raw)

	i := strings.LastIndex(s, mcpPwdMarker)
	markerLen := len(mcpPwdMarker)
	if i < 0 {
		i = strings.LastIndex(s, mcpPwdMarkerAlt)
		markerLen = len(mcpPwdMarkerAlt)
	}

	if i < 0 {
		return strings.TrimSpace(s), "", false
	}

	body = strings.TrimRight(s[:i], "\r\n")

	rest := s[i+markerLen:]
	rest = strings.TrimLeft(rest, "\r\n")
	if rest == "" {
		return body, "", false
	}

	firstLine, _, _ := strings.Cut(rest, "\n")
	firstLine = strings.TrimSpace(strings.TrimRight(firstLine, "\r"))
	if firstLine == "" {
		return body, "", false
	}

	pwd = firstLine
	return body, pwd, true
}

func appendExportsFromCommand(ls *liveSession, cmd string) {
	ls.stateMu.Lock()
	defer ls.stateMu.Unlock()
	for _, line := range strings.Split(cmd, "\n") {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "export ") {
			ls.envLines = append(ls.envLines, t)
		}
	}
}

func clientConfig(host domain.Host, agentConn net.Conn) (*ssh.ClientConfig, error) {
	auth, err := authMethods(host, agentConn)
	if err != nil {
		return nil, err
	}
	if len(auth) == 0 {
		return nil, domain.ErrValidation
	}
	cb := ssh.InsecureIgnoreHostKey()
	if hcb, err := knownHostsCallback(); err == nil {
		cb = hcb
	}
	return &ssh.ClientConfig{
		User:            host.Username,
		Auth:            auth,
		HostKeyCallback: cb,
		Timeout:         15 * time.Second,
	}, nil
}

func knownHostsCallback() (ssh.HostKeyCallback, error) {
	path, err := userKnownHostsPath()
	if err != nil {
		return nil, err
	}
	return knownhosts.New(path)
}

func useSSHAgent(host domain.Host) bool {
	if !sshAgentUsable() {
		return false
	}
	switch host.AuthKind {
	case domain.HostAuthAgent:
		return true
	case domain.HostAuthNone:
		return true
	default:
		return false
	}
}

func authMethods(host domain.Host, agentConn net.Conn) ([]ssh.AuthMethod, error) {
	switch host.AuthKind {
	case domain.HostAuthNone:
		var methods []ssh.AuthMethod
		if agentConn != nil {
			signers, err := agent.NewClient(agentConn).Signers()
			if err != nil {
				return nil, err
			}
			if len(signers) > 0 {
				methods = append(methods, ssh.PublicKeys(signers...))
			}
		}
		fileSigners, err := loadDefaultIdentitySigners()
		if err != nil {
			return nil, err
		}
		if len(fileSigners) > 0 {
			methods = append(methods, ssh.PublicKeys(fileSigners...))
		}
		if len(methods) == 0 {
			return nil, domain.ErrValidation
		}
		return methods, nil
	case domain.HostAuthPassword:
		return []ssh.AuthMethod{ssh.Password(host.Password)}, nil
	case domain.HostAuthAgent:
		if agentConn == nil {
			return nil, domain.ErrValidation
		}
		signers, err := agent.NewClient(agentConn).Signers()
		if err != nil {
			return nil, err
		}
		if len(signers) == 0 {
			return nil, domain.ErrValidation
		}
		return []ssh.AuthMethod{ssh.PublicKeys(signers...)}, nil
	default:
		return nil, domain.ErrValidation
	}
}

var _ ports.SSHGateway = (*Gateway)(nil)
