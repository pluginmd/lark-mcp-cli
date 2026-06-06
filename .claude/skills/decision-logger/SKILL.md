---
name: decision-logger
description: Detect decisions in IM/minutes and promote them to a structured Base table. Triggers "decision log", "chốt cái này", "log lại decision".
version: 1.0.0
last_updated: 2026-05-11
---

# decision-logger

Surfaces decisions from IM/meetings and persists them into a Base
table — so "we chốt cái này nhưng không ai nhớ ở đâu" stops happening.

**Verbs** (`.claude/SHORTCUTS.generated.md`): `im +messages-search`,
`im +threads-messages-list`, `vc +search`, `vc +notes`,
`base +record-batch-create`, `base +record-search`,
`base +table-create`, `docs +fetch`. AI minutes artifacts → atomic
`lark-minutes` skill (no raw `+summary` verb). For atomic command
shapes + token flags see `.claude/RECIPES.md`.

## Prerequisite — decisions Base table

Default schema: Title (Text), Decision (LongText), Decided by
(Member), Date decided (DateTime), Context source (URL), Affected
area (SingleSelect: Engineering/Sales/Product/Ops/Other), Stakeholders
(Member), Status (SingleSelect: Active/Superseded/Reversed), Supersedes
(Link → Decision), Review date (DateTime), Tags (MultiSelect).

Auto-create via `base +table-create` if missing — confirm first.

## Mode A — explicit promote

1. **Capture source.** IM thread → `im +threads-messages-list
   --thread <id> --jq '.data'` (peek shape). Meeting →
   `vc +notes --meeting-ids <id>`. Doc → `docs +fetch --api-version v2
   --doc <url> --limit <n> --offset <n>` (no keyword/section mode —
   only `--limit`/`--offset` paging).
2. **Extract.** Decision-keyword sentences ("chốt", "decided",
   "approved", "OK đi", "we'll go with", "agreed on"); decider name;
   affected area from group/meeting name.
3. **Echo intent** — show Title/Decision/Decided by/Date/Source/Area/
   Stakeholders, then confirm 3 things: Status, Review date (default
   90d), Supersedes? (search → pick / no).
4. **On confirm:** `base +record-batch-create --base-token <t>
   --table-id <t> --json '{...}'` → pre-mutate hook + audit log.
5. **Cross-link** (with confirm): reply in source IM thread with Base
   record URL; @mention stakeholders.

## Mode B — scan & propose

1. Read `priority-topics.md`.
2. Sweep 7d: `im +messages-search --query "chốt OR decided OR approve"
   --page-size 50 --jq '.data'` (peek shape; needs `search:message`
   scope) + `vc +search --start <t> --end <t>
   --jq '.data.items[]|{id,info:.display_info}'` (`data` is a dict —
   meetings in `.data.items[]`).
3. Filter & cluster: drop noise (require ≥1 priority-topic match OR
   stakeholder count ≥3); same topic+day+group → one candidate; rank
   by stakeholder count.
4. Present candidates list with per-item Log/skip/merge + bulk action.
5. Per-confirm: same `base +record-batch-create` flow.

## Hard rules

1. Never auto-create record without echo+confirm (do-not.md §6).
2. Source URL required — every decision links back to IM/meeting/doc.
3. Resolve names, never raw open_ids.
4. No @mention of stakeholders until explicit confirm.
5. Supersedes: set prior record Status=Superseded + link; never delete.
6. Review date default 90d (override in optional `decision-policy.md`).
7. Don't log private 1:1 decisions without explicit confirm.

## Edge cases

- No table → prompt create with default schema (confirm).
- Keyword in joke context → require priority-topic match OR ≥3 stakeholders.
- Ambiguous decider → list all, user picks primary.
- Spans multiple meetings → separate records linked via Supersedes/extends.
- Reversal → search prior in Base, mark Status=Reversed, log linked new record.

## Memory

Required: `do-not.md`. Recommended: `priority-topics.md`,
`team-roster.md`, optional `decision-policy.md`.

## Cadence

On-demand (explicit); weekly Mode B sweep Monday (via `morning-brief`);
post-meeting offer from `meeting-prep`. Also called by `incident-retro`
and `sprint-retro`. Base over Wiki because decisions need queryability.
