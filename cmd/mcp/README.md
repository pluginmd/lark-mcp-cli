# lark-cli MCP server

A built-in [Model Context Protocol](https://modelcontextprotocol.io/)
server that turns `lark-cli` into a local MCP endpoint for Claude
Desktop, Claude Code, OpenClaw, Cursor, Zed, or any other MCP-aware
host.

> One binary. Zero extra runtime. No Node.js, no Python.
> Two transports: **stdio** (default, for local hosts) and **HTTP**
> streamable-http (for containerized hosts like OpenClaw in Docker).
> Exposes **21 curated Lark/Feishu tools** — IM, Mail, Calendar, Docs,
> Base, Contact, Task, Drive, Sheets, VC, Minutes, OKR, plus a
> generic API passthrough. **12 read-type tools** accept a `jq` arg
> for server-side projection; **8 mutating tools** accept `dry_run`
> for preview-without-commit; optional `--audit-log` writes per-call
> NDJSON for replay/cost analysis; opt-in `--max-concurrency` lets
> orchestrator hosts dispatch tools/call in parallel on stdio.

---

## Table of contents

1. [How it works](#how-it-works)
2. [Quick start](#quick-start)
3. [Build from source](#build-from-source)
4. [Authentication](#authentication)
5. [Wire up your MCP host](#wire-up-your-mcp-host)
   - [Claude Desktop](#claude-desktop)
   - [Claude Code](#claude-code)
   - [Cursor / Zed / generic](#cursor--zed--generic)
6. [Tool catalogue](#tool-catalogue)
   - [Shared arguments](#shared-arguments) (`jq`, `dry_run`, `as`)
   - [Error envelope](#error-envelope)
7. [Operational flags](#operational-flags) (`--audit-log`, `--max-concurrency`)
8. [Verify the bridge](#verify-the-bridge)
9. [Troubleshooting](#troubleshooting)
10. [Adding a new tool](#adding-a-new-tool)
11. [Architecture notes](#architecture-notes)

---

## How it works

```
┌──────────────────┐   stdio JSON-RPC 2.0   ┌─────────────────────────┐
│  Claude Desktop  │ ────────────────────▶ │ lark-cli mcp serve      │
│  (MCP host)      │ ◀──────────────────── │  ├─ initialize          │
└──────────────────┘                       │  ├─ tools/list          │
                                           │  └─ tools/call          │
                                           │       │                 │
                                           │       ▼ exec subprocess │
                                           │   lark-cli <shortcut>   │
                                           │       │                 │
                                           │       ▼ HTTPS           │
                                           │  open.feishu.cn /       │
                                           │  open.larksuite.com     │
                                           └─────────────────────────┘
```

- The MCP host (Claude Desktop, etc.) spawns `lark-cli mcp serve` and
  exchanges newline-delimited JSON-RPC messages over stdin/stdout.
- Every `tools/call` is translated into a `lark-cli <shortcut>`
  subprocess so all the existing auth, profile, output, and rate-limit
  machinery is reused as-is.
- Stdout is reserved strictly for JSON-RPC frames. All logs go to
  stderr.

## Quick start

The fastest path is the bootstrap script:

```bash
git clone https://github.com/larksuite/cli.git
cd lark-cli
./scripts/setup-mcp.sh                 # builds + installs to ~/bin/lark-cli
lark-cli auth login                    # one-time browser auth
```

Then pick the matching snippet from [`examples/mcp-hosts/`](../../examples/mcp-hosts/)
and merge it into your MCP host's config file (Claude Desktop, Claude
Code, OpenClaw, Cursor, Zed, Cline are all covered).

### Or do it manually

```bash
# 1. Build (Go 1.23+ required)
git clone https://github.com/larksuite/cli.git
cd lark-cli
go build -o lark-cli .

# 2. Move to PATH
sudo mv lark-cli /usr/local/bin/        # macOS/Linux
# or:
mv lark-cli "$HOME/bin/"                # no sudo; remember to add ~/bin to PATH

# 3. Authenticate
lark-cli config init
lark-cli auth login

# 4. Sanity-check the MCP bridge
lark-cli mcp tools

# 5. Add to your MCP host config (see below) and restart it.
```

## Build from source

### Requirements

- Go **1.23 or newer** (`go version` to check).
- Network access to `open.feishu.cn` / `open.larksuite.com`.
- A Lark/Feishu account (personal or workspace).

### Build commands

```bash
go build -o lark-cli .                  # development build
make build                              # uses goreleaser-style flags
./build.sh                              # multi-arch release build
```

### Cross-compile (optional)

```bash
GOOS=darwin  GOARCH=arm64 go build -o dist/lark-cli-darwin-arm64 .
GOOS=darwin  GOARCH=amd64 go build -o dist/lark-cli-darwin-amd64 .
GOOS=linux   GOARCH=amd64 go build -o dist/lark-cli-linux-amd64 .
GOOS=windows GOARCH=amd64 go build -o dist/lark-cli-windows-amd64.exe .
```

### Install paths cheatsheet

| OS      | Recommended path                | PATH already set? |
| ------- | ------------------------------- | ----------------- |
| macOS   | `/usr/local/bin/lark-cli`       | Yes (GUI apps)    |
| macOS   | `~/bin/lark-cli`                | Only if you added it |
| Linux   | `/usr/local/bin/lark-cli`       | Yes               |
| Windows | `C:\Program Files\lark-cli\lark-cli.exe` | Add manually |

> Claude Desktop on macOS launches from Finder which **does not** load
> your shell's PATH (`.zshrc`, `.bashrc`). When in doubt, use the
> absolute path in `claude_desktop_config.json`.

## Authentication

The MCP server reuses the existing CLI auth flow — there is **no**
separate login for MCP.

```bash
lark-cli config init                    # first-time profile setup
lark-cli auth login                     # opens browser, stores tokens
lark-cli auth status                    # verify token is valid
```

Tokens are stored in the OS keychain (macOS Keychain, Windows Credential
Manager, libsecret on Linux) under the active profile. The MCP server
subprocess inherits the same profile automatically.

Multiple identities? Use `--as user` or `--as bot` on each tool call —
every exposed MCP tool accepts an optional `as` argument that maps to
`lark-cli --as <identity>`.

## Wire up your MCP host

### Claude Desktop

1. Open **Claude Desktop → Settings → Developer → Edit Config**.
2. Add a `lark-cli` entry under `mcpServers`:

#### macOS / Linux

```json
{
  "mcpServers": {
    "lark-cli": {
      "command": "/usr/local/bin/lark-cli",
      "args": ["mcp", "serve"]
    }
  }
}
```

#### Windows

```json
{
  "mcpServers": {
    "lark-cli": {
      "command": "C:\\Program Files\\lark-cli\\lark-cli.exe",
      "args": ["mcp", "serve"]
    }
  }
}
```

#### With environment overrides

```json
{
  "mcpServers": {
    "lark-cli": {
      "command": "/usr/local/bin/lark-cli",
      "args": ["mcp", "serve"],
      "env": {
        "LARK_CLI_PROFILE": "work",
        "NO_COLOR": "1"
      }
    }
  }
}
```

3. Restart Claude Desktop. Look for `lark-cli` under **Settings →
   Developer → Local MCP servers**.

### Claude Code

```bash
claude mcp add lark-cli /usr/local/bin/lark-cli mcp serve
```

Or edit `~/.claude.json` directly:

```json
{
  "mcpServers": {
    "lark-cli": {
      "type": "stdio",
      "command": "/usr/local/bin/lark-cli",
      "args": ["mcp", "serve"]
    }
  }
}
```

### OpenClaw

[OpenClaw](https://openclaw.ai) is a personal AI gateway that bridges
channel chats (Telegram, Slack, Discord, Feishu, …) to MCP. It
maintains its own MCP server registry under `mcp.servers` in
`~/.openclaw/openclaw.json`. Pick the variant that matches how
OpenClaw runs.

#### Variant A — OpenClaw on the host (same machine as lark-cli)

Use the stdio transport. Easiest, no network exposure.

```bash
openclaw mcp set lark-cli '{"command":"/usr/local/bin/lark-cli","args":["mcp","serve"],"env":{"NO_COLOR":"1"}}'
openclaw mcp show lark-cli --json
```

#### Variant B — OpenClaw in Docker, lark-cli on the host (Recommended for production)

Use the HTTP transport. The container reaches the host binary over
`host.docker.internal`. Auth keeps using the host's keychain — no
bind-mount, no cross-build, no credential plumbing.

**1. Run lark-cli in HTTP mode on the host:**

```bash
lark-cli mcp serve --transport http --addr 127.0.0.1:3000
```

> Bind to `127.0.0.1`, not `0.0.0.0`. Docker Desktop maps
> `host.docker.internal` to the host's loopback by default — no need
> to expose the port on the LAN.

**2. Register the URL in OpenClaw:**

```bash
openclaw mcp set lark-cli '{"url":"http://host.docker.internal:3000","transport":"streamable-http","connectionTimeoutMs":10000}'
```

**3. Keep the lark-cli HTTP server running.** Add a launchd plist
(`~/Library/LaunchAgents/com.lark-cli.mcp.plist`) or a systemd user
unit so it survives logout. Sample plist:

```xml
<plist version="1.0"><dict>
  <key>Label</key><string>com.lark-cli.mcp</string>
  <key>ProgramArguments</key><array>
    <string>/usr/local/bin/lark-cli</string>
    <string>mcp</string><string>serve</string>
    <string>--transport</string><string>http</string>
    <string>--addr</string><string>127.0.0.1:3000</string>
  </array>
  <key>RunAtLoad</key><true/>
  <key>KeepAlive</key><true/>
  <key>StandardErrorPath</key><string>/tmp/lark-cli-mcp.err.log</string>
</dict></plist>
```

```bash
launchctl load ~/Library/LaunchAgents/com.lark-cli.mcp.plist
```

#### Other variants

- **OpenClaw in Linux on the same host** (not containerized): use
  Variant A with the absolute Linux path.
- **OpenClaw and lark-cli BOTH inside the same container/image**:
  install `lark-cli` in the image and use Variant A with its
  in-container path. Auth becomes tricky — bake tokens via env
  vars (`LARKSUITE_CLI_*`) or use a bot identity.
- **Remote OpenClaw, lark-cli on a separate host**: same as Variant B
  but use the public hostname and put TLS in front (Caddy, nginx).
  Add `headers` with a bearer token if you want auth — the current
  lark-cli HTTP server is open by default, so don't expose it to the
  internet without a proxy.

**Env safety filter**: OpenClaw rejects interpreter-startup env keys
(`NODE_OPTIONS`, `PYTHONSTARTUP`, `PYTHONPATH`, `PERL5OPT`, etc.) in
the `env` block. `lark-cli` does not need any of these — safe.

### Cursor / Zed / Cline / generic

Any MCP host that supports stdio servers works. Point it at
`lark-cli mcp serve` with the absolute binary path. Drop-in snippets:

- [`examples/mcp-hosts/cursor.json`](../../examples/mcp-hosts/cursor.json)
- [`examples/mcp-hosts/zed.json`](../../examples/mcp-hosts/zed.json)
- [`examples/mcp-hosts/cline.json`](../../examples/mcp-hosts/cline.json)

## Tool catalogue

`lark-cli mcp tools` prints the live catalogue. Today:

| MCP tool                  | Description                                    | Underlying shortcut          | jq | dry_run |
| ------------------------- | ---------------------------------------------- | ---------------------------- | -- | ------- |
| `lark_im_send`            | Send IM (text/markdown) to user/chat           | `im +messages-send`          |    | ✓       |
| `lark_im_card_send`       | Send Feishu interactive card (YAML spec)       | `im +card-send`              |    | ✓       |
| `lark_im_search`          | Search IM messages by keyword                  | `im +messages-search`        | ✓  |         |
| `lark_mail_send`          | Send email (draft by default; `confirm_send=true` to commit) | `mail +send`   |    | ¹       |
| `lark_mail_draft_create`  | Create mail draft (never auto-sends)           | `mail +draft-create`         |    |         |
| `lark_calendar_agenda`    | List upcoming calendar events                  | `calendar +agenda`           | ✓  |         |
| `lark_calendar_create`    | Create a calendar event                        | `calendar +create`           |    | ✓       |
| `lark_doc_create`         | Create a Lark Doc (optional markdown body)     | `docs +create`               |    | ✓       |
| `lark_doc_search`         | Search documents in user's drive               | `docs +search`               | ✓  |         |
| `lark_doc_fetch`          | Fetch a Lark Doc as markdown                   | `docs +fetch`                | ✓  |         |
| `lark_base_search`        | Search Base (Bitable) records                  | `base +record-search`        | ✓  |         |
| `lark_contact_search`     | Search the org directory                       | `contact +search-user`       | ✓  |         |
| `lark_task_my`            | List the authenticated user's tasks            | `task +get-my-tasks`         | ✓  |         |
| `lark_task_create`        | Create a Lark Task (todo)                      | `task +create`               |    | ✓       |
| `lark_drive_upload`       | Upload a local file to Drive                   | `drive +upload`              |    | ✓       |
| `lark_sheets_read`        | Read cells from a Lark Sheet                   | `sheets +read`               | ✓  |         |
| `lark_sheets_append`      | Append rows to a Lark Sheet                    | `sheets +append`             |    | ✓       |
| `lark_vc_search`          | Search past video meetings                     | `vc +search`                 | ✓  |         |
| `lark_minutes_search`     | Search Lark Minutes recordings/transcripts     | `minutes +search`            | ✓  |         |
| `lark_okr_cycle_list`     | List OKR cycles for a user                     | `okr +cycle-list`            | ✓  |         |
| `lark_api`                | Generic Open API passthrough (escape hatch)    | `api <METHOD> <PATH>`        | ✓  | ✓       |

¹ `lark_mail_send` uses its own `confirm_send=false → draft` safety
pattern (predates the universal `dry_run` arg). The underlying shortcut
also supports `--dry-run` for previewing the API request shape.

### Shared arguments

Every tool accepts these optional args alongside its tool-specific
fields:

| Arg        | Type     | Purpose                                                                  |
| ---------- | -------- | ------------------------------------------------------------------------ |
| `as`       | enum     | `"user"` or `"bot"` — switch identity for this call only.                |
| `jq`       | string   | (read-type tools) Server-side jq projection. Cuts token cost of large responses 10-50×. Mutually exclusive with `--format`. |
| `dry_run`  | boolean  | (mutating tools) Preview the would-be request as JSON without committing. Recommended before any mutating call. |

### Error envelope

When a tool call fails (build-layer validation, unknown tool, exec
failure, shortcut non-zero exit), `content[0].text` carries a uniform
classification JSON the host model switches on:

```json
{
  "error_type": "validation|auth_expired|rate_limit|permission_denied|api_error|network|unknown",
  "message":    "<human-readable>",
  "retry_advice": "<actionable next step>",
  "code":       1,
  "log_id":     "<lark log_id when present>",
  "raw":        "<original stderr verbatim>"
}
```

Agents should branch on `error_type` rather than grep'ing `raw`:
- `validation` → fix args, do NOT retry blindly
- `auth_expired` → surface `lark-cli auth login` to user
- `rate_limit` → wait + retry with backoff
- `permission_denied` → check scope, try `as: "bot"` if applicable
- `api_error` with `log_id` → Lark-side, retry once then escalate
- `network` → retry once, then surface
- `unknown` → consult `raw`

### Example: ask Claude to send an IM (with dry-run gate)

> "Send a message to Alice saying the deploy is done."

Claude's expected flow:
1. `lark_contact_search` to resolve `alice` → `open_id` (project only the open_id with `jq: ".users[0].open_id"`).
2. `lark_im_send` with `user_id` + `text` + **`dry_run: true`** — preview the planned API call.
3. Echo the preview to the user, await confirmation.
4. Re-call `lark_im_send` without `dry_run` to commit.

### Example: agenda + tasks brief (with projection)

> "What's on my calendar today and which tasks are due?"

Claude calls `lark_calendar_agenda` with
`jq: ".events[] | {summary, start, attendees: [.attendees[].display_name]}"`
to cut a full event payload (often 5-20 KB) down to ~500 tokens.
Then `lark_task_my` with
`jq: ".tasks[] | select(.due) | {id, summary, due}"`
to filter and project. Synthesise both summaries.

### Example: low-level API call

If a needed endpoint is not in the curated list, Claude (or you) can
use `lark_api`. It supports both `jq` and `dry_run`:

```json
{
  "name": "lark_api",
  "arguments": {
    "method": "GET",
    "path": "/open-apis/calendar/v4/calendars/primary/events",
    "params": { "page_size": 5 },
    "jq": ".data.items[] | {summary, start_time}"
  }
}
```

For a mutating raw call, preview first:

```json
{
  "name": "lark_api",
  "arguments": {
    "method": "POST",
    "path": "/open-apis/im/v1/messages",
    "params": {"receive_id_type": "open_id"},
    "data": {"receive_id": "ou_xxx", "msg_type": "text", "content": "{\"text\":\"hi\"}"},
    "dry_run": true
  }
}
```

## Operational flags

`lark-cli mcp serve` exposes a small set of operator-level flags that
don't affect tool semantics but shape how the server runs:

| Flag                  | Default      | Purpose                                                                       |
| --------------------- | ------------ | ----------------------------------------------------------------------------- |
| `--transport`         | `stdio`      | `stdio` (default, for local hosts) or `http` (for containerized hosts).       |
| `--addr`              | `127.0.0.1:3000` | HTTP listen address. Only used with `--transport http`.                  |
| `--audit-log <path>`  | (disabled)   | Path to write NDJSON audit log. One line per `tools/call` with ts, session, tool, args_hash, exit_code, latency_ms, byte counts, is_error. |
| `--audit-log-args`    | `false`      | Include raw tool arguments alongside the hash. Default hashes only to protect PII (recipient emails, mail bodies, etc.). |
| `--max-concurrency`   | `4`          | Maximum in-flight tool dispatches on stdio. `0` or `1` = serial (legacy). Higher = worker pool. HTTP transport is always concurrent. |

### Audit log

```bash
lark-cli mcp serve --audit-log ~/.lark-mcp/audit.ndjson
```

The file is append-only (`O_APPEND`), so multiple `mcp serve` runs
against the same path accumulate history. Tail and parse with `jq`:

```bash
# Live error feed
tail -f ~/.lark-mcp/audit.ndjson | jq 'select(.is_error) | {ts, tool, error_type: .raw[0:80]}'

# Cost/latency analytics for the last hour
jq -s 'map(select(.ts > "'$(date -u -v-1H +%FT%TZ)'")) | {n: length, p95_ms: (map(.latency_ms) | sort | .[((length*0.95)|floor)])}' ~/.lark-mcp/audit.ndjson
```

For dev/debug, opt in to raw args (PII risk — never on production):

```bash
lark-cli mcp serve --audit-log /tmp/dev.ndjson --audit-log-args
```

### Concurrency

The stdio transport processes `tools/call` requests through a worker
pool (default 4 workers). JSON-RPC 2.0 does NOT require response
ordering — host MUST match by `id` — so out-of-order replies are
spec-compliant. `writeMu` serializes stdout writes so frames stay
well-formed under concurrency.

For orchestrator use cases (one Claude session spawning N parallel
worker subagents), raise concurrency:

```bash
lark-cli mcp serve --max-concurrency 8
```

For debugging or hosts that expect strict ordering, force serial:

```bash
lark-cli mcp serve --max-concurrency 1
```

## Verify the bridge

### List tools as a human

```bash
lark-cli mcp tools
```

### Drive the stdio protocol manually

```bash
printf '%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"t","version":"0"}}}' \
  '{"jsonrpc":"2.0","method":"notifications/initialized"}' \
  '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' \
  | lark-cli mcp serve
```

Expected: two JSON lines on stdout. Initialize result first, then the
tools list.

### Call a tool end-to-end

```bash
printf '%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"t","version":"0"}}}' \
  '{"jsonrpc":"2.0","method":"notifications/initialized"}' \
  '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"lark_contact_search","arguments":{"query":"alice"}}}' \
  | lark-cli mcp serve
```

A non-error result with a `content[0].text` JSON blob means everything
is wired correctly.

## Troubleshooting

### "command not found" inside Claude Desktop

Claude Desktop on macOS does not load your shell's PATH. Use the
absolute path in `claude_desktop_config.json`:

```bash
which lark-cli                          # copy this path
```

### Tool calls return `isError: true`

The MCP response carries a classified error envelope in
`content[0].text` (see [Error envelope](#error-envelope) above). Parse
the JSON, switch on `error_type`, and act according to the
`retry_advice` field.

To reproduce a shortcut-layer failure directly:

```bash
lark-cli contact +search-user --query "alice"
```

The CLI's stderr (visible verbatim in the envelope's `raw` field) is
the source of truth for the failure cause.

### "permission denied" / scope errors

Run `lark-cli auth login` again and ensure the OAuth consent screen
includes the scopes the shortcut needs. Or retry the tool call with
`as: "bot"` if the flow only works with bot identity.

### Garbled output / Claude disconnects immediately

Something is writing to stdout that isn't JSON-RPC. Symptoms include
ANSI colour codes or progress bars from the CLI. Mitigations:

```json
"env": { "NO_COLOR": "1", "LARK_CLI_FORMAT": "json" }
```

If you ever wrap `lark-cli mcp serve` in a shell script, route all
diagnostic prints to stderr (`>&2`).

### Long-running tool calls

Calls inherit the MCP host's timeout. For uploads/exports over a few
seconds, prefer the dedicated `drive +upload` flag set (auto-chunk)
over `lark_api` and watch the host's MCP timeout setting.

## Adding a new tool

1. Edit [`tools.go`](tools.go) and append a `tool{...}` entry from a new
   `toolXxx()` factory.
2. Register it in `allTools()`.
3. Define the JSON Schema in the `Schema` field (raw JSON literal so
   the wire format is exact). For a **read-type** tool, include the
   `jq` property; for a **mutating** tool whose shortcut implements
   `DryRun`, include the `dry_run` property.
4. In the `Build` closure, translate MCP arguments into `lark-cli`
   argv. Reuse helpers in `tools.go`:
   - `argString(args, "key")` / `mustString(args, "key")`
   - `argInt(args, "key")`
   - `appendFlag(argv, "flag-name", value)`
   - `appendBoolFlag(argv, args, "key", "flag-name")`
   - `appendJq(argv, args)` — only for read-type tools
   - `appendIdentity(argv, args)` for the `--as` flag
   - `requireOneOf(args, "a", "b")` for mutually-exclusive required pairs
5. Write a unit test in [`tools_test.go`](tools_test.go) — add the new
   tool to `TestReadToolsExposeJq` or `TestMutatingToolsExposeDryRun`
   as appropriate, and update `minimalArgsFor()`.
6. Add a description-quality entry in
   [`descriptions_test.go`](descriptions_test.go) if the tool is
   mutating (mention safety mechanism) or part of a confusable pair.
7. Rebuild (`go build .`), run `go test -race ./cmd/mcp/...`, and
   verify with `lark-cli mcp tools`.

Read-type example skeleton (with `jq`):

```go
func toolFooList() tool {
    return tool{
        Name:        "lark_foo_list",
        Description: "List foo items. Use when the user asks ... Pair with jq to project only id/title for large lists.",
        Schema: `{
  "type": "object",
  "properties": {
    "filter": {"type": "string", "description": "Optional filter expression."},
    "jq":     {"type": "string", "description": "Optional jq projection (e.g. '.items[] | {id, title}')."},
    "as":     {"type": "string", "enum": ["user", "bot"]}
  }
}`,
        Build: func(args map[string]interface{}) ([]string, error) {
            argv := []string{"foo", "+list"}
            argv = appendFlag(argv, "filter", mustString(args, "filter"))
            argv = appendJq(argv, args)
            argv = appendIdentity(argv, args)
            return argv, nil
        },
    }
}
```

Mutating example skeleton (with `dry_run`):

```go
func toolFooCreate() tool {
    return tool{
        Name:        "lark_foo_create",
        Description: "Create a foo. Use when ... ALWAYS preview with dry_run=true first to verify before commit.",
        Schema: `{
  "type": "object",
  "properties": {
    "name":    {"type": "string", "description": "Foo name."},
    "dry_run": {"type": "boolean", "description": "Preview the would-be request without committing."},
    "as":      {"type": "string", "enum": ["user", "bot"]}
  },
  "required": ["name"]
}`,
        Build: func(args map[string]interface{}) ([]string, error) {
            name, ok := argString(args, "name")
            if !ok {
                return nil, fmt.Errorf("name is required")
            }
            argv := []string{"foo", "+create", "--name", name}
            argv = appendBoolFlag(argv, args, "dry_run", "dry-run")
            argv = appendIdentity(argv, args)
            return argv, nil
        },
    }
}
```

> The underlying shortcut MUST implement the `DryRun` field
> (`shortcuts/<domain>/<shortcut>.go`) for `dry_run=true` to do
> anything meaningful — otherwise lark-cli returns
> `--dry-run is not supported for X +Y`.

## Architecture notes

| Concern             | Choice                                                                                  |
| ------------------- | --------------------------------------------------------------------------------------- |
| Transport           | stdio (default) or streamable-http, newline-delimited JSON-RPC 2.0                      |
| Protocol version    | Advertised: `2025-06-18`. Negotiated down to `2024-11-05` for older clients.            |
| MCP library         | None — hand-rolled to keep `go.mod` lean                                                |
| Tool execution      | Subprocess (`os/exec` of the same binary)                                               |
| Auth                | Inherited from the active CLI profile + keychain                                        |
| Output channel      | stdout = JSON-RPC frames only (mutex-serialized); stderr = `[mcp]` logs                 |
| Stdio concurrency   | Worker pool (`--max-concurrency`, default 4). Out-of-order responses are spec-compliant. |
| HTTP concurrency    | Native Go `net/http` (each request its own goroutine)                                   |
| Error envelope      | Classified JSON in `content[0].text` (see [Error envelope](#error-envelope))            |
| Audit log           | Optional NDJSON via `--audit-log` (privacy: hashed args by default)                     |

### File map

```
cmd/mcp/
├── README.md           ← you are here
├── mcp.go              ← `lark-cli mcp` parent command
├── serve.go            ← `lark-cli mcp serve` stdio loop + worker pool + flags
├── protocol.go         ← JSON-RPC + MCP wire types
├── tools.go            ← 21 tool definitions + dispatch + helpers (appendJq, etc.)
├── runner.go           ← subprocess execution helper
├── http.go             ← streamable-http transport
├── errors.go           ← stderr → classified envelope (T6 contract)
├── audit.go            ← NDJSON audit log writer (T3 contract)
├── tools_test.go       ← jq + dry_run + schema contracts
├── descriptions_test.go ← selection-guidance + pair-disambiguation + safety mention
├── errors_test.go      ← classifier heuristics + formatResult wire contract
├── audit_test.go       ← privacy + concurrent writes + append mode
└── concurrency_test.go ← worker pool + serial path + frame integrity
```

The repository also ships an **offline eval suite** for contract
regression testing — see [`demo/pattern/eval/`](../../demo/pattern/eval/).
Run `./demo/pattern/eval/run.sh` for a 10-scenario, 86-assertion gate
that exercises tools/list shape, build-layer validation, classified
errors, dry_run propagation, confusable-pair discovery, and more.

### Why subprocess instead of in-process?

- Reuses the CLI's already-tested auth, profile, output, and pagination
  flows — no need to duplicate them in the MCP layer.
- Isolation: a panic in a shortcut cannot bring down the long-running
  MCP server.
- Fork cost is ~10 ms on a modern Mac; negligible next to the Lark API
  round-trip.

### Why out-of-order responses on stdio?

JSON-RPC 2.0 §4.2 specifies that the host MUST match responses to
requests by `id`; ordering is unspecified. The worker-pool path takes
advantage of this to interleave slow Lark calls — e.g. four parallel
research subagents can each fire a `tools/call` and the server
dispatches all four concurrently, returning results as they arrive
rather than blocking on the slowest. `writeMu` ensures each response
stays a complete JSON-RPC frame.

If your host depends on strict ordering for any reason, set
`--max-concurrency 1` to force the legacy serial path.

## License

MIT — same as the rest of `lark-cli`.
