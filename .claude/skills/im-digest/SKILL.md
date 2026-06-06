---
name: im-digest
description: Triage IM groups — classify N latest messages per group into cần action / cần biết / bỏ qua. Triggers "47 group có gì", "im digest", "tóm tắt chat".
version: 1.0.0
last_updated: 2026-05-11
---

# im-digest

"47 groups, 12 mentions — what do I actually care about." 90 min of
scrolling → one 15-line summary. Mail-style triage → `inbox-zero`.
Real-time event-driven → atomic `lark-event` skill.

**Verbs** (`.claude/SHORTCUTS.generated.md`): `im +chat-search`
(`--query/--page-size/--sort-by` — NO unread filter; use
`--sort-by update_time_desc` to surface active chats, or atomic
`lark-im` skill for the chat list), `im +chat-messages-list`
(`--chat-id/--start/--end/--page-size`), `im +messages-search`
(`--is-at-me` for @me). Atomic shapes + token flags: `.claude/RECIPES.md`.

## Workflow

1. **Read memory.** `do-not.md`, `priority-people.md`,
   `priority-topics.md`, `team-roster.md` (decision vs social groups).
2. **Inventory active groups.** `im +chat-search
   --sort-by update_time_desc --page-size 50
   --jq '.data.chats[]|{chat_id,name}'` (`data` is a dict — chats in
   `.data.chats[]`), or atomic `lark-im` chat list. For @me items
   directly: `im +messages-search --is-at-me --start <last-check>
   --jq '.data'` (peek shape; `--start` needs a timezone offset;
   needs `search:message` scope).
3. **Per group, fetch recent** (parallel batch ~20):
   `im +chat-messages-list --chat-id <id> --start <last-check>
   --page-size 50 --jq '.data'` (peek shape).
4. **Score each group:** VIP sender +3 · @mention of user +5 ·
   decision keyword (chốt/approve/decided/OK đi/deploy) +3 · priority
   topic +2 · volume >50 msgs −1 · social group −2.
5. **Classify 3 buckets:** 🔴 CẦN HÀNH ĐỘNG (score ≥5) · 🟡 CẦN BIẾT
   (2-4) · ⚪ CÓ THỂ BỎ QUA (<2).
6. **Summarize:** action = 1 line context + "needs <verb>"; FYI =
   1 line digest; skip = count only.
7. **Resolve sender names** via `lark-contact` (no raw open_ids).

## Output

```
💬 IM DIGEST — <date> · last check <time>
🔴 CẦN HÀNH ĐỘNG (<n>) — #<group>: <sender> hỏi <topic> · cần <verb> trước <deadline>
🟡 CẦN BIẾT (<n>) — #<group>: <1-line summary>
⚪ CÓ THỂ BỎ QUA (<n>) — <social>/<resolved>/<noise> counts
Auto-reply 'noted' cho FYI mention? (y/n)
```

## Hard rules

1. Read-only by default; mark-as-read only with explicit confirm.
2. No auto-reply unless user confirms specific message (do-not.md §7).
3. Resolve names — "CEO Hùng", not open_id.
4. Truncate per-group preview to 1 line, max 50 chars.
5. Skip private 1:1 chats in bucket counts (→ `inbox-zero` territory).
6. View-only groups → metadata + summarized topic, no content dump.

## Edge cases

- First run (no last-check) → 24h window.
- 100+ groups → top 30 by recent activity, note truncation.
- No priority signal → "có thể bỏ qua" bucket.
- Active fire (score >10 OR `red-flags.md` keyword) → bypass digest,
  surface immediately.

## Memory

Required: `priority-people.md`, `priority-topics.md`, `do-not.md`.
Recommended: `team-roster.md`, `red-flags.md`.

## Cadence

Morning: full digest via `morning-brief`. Mid-day: 🔴 delta only.
EOD: folded into `daily-digest`.
