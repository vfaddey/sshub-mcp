# sshub-mcp

HTTP [Model Context Protocol](https://modelcontextprotocol.io/) server: projects, SSH targets, API tokens, and tools to open SSH sessions and run commands. Data is stored in **SQLite**. Intended to run **locally**.

Private SSH keys are **not** stored in the database. Authentication uses optional **password** (DB), **ssh-agent** (`SSH_AUTH_SOCK`), or default keys under **`~/.ssh`** (`id_ed25519`, `id_ecdsa`, `id_rsa`) for `auth_kind: none`.

## Requirements

- **Go 1.25+**
- Remote hosts must be reachable from the machine running sshub-mcp.

## Build & run

```bash
go build -o sshub-mcp ./cmd/sshub-mcp
./sshub-mcp
```

Or:

```bash
go run ./cmd/sshub-mcp
```

Default listen address: **`127.0.0.1:8787`** (localhost only).

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `SSHUB_MCP_HTTP_ADDR` | `127.0.0.1:8787` | Listen address |
| `SSHUB_MCP_DB` | platform path (see below) | SQLite database file |
| `SSHUB_MCP_SESSION_TTL` | `10m` | TTL for logical SSH sessions (Go duration) |

The same settings still work under the legacy names `SSH_MCP_HTTP_ADDR`, `SSH_MCP_DB`, `SSH_MCP_SESSION_TTL` if the `SSHUB_MCP_*` variable is not set.

Default database path if neither `SSHUB_MCP_DB` nor `SSH_MCP_DB` is set:

- **Linux / *BSD:** `$XDG_DATA_HOME/sshub-mcp/sshub-mcp.db` or `~/.local/share/sshub-mcp/sshub-mcp.db`
- **macOS:** `~/Library/Application Support/sshub-mcp/sshub-mcp.db`

The data directory is created automatically on first start.

HTTP routes (defaults):

| Path | Purpose |
|------|---------|
| `/mcp` | MCP (Streamable HTTP), **requires** `Authorization: Bearer <token>` |
| `/admin` | Web UI + JSON API (**no authentication**) |

Both `/mcp` and `/mcp/` are accepted for MCP.

## MCP tools

All tool calls are scoped by the token’s attached projects.

| Tool | Description |
|------|-------------|
| `list_projects` | List allowed projects |
| `list_hosts` | List hosts for a `project_id` |
| `ssh_create_session` | Open SSH session (`project_id`, `host_id`); returns `session_id` |
| `ssh_exec` | Run a command in that session (`session_id`, `command`) |
| `ssh_close_session` | Close session (`session_id`) |
| `ssh_list_sessions` | List open sessions for a `project_id` |

Within one `session_id`, **`ssh_exec` keeps shell-like state** between calls: working directory (via tracked `pwd -P`) and prior `export …` lines are replayed before each command. Each exec still uses a new remote `bash -s` script; this is not a full interactive PTY.

## Cursor / `mcp.json`

Example (global: `~/.cursor/mcp.json` or project `.cursor/mcp.json`):

```json
{
  "mcpServers": {
    "sshub-mcp": {
      "url": "http://127.0.0.1:8787/mcp",
      "headers": {
        "Authorization": "Bearer ${env:SSHUB_MCP_TOKEN}"
      }
    }
  }
}
```

Create a token in the admin UI (`POST /admin/api/tokens`) and set `SSHUB_MCP_TOKEN` (or any env name you reference in `mcp.json`) to the **full** secret string (shown once). Restart the editor after changing `mcp.json`.

## Admin HTTP API

Mounted under **`/admin`** (e.g. `http://127.0.0.1:8787/admin/`). JSON bodies, `Content-Type: application/json`.

| Method | Path | Body / notes |
|--------|------|--------------|
| `GET` | `/api/projects` | List projects |
| `POST` | `/api/projects` | `{"name":"..."}` |
| `GET` | `/api/projects/{id}/hosts` | List hosts |
| `POST` | `/api/projects/{id}/hosts` | `name`, `address`, `port`, `username`, `auth_kind`, optional `password` |
| `POST` | `/api/tokens` | `label`, `project_ids` — response includes `token` **once** |

`auth_kind`: `none` | `password` | `agent`.

## Host authentication

- **`password`**: password stored in SQLite (local trust model).
- **`agent`**: requires a **valid** `SSH_AUTH_SOCK` (socket must exist).
- **`none`**: uses agent if `SSH_AUTH_SOCK` is usable; otherwise tries **`~/.ssh/id_ed25519`**, **`id_ecdsa`**, **`id_rsa`** (unencrypted keys only; encrypted keys are skipped).

Host keys are checked against **`$HOME/.ssh/known_hosts`** (with `HOME` unset, Go’s user home). If the file is missing or cannot be loaded, verification falls back to **insecure** accept (see code).

## Security notes

- **Admin has no login** — bind to `127.0.0.1` (default) or do not expose the port to untrusted networks.
- MCP is gated only by **Bearer tokens** tied to project lists.
- Passwords and DB file are only as safe as the host filesystem and process environment.
- Run sshub-mcp in the same environment as your SSH tooling if you rely on **agent** or **`~/.ssh`** keys.

## Releases & packaging

CI runs on **GitHub Actions** (`.github/workflows/ci.yml`). Pushing a tag `v*` builds tarballs and `.deb` packages and creates a **GitHub Release** with those assets (`.github/workflows/release.yml`).

### Manual install (tarball from GitHub Release)

Each tarball (`sshub-mcp_linux_amd64.tar.gz`, `sshub-mcp_darwin_arm64.tar.gz`, etc.) contains:

- `sshub-mcp` — binary
- `sshub-mcp.service` (Linux) or `sshub-mcp.plist` (macOS)
- `install.sh` — copies binary to `/usr/bin` (Linux) or `/opt/homebrew/bin`/`/usr/local/bin` (macOS), installs systemd/launchd unit

```bash
tar xzf sshub-mcp_linux_amd64.tar.gz
./install.sh
systemctl --user enable --now sshub-mcp
```

macOS:

```bash
tar xzf sshub-mcp_darwin_arm64.tar.gz
./install.sh
launchctl load ~/Library/LaunchAgents/sshub-mcp.plist
```

### Debian package

`.deb` installs binary to `/usr/bin/sshub-mcp` and systemd user unit. After `apt install sshub-mcp`:

```bash
systemctl --user enable --now sshub-mcp
```

### Homebrew

Use `packaging/brew/sshub-mcp.rb` as a template in your tap; point `url`/`sha256` at files from the GitHub release. The formula installs to Homebrew’s prefix and provides `brew services start sshub-mcp`.

## License

Specify your license in a `LICENSE` file in the repository root.
