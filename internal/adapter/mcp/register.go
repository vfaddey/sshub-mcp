package mcpadapter

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"sshub-mcp/internal/usecase/closesshsession"
	"sshub-mcp/internal/usecase/execssh"
	"sshub-mcp/internal/usecase/listhosts"
	"sshub-mcp/internal/usecase/listprojects"
	"sshub-mcp/internal/usecase/listsshsessions"
	"sshub-mcp/internal/usecase/opensshsession"
)

type ToolSet struct {
	ListProjects    *listprojects.UseCase
	ListHosts       *listhosts.UseCase
	OpenSSHSession  *opensshsession.UseCase
	ExecSSH         *execssh.UseCase
	CloseSSHSession *closesshsession.UseCase
	ListSSHSessions *listsshsessions.UseCase
}

func RegisterTools(s *mcp.Server, t *ToolSet) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_projects",
		Description: "List projects the user has access to. Call this first to discover available projects before listing hosts or creating SSH sessions. Returns project IDs and names.",
	}, listProjects(t))
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_hosts",
		Description: "List SSH hosts for a project. Requires project_id from list_projects. Returns hosts with their IDs, names, addresses, and usernames — use host_id when opening a session.",
	}, listHosts(t))
	mcp.AddTool(s, &mcp.Tool{
		Name:        "ssh_create_session",
		Description: "Open an SSH session to a host. Requires project_id and host_id from list_hosts. Returns session_id — use it with ssh_exec to run commands. One session per host; shell state (cwd, env) is preserved between ssh_exec calls.",
	}, createSession(t))
	mcp.AddTool(s, &mcp.Tool{
		Name:        "ssh_exec",
		Description: "Execute a shell command in an existing SSH session. Requires session_id from ssh_create_session and the command string. Output is returned as plain text. Working directory and exported env vars are preserved between calls.",
	}, execSSH(t))
	mcp.AddTool(s, &mcp.Tool{
		Name:        "ssh_close_session",
		Description: "Close an SSH session and release resources. Requires session_id. Call when done with a host to free the connection.",
	}, closeSession(t))
	mcp.AddTool(s, &mcp.Tool{
		Name:        "ssh_list_sessions",
		Description: "List active SSH sessions for a project. Requires project_id. Use to see which sessions are open or to find a session_id for ssh_exec or ssh_close_session.",
	}, listSessions(t))
}

func toJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return `{"error":"encode"}`
	}
	return string(b)
}

func listProjects(t *ToolSet) mcp.ToolHandlerFor[struct{}, any] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		list, err := t.ListProjects.Execute(ctx)
		if err != nil {
			return nil, nil, err
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: toJSON(list)}}}, nil, nil
	}
}

type listHostsIn struct {
	ProjectID int64 `json:"project_id" jsonschema:"Numeric ID of the project (from list_projects)"`
}

func listHosts(t *ToolSet) mcp.ToolHandlerFor[listHostsIn, any] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in listHostsIn) (*mcp.CallToolResult, any, error) {
		list, err := t.ListHosts.Execute(ctx, in.ProjectID)
		if err != nil {
			return nil, nil, err
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: toJSON(list)}}}, nil, nil
	}
}

type createSessionIn struct {
	ProjectID int64 `json:"project_id" jsonschema:"Numeric ID of the project"`
	HostID    int64 `json:"host_id" jsonschema:"Numeric ID of the host (from list_hosts)"`
}

func createSession(t *ToolSet) mcp.ToolHandlerFor[createSessionIn, any] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in createSessionIn) (*mcp.CallToolResult, any, error) {
		sess, err := t.OpenSSHSession.Execute(ctx, in.ProjectID, in.HostID)
		if err != nil {
			return nil, nil, err
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: toJSON(sess)}}}, nil, nil
	}
}

type execIn struct {
	SessionID string `json:"session_id" jsonschema:"ID of the SSH session (from ssh_create_session)"`
	Command   string `json:"command" jsonschema:"Shell command to execute (e.g. ls -la, cat /etc/hostname)"`
}

func execSSH(t *ToolSet) mcp.ToolHandlerFor[execIn, any] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in execIn) (*mcp.CallToolResult, any, error) {
		out, err := t.ExecSSH.Execute(ctx, in.SessionID, in.Command)
		if err != nil {
			return nil, nil, err
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: out}}}, nil, nil
	}
}

type closeSessionIn struct {
	SessionID string `json:"session_id" jsonschema:"ID of the SSH session to close"`
}

func closeSession(t *ToolSet) mcp.ToolHandlerFor[closeSessionIn, any] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in closeSessionIn) (*mcp.CallToolResult, any, error) {
		if err := t.CloseSSHSession.Execute(ctx, in.SessionID); err != nil {
			return nil, nil, err
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: `{"ok":true}`}}}, nil, nil
	}
}

type listSessionsIn struct {
	ProjectID int64 `json:"project_id" jsonschema:"Numeric ID of the project to list sessions for"`
}

func listSessions(t *ToolSet) mcp.ToolHandlerFor[listSessionsIn, any] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in listSessionsIn) (*mcp.CallToolResult, any, error) {
		list, err := t.ListSSHSessions.Execute(ctx, in.ProjectID)
		if err != nil {
			return nil, nil, err
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: toJSON(list)}}}, nil, nil
	}
}
