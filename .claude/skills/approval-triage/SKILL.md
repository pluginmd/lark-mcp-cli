---
name: approval-triage
description: Triage pending Lark Approval queue — recommend approve/reject/ask-more per item with policy citation. Triggers "approval pending", "duyệt expense", "queue duyệt", "có gì cần duyệt".
version: 1.0.0
last_updated: 2026-05-11
---

# approval-triage

"8 approvals pending, 5 min each" → "1 batch review, 30 seconds."
Agent reads; user decides.

**Verbs**: raw `approval` exposes only `instances`/`tasks` subgroups —
NO `+list`/`+get`/`+approve`/`+reject` shortcut. Use the atomic
`lark-approval` skill (wraps the API). Workflow below is conceptual
flow; call the atomic skill, not raw bash. `docs +search` (for policy
docs) takes `--query` + `--filter` JSON, no `--type` flag; needs
`search:docs:read` scope. Atomic shapes + token flags: `.claude/RECIPES.md`.

## Workflow

1. **Read memory.** `do-not.md`, `policies.md` (thresholds/categories),
   `team-roster.md` (requester level + manager line),
   `delegation-rules.md` (exec — financial thresholds).
2. **List pending** via atomic `lark-approval` skill: `list-pending
   --assignee me`.
3. **Per approval, fetch detail** (parallel ~5) via atomic skill:
   `get-instance <id>` → requester, type, amount, reason, attachments.
4. **Cross-check policy.** Look up rule in `policies.md` (or
   `docs +search --query "<policy keyword>" --jq '.data'` to find the
   doc — peek the shape, don't guess nested paths); compare requester
   pattern + amount vs `delegation-rules.md` threshold.
5. **Recommend per item:**
   - **✅ APPROVE** — within policy, clean history, valid category,
     below auto-decide threshold.
   - **⚠️ NEEDS CHECK** — 10-20% over threshold, unusual pattern,
     missing context.
   - **❌ REJECT** — clear policy violation, over hard limit, abuse pattern.
6. **Format batch report** — APPROVE list (batch-approvable) / NEEDS
   CHECK (concern + link) / REJECT (policy reason) + action checklist.
7. **On confirm "approve lô N"** — iterate approved batch via atomic
   skill `approve <id> --comment "..."` per item. Each triggers
   `pre-mutate.sh` → audit log.

## Hard rules

1. Never auto-approve without explicit confirm (do-not.md §4).
2. Show rationale for every recommendation.
3. Batch ≤10 per confirm — split if more.
4. Each rationale cites a §N from `policies.md` (or "policy unclear").
5. Reject comment always includes reason + how to fix.
6. Audit every approve/reject (automatic via hook).
7. Refuse if user lacks approval authority (verify via `team-roster.md`).

## Edge cases

- No policy in memory → "needs check" for all amount-based decisions.
- Approval expired while reviewing → surface + skip.
- Self-approval → refuse (policy violation).
- Threshold > user authority → route to manager (`team-roster.md`).

## Memory

Required: `policies.md`, `team-roster.md`.
Recommended: `delegation-rules.md` (exec), `incident-log.md`.

## Cadence

Morning brief: count + top 3 by amount. On-demand: full triage.
Friday 16h: reminder if backlog >5.
