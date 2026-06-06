---
name: task-prioritizer
description: Rank tasks by deadline×risk + blocking + OKR-linked + assigner weight. Surface top 5 today. Triggers "việc nào quan trọng", "top 5 today".
version: 1.0.0
last_updated: 2026-05-11
---

# task-prioritizer

"32 open tasks, 30 min scrolling" → "top 5 to focus today."

**Verbs** (`.claude/SHORTCUTS.generated.md`): `task +get-my-tasks`,
`task +get-related-tasks`, `okr +progress-list`, `okr +progress-get`,
`im +messages-search`. Atomic shapes + token flags: `.claude/RECIPES.md`.

## Workflow

1. **Read memory.** `do-not.md`, `team-roster.md` (sếp-assigned
   patterns), `project-map.md` + `velocity-history.md` (pm only).
2. **Fetch open tasks.** `task +get-my-tasks --complete=false
   --page-limit 20 --jq '.data.items[]|{summary,due:.due_at,guid,url}'`
   (`data` is a dict — tasks in `.data.items[]`; never combine `--jq`
   with `--format`). The CLI item carries only `summary/due_at/guid/
   created_at/url` — NO `followers`/`creator`; if a scoring factor
   needs them, pull via atomic `lark-task` skill.
3. **Enrich** (only where signal matters, parallel):
   - OKR link: `okr +progress-get --progress-id <id>` (or
     `okr +progress-list --target-id <kr_id> --target-type key_result`)
     if the task references a KR.
   - Blocking signal: `im +messages-search --query "<task keyword>
     đợi" --page-size 10 --jq '.data'` (peek shape; needs
     `search:message` scope).
4. **Score each task** (priority_score):

   | Factor | Weight | Computation |
   |---|---|---|
   | Deadline proximity | 1.0 | (1 / max(days_until_due, 0.5)) × hardness |
   | deadline_hardness | — | hard=2.0, soft=1.0 (from tag/description) |
   | Blocking N people | 1.0 | followers_waiting×2 + im_mentions×1 |
   | Sếp-assigned | 2.0 | creator in priority-people Tier 1 |
   | Manager-assigned | 1.5 | creator is user's manager (team-roster) |
   | OKR-linked | 1.5 | KR link AND KR at-risk |
   | Quick-win | 0.5 | estimated <30 min |
   | Stale (>7d untouched) | 0.7 | discount factor |
   | Overdue | +5 flat | bonus |

5. **Rank top 5**, format with one-line reason per task + link.
   Other tasks → bucket counts: defer / delegate (suggest assignee) /
   close (>30d, no signal). Offer drill-down.

## Hard rules

1. Read-only by default — don't auto-update task status.
2. Suggest delegate, don't auto-assign (do-not.md §10).
3. Don't auto-close stale tasks — flag for review only.
4. Names not IDs (resolve via `contact +search-user` / atomic `lark-contact`).
5. Honor user time budget — if "chỉ có 4 tiếng", fit suggestions.
6. Don't double-count blocking — IM mention vs follower may be same person.

## Edge cases

- No tasks → "Inbox zero on tasks 🎉" + upcoming deadlines.
- All overdue → "12 task quá hạn — cần triage tổng thể."
- User OOO (working-hours) → "for return-from-leave" framing.

## Memory

Required: `priority-people.md`, `team-roster.md`.
Recommended (pm): `project-map.md`, `velocity-history.md`.

## Cadence

Morning brief: top 5. Mid-day: top 3 not started. EOD: which top 5
done / carried over.
