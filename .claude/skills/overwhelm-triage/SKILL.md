---
name: overwhelm-triage
description: Meta-router for "I'm overwhelmed" — disambiguate between mail/IM/task/meeting pain in one question, then dispatch to the right skill.
version: 1.0.0
last_updated: 2026-05-11
---

# overwhelm-triage

When the user says "tôi quá tải", "burnt out", "overload", or
similar, four different skills compete for the right entry point:

- `inbox-zero` for mail backlog
- `im-digest` for IM group overload
- `task-prioritizer` for task pile
- `calendar-optimizer` for meeting bloat

This skill asks one short question, picks one, and hands off.

## When to activate

- "Tôi quá tải", "overwhelmed", "burnt out"
- "Quá nhiều thứ phải làm", "không biết bắt đầu từ đâu"
- "Help me triage", "tôi cần thoát overload"

If the user already named the surface ("inbox quá nhiều" → use
`inbox-zero` directly), DO NOT activate this skill.

## Workflow

1. **Ask exactly one question**:

```
🔥 Bạn đang quá tải nhất ở đâu? (chọn 1)

  📧 Mail backlog            →  inbox-zero
  💬 IM groups (47+ groups)  →  im-digest
  ✅ Task pile (open tasks)  →  task-prioritizer
  📅 Meeting bloat (5h/day)  →  calendar-optimizer
  🌀 Cả 4 — combo            →  morning-brief (parallel triage)
```

2. **Dispatch** based on the answer. Each option spawns the
   corresponding skill via Task tool with the relevant scope.

3. **Do not aggregate** — let the chosen skill own the screen.
   The user picked one surface; respect that.

## Hard rules

1. **One question only**. No follow-up questions until the
   downstream skill takes over.
2. **No analysis paralysis** — if the user already mentions a
   surface, skip this skill entirely.
3. **Read-only**. This skill never mutates anything; the
   downstream skill handles the actual work.

## Calling pattern

```
1. Detect "I'm overwhelmed" pattern in user input
2. Confirm scope is genuinely ambiguous (else hand off directly)
3. Print the 5-option chooser
4. Wait for user reply
5. Task tool → spawn skill matching the chosen option
6. Exit — let downstream skill drive
```

## Why this exists

`morning-brief` does parallel triage across all four surfaces but
takes >30s for a 15-line wrap. An acutely overwhelmed user wants one
screen, not four — this skill is the "first ask" before a deep dive.

## Cadence

On-demand only. Common right after `morning-brief` when one bucket
dominates and the user reacts "ugh, too much".

## Memory

None required — reads memory only if the downstream skill needs it.
