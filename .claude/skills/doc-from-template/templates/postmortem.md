<!--
template: postmortem
target: docx (v2)
vars:
  - incident: short title, e.g. "Login outage 2026-05-10"
  - date: ISO date of incident
  - duration: e.g. "47 min"
  - severity: SEV1 | SEV2 | SEV3
  - author: resolved name (postmortem owner)
  - impact: list of {what, magnitude} — what broke and how much
  - timeline: list of {time, event} — UTC or local, consistent
  - root_cause: 1-2 paragraphs
  - contributing_factors: list of strings
  - what_went_well: list of strings
  - what_went_wrong: list of strings
  - actions: list of {what, owner, due, priority}
-->

# Postmortem: {{incident}}

> **Date:** {{date}} · **Duration:** {{duration}} · **Severity:** {{severity}}
> **Author:** {{author}} · **Status:** Draft → Review → Final

> This is a blameless postmortem. Focus on systems, not people.

## Summary

_3-sentence executive summary. What happened, what was the impact, what
have we done._

## Impact

{{#impact}}
- **{{what}}** — {{magnitude}}
{{/impact}}

## Timeline

| Time | Event |
|------|-------|
{{#timeline}}
| {{time}} | {{event}} |
{{/timeline}}

## Root cause

{{root_cause}}

## Contributing factors

{{#contributing_factors}}
- {{.}}
{{/contributing_factors}}

## What went well

{{#what_went_well}}
- ✅ {{.}}
{{/what_went_well}}

## What went wrong

{{#what_went_wrong}}
- ❌ {{.}}
{{/what_went_wrong}}

## Action items

| Priority | Action | Owner | Due |
|----------|--------|-------|-----|
{{#actions}}
| {{priority}} | {{what}} | {{owner}} | {{due}} |
{{/actions}}

## Lessons learned

_What should the org take away that goes beyond the action items?_

-
