---
name: inbox-zero
description: Full mail-triage workflow targeting an empty inbox. Triggers "clear inbox", "inbox zero", "xử lý hết mail tồn đọng".
version: 1.0.0
last_updated: 2026-05-11
---

# inbox-zero

Pragmatic workflow for processing a mail backlog down to zero unread.
Quick triage only (not full clear) → `lark-mail-triage` subagent or
`/inbox`.

**Verbs** (`.claude/SHORTCUTS.generated.md`): `mail +triage` is the
list/search verb (date/from/subject/message_id; `--query` full-text,
`--filter` exact-match JSON; prints a dense table natively — no
`--format` needed). Also `mail +messages` (read bodies by IDs),
`mail +draft-create`, `mail +reply`, `mail +send`. There is NO
`+archive` verb. Folder discovery → atomic `lark-mail` skill. Atomic
shapes + token flags: `.claude/RECIPES.md`.

## Workflow

1. **Survey.** `mail +triage --max <n> --filter '{"folder":"INBOX"}'`
   to scan unread; or atomic `lark-mail` skill (`is:unread` query).
   Folders → Lark UI to find names, pass as `--filter` JSON.
2. **Bulk-archive noise.** Identify newsletters/automated/no-action.
   Show count + sample, ask "archive these <N>? (y/n)". On yes →
   atomic `lark-mail` skill (no `+archive` verb in raw CLI; label/move
   via the atomic skill).
3. **Classify remaining** via `lark-mail-triage` subagent: Urgent
   (today) / Important (this week) / FYI (read only).
4. **Process Urgent** — draft reply per item via subagent, user
   confirms each send.
5. **Defer Important** — offer `lark-task` items "Reply to <sender>
   re: <subject>" + due date.
6. **Skim FYI** — collapse to one-line summary, mark read.
7. **Report:** started / archived / replied / tasked / read / remaining.

## Hard rules

- Confirmation gate at every batch action (archive, reply, task) —
  never bulk-mutate without echo + confirm.
- Resolve senders via `lark-contact` before showing names.
- Archive in chunks ≤20 with confirmation between batches.
- Stop on auth error — surface `lark-cli auth login`, exit cleanly.
- Don't classify VIP mail as noise — sender in contact-frequency
  top-10 (30d) → bump to Important minimum.

## Output

End state: one short report block. Not a wall of mail.
