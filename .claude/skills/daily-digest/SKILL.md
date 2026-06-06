---
name: daily-digest
description: End-of-day digest — today's meetings + completed tasks + inbox highlights. Triggers "tổng kết hôm nay", "end of day digest", "wrap up the day".
version: 1.0.0
last_updated: 2026-05-11
---

# daily-digest

The end-of-day "what happened today" wrap. Short, scannable, useful
before logging off.

## When to activate

- "Tổng kết hôm nay", "wrap up", "EOD digest"
- "Hôm nay tôi đã làm gì"
- "Daily summary"
- "End-of-day report"

For Friday wrap, use `weekly-review` instead.

## Workflow

Verbs (`.claude/SHORTCUTS.generated.md`): `calendar +agenda`
(defaults to today), `task +search` (`--completed/--due`),
`task +get-my-tasks` (`--complete=false/--due-end`). Atomic shapes +
token flags: `.claude/RECIPES.md`.

1. **Range**: today (00:00–now), user's timezone.
2. **Pull** (read-only, output-projected — `--jq` and `--format` are
   mutually exclusive, use `--jq` only):
   - `calendar +agenda --jq '.data[]|{summary,start:.start_time.datetime,end:.end_time.datetime}'`
     (defaults to today; `start_time`/`end_time` are nested
     `{datetime,timezone}` objects).
   - `task +search --completed --due <today-00:00>,<now>
     --jq '.data.items[]|{summary}'` (completed today; `data` is a
     dict — tasks in `.data.items[]`).
   - `task +get-my-tasks --complete=false --due-end <today>
     --jq '.data.items[]|{summary,due:.due_at}'` (today's pending).
3. **Optional mail snapshot** (only if user says "include mail") —
   spawn `lark-mail-triage` for unread count + top 3.
4. **Format**:

```
🌙 DIGEST · <weekday>, <date>

📅 Meetings (<count>)
  • <time> <title> · <duration> · <note if any>
  ...

✅ Done today (<count>)
  • <task title>
  ...

⏳ Still open (<count>) — due today
  • <task title>
  ...

📧 Inbox (optional)
  Unread: <N> · Urgent: <X>

💭 Reflection prompts:
  - Top win hôm nay là gì?
  - Có gì kéo dài sang ngày mai?
```

## Hard rules

- **Read-only**. Don't auto-update task status. Don't send anything.
- **Compact** — total output under 25 lines.
- **No OKR** here. That belongs to `weekly-review`.
- **Don't include mail by default** — it adds noise; only include if
  user explicitly asks.
- **Names** for assignees/attendees, not open_ids.

## Useful follow-ups

After producing the digest:
- "Lưu vào nhật ký hôm nay?" → handoff to `lark-doc-author` to
  append to a journal doc.
- "Báo cáo cho sếp?" → handoff to `/draft-mail` with digest as body.
- "Carryover task sang mai?" → handoff to update task due dates
  (confirmation required).
