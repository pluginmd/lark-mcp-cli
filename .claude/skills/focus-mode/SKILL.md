---
name: focus-mode
description: Block calendar time + IM DND + notify team. Triggers "focus mode", "block <X> tiếng", "DND", "không làm phiền", "deep work".
version: 1.0.0
last_updated: 2026-05-11
---

# focus-mode

"Leave me alone for <N> hours." Three steps: block calendar, set IM
DND, optionally post to team.

**Verbs** (`.claude/SHORTCUTS.generated.md`): `calendar +create`
(`--summary/--start/--end/--description` — note: CLI has NO
free-busy-status or visibility flag), `im +messages-send`
(`--chat-id/--text`). Atomic shapes + token flags: `.claude/RECIPES.md`.

## Limitation — DND not settable via CLI

`im` has no `+set-status`/`+set-dnd` verb. DND must be set manually in
Lark IM UI (Settings → Notifications → Do Not Disturb). This skill:
creates the calendar block, optionally posts to team, and **suggests**
the user set DND manually.

## Workflow

1. **Parse duration + start** from message: "block 2 tiếng chiều nay"
   → start now+30min, 2h; "DND tới 17h" → start now, end 17:00;
   explicit window → use as-is. Ambiguous → ask.
2. **Parse topic** (optional) → event title "Focus — <topic>", else
   "Focus block".
3. **Echo confirm:** calendar block window + title, optional team
   post, DND reminder. Wait y/n.
4. **Execute** (after confirm):
   ```
   calendar +create --summary "Focus — <topic>" --start "<t>" --end "<t>"
   # optional:
   im +messages-send --chat-id <team-group> --text "🎧 Focus mode tới <end>. Urgent thì DM."
   ```
5. **Output:** block created until <end> + calendar URL + team
   notified/skipped + reminder to set DND in Lark UI.

## Hard rules

1. Confirmation gate — echo intent + wait y/n, especially the team post.
2. Don't auto-decline existing meetings in the focus window — surface
   conflict, let user pick decline/skip/re-block.
3. DND isn't silent — IM still receives, just no notify; user should re-check.
4. Don't extend silently — window ends, focus ends. Re-invoke to extend.
5. Don't notify group if user is solo — check `memory/team-roster.md`.

## Edge cases

- Mid-meeting trigger → start = meeting end + 5min, not now.
- Across midnight → split into 2 events or multi-day.
- Weekend/off-hours (`working-hours.md`) → warn "ngoài giờ làm, vẫn block?".

## Memory

Optional: `working-hours.md` (off-hours warning), `team-roster.md`
(which group to notify). Missing → skip optional steps.
