---
name: calendar-optimizer
description: Audit 30-day meeting patterns — surface decline/merge/async candidates. Triggers "tôi họp quá nhiều", "meeting bloat", "calendar review", "phân tích lịch họp".
version: 1.0.0
last_updated: 2026-05-11
---

# calendar-optimizer

"5 hours/day in meetings" → "audit them, cut 30%." 30-day pattern
analysis → specific decline/merge/async recommendations.

**Verbs** (`.claude/SHORTCUTS.generated.md`): `calendar +agenda`
(`--start/--end/--calendar-id`), `vc +notes` (`--meeting-ids`).
Atomic shapes + token flags: `.claude/RECIPES.md`.

## Workflow

1. **Read memory.** `working-hours.md` (deep-work blocks),
   `priority-people.md` (VIP weight), `delegation-rules.md` (exec).
2. **Fetch 30-day history.** `calendar +agenda --start <t-30d>
   --end <today> --jq '.data[]|{summary,start:.start_time.datetime,end:.end_time.datetime,attendees}'`
   (`--jq` and `--format` are mutually exclusive — `--jq` projects;
   `start_time`/`end_time` are nested `{datetime,timezone}` objects).
3. **Classify each meeting:** recurring vs one-off (repeat pattern);
   user role (chair/required/optional/FYI); decision vs status (check
   minutes for "decided/approved/chốt" via `vc +notes
   --meeting-ids <id>`); outcome quality (action items vs none); duration.
4. **Value score:** chair +3 · VIP attendee +2 · decision keyword in
   minutes +3 · has action items +2 · zero outcome 3+ in a row −3 ·
   replaceable status sync −2 · conflicts deep-work block −2.
5. **Bucket + report:** DECLINE (score <0) · SHORTEN (60-min finishing
   in 30) · CONVERT-TO-ASYNC (status sync) · KEEP. Include numbers
   header (total meeting h / %, recurring count, deep-work h left) +
   projected savings + action checklist.

## Hard rules

1. Never auto-decline — drafts only, user confirms.
2. Decline politely — include "thanks, please share minutes", never ghost.
3. Don't decline VIP meetings even if low value — flag for manual review.
4. Don't suggest cutting 1:1 meetings (managerial).
5. Only audit user's own calendar.

## Edge cases

- <10 meetings/month → "you're under-meeting, skip optimization".
- All 1:1 → "100% your calendar is 1:1 — intentional?".

## Memory

Required: `working-hours.md`, `priority-people.md`.

## Cadence

Monthly (1st of month) or when user feels meeting-overwhelmed.
