---
name: incident-retro
description: Build a blameless postmortem from an on-call IM timeline. Triggers "incident retro", "postmortem cho SEV<X>", "viết postmortem".
version: 1.0.0
last_updated: 2026-05-11
---

# incident-retro

"Fire is out, now we learn." Converts on-call IM chaos + your memory
into a structured postmortem doc. Sprint-level retro → `sprint-retro`.

**Verbs** (`.claude/SHORTCUTS.generated.md`): `im +messages-search`,
`im +chat-messages-list`, `docs +search`, `task +create`. Atomic
shapes + token flags: `.claude/RECIPES.md`.

## Workflow

1. **Identify incident.** User gives date + description, OR ref an
   incident ID in `memory/incident-log.md` / incident Base, OR search:
   `im +messages-search --query "SEV1 OR SEV2 OR outage OR đứt"
   --chat-id <oncall-group> --start <t> --end <t> --jq '.data'`
   (peek `.data` shape — varies per verb; needs `search:message` scope).
2. **Gather timeline.** `im +chat-messages-list --chat-id <group>
   --start <t> --end <t> --jq '.data'` (peek shape first).
   Filter to key actors (oncall lead, eng lead, SRE). Map events:
   detection → page → triage → mitigation → resolution.
3. **Cross-reference.** Dashboards/logs URLs mentioned; PRs deployed
   pre-incident (`docs +search --query "<service>" --jq '.data'` —
   peek shape; needs `search:docs:read` scope); support tickets (if
   helpdesk skill); status-page updates.
4. **Synthesize structure.** Summary (3 sentences); Impact (users,
   duration, $); Timeline (chronological + timestamps); Root cause
   (best understanding, don't fake certainty); Contributing factors;
   What went well / wrong; Action items (owner + by-when + priority).
5. **Render doc.** Spawn `lark-doc-author` with `doc-from-template`
   template `postmortem.md`. Save to eng wiki "Postmortems" folder.
6. **Create follow-up tasks** (after confirm). Per action item:
   `task +create --summary "<action>" --assignee <open_id>
   --due <date>` in tasklist "Postmortem actions" (priority → in
   `--description` or `--data` JSON; no `--priority` flag).

## Hard rules

1. Blameless tone — focus systems, not individuals.
2. Don't quote IM verbatim — paraphrase (do-not.md §20; chatters
   didn't opt into being quoted).
3. Anonymize non-DRI mentions ("oncall lead", not names) unless DRI
   role must be tracked OR user asks + recipient consents.
4. Action items always have owner + due (default due = end-of-sprint
   per `velocity-history.md`).
5. Severity from incident convention (incident Base schema), not opinion.
6. Don't speculate on "what users felt" — quote support ticket or
   leave empty.

## Output

Header (title / severity / duration / date) + sections: SUMMARY,
TIMELINE, ROOT CAUSE, CONTRIBUTING FACTORS, WHAT WENT WELL, WHAT WENT
WRONG, ACTION ITEMS ([P1] action — owner — due), LESSONS LEARNED.
Ends with confirm gate: save to wiki / create tasks / update incident-log.

## Edge cases

- Very recent (<2h since resolution) → partial draft, flag "complete
  after war-room ends".
- Multi-day → split timeline by day.
- External-vendor cause → include vendor contact + their postmortem ref.
- Security incident → mark doc Restricted (`policies.md` §1.4),
  don't share details publicly.

## Memory

Required: `INDEX.md` (incident Base + wiki space location).
Recommended: `team-roster.md`, `incident-log.md`, `policies.md`.

Template: `doc-from-template` → `postmortem.md` (already in
`.claude/skills/doc-from-template/templates/`).
