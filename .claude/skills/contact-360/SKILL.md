---
name: contact-360
description: 360° relationship brief before a meeting/call — aggregates IM, mail, meetings, docs, tasks for one person. Triggers "tôi sắp gặp <name>", "context với <name>", "contact 360".
version: 1.0.0
last_updated: 2026-05-11
---

# contact-360

"I'm about to call them in 10 min — who are they to me?" Aggregates
every touchpoint with a person into one brief.

**Verbs** (`.claude/SHORTCUTS.generated.md`): `contact +search-user`
(resolve name), `contact +get-user` (details by `--user-id`),
`im +messages-search`, `calendar +agenda`, `vc +search`,
`docs +search`, `task +search`, `base +record-search`. Mail has no
search verb in raw CLI — use atomic `lark-mail` skill. Atomic command
shapes + token flags: `.claude/RECIPES.md`.

## Workflow

1. **Resolve identity.** `contact +search-user --query "<name>"
   --jq '.data.users[]|{name:.localized_name,open_id,email}'` → pick
   open_id (`data` is a dict `{users,has_more}`). Then
   `contact +get-user --user-id <open_id> --jq '.data.user'` for
   dept/title/manager.
2. **Read memory.** `people.md` (personal notes), `priority-people.md`
   (VIP tier), `team-roster.md`, `accounts.md` (sales).
3. **Parallel fetch touchpoints** — all output-projected:
   - IM: `im +messages-search --sender <open_id> --start <t-90d>
     --end <t> --page-size 50 --jq '.data'` (peek shape — needs
     `search:message` scope; CLI has no recipient-direction filter,
     sender-only; `--start/--end` need a timezone offset).
   - Mail: via atomic `lark-mail` skill (no raw mail search verb).
   - Meetings: `vc +search --participant-ids <open_id> --start <t-90d>
     --end <t> --jq '.data.items[]|{id,info:.display_info}'` (`data`
     is a dict — meetings in `.data.items[]`). Calendar `+agenda` has
     NO attendee filter — skip or scan agenda titles only.
   - Docs: `docs +search --query "<name>" --jq '.data'` (peek shape;
     needs `search:docs:read` scope; keyword match on name only).
   - Tasks: `task +search --follower <open_id> --page-limit 5
     --jq '.data.items[]|{summary,due:.due_at}'` (`data` is a dict —
     tasks in `.data.items[]`; also try `--assignee`).
   - CRM (sales): `base +record-search --base-token <t> --table-id
     <accounts> --json '{"keyword":"<name>","search_fields":["Owner"]}'`.
4. **Synthesize.** Last interaction (most-recent IM/mail/meeting);
   recurring themes (top 3 keywords); open threads (meetings w/o
   follow-up, unanswered mail); past decisions (keyword match);
   outstanding commitments (tasks where person is follower);
   keyword count of "urgent/escalate/critical" (count only — no labels).

## Output

```
👤 CONTACT 360 — <Name> · <Title> · <Dept> · Manager: <Manager>
📅 LAST INTERACTION — <date>: <type> · "<topic>"
🔄 RECURRING THEMES (90d) — <theme>: <N> · ...
⏳ OPEN THREADS — unanswered mail / meeting w/o action / stale task
✅ PAST DECISIONS (90d) — <date>: <summary> (source)
📌 COMMITMENTS — you owe / they owe
🧠 SUGGESTED OPENERS (3) — each anchored to a real thread
⚠️ NOTES FROM people.md — <if exists>
Action sau gặp: update people.md / tạo task follow-up / log decision
```

## Hard rules

1. Read-only against Lark. Only `memory/people.md` may be appended
   (after explicit confirm) — not a Lark mutation.
2. Resolve names, never raw open_ids in output.
3. Truncate per source to top 5 — full brief <30 lines.
4. 0 IM + 0 mail + 0 calendar in 90d → warn "no recent touchpoints".
5. No psychoanalysis — count keyword occurrences, no sentiment labels.
6. CRM lookup only for sales role.
7. Suggested openers must reference actual data — never generic.

## Edge cases

- Internal but new (<30d) → "first-time meeting — read team-roster.md".
- External (no open_id) → skip IM, use mail + calendar + CRM.
- Name collision (`+search-user` >1 hit) → prompt user to pick.
- User asks about self → redirect to `morning-brief` / `weekly-review`.

## Memory

Required: `do-not.md`. Recommended: `people.md`, `priority-people.md`,
`team-roster.md`, `accounts.md` (sales).

## Integration

On-demand or auto from `meeting-prep` (1:1 / external attendee) and
`deal-update` (pre-call). After meeting, offers to append a 1-2 line
note to `memory/people.md` — closes the "remember context" loop.
