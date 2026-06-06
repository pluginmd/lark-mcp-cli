---
name: weekly-review
description: Weekly view of calendar + tasks + OKR as one narrative. Triggers "weekly review", "báo cáo tuần", "tổng kết tuần".
version: 1.0.0
last_updated: 2026-05-11
---

# weekly-review

Weekly wrap-up workflow. Produces one coherent narrative across
calendar, tasks, and OKR — not three disconnected dumps.

## When to activate

- "Weekly review", "báo cáo tuần", "tuần này có gì"
- "Tổng kết tuần", "weekly digest", "Friday wrap"
- "OKR progress tuần này"

For daily summary, use `daily-digest` skill instead.
For just a quick agenda, use `/agenda week`.

## Workflow

Verbs (`.claude/SHORTCUTS.generated.md`): `calendar +agenda`
(`--start/--end`), `task +search` (`--completed/--due`),
`task +get-my-tasks` (`--complete=false`), `okr +cycle-list`
(`--time-range`), `okr +progress-list`. Atomic shapes + token flags:
`.claude/RECIPES.md`.

1. **Week range** — default current ISO week (Mon–Fri); user override.
2. **Pull data in parallel** via `lark-week-planner` subagent, all
   output-projected:
   - `calendar +agenda --start <Mon> --end <Fri>
     --jq '.data[]|{summary,start:.start_time.datetime}'`.
   - `task +search --completed --due <Mon>,<Fri> --jq '.data'`.
   - `task +get-my-tasks --complete=false --page-limit 20
     --jq '.data.items[]|{summary,due:.due_at}'` (flag overdue client-side).
   - `okr +cycle-list --time-range <YYYY-MM--YYYY-MM>` →
     `okr +progress-list` per KR.
3. **Synthesize narrative**: don't just list — surface signal.
   - Time allocation: % in meetings vs focus blocks.
   - Completion rate: done / planned.
   - OKR velocity: KR progress delta vs week start.
   - Theme: what dominated this week? (use top tags/projects)
4. **Format**:

```
WEEKLY REVIEW · Tuần <ISO> (<Mon date> – <Fri date>)

╔══ THEMES ══╗
  • <theme 1 with 1-line elaboration>
  • <theme 2>

╔══ ALLOCATION ══╗
  Meetings:    <N>h (<X>% of working hours)
  Focus time:  <N>h
  Top 3 meetings by time spent:
    1. <title> — <h>h
    2. <title> — <h>h
    3. <title> — <h>h

╔══ EXECUTION ══╗
  Tasks done:    <count> ✅
  Tasks open:    <count> (<overdue> overdue 🔴)
  Highlights:
    • <key delivery>
    • <key delivery>

╔══ OKR ══╗
  Period: <name>
  • <KR title>: <% start> → <% now> (Δ +<X>%)
  • ...

╔══ REFLECTION ══╗
  Wins:        <1-2 items>
  Friction:    <1-2 items>
  Next week:   <suggested focus>
```

5. **Offer follow-ups**:
   - Save to a wiki page? → hand off to `lark-doc-author`.
   - Send via mail? → hand off to `/draft-mail`.
   - Update OKR KR? → confirm before running `okr +progress-update`.

## Hard rules

- **Read-only** during data gathering. Any write (save report, send,
  update OKR) requires explicit confirmation.
- **Names not IDs** — resolve all attendees/assignees.
- **No fabrication**: if a KR has no recent updates, say so. Don't
  assume progress.
- **Compact output** — full report fits on one screen (~40 lines).
- **Don't double-count tasks** completed multiple times due to
  re-opening.

## Variants

- `weekly-review --light` — skip OKR section if user doesn't run OKRs.
- `weekly-review --send-to <person>` — finish with a draft mail
  containing the report (still requires send confirmation).
- `weekly-review --save-wiki <space>` — auto-offer to save into wiki.
