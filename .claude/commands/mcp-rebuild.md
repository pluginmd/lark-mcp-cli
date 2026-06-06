---
description: Rebuild lark-cli and reinstall the binary so Claude Desktop picks up changes
argument-hint: [install-path, default ~/bin/lark-cli]
allowed-tools: Bash, Read
---

Rebuild the local binary after MCP changes and (re)install it where
the MCP host will find it.

## 1. Pick the install path

Use `$1` if supplied. Otherwise default to `~/bin/lark-cli`.
If the user explicitly wants `/usr/local/bin/lark-cli`, warn that
sudo will be required and ask for confirmation.

## 2. Build

```bash
go build -o /tmp/lark-cli .
ls -la /tmp/lark-cli
```

If `go build` fails, stop and surface the error.

## 3. Smoke-check the build

```bash
/tmp/lark-cli --version
/tmp/lark-cli mcp tools | tail -3
```

Both must succeed. The `mcp tools` output should show "20 tools total"
(or however many are defined in `cmd/mcp/tools.go`).

## 4. Install

```bash
cp /tmp/lark-cli <install-path>
<install-path> --version
```

## 5. Remind to restart

After installing, the MCP host (Claude Desktop, Claude Code) still
holds a handle to the **old** binary process. Tell the user:

- Claude Desktop: Cmd+Q (full quit) and reopen.
- Claude Code: the next session picks up the new binary; no restart
  needed for ad-hoc CLI use.

## 6. Report

- Install path.
- Binary version string and modification time.
- Tool count.
- Whether a restart is needed for the active host.

Do not edit any source. This command is build-and-install only.
