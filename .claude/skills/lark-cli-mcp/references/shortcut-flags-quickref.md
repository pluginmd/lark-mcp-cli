# Shortcut flag quick reference

Common flags across lark-cli shortcuts. Useful when drafting MCP
tool builders. **Always confirm against the actual shortcut source
before shipping** — this list is a starting point, not authoritative.

## Identity & output (global)

| Flag                  | Meaning                                                       |
| --------------------- | ------------------------------------------------------------- |
| `--as user|bot`       | Identity for the call. Defaults vary per shortcut.            |
| `--format json|ndjson|table|csv|pretty` | Output format. **Always JSON for MCP.**     |
| `--params <json>`     | URL/query parameters (generic `api` command).                 |
| `--data <json>`       | Request body (generic `api` command).                         |
| `--page-all`          | Auto-paginate through all pages.                              |
| `--page-size N`       | Page size.                                                    |
| `--page-limit N`      | Max pages when `--page-all` (default 10).                     |
| `--dry-run`           | Print the request without executing.                          |
| `--jq <expr>` / `-q`  | jq filter on JSON output. Mutually exclusive with a non-JSON `--format` — use only with `--format json` (the default). Not every shortcut exposes it (e.g. `base +record-search`, `base +data-query` do not). |

## IM (`shortcuts/im/`)

| Shortcut                | Common flags                                                                          |
| ----------------------- | ------------------------------------------------------------------------------------- |
| `im +messages-send`     | `--chat-id`, `--user-id`, `--text`, `--markdown`, `--msg-type`, `--content`, `--idempotency-key`, `--image`, `--file`, `--video`, `--video-cover`, `--audio` |
| `im +messages-search`   | `--query`, `--page-size`                                                              |
| `im +messages-reply`    | `--message-id`, `--text`, `--markdown`, `--content`                                   |
| `im +chat-search`       | `--query`                                                                             |
| `im +chat-create`       | `--name`, `--description`, `--users` (open_ids, comma-separated), `--bots`, `--owner`, `--type` (private\|public) |

> ⚠️ Verified against live `lark-cli <domain> +<verb> --help` on
> 2026-05-15. If a flag mismatch shows up in a tool call, `--help`
> (and the Go source under `shortcuts/<domain>/*.go`) is
> authoritative — update this table to match, not the other way.

## Calendar (`shortcuts/calendar/`)

| Shortcut                | Common flags                                                                          |
| ----------------------- | ------------------------------------------------------------------------------------- |
| `calendar +agenda`      | `--start`, `--end`, `--calendar-id` (defaults to today on primary calendar)           |
| `calendar +create`      | `--summary`, `--start`, `--end`, `--description`, `--attendee-ids`, `--calendar-id`, `--rrule` |
| `calendar +freebusy`    | `--user-id`, `--start`, `--end`                                                       |
| `calendar +suggestion`  | `--attendee-ids`, `--duration-minutes`, `--start`, `--end`, `--exclude`, `--timezone`  |

## Docs (`shortcuts/doc/`, Service: `docs`)

| Shortcut                | Common flags                                                                          |
| ----------------------- | ------------------------------------------------------------------------------------- |
| `docs +create`          | `--api-version` (default `v1`; pass `v2` for current schema), `--title`, `--markdown`, `--folder-token`, `--wiki-node`, `--wiki-space` |
| `docs +fetch`           | `--api-version` (default `v1`; pass `v2` for current schema), `--doc` (URL or token), `--limit`, `--offset` |
| `docs +search`          | `--query`, `--filter`, `--page-size` (default 15, max 20)                             |
| `docs +update`          | `--api-version` (default `v1`; pass `v2` for current schema), `--doc`, plus operation-specific flags |

## Base / Bitable (`shortcuts/base/`)

| Shortcut                | Common flags                                                                          |
| ----------------------- | ------------------------------------------------------------------------------------- |
| `base +record-search`   | `--base-token`, `--table-id`, `--json '{"keyword":"...","search_fields":[...]}'` (REQUIRED) |
| `base +record-list`     | `--base-token`, `--table-id`, `--view-id`, `--field-id` (repeatable), `--limit`, `--offset` |
| `base +record-batch-create` | `--base-token`, `--table-id`, `--json <batch create JSON object>`                 |
| `base +record-batch-update` | `--base-token`, `--table-id`, `--json '{"record_id_list":[...],"patch":{...}}'`    |
| `base +data-query`      | `--base-token`, `--dsl '<LiteQuery JSON>'` (server-side count/sum/group-by)        |
| `base +field-list`      | `--base-token`, `--table-id`                                                          |

