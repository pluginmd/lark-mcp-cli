# Drop-in MCP host configs

Each file under `mcp-hosts/` is a ready-to-paste snippet for a
specific MCP-aware host. Replace `<BINARY-PATH>` with the absolute
path returned by `which lark-cli` (or `~/bin/lark-cli` if you used
the default install).

| Host                 | File                                | Config location                                                         | Transport |
| -------------------- | ----------------------------------- | ----------------------------------------------------------------------- | --------- |
| Claude Desktop       | `mcp-hosts/claude-desktop.json`     | `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) | stdio   |
| Claude Code          | `mcp-hosts/claude-code.json`        | `~/.claude.json`                                                        | stdio     |
| OpenClaw (on host)   | `mcp-hosts/openclaw.json` (A)       | `~/.openclaw/openclaw.json` (`mcp.servers.lark-cli`)                    | stdio     |
| OpenClaw (Docker)    | `mcp-hosts/openclaw.json` (B)       | same                                                                    | **http**  |
| Cursor               | `mcp-hosts/cursor.json`             | `~/.cursor/mcp.json`                                                    | stdio     |
| Zed                  | `mcp-hosts/zed.json`                | `~/.config/zed/settings.json` (`context_servers`)                       | stdio     |
| Cline (VS Code ext)  | `mcp-hosts/cline.json`              | `~/.cline/mcp_settings.json`                                            | stdio     |

> **stdio** = MCP host spawns `lark-cli` as a subprocess and talks
> over stdin/stdout. Use for local hosts.
>
> **http** = MCP host POSTs JSON-RPC to a running `lark-cli mcp serve
> --transport http` process. Use when the host runs in a container
> and the binary lives on the surrounding macOS/Linux host (auth
> stays in the host's keychain).

> **Windows**: replace `command` with the full path to
> `lark-cli.exe` (e.g. `C:\\Program Files\\lark-cli\\lark-cli.exe`).
> **Linux**: typically `/usr/local/bin/lark-cli` or `~/bin/lark-cli`.

## After editing

Restart the host so it re-reads its config:

- Claude Desktop: Cmd+Q and reopen.
- Claude Code: next session picks up changes.
- Cursor: Cmd+Shift+P → "Reload Window".
- Zed: Cmd+Q and reopen.
- OpenClaw: restart whichever runtime adapter consumes the MCP
  registry (embedded Pi, etc.). The CLI's `openclaw mcp set` does
  not auto-restart those.

## Verify

```bash
lark-cli mcp tools                          # 20 tools should list
~/bin/lark-cli mcp tools                    # if not on PATH yet
```

Then issue a real request from the host. For Claude Desktop, ask:
*"List my upcoming calendar events"* — Claude should call
`lark_calendar_agenda`.
