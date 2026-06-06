# MCP bridge — architecture reference

## File layout

```
cmd/mcp/
├── README.md     ← user-facing setup guide (Claude Desktop config, etc.)
├── mcp.go        ← `lark-cli mcp` parent cobra command
├── serve.go      ← `mcp serve` (transport-agnostic core, stdio loop) + `mcp tools` listing
├── http.go       ← streamable-http transport (`--transport http`)
├── protocol.go   ← JSON-RPC 2.0 + MCP wire types (initializeResult, toolDescriptor, …)
├── tools.go      ← 20 tool definitions + dispatch
└── runner.go     ← os/exec helper for spawning lark-cli subprocesses
```

## Transports

`lark-cli mcp serve --transport stdio` (default) — newline-delimited
JSON-RPC for local hosts like Claude Desktop. `--transport http`
(alias `streamable-http`, `--addr` default `127.0.0.1:3000`) — same
JSON-RPC core over a single HTTP endpoint. Both route through the
transport-agnostic `process()` in `serve.go`.

## Wire protocol

JSON-RPC 2.0 (newline-delimited on stdio; one request/response per
HTTP POST on the http transport). Advertised protocol version
`2025-06-18`; minimum accepted `2024-11-05`. Methods handled:

| Method                          | Behaviour                                            |
| ------------------------------- | ---------------------------------------------------- |
| `initialize`                    | Returns server info + tools capability.              |
| `initialized` / `notifications/initialized` | Logged, no response.                     |
| `ping`                          | Empty result.                                        |
| `tools/list`                    | Returns the cached descriptor list.                  |
| `tools/call`                    | Dispatches via `dispatch()` → subprocess → MCP result. |
| `shutdown`                      | Empty result (host closes stdin to exit).            |
| _unknown notification_          | Silently ignored (resources/, prompts/).             |
| _unknown request_               | Returns `code: -32601` method not found.             |

## Concurrency model

- Stdio transport: one scanner reads stdin line by line.
- `dispatch()` is synchronous — one tool call at a time.
- `writeMu` (a `sync.Mutex` on `server`) guards stdout writes in
  `writeStdio()` so future async paths cannot interleave bytes
  mid-frame. The HTTP transport writes through Go's
  `http.ResponseWriter` instead and does not use `writeMu`.
- Subprocess `cmd.Run()` blocks the scanner. If you ever introduce
  long-running tool calls (uploads), upgrade to goroutine dispatch
  with proper request-id correlation.

## Subprocess execution

`runner.run(ctx, argv)`:

1. `exec.CommandContext(ctx, exe, argv...)` — `exe` resolved via
   `os.Executable()`, so `lark-cli mcp serve` forks _itself_.
2. Inherits `os.Environ()` — auth tokens, profile selection, brand
   resolution all "just work".
3. Captures stdout and stderr separately. Non-zero exit becomes an
   MCP result with `isError: true` and stderr as the text block.

## Why subprocess (vs in-process command tree)

- Auth, profile, output, pagination, dry-run, vfs, validate — all
  in `cmd/api`, `cmd/service`, and `shortcuts/`. Calling them
  in-process would require rebuilding the cobra tree per request
  or sharing a long-lived factory (which holds tokens, http
  clients, etc.).
- Subprocess isolation: a panic in one shortcut cannot bring down
  the long-lived server.
- ~10 ms fork cost on Apple Silicon. Negligible next to a 100–500 ms
  Lark API round-trip.

## What goes where

| Concern                     | Lives in       |
| ---------------------------- | -------------- |
| Stdio wire framing + core    | `serve.go`     |
| HTTP transport               | `http.go`      |
| Wire types                   | `protocol.go`  |
| Tool schemas + builders      | `tools.go`     |
| Subprocess plumbing          | `runner.go`    |
| Cobra command surface        | `mcp.go`       |

## Extension points

- **Resources / prompts**: not implemented. Stubs would go in
  `serve.go` (`handleResourcesList`, etc.) and surface from new
  fields in `serverCapabilities`.
- **Streaming**: MCP supports progress notifications. The current
  server does not emit them. Add via `s.writeStdio(...)` with a
  notification-shaped `rpcMessage` (no `id`, with `method:
  "notifications/progress"`).
- **Cancellation**: would require request-id → ctx cancel map and
  a goroutine per call.
