---
description: Walk through adding a new tool to the lark-cli MCP bridge
argument-hint: <tool-name> [underlying shortcut, e.g. "im +messages-pin"]
allowed-tools: Bash, Read, Edit, Write, Grep, Glob
---

You will add a new MCP tool to `cmd/mcp/tools.go`. Follow these
steps exactly; do not skip the verification.

## 1. Resolve the shortcut

If the user supplied the underlying shortcut as the second argument,
use it directly. Otherwise, ask the user OR delegate to the
`shortcut-explorer` agent to find the right service/verb pair (conceptual `service +command`).

Confirm the shortcut's exact flag names by reading
`shortcuts/<domain>/<file>.go`. **Never guess flags.**

## 2. Pick a tool name

Format: `lark_<domain>_<verb>` — e.g. `lark_im_pin`, `lark_doc_export`.
Stay consistent with the existing 20 tools.

## 3. Edit `cmd/mcp/tools.go`

1. Append a new `toolXxx()` factory near the bottom of the file.
2. Register it in `allTools()`, grouped with siblings of the same
   domain.
3. Define the JSON Schema as a raw string literal. Required fields
   go in the top-level `required` array.
4. In the `Build` closure:
   - Use `argString(args, "key")` for required strings.
   - Use `appendFlag(argv, "flag-name", value)` for optional flags.
   - Add `appendIdentity(argv, args)` at the end so callers can
     pass `as: "bot"`.
   - Map MCP arg names with `_` to CLI flag names with `-`.

## 4. Verify

```bash
go build -o ~/bin/lark-cli .
~/bin/lark-cli mcp tools | grep <tool-name>
```

Then run the smoke handshake (from `cmd/mcp/README.md` §Verify) and
issue a real `tools/call` for the new tool with safe arguments
(`--dry-run` flag if the shortcut supports it).

## 5. Update docs

- Append the tool to the catalogue table in `cmd/mcp/README.md`.

## 6. Report

Summarise:

- Tool name + 1-line description.
- Exact CLI invocation it generates.
- Verification output (build status, mcp tools listing line, smoke
  test result).
- Any flag names you could not confirm from source.

## Hard rules

- No third-party MCP library imports.
- No `fmt.Println` or any stdout write inside `cmd/mcp/`. Stdout
  is JSON-RPC only.
- One MCP tool wraps exactly one shortcut. Composite flows go to
  `lark_api` or are left to the model to chain.
