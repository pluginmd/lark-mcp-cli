---
name: one-on-one-prep
description: Brief for an upcoming 1:1 — OKR, recent tasks, prior notes, suggested questions. Triggers "1:1 prep với <person>", "brief 1:1".
version: 1.0.0
last_updated: 2026-05-11
---

# one-on-one-prep

"1:1 in 30 min — what do I need to know." For managers prepping to
meet a direct report, or ICs prepping to meet their manager. Regular
meeting (not 1:1) → `meeting-prep`.

**Verbs** (`.claude/SHORTCUTS.generated.md`): `contact +search-user`
(resolve name), `okr +cycle-list` (`--user-id/--time-range`),
`okr +progress-list` (`--target-id/--target-type`), `task +search`
(`--assignee/--completed/--due`), `docs +search` (`--query` only —
no creator/date filter). No `--status`/`--author`/`--creator-ids`/
`--*-after` flags exist on these — use the real ones below. Atomic
shapes + token flags: `.claude/RECIPES.md`.

## Workflow

1. **Resolve person.** `memory/people.md` first, else
   `contact +search-user --query "<name>"
   --jq '.data.users[]|{name:.localized_name,open_id}'` (`data` is a
   dict `{users,has_more}`). No match → ask user to confirm.
2. **Role context** from `team-roster.md`: direct report → angle
   support/unblock/growth; manager → status update / ask for help;
   peer → collab.
3. **Parallel fetch** (via `lark-pm` subagent, scope = 1:1):
   - OKR: `okr +cycle-list --user-id <open_id>
     --time-range <YYYY-MM--YYYY-MM>` then `okr +progress-list
     --target-id <kr_id> --target-type key_result` for KR detail (or
     atomic `lark-okr` skill).
   - Closed tasks (14d): `task +search --assignee <open_id>
     --completed --due <t-14d>,<t>
     --jq '.data.items[]|{summary,completed_at}'` (`data` is a dict —
     tasks in `.data.items[]`).
   - Open/overdue: `task +search --assignee <open_id> --completed=false
     --jq '.data.items[]|{summary,due:.due_at}'` (flag overdue client-side).
   - Recent docs: `docs +search --query "<person name>" --jq '.data'`
     (peek shape; keyword match only, no author/date filter; needs
     `search:docs:read` scope).
4. **Prior 1:1 notes.** `docs +search --query "1:1 <user> <person>"`,
   take latest, extract "Action items" + "Carry over".
5. **Synthesize signals:** KR stalled >2 weeks? task blocked >7d?
   last doc update? prior 1:1 action items done?
6. **Format brief** (≤40 lines): ROLE · OKR (current period) · Recent
   delivery (done/blocked/docs) · From last 1:1 (open action items) ·
   Questions to bring (3, each anchored to a signal).
7. **Offer:** "Tạo doc 1:1 từ template?" → `doc-from-template`
   (`one-on-one`) with context.

## Hard rules

1. Read-only — no task create, no OKR update, no reminders to target.
2. No surveillance tone — "context"/"history", not "monitoring"/"audit".
3. No OKR/task setup → output brief with placeholders, don't fail.
4. Target is manager/peer (not report) → list KR names only, no
   detail unless public.
5. Compact — ≤40 lines.
6. Questions must be actionable — anchored to a real signal.

## Memory

Required: `people.md` (name resolution).
Recommended: `team-roster.md` (role + relationship).
