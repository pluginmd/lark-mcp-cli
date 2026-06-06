---
description: List the MCP tools currently exposed by `lark-cli mcp serve`
allowed-tools: Bash
---

Run `lark-cli mcp tools` against the active install and print the
catalogue. This is a one-shot read; no edits.

```bash
lark-cli mcp tools 2>&1 || ~/bin/lark-cli mcp tools 2>&1
```

If neither works, surface the error verbatim — the binary is likely
not on PATH yet. Suggest running `/mcp-rebuild` to (re)install.

After printing the catalogue, summarise in one sentence which tool
domains are covered (IM, Calendar, Docs, Base, Contact, Task, Drive,
plus the generic `lark_api` passthrough — note any gaps).
