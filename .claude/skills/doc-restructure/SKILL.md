---
name: doc-restructure
description: Audit a Lark Wiki — flag stale/orphan/duplicate pages and propose archive/merge/re-parent. Triggers "wiki bị bừa", "restructure wiki", "wiki cleanup".
version: 1.0.0
last_updated: 2026-05-11
---

# doc-restructure

Read-only audit of a messy wiki — proposes archive/merge/re-parent
batches. "5000 pages, 60% outdated, no one knows what to keep."

**Verbs** (`.claude/SHORTCUTS.generated.md`): raw `wiki` exposes only
`+node-create`, `+move`, `+delete-space`. Node listing + content
fetch → atomic `lark-wiki` skill. `docs +search --filter` (wiki type
via JSON filter), `docs +fetch --api-version v2 --doc <url>` for
content, `contact +get-user` for owner resolution. Atomic shapes +
token flags: `.claude/RECIPES.md`.

## Workflow

1. **Read memory.** `taxonomy.md`, `quality-rules.md` (staleness
   threshold, owner rules), `INDEX.md`, `team-roster.md`.
2. **Inventory space.** Atomic `lark-wiki` skill: space-list →
   node-list recursive. Build parent→children tree with depth + metadata.
3. **Per-node enrich** (parallel batch ~20). Atomic `lark-wiki` node
   metadata gives title/owner/last_modified. Use `docs +fetch
   --api-version v2 --doc <url> --limit 1` only when content_length /
   link_count is needed (no metadata-only mode in raw CLI — keep reads minimal).
4. **Detect issues:**

   | Rule | Trigger | Severity |
   |---|---|---|
   | Stale | last_modified >180d | warn |
   | Very stale | >365d | high |
   | Orphan owner | owner not in team-roster | high |
   | Broken link | links to deleted doc | high |
   | Duplicate title | exact match >1 | warn |
   | Near-duplicate | title similarity >90% | info |
   | Empty content | <100 chars | high |
   | Wrong parent | tagged taxonomy ≠ actual parent | warn |
   | Deep nesting | depth >5 | info |
   | Single-page section | parent has 1 child >90d | info |

5. **Bucket into batches:** ARCHIVE (very stale + orphan owner),
   MERGE (near-dup titles → keep newer + redirect), RE-PARENT (wrong
   taxonomy), RE-ASSIGN (orphan owner → team-roster suggestion),
   FLATTEN (single-page section), DELETE-CANDIDATE (empty + >90d).

## Output

Report header (total / healthy / issues / last audit) + per-batch
lists (ARCHIVE, MERGE, RE-PARENT, RE-ASSIGN, FLATTEN, DELETE) +
projected "if you execute all" stats + action checklist with per-batch
approve gates.

## Hard rules

1. Never auto-delete — DELETE candidates flagged for human only.
2. Never auto-archive without explicit batch confirm.
3. Redirect after archive — moved pages leave a redirect stub.
4. Owner re-assign drafts a notification mail — never silent re-assign.
5. Propose-only by default; mutations only on "approve batch N".
6. Archive = move, never hard-delete (preserve history).
7. Skip docs the user can't read — don't infer "should delete".
8. Re-parent targets must exist in taxonomy.md.

## Edge cases

- No taxonomy.md → structural-only issues (broken link, orphan,
  empty), skip re-parent.
- Cross-space links → detect, don't auto-fix.
- Permission denied on subtree → report count, skip detail.
- First audit → present as baseline, suggest piloting one batch first.

## Memory

Required (librarian): `taxonomy.md`, `quality-rules.md`.
Recommended: `INDEX.md`, `team-roster.md`.

## Cadence

Quarterly full audit (mid-Q, not during planning); monthly mini
(ARCHIVE + RE-ASSIGN only); triggered when a team can't find anything.
On approve: `wiki +move` (re-parent) / atomic `lark-wiki` archive,
one item per call (pre-mutate hook fires each). Companion:
`doc-from-template` for new docs — together = author → maintain → archive.
