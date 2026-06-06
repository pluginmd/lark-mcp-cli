---
name: morning-brief
description: First-of-day brief orchestrator. Spawns mail-triage/im-digest/approval-triage/task-prioritizer in parallel, merges ≤15 lines. 5 variants: default,ic,exec,pm,sales. Triggers "morning", "/morning".
version: 1.0.0
last_updated: 2026-05-11
---

# morning-brief (orchestrator)

One message replacing the overnight catch-up. Fans out composition
skills in parallel, merges into one brief. Does **not** call
`lark-cli` directly — each spawned skill owns its verb mapping.

End-of-day → `daily-digest`. Full week → `weekly-review`.

## Variants

| Variant | Audience | Length | Skills spawned |
|---|---|---|---|
| `default` | IC / manager | ≤15 lines | mail-triage, week-planner, im-digest, task-prioritizer |
| `ic` | Engineer/designer/analyst | ≤12 lines | mail-triage (lite), week-planner, im-digest (mention-only), task-prioritizer (top 3 + time-budget) |
| `exec` | C-level | ≤5 lines | + approval-triage |
| `pm` | Project manager | ≤18 lines | + decision-logger (Mon, scan mode) |
| `sales` | Sales | ≤15 lines | + pipeline-review (Mon), client-followup (Fri) |

Variant from arg (`/morning exec`); else infer from active persona; default `default`.

IC variant biases: im-digest = @mention + decision keyword only;
task-prioritizer top 3 + asks focus-hours budget; calendar flags
deep-work conflicts (propose decline); skip approval + decision sweep.

## Workflow

**Phase 1 — setup (sequential).** Read memory (one Read call each,
parallel): `priority-people.md`, `priority-topics.md`,
`working-hours.md`, `do-not.md`. exec also: `red-flags.md`,
`delegation-rules.md`, `tone-preference.md`. pm also: `team-roster.md`,
`velocity-history.md`. sales also: `accounts.md`, `playbooks.md`.
Detect day-of-week + persona.

**Phase 2 — parallel fan-out.** Spawn all sub-tasks in ONE message
with multiple Task calls (never serialize — each has its own context
budget):

| Skill | Prompt | Returns |
|---|---|---|
| mail-triage / inbox-zero | "Overnight unread, filter by priority-people + priority-topics. 5-line summary." | 5-bucket counts + 3 top mails |
| lark-week-planner (agent) | "Today only: agenda + due-today + overdue." | Calendar + tasks |
| im-digest | "Groups unread since last check. 3-bucket." | Action/FYI/skip + top 3 action |
| task-prioritizer | "Top 5 today, all open tasks. Honor priority-people." | Top 5 ranked |
| approval-triage (exec) | "Pending count + top 3 by amount." | Approve/Check/Reject counts |

**Phase 3 — merge + format.** Aggregate into variant output, drop
empty sections, truncate hard.

**Phase 3.5 — telemetry.** Append one line to `.claude/hooks/audit.log`
for `/roi` (plain Bash append, skips mutate hook). Skip if
`LARKSUITE_CLI_HOOKS_BYPASS=1`.

```bash
ts=$(date -u +%Y-%m-%dT%H:%M:%SZ)
cat >> .claude/hooks/audit.log <<JSON
{"ts":"$ts","action":"skill_morning-brief","variant":"<variant>","buckets":{"mail":<n>,"im":<n>,"approval":<n>,"task":<n>},"meetings":<n>,"actor":"skill","status":"completed"}
JSON
```

## Output shapes

`default` — weekday/date + load signal · 📅 HÔM NAY (count, total h) ·
🎯 TOP 5 TASK · 📧 MAIL ƯU TIÊN (vip/total) · 💬 IM ACTION · 🔵 FOCUS
BLOCK · "Đề xuất bắt đầu: <action>".

`exec` — weekday/date · DECISIONS TODAY · ESCALATION · CALENDAR ·
APPROVALS (counts) · HANDLED OVERNIGHT (mail archived / IM noise / task carried).

`pm` — + Team weekly index · standup time · 3 BLOCKER · SPRINT HEALTH
(velocity vs avg, at-risk OKR, decisions to log) · "1:1 today? → /oneonone".

`ic` — + deep-work hours available · meeting/deep-work conflict flags ·
TOP 3 FIT TIME BUDGET (~Xh each) · DEFER count · IM mention-only · FOCUS BLOCK.

`sales` — + pipeline stage count/amount · calls/demos · TOP 3 DEAL ·
SẮP CALL (→ contact-360) · khách mail/chat counts · (Mon) pipeline-review · (Fri) client-followup.

## Hard rules

1. Read-only. No send, no task create, no RSVP.
2. One message, no scroll (~15 lines default, ~5 exec).
3. Parallel fan-out — all Task calls in ONE message.
4. Don't auto-run before working hours — user must ask.
5. Truncate: meeting title >40 chars, mail subject >60 chars.
6. Names not IDs — resolve via `lark-contact` or memory.
7. Drop empty sections (no padding).
8. End with one concrete suggested next step.

## Memory

Required: `priority-people.md`, `priority-topics.md`,
`working-hours.md`, `do-not.md`. Variant extras listed in Phase 1.
Empty memory → still runs, emits one warning line suggesting setup.

## Test scenarios

- Empty inbox + clear calendar → "🌅 Sáng yên ả. Top 1 task: <X>. Đề xuất focus block 90 phút."
- 200 unread + 4 meetings + 12 overdue → triage report + focus suggestion.
- Friday afternoon → redirect to `daily-digest`.
- Weekend → "Cuối tuần. Brief về Monday không?"
