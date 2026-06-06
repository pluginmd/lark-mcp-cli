<!--
template: meeting-notes
target: docx (v2)
vars:
  - title: meeting title
  - date: ISO date + time
  - duration: e.g. "60 min"
  - location: room name or "Online (Lark VC)"
  - attendees: list of {name, role}
  - facilitator: resolved name
  - notetaker: resolved name (default = facilitator if unspecified)
  - agenda: list of {item, owner, minutes}
  - prereads: list of {title, url} — optional
-->

# {{title}}

> **Date:** {{date}} · **Duration:** {{duration}} · **Location:** {{location}}
> **Facilitator:** {{facilitator}} · **Note-taker:** {{notetaker}}

## Attendees

{{#attendees}}
- {{name}} — {{role}}
{{/attendees}}

{{#prereads}}
## Pre-reads

{{#prereads}}
- [{{title}}]({{url}})
{{/prereads}}
{{/prereads}}

## Agenda

| # | Topic | Owner | Time |
|---|-------|-------|------|
{{#agenda}}
| {{@index1}} | {{item}} | {{owner}} | {{minutes}} min |
{{/agenda}}

---

## Discussion

_Notes captured during meeting. Group by agenda item._

### {{agenda.0.item}}

-

### {{agenda.1.item}}

-

---

## Decisions

- [ ]

## Action items

| Owner | Action | Due |
|-------|--------|-----|
|       |        |     |

## Parking lot

_Topics raised but not resolved. Carry to next meeting._

-
