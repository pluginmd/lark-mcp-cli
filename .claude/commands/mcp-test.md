---
description: Smoke-test the lark-cli MCP server (initialize + tools/list + optional tools/call)
argument-hint: [tool-name] [json-args]
allowed-tools: Bash, Read
---

Verify the MCP bridge end-to-end without needing Claude Desktop.

## 1. Locate the binary

Prefer `~/bin/lark-cli` (development build). Fall back to
`/usr/local/bin/lark-cli` if that's where the user installed it.
Print the path you picked.

## 2. Run the handshake

```bash
printf '%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"smoke","version":"0"}}}' \
  '{"jsonrpc":"2.0","method":"notifications/initialized"}' \
  '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' \
  | <binary-path> mcp serve 2>/tmp/mcp-smoke.log
```

Expected: two JSON-RPC response lines on stdout. The first must
contain `"protocolVersion":"2024-11-05"`. The second must contain
`"tools":[`.

## 3. (Optional) Real tool call

If the user supplied `$1` (tool name), build a third request:

```bash
printf '%s\n' \
  <initialize line> \
  <initialized notification> \
  '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"<tool>","arguments":<args>}}' \
  | <binary-path> mcp serve
```

Use `$2` (JSON object) for arguments, or `{}` if not provided.

## 4. Report

- Binary path used.
- Did initialize succeed? (Y/N)
- Tool count in `tools/list` response.
- If `tools/call` run: did it return `isError: true` or success?
- Tail of `/tmp/mcp-smoke.log` for the server's stderr logs.

Stop after reporting. Do not modify any files.
