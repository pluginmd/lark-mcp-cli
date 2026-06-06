---
name: deal-update
description: Post-call CRM update — pull minutes, extract pain/budget/timeline, update Base record, draft follow-up mail (drafts only). Triggers "cập nhật deal sau gọi", "deal update".
version: 1.0.0
last_updated: 2026-05-11
---

# deal-update

"I just got off a call, update everything." Saves ~15 min of manual
CRM updating per call.

**Verbs** (`.claude/SHORTCUTS.generated.md`): `vc +search`,
`vc +notes`, `minutes +search`, `base +record-search`,
`base +record-get`, `base +record-batch-update`. AI minutes artifacts
→ atomic `lark-minutes` skill. Atomic shapes + token flags:
`.claude/RECIPES.md`.

## Workflow

1. **Identify meeting.** `vc +search --participant-ids <open_id>
   --start <t-7d> --end <t> --jq '.data.items[]|{id,info:.display_info}'`
   → most recent (`data` is a dict — meetings in `.data.items[]`,
   meeting id is `.id`). Or user gives meeting ID directly.
2. **Fetch minutes.** `minutes +search --query "<customer>"
   --start <t-7d> --jq '.data'` → peek shape, then pull minute-token
   (needs `minutes:minutes.search:read` scope). Then atomic
   `lark-minutes` skill for summary/todos (transcript only if user
   wants deep). Or `vc +notes --minute-tokens <T>`.
3. **Extract fields** from minutes: pain points, competitor mentions,
   budget, timeline, decision-maker/champion, next step, risks/objections.
4. **Resolve CRM record.** `base +record-search --base-token <t>
   --table-id <CRM> --json '{"keyword":"<customer>","search_fields":["Company"]}'`
   → record ID. `base +record-get` current state to diff.
5. **Echo proposed update** — show stage change + reason, last touch,
   next step, notes append block. Wait y/n.
6. **On confirm:** `base +record-batch-update --base-token <t>
   --table-id <CRM> --json '{"record_id_list":["<id>"],"patch":{"Stage":"<new>","Last touch":"<today>","Next step":"<text>","Notes":"<existing>\n\n<new entry>"}}'`.
7. **Offer follow-up mail draft.** On yes → spawn `lark-mail-triage`
   with context, drafts mail (NEVER sends — do-not.md §7). Show draft
   ID, user reviews + sends in Lark Mail UI.
8. **Update `memory/accounts.md`** entry (propose, then edit on confirm).

## Output

Header (customer · call date) + Extracted block (pain / competitor /
budget / timeline / next) + CRM update preview (per-field diff) + mail
draft preview (to / subject / first 200 chars) + confirm checklist
(update CRM / create task / save draft / update accounts.md).

## Hard rules

1. NEVER auto-send mail — drafts only (do-not.md §7).
2. Echo before every update — confirm each action.
3. Quote source — each extracted field refs a minutes timestamp/speaker.
4. Don't fabricate — field not in minutes → "không discussed".
5. Resolve champion via `memory/people.md` → `lark-contact`.
6. Update memory last, after CRM record updated successfully.

## Edge cases

- No minutes (VC not transcribed yet) → wait + retry, or user pastes.
- Multiple CRM records same company → ask user to pick.
- No CRM record (first call) → offer `+record-upsert` instead, confirm.
- Minutes <5 min → warn "ngắn, có thể thiếu signal — vẫn extract?".

## Memory

Required: `accounts.md` (tone + champion resolution).
Recommended: `playbooks.md`, `competitors.md`.
