# Lark/Feishu × MCP — quick start

Drive your Lark/Feishu workspace from any MCP host (Claude Desktop,
Claude Code, OpenClaw, Cursor, Zed, Cline, …) in three commands.

```bash
git clone https://github.com/larksuite/cli.git
cd lark-cli
./scripts/setup-mcp.sh && lark-cli auth login
```

Then pick the snippet for your host from
[`examples/mcp-hosts/`](examples/mcp-hosts/) and paste it into the
host's config file. **Done.**

---

## What you get

A built-in MCP stdio server (`lark-cli mcp serve`) that exposes 20
curated tools across Lark/Feishu:

| Domain    | Tools                                                       |
| --------- | ----------------------------------------------------------- |
| IM        | `lark_im_send`, `lark_im_search`                            |
| Mail      | `lark_mail_send`, `lark_mail_draft_create`                  |
| Calendar  | `lark_calendar_agenda`, `lark_calendar_create`              |
| Docs      | `lark_doc_create`, `lark_doc_search`, `lark_doc_fetch`      |
| Base      | `lark_base_search`                                          |
| Contact   | `lark_contact_search`                                       |
| Task      | `lark_task_my`, `lark_task_create`                          |
| Drive     | `lark_drive_upload`                                         |
| Sheets    | `lark_sheets_read`, `lark_sheets_append`                    |
| Meetings  | `lark_vc_search`, `lark_minutes_search`                     |
| OKR       | `lark_okr_cycle_list`                                       |
| Generic   | `lark_api` (any Open API endpoint as escape hatch)          |

After install, run `lark-cli mcp tools` to see the live catalogue.

## Supported MCP hosts

| Host                       | Transport | Drop-in config                                                                       |
| -------------------------- | --------- | ------------------------------------------------------------------------------------ |
| Claude Desktop             | stdio     | [`examples/mcp-hosts/claude-desktop.json`](examples/mcp-hosts/claude-desktop.json)   |
| Claude Code                | stdio     | [`examples/mcp-hosts/claude-code.json`](examples/mcp-hosts/claude-code.json)         |
| OpenClaw (on host)         | stdio     | [`examples/mcp-hosts/openclaw.json`](examples/mcp-hosts/openclaw.json) — Variant A   |
| OpenClaw (in Docker)       | **http**  | [`examples/mcp-hosts/openclaw.json`](examples/mcp-hosts/openclaw.json) — Variant B   |
| Cursor                     | stdio     | [`examples/mcp-hosts/cursor.json`](examples/mcp-hosts/cursor.json)                   |
| Zed                        | stdio     | [`examples/mcp-hosts/zed.json`](examples/mcp-hosts/zed.json)                         |
| Cline (VS Code)            | stdio     | [`examples/mcp-hosts/cline.json`](examples/mcp-hosts/cline.json)                     |
| Continue.dev               | stdio     | same shape as Claude Code; add to `~/.continue/config.json`                          |
| Any stdio MCP host         | stdio     | `command = lark-cli`, `args = ["mcp","serve"]`                                       |
| Any streamable-http host   | http      | `url = http://<host>:3000`, run server with `lark-cli mcp serve --transport http`    |

### When to use which transport

- **stdio** — host spawns lark-cli as a subprocess. Default. Use for
  every local MCP host (Claude Desktop, Code, Cursor, Zed, Cline).
- **http** — host POSTs JSON-RPC to a running `lark-cli mcp serve
  --transport http` process. Use when the host runs in Docker and
  the lark-cli binary lives on the surrounding macOS/Linux host, so
  the host's keychain auth stays usable. The container reaches
  the host over `host.docker.internal:<port>`.

See [`cmd/mcp/README.md`](cmd/mcp/README.md) § *Wire up your MCP
host* → *OpenClaw* for the full Docker walk-through (including a
launchd plist to keep the HTTP server alive).

## Adding a new tool

Open the project in any Claude-aware IDE and type `/mcp-add <name>`.
The slash command walks through schema, builder, build, and verify.
See [`cmd/mcp/README.md`](cmd/mcp/README.md) for the full architecture,
and [`.claude/skills/lark-cli-mcp/references/tool-builder-cookbook.md`](.claude/skills/lark-cli-mcp/references/tool-builder-cookbook.md)
for the canonical recipe.

## What's in this repo

```
.
├── MCP_QUICKSTART.md           ← you are here
├── README.md                   ← upstream lark-cli docs
├── CLAUDE.md                   ← project memory for Claude (role + scope)
├── AGENTS.md                   ← repo conventions for AI agents (build/test/PR gates)
├── .claude/                    ← Claude workspace kit (agents, commands, skill)
├── cmd/mcp/                    ← the MCP bridge source (Go)
│   └── README.md               ← MCP bridge architecture + per-host wiring
├── examples/mcp-hosts/         ← drop-in JSON snippets for each host
├── scripts/setup-mcp.sh        ← one-shot build + install
└── shortcuts/                  ← lark-cli shortcuts (the CLI itself)
```

## Troubleshooting

- **"command not found" inside Claude Desktop on macOS**: GUI apps
  don't load shell PATH. Use the absolute path from `which lark-cli`
  in the config.
- **Tool calls return `isError: true`**: reproduce the equivalent
  `lark-cli ...` call in a terminal — the underlying shortcut output
  shows the real reason.
- **Garbled output / host disconnects**: something is writing to
  stdout that isn't JSON-RPC. Set `"env": {"NO_COLOR":"1"}` in your
  host config. Never wrap `lark-cli mcp serve` in a shell script that
  prints to stdout.

Full troubleshooting in [`cmd/mcp/README.md`](cmd/mcp/README.md).

## License & attribution

MIT — same as upstream `lark-cli`.

This MCP bridge + the [`.claude/`](./.claude/) Cowork workspace it
ships with are built by **[Transform Group](https://www.transform.group/)**
— Lark Platinum Partner. See [docs/README.md](./docs/README.md) for
full documentation index.
