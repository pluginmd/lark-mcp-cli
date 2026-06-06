---
name: sprint-retro
description: End-of-sprint retro draft — closed tickets, velocity, retro form, blockers. Triggers "sprint retro", "retro tuần này", "/retro".
version: 1.0.0
last_updated: 2026-05-11
---

# sprint-retro

End-of-sprint "what went well / didn't / change" doc draft, built on
data not vibes. Continuous standup → `lark-workflow-standup-report`.

**Verbs** (`.claude/SHORTCUTS.generated.md`): `task +search`
(`--completed` bool, `--due` range — no `--completed-between`),
`base +record-list` (`--base-token/--table-id/--view-id` — scope via
a pre-filtered view, no `--filter` flag), `base +data-query --dsl`
(server-side counts), `im +messages-search` (`--query/--chat-id/--start`).
Atomic shapes + token flags: `.claude/RECIPES.md`.

## Workflow

1. **Sprint window.** `memory/velocity-history.md` last entry end +1d
   = current start. Default 14d if no history. User can override.
2. **Pull data (parallel):**
   - Closed tickets: `task +search --completed --due <start>,<end>
     --page-limit 20 --jq '.data.items[]|{summary,completed_at}'`
     (`data` is a dict — tasks in `.data.items[]`). Base tracker
     variant: `base +record-list --base-token <t> --table-id
     <tracker> --view-id closed-this-sprint --field-id name --field-id points`.
   - Velocity: closed count this sprint vs 3 prior (`velocity-history.md`),
     %change. Or `base +data-query --dsl` for server-side count.
   - Retro form: `base +record-list --base-token <t> --table-id
     retro_feedback --view-id <current-sprint-view>`; bucket by sentiment.
   - Blocker signal: `im +messages-search --query "blocker OR stuck OR
     chờ OR wait OR hold" --chat-id <dev-group> --start <start>
     --jq '.data'` (peek shape; needs `search:message` scope); group
     by keyword, count.
3. **Synthesize:** velocity delta + signal · top 3 recurring blockers ·
   win patterns (biggest/hardest closed) · 3 representative retro quotes.
4. **Generate doc draft** via `doc-from-template` (custom retro shape):
   Numbers / What went well / What didn't / Try next sprint
   (checkboxes) / Open questions. ≤1 page.
5. **Offer:** save to wiki → `lark-doc-author`; update
   `velocity-history.md` (confirm); schedule retro meeting → `lark-assistant`.

## Hard rules

1. Anonymize form feedback by default (unless manager + small team + opt-in).
2. Don't quote IM verbatim — paraphrase.
3. Every number cites a source (Base table / task filter + timestamp).
4. Read-only data — doc draft is the only output, no task create.
5. Compact — retro doc ≤1 page, bullets not paragraphs.

## Memory

Required: `velocity-history.md` (or 14d default).
Recommended: `team-roster.md` (dev group), `project-map.md` (trackers scope).
