---
name: client-followup
description: Detect dormant CRM contacts (>21d) and draft personalized re-engagement mails (drafts only — never auto-sends). Triggers "khách im lặng", "follow-up khách".
version: 1.0.0
last_updated: 2026-05-11
---

# client-followup

"Who haven't I talked to in a while." Surfaces dormant clients +
drafts personalized re-engagement mails. Drafts only — user sends
manually. Full pipeline view → `pipeline-review`.

**Verbs** (`.claude/SHORTCUTS.generated.md`): `drive +search
--doc-types bitable` (find CRM Base), `base +record-list`
(`--base-token/--table-id/--view-id` — dormant via a filtered view,
no `--filter` flag), `base +record-get`, `vc +search`
(`--participant-ids/--start/--end`), `mail +draft-create`
(`--to/--subject/--body` — creates a draft, never sends). Mail's list
verb is `mail +triage` (date/from/subject/message_id; `--query` for
search) — inbox/folder survey also via atomic `lark-mail` skill.
Atomic shapes + token flags: `.claude/RECIPES.md`.

## Workflow

1. **Discover CRM table.** `memory/accounts.md` has Base token, else
   `drive +search --doc-types bitable --query "CRM" --jq '.data'`
   (peek shape — needs `search:docs:read` scope) → ask user to pick.
2. **Threshold.** Default 21d no-touch (arg `--days N` overrides).
3. **Pull dormant clients.** `base +record-list --base-token <t>
   --table-id <CRM> --view-id dormant-active --field-id name
   --field-id company --field-id last_touch --field-id email --limit 50`
   (set up a filtered "dormant-active" view in the Base, or pull
   active and filter `last_touch < threshold` client-side).
4. **Per client, gather context** (parallel batch): mail history via
   atomic `lark-mail` skill; `vc +search --participant-ids <open_id>
   --start <t-180d> --end <t> --jq '.data.items[]|{id,info:.display_info}'`
   (`vc +search` `data` is a dict — meetings live in `.data.items[]`);
   CRM record notes via `base +record-get`.
5. **Synthesize per client:** last topic discussed, pain point stated,
   a genuine reason to re-engage.
6. **Draft mail per client.** `mail +draft-create --to <email>
   --subject "<personalized>" --body "<refs last topic + reason>"` →
   creates a draft (never sends). Capture draft info.
7. **Report:** dormant count + per-client block (name/company, last
   touch + days ago, last topic, suggested angle, draft preview).
   End: "Review each draft in Lark Mail UI before sending — NOT sent."

## Hard rules

1. NEVER auto-send — `mail +draft-create` only. Drafts live in Lark
   Mail UI; user reviews + sends there.
2. Personalize — each draft refs ≥1 concrete history item. No generic
   "checking in".
3. Tone gentle, not pushy — no "limited time" / "before it's too late".
4. Respect per-account tone from `memory/accounts.md`.
5. Don't re-draft if followed up recently — check mail history; a
   draft from last 14d not yet sent → surface it instead.
6. Batch ≤20 drafts per run — split if more.

## Memory

Required: `accounts.md` (tone + last context per VIP client).
Recommended: `playbooks.md`, `won-deals.md`. Empty → drafts more
generic, emits a setup warning.
