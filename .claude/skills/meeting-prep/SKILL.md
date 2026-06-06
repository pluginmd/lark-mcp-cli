---
name: meeting-prep
description: Two-phase meeting flow — before (pull context) and after (extract action items into tasks). Triggers "chuẩn bị họp X", "action items từ meeting".
version: 1.0.0
last_updated: 2026-05-11
---

# meeting-prep

End-to-end meeting workflow, two phases. Quick agenda lookup → `/agenda`.

**Verbs** (`.claude/SHORTCUTS.generated.md`): `calendar +agenda` (only
`--start/--end/--calendar-id` — NO keyword filter; scan titles
client-side), `docs +fetch`, `contact +search-user` (resolve names),
`im +messages-search`, `minutes +search` (`--query/--start/--end`,
no `--recent`), `vc +notes`, `task +create`. Mail has no search verb
— use atomic `lark-mail` skill for prior mail threads. Atomic shapes +
token flags: `.claude/RECIPES.md`.

## Phase 1 — BEFORE

1. **Identify meeting.** `calendar +agenda --start <t> --end <t>
   --jq '.data[]|{summary,start:.start_time.datetime,end:.end_time.datetime,event_id}'`,
   match title client-side. Explicit ID → use directly. (`data` is a
   list; `start_time` is a nested `{datetime,timezone}` object; no
   `attendees`/`location` field — see RECIPES.)
2. **Pull materials** via `lark-meeting-scout` subagent: linked
   pre-read doc, attendee list (resolve names via `contact +search-user`),
   prior mail/IM thread with same attendees + topic (last 14d).
3. **Auto-trigger `contact-360`** for high-stakes attendees: 1:1
   (exactly 2 attendees), external (domain ≠ org), Tier 1 from
   `priority-people.md`. Append one line per: "📌 Relationship:
   <last interaction> · <open thread count>".
4. **Compose brief** (<200 words): title/time/location · attendees ·
   pre-read summary · prior context · suggested questions.
5. **Offer:** "Lưu brief vào lark-doc?"

## Phase 2 — AFTER

1. **Find minutes.** `minutes +search --query "<title>" --start <t-1d>
   --jq '.data'` (peek shape; needs `minutes:minutes.search:read` scope).
2. **Fetch AI artifacts** (no `+summary`/`+todos` verbs) — atomic
   `lark-minutes` skill (summary/todos/chapters actions) or
   `vc +notes --minute-tokens <T>`.
3. **Extract action items** (who/what/by-when). Resolve assignees via
   `contact +search-user`. Default due = end of week if unspecified.
4. **Report:** title/date/duration · summary (2-3 sentences) · action
   items ([assignee] action · due) · key decisions.
5. **Offer:** "Tạo task cho từng action item?" On yes → `task +create`
   batch with confirmation.
6. **Auto-trigger `decision-logger`** if summary/transcript has
   keyword `chốt|decided|approved|"OK đi"|"we agreed"|"we'll go with"`
   → echo decisions + offer to log to Base. On confirm hand off
   `decision-logger` Mode A with meeting URL as source.
7. **Auto-update `people.md`** for 1:1 / external — offer 1-line note
   with `last_met: <date>`.

## Hard rules

- Phase 1 stays compact — user is walking into a meeting.
- Phase 2 always names assignees; unnamed → "unassigned", ask user.
- Confirmation gate before creating tasks (show all, user picks).
- No fabricated context — no prior thread found → say so.

## Cross-skill integration

Hub skill. Calls `contact-360` (Phase 1 attendee context),
`decision-logger` (Phase 2 decisions), `lark-meeting-scout` agent
(context gathering), atomic `lark-task` (batch create). Canonical
"one prompt → multi-skill flow" pattern.
