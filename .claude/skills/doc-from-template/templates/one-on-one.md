<!--
template: one-on-one
target: docx (v2)
vars:
  - manager: resolved name
  - report: resolved name (the direct report)
  - date: ISO date
  - cadence: weekly | biweekly | monthly (default: biweekly)
  - prev_doc_url: optional — link to last 1:1 doc
  - topics: list of strings — agenda items from report
  - manager_topics: list of strings — agenda from manager
  - career_thread: optional string — long-running career topic
-->

# 1:1 — {{manager}} & {{report}}

> **Date:** {{date}} · **Cadence:** {{cadence}}
{{#prev_doc_url}}
> **Previous:** [{{date}}]({{prev_doc_url}})
{{/prev_doc_url}}

## Topics from {{report}}

{{#topics}}
- [ ] {{.}}
{{/topics}}

## Topics from {{manager}}

{{#manager_topics}}
- [ ] {{.}}
{{/manager_topics}}

## Career & growth (ongoing)

{{#career_thread}}
{{.}}
{{/career_thread}}
{{^career_thread}}
_Nothing carried over. Add long-running development items here._
{{/career_thread}}

## Notes from this session

_Fill during the conversation. Use bullets, not paragraphs._

-

## Action items

| Owner | Action | Due |
|-------|--------|-----|
|       |        |     |

## Next 1:1

- **Date:**
- **Topics to carry over:**
