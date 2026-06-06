# Tool builder cookbook

Step-by-step recipe for adding a new MCP tool to `cmd/mcp/tools.go`.

## Anatomy of a tool

```go
type tool struct {
    Name        string                                                  // MCP tool name (e.g. "lark_im_history")
    Description string                                                  // What the host model sees
    Schema      string                                                  // Raw JSON Schema literal
    Build       func(args map[string]interface{}) ([]string, error)     // Args → lark-cli argv
}
```

Three pieces. The schema teaches the model how to call the tool;
the builder translates those args into a CLI invocation.

## Recipe

### 1. Confirm the shortcut exists and the flag names

Never trust the README or your memory. Open
`shortcuts/<domain>/<file>.go` and read the `Flags []Flag{...}`
block. Or run `lark-cli <domain> +<command> --help` and copy the
flag names verbatim.

### 2. Name the tool

`lark_<domain>_<verb>`. Examples that already exist:
`lark_im_send`, `lark_calendar_agenda`, `lark_doc_fetch`,
`lark_base_search`. Stay consistent.

### 3. Draft the schema

JSON Schema, embedded as a raw string literal. Conventions:

- Top-level `"type": "object"`.
- `properties` keyed by **snake_case** MCP argument names.
- `required` array for hard-required arguments.
- `enum` for closed sets.
- Always include `"as": {"type": "string", "enum": ["user", "bot"]}`
  if the underlying shortcut supports both identities.

```go
Schema: `{
  "type": "object",
  "properties": {
    "chat_id": {"type": "string", "description": "Target chat (oc_xxx)"},
    "as":      {"type": "string", "enum": ["user", "bot"]}
  },
  "required": ["chat_id"]
}`,
```

### 4. Write the Build closure

```go
Build: func(args map[string]interface{}) ([]string, error) {
    chatID, ok := argString(args, "chat_id")
    if !ok {
        return nil, fmt.Errorf("chat_id is required")
    }
    argv := []string{"im", "+chat-messages-list", "--chat-id", chatID}
    argv = appendIdentity(argv, args)
    return argv, nil
},
```

Available helpers (defined in `tools.go`):

| Helper                         | Use                                                |
| ------------------------------ | -------------------------------------------------- |
| `argString(args, "key")`              | Read a string, returns `("", false)` when missing/empty. |
| `argInt(args, "key")`                 | Read a number; "present" signal for int fields.   |
| `appendFlag(argv, "flag", v)`         | Add `--flag value` only if `v` is non-empty.      |
| `appendBoolFlag(argv, args, "key", "flag")` | Add bare `--flag` only when `args["key"]` is literal `true` (omit-to-disable). |
| `appendIdentity(argv, args)`          | Add `--as <user|bot>` from `args["as"]`.          |
| `appendFormat(argv, args)`            | Add `--format <fmt>` from `args["format"]` (rarely needed). |
| `requireOneOf(args, "a", "b", …)`     | Error if none of the listed keys has a value.     |

### 5. Register

In `allTools()`:

```go
func allTools() []tool {
    return []tool{
        toolIMSend(),
        toolIMSearch(),
        toolIMHistory(),           // ← new, grouped with siblings
        // …
    }
}
```

### 6. Verify

```bash
go build -o ~/bin/lark-cli .
~/bin/lark-cli mcp tools | grep lark_im_history
```

Then JSON-RPC handshake (see `architecture.md` or `cmd/mcp/README.md`).
If the shortcut supports `--dry-run`, issue a real `tools/call` with
those args to confirm argv translation.

### 7. Document

Append a row to the catalogue table in `cmd/mcp/README.md` so users
discover the new tool.

## Mapping conventions

| MCP arg          | CLI flag                  |
| ---------------- | ------------------------- |
| `chat_id`        | `--chat-id`               |
| `user_id`        | `--user-id`               |
| `page_size`      | `--page-size`             |
| `app_token`      | `--app-token`             |
| `as`             | `--as`                    |
| `format`         | `--format`                |

The underscore → dash translation is **manual** — `appendFlag`
takes the CLI flag name explicitly. Don't auto-translate; some
flags have non-obvious names (e.g. `--doc-token` not `--doc-id`).

## Pitfalls

- **Schema is invalid JSON**: `descriptors()` validates at boot and
  fails loudly. If `mcp serve` exits with "invalid schema for tool",
  re-check your literal — likely a trailing comma or unescaped quote.
- **Flag name wrong**: tool call succeeds at the MCP layer but the
  subprocess exits non-zero. Read the stderr block in the MCP
  response; it points at the bad flag.
- **Stdout pollution**: if the new tool's stdout has anything other
  than JSON, the host may render garbage. Always pipe through
  `--format json` (the lark-cli default) and avoid `--format pretty`.
- **Risk: high-risk-write**: think twice before exposing destructive
  shortcuts (e.g. `+record-delete`, `+chat-delete`). If you do,
  surface the risk in the tool description.
