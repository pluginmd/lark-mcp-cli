---
name: lark-cli-mcp
description: Build/debug/extend the lark-cli MCP bridge in cmd/mcp/. Triggers MCP tool add/remove/refine, Claude Desktop disconnects, tool schema changes.
version: 1.0.0
last_updated: 2026-05-11
---

# lark-cli-mcp

A focused skill for working on the **built-in MCP server** that ships
inside `lark-cli` (added in `cmd/mcp/`).

## When to activate

Activate this skill whenever the user is working on:

- The MCP bridge itself: `cmd/mcp/*.go`.
- Tool definitions: adding, removing, renaming, or fixing schemas.
- The JSON-RPC stdio loop (initialize / tools/list / tools/call).
- Connecting Claude Desktop, Claude Code, Cursor, or Zed to the
  bridge.
- Debugging "the host can see lark-cli but tools fail" — usually a
  flag-name mismatch between `Build` closure and the actual shortcut.

Do **not** activate for general lark-cli work (other `cmd/*`,
`shortcuts/*`). That's a job for the `shortcut-explorer` agent or
plain code reading.

## The mental model in 60 seconds

```
MCP host  ──stdio JSON-RPC──▶  lark-cli mcp serve
                                  │
                                  │ os/exec
                                  ▼
                              lark-cli <shortcut>
                                  │
                                  │ HTTPS
                                  ▼
                              Lark/Feishu Open API
```

The bridge is **stateless**. Every `tools/call` spawns a fresh
subprocess. Auth, profile, rate limiting, output formatting all
live in the spawned process — the MCP layer just translates args
and forwards stdout.

## Hard rules (memorise)

1. **Stdout = JSON-RPC only.** Never `fmt.Println`, `log.Print`, or
   any direct write to stdout from inside `cmd/mcp/`. Use `s.logf`
   (writes to stderr) instead.
2. **No third-party MCP library.** `cmd/mcp/protocol.go` rolls the
   wire types by hand. Keep `go.mod` clean.
3. **One MCP tool ≈ one shortcut.** Don't compose flows. If a use
   case needs two calls, let the host model chain them.
4. **Verify flag names against `shortcuts/<domain>/*.go`.** The CLI
   parser will exit non-zero on unknown flags and the host model
   sees a confusing error. Confirm before shipping.

## Quick reference

| Need                                  | File / command                          |
| ------------------------------------- | --------------------------------------- |
| Add a tool                            | `cmd/mcp/tools.go` + `/mcp-add` command |
| See current catalogue                 | `lark-cli mcp tools`                    |
| Run handshake without Claude          | `/mcp-test`                             |
| Rebuild + install                     | `/mcp-rebuild`                          |
| Architecture deep-dive                | `references/architecture.md`            |
| Cookbook for new tools                | `references/tool-builder-cookbook.md`   |
| Common shortcut flag patterns         | `references/shortcut-flags-quickref.md` |

## Verification recipe (always run after edits)

```bash
go build -o ~/bin/lark-cli .
~/bin/lark-cli mcp tools                                  # 1) catalogue parses
printf '%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"t","version":"0"}}}' \
  '{"jsonrpc":"2.0","method":"notifications/initialized"}' \
  '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' \
  | ~/bin/lark-cli mcp serve                              # 2) JSON-RPC handshake
```

Both must succeed. If you changed a tool, also issue a `tools/call`
for it with safe arguments (preferring shortcuts that support
`--dry-run`).
