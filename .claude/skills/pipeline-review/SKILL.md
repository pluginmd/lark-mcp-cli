---
name: pipeline-review
description: Weekly pipeline scan — by stage, stuck deals, closing soon, win-rate trend. Triggers "pipeline review", "tổng quan deal", "weekly pipeline".
version: 1.0.0
last_updated: 2026-05-11
---

# pipeline-review

Monday-morning sales scan. Surfaces what needs attention this week
without dumping the full CRM.

**Verbs** (`.claude/SHORTCUTS.generated.md`): `base +record-list`
(`--base-token/--table-id/--view-id/--field-id` — project only needed
fields), `base +data-query --dsl` (server-side count/sum/group-by —
use this for stage rollups + win-rate, not client-side aggregation),
`drive +search --doc-types bitable` (locate CRM Base by name).
Atomic shapes + token flags: `.claude/RECIPES.md`.

## Workflow

1. **Locate CRM Base.** `memory/accounts.md` / `INDEX.md` should have
   the Base token. Fallback: `drive +search --doc-types bitable
   --query "CRM" --jq '.data'` (peek shape; needs `search:docs:read` scope).
2. **Stage rollup — server-side.** `base +data-query --base-token <t>
   --dsl '<group-by stage, count + sum value>'` returns the computed
   table directly (don't pull all rows then aggregate).
3. **Pull only attention-worthy rows.** `base +record-list
   --base-token <t> --table-id <CRM> --view-id all_active --field-id
   name --field-id stage --field-id value --field-id owner --field-id
   last_touch --field-id close_date --limit 100 --format json`.
4. **Filter** (in agent): `last_touch < today-14d` → stuck; `close_date
   < today+30d AND stage IN (negotiation, proposal)` → closing soon.
   Win-rate via a second `base +data-query` (won / (won+lost) per month).
5. **Cross-check memory.** Tier 1 accounts from `accounts.md` always
   shown; dormant (last_touch >21d) highlighted.
6. **Format report:** BY STAGE (count + $) · STUCK (>14d) · CLOSING
   SOON (<30d, with risk) · WIN RATE (this vs last month) · TIER 1
   STATUS · RECOMMENDED FOCUS (top 3 actions).
7. **Offer:** draft re-engage mail for stuck deals → `client-followup`;
   create follow-up tasks for top 3 → confirm → `task +create`; save
   report → `lark-doc-author`.

## Hard rules

1. Read-only scan — no deal-stage update or task create without confirm.
2. >100 deals → summary stats + top 10 per bucket, no full list.
3. ACV/value shown only if user is account owner or sales lead
   (check `team-roster.md` role).
4. Tone factual — statement + suggested action, no editorializing.
5. Last_touch source: prefer CRM field; fallback via atomic
   `lark-mail` skill (`mail +triage` to scan, then `mail +thread`) —
   `mail +messages` is fetch-by-ID, not a history/search verb.

## Edge cases

- Non-standard CRM schema → ask user to map once, save to
  `accounts.md` schema section.
- Custom stages (Demo/POC) → support user-defined, don't hardcode.
- No CRM Base → fail gracefully, suggest creating from template.

## Memory

Required: `accounts.md` (Tier highlight + tone).
Recommended: `INDEX.md` (CRM Base token), `team-roster.md` (ACV visibility).

## Cadence

Monday morning. Pairs with `morning-brief` sales variant.