## Contact (`shortcuts/contact/`)

| Shortcut                | Common flags                                                                          |
| ----------------------- | ------------------------------------------------------------------------------------- |
| `contact +search-user`  | `--query` OR `--user-ids`, `--page-size` (1-30, default 20) — requires `--as user`   |
| `contact +get-user`     | `--user-id` (single open_id)                                                          |

## Task (`shortcuts/task/`)

| Shortcut                | Common flags                                                                          |
| ----------------------- | ------------------------------------------------------------------------------------- |
| `task +create`          | `--summary`, `--description`, `--due`, `--assignee` (SINGULAR, one per call), `--follower`, `--tasklist-id` |
| `task +get-my-tasks`    | `--query`, `--complete`, `--created_at`, `--due-start`, `--due-end`, `--page-limit`   |
| `task +complete`        | `--task-id`                                                                           |
| `task +search`          | `--query`, `--completed` (bool), `--assignee`, `--creator`, `--follower`, `--due` (start,end) |
| `task +get-related-tasks` | `--created-by-me`, `--followed-by-me`, `--include-complete`, `--page-limit`          |

## Mail (`shortcuts/mail/`)

| Shortcut                | Common flags                                                                          |
| ----------------------- | ------------------------------------------------------------------------------------- |
| `mail +send`            | `--to`, `--subject`, `--body`, `--cc`, `--bcc`, `--from`, `--attach`, `--confirm-send` (without it, creates a draft) |
| `mail +draft-create`    | `--to`, `--subject`, `--body`, `--cc`, `--bcc`, `--attach`                            |
| `mail +reply`           | `--body`, `--cc`, `--attach`, `--confirm-send` (without it, saves a draft); separate verb `mail +reply-all` |
| `mail +message`         | `--message-id` (SINGULAR, required), `--mailbox`, `--html`                            |

## IM (`shortcuts/im/`)

| Shortcut                | Common flags                                                                          |
| ----------------------- | ------------------------------------------------------------------------------------- |
| `im +messages-send`     | `--chat-id` OR `--user-id`, `--text` / `--markdown` / `--content`                     |
| `im +messages-reply`    | `--message-id`, `--text` / `--markdown`                                               |
| `im +messages-search`   | `--query`, `--page-size`                                                              |
| `im +chat-messages-list`| `--chat-id` XOR `--user-id`, `--start`, `--end`, `--page-size`, `--sort`              |
| `im +threads-messages-list` | `--thread`, `--page-size`, `--sort`                                              |

## Drive (`shortcuts/drive/`)

| Shortcut                | Common flags                                                                          |
| ----------------------- | ------------------------------------------------------------------------------------- |
| `drive +upload`         | `--file` (REQUIRED), `--folder-token` XOR `--wiki-token`, `--name`                    |
| `drive +create-folder`  | `--name`, `--parent-token`                                                            |
| `drive +download`       | `--file-token`, `--output`                                                            |
| `drive +search`         | `--query`, `--doc-types` (doc,sheet,bitable,file,wiki,docx,folder,slides…), `--folder-tokens`, `--space-ids`, `--creator-ids`, `--mine` |

## Patterns to memorise

1. **IDs end in `-id` or `-token`**: `--chat-id`, `--user-id`,
   `--app-token`, `--doc-token`, `--file-token`.
2. **Lists use comma-separated strings**, not arrays: `--user-ids
   "ou_a,ou_b"`.
3. **Bodies passed as JSON strings**: `--data '{"k":"v"}'`,
   `--fields '{"Name":"…"}'`.
4. **Inputs from disk**: many shortcuts accept `@/path/to/file` as
   the flag value (look for `Input: []string{File}` in the Flag
   definition).
5. **stdin**: `Input: []string{Stdin}` enables `-` as the value.

## Verification command

When in doubt:

```bash
lark-cli <domain> +<command> --help
```

prints the live, authoritative flag list for the installed binary.
