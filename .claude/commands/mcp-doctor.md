---
description: One-page health report for the MCP bridge — build, handshake, sample call, host config, logs
allowed-tools: Bash, Read, Grep
---

Aggregate every layer of the bridge into a single, scannable report.
Run when the user says "is everything ok", "status check", "what's
the state of MCP", or before a release.

## Sections (run in order, stop only on hard failure)

### 1. Binary

```bash
BIN=$(command -v lark-cli || echo "$HOME/bin/lark-cli")
[ -x "$BIN" ] || { echo "no binary"; exit 1; }
"$BIN" --version
```

### 2. Catalogue

```bash
"$BIN" mcp tools 2>&1 | tail -3
```

Count tools. Compare to the count of `func toolXxx() tool {`
in `cmd/mcp/tools.go` — should match.

### 3. Handshake

Reuse the recipe from `/mcp-test`. Capture stdout line count
(expect 2) and stderr length.

### 4. Sample `tools/call`

Pick the safest tool from the catalogue (read-only, no required
arg): `lark_calendar_agenda` with `{}` (today is the default), or
`lark_task_my` with `{}`. Both are safe. If neither exists, skip
this step and note "no safe sample".

### 5. Host config

```bash
HOST_CFG="$HOME/Library/Application Support/Claude/claude_desktop_config.json"
if [ -f "$HOST_CFG" ]; then
  python3 -c "
import json
c = json.load(open('$HOST_CFG'))
servers = c.get('mcpServers', {})
for name, s in servers.items():
    if 'lark' in name.lower():
        print(name, '→', s.get('command'), s.get('args'))
"
fi
```

### 6. Recent host log

`tail -20` of the latest `~/Library/Logs/Claude/mcp*.log` if any.

### 7. Auth

`lark-cli auth status` first 3 lines + `lark-cli profile list` head.

## Report shape

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 MCP doctor — <timestamp>
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Binary       <path> <version>                    ✅ / ❌
Catalogue    <N> tools (source: <M>)              ✅ / ⚠ drift / ❌
Handshake    initialize + tools/list              ✅ / ❌
Sample call  <tool> → isError=<bool>              ✅ / ❌
Host config  registered as "<name>" → <path>      ✅ / ❌ / not-installed
Host log     last error: "<line or —>"
Auth         profile=<name> identity=<u/b>        ✅ / ❌
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
VERDICT      HEALTHY / DEGRADED / BROKEN
NEXT ACTION  <one sentence, only if not HEALTHY>
```

Then stop. Do not start work — the user might just be checking.

## Hard rules

- Pure observation. Do not mutate any state — no `auth login`, no
  host config edits, no rebuilds.
- If a step requires sudo or a destructive command to even read
  (rare here), skip it and note "skipped, requires elevation".
- Always clean up temp files: `rm -f /tmp/mcp-doctor.*`.
