---
description: Invoke one MCP tool end-to-end (handshake + tools/call) and print the result
argument-hint: <tool-name> [json-args, default {}]
allowed-tools: Bash, Read
---

One-shot wrapper around `lark-cli mcp serve` for fast dev iteration.
Use after editing `cmd/mcp/tools.go` to confirm a tool actually
works without launching Claude Desktop.

## 1. Resolve inputs

- Tool name: `$1`. Required. If empty, abort and tell the user the
  usage: `/mcp-call <tool-name> '<json-args>'`.
- Args: `$2`. Default `{}` if not supplied.
- Validate `$2` parses as JSON before spending the handshake:

```bash
echo "$2" | python3 -c "import json,sys;json.loads(sys.stdin.read())" \
  || { echo "Args is not valid JSON"; exit 1; }
```

## 2. Locate the binary

Prefer `~/bin/lark-cli` (dev build). Fall back to `which lark-cli`.
Print the absolute path you picked.

## 3. Confirm the tool is in the catalogue

```bash
$BIN mcp tools 2>&1 | grep -F "$1" \
  || { echo "Tool '$1' not in catalogue — rebuild via /mcp-rebuild"; exit 1; }
```

## 4. Send the call

```bash
printf '%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"mcp-call","version":"0"}}}' \
  '{"jsonrpc":"2.0","method":"notifications/initialized"}' \
  "{\"jsonrpc\":\"2.0\",\"id\":2,\"method\":\"tools/call\",\"params\":{\"name\":\"$1\",\"arguments\":$ARGS}}" \
  | $BIN mcp serve 2>/tmp/mcp-call.stderr
```

Where `$ARGS` is `$2` (already validated JSON).

## 5. Parse and present

Three lines on stdout. Discard the first (init response) and
focus on the third (the `tools/call` result):

```bash
... | python3 -c "
import json, sys
lines = [l for l in sys.stdin if l.strip()]
resp = json.loads(lines[-1])
result = resp.get('result', {})
err = result.get('isError')
print('isError:', bool(err))
for block in result.get('content', []):
    print(block.get('text', ''))
"
```

## 6. Report

- Tool called + args used.
- `isError: true/false`.
- Result content (truncated to ~40 lines).
- If `isError: true`: tail `/tmp/mcp-call.stderr` for the stderr
  the subprocess emitted — that's where the real error is.

## Safety

- **Refuse destructive args** unless the user explicitly types
  "yes, run it for real". Specifically: `lark_im_send` with a real
  `chat_id`, any `*-delete` tool, anything where the shortcut does
  not advertise `--dry-run`. Default to `--dry-run` if the
  underlying shortcut supports it (delegate to `shortcut-explorer`
  to check).
- Always `rm -f /tmp/mcp-call.stderr` after.
