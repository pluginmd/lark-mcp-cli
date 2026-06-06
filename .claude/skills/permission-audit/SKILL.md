---
name: permission-audit
description: Read-only scan of Drive/Doc/Wiki/Base for risky permissions (public, external, PII). Triggers "permission audit", "quét quyền", "PII leak".
version: 1.0.0
last_updated: 2026-05-11
---

# permission-audit

Read-only scan → one severity-ranked report. No auto-fix, no
auto-revoke, no notification.

**Verbs** (`.claude/SHORTCUTS.generated.md`): `drive +search`
(`--doc-types/--folder-tokens/--space-ids/--creator-ids` — inventory
files; NO permission flag), `docs +fetch` (content for PII scan),
`base +table-list`. Raw CLI has **no verb that returns per-file
share/permission state** — get permission metadata via the atomic
`lark-drive` skill (or note the gap in the report). Atomic shapes +
token flags: `.claude/RECIPES.md`.

## Workflow

1. **Scope.** Default: all Drive/Wiki/Base the user admins. Narrow via
   `--folder-tokens`, `--space-ids`, `--creator-ids`. >1000 items →
   warn + suggest batching.
2. **Inventory.** `drive +search --doc-types doc,sheet,bitable,file
   --folder-tokens <t> --jq '.data'` (peek shape; + `--space-ids` for
   wiki; needs `search:docs:read` scope). Permission state → atomic
   `lark-drive` skill per item (batch). Cache for the session — no re-fetch.
3. **Classify by memory rule:** `public-internet` → HIGH (CRIT if
   strict policy) · `anyone-with-link` → MED · `external_users >0` →
   MED (HIGH if folder is Confidential/HR/Finance) · `inactive_owner`
   → MED.
4. **PII/DLP scan on HIGH+ items.** `docs +fetch --doc <url>
   --limit 1` (cap content), regex `memory/regulated-data-types.md`:
   VN CCCD `\b\d{9}\b|\b\d{12}\b`, Luhn-valid 13-19 digit cards,
   email+phone batch (>10/file), API keys (`sk-`, `ghp_`, `AKIA`,
   `xoxb-`). Each match: severity +1, mask 50% digits.
5. **Dedupe vs `incident-log.md`** — already-raised → mark RECURRENT.
6. **Format report** — SUMMARY (counts by severity) + CRIT/HIGH
   findings (token/url/owner/share/PII/policy ref/action/status) +
   OWNER EXPOSURE top 5 + RECOMMENDED ACTIONS. Ends "read-only output,
   no changes made."

## Hard rules

1. READ-ONLY — never call `+update`/`+apply-permission`/`+delete`.
   "Fix luôn" → redirect to admin Lark UI or offer a task (confirm).
2. Mask sensitive data — `4111-****-****-1234`, CCCD 3+3 digits only.
3. Don't dump file contents — report token + URL only.
4. Scan only where user has admin. `403` → skip silent, no brute force.
5. Severity from policy — every finding refs a `memory/policies.md` §.
6. Report to admin privately — 1:1 chat, never wide group.

## Memory

Required: `policies.md` (sharing rules + severity map),
`regulated-data-types.md` (PII patterns).
Recommended: `incident-log.md` (dedupe recurrent).

Empty memory → conservative defaults (anyone-with-link = HIGH) +
warning to set up `policies.md`.
