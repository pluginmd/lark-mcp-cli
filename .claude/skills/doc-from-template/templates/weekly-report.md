<!--
template: weekly-report
target: docx (v2)
vars:
  - week: ISO week number, e.g. "2026-W19"
  - date_range: "2026-05-11 → 2026-05-15"
  - themes: list of {title, note} — top themes of the week
  - allocation: {meetings_hours, focus_hours, meetings_percent}
  - top_meetings: list of {title, hours} — top 3 by time spent
  - tasks_done: count
  - tasks_open: count
  - tasks_overdue: count
  - highlights: list of strings — key deliveries
  - okrs: list of {kr, start_pct, now_pct, delta_pct}
  - wins: list of strings
  - friction: list of strings
  - next_week_focus: string
-->

# Weekly Report — Tuần {{week}}

> **Khoảng thời gian:** {{date_range}}

## Themes

{{#themes}}
- **{{title}}** — {{note}}
{{/themes}}

## Time Allocation

| Phân loại | Số giờ | % |
|-----------|--------|---|
| Meetings | {{allocation.meetings_hours}}h | {{allocation.meetings_percent}}% |
| Focus time | {{allocation.focus_hours}}h | — |

**Top 3 meetings by time:**

{{#top_meetings}}
1. {{title}} — {{hours}}h
{{/top_meetings}}

## Execution

- **Tasks done:** {{tasks_done}} ✅
- **Tasks open:** {{tasks_open}} (trong đó {{tasks_overdue}} quá hạn 🔴)

**Highlights:**

{{#highlights}}
- {{.}}
{{/highlights}}

## OKR Progress

| KR | Đầu tuần | Hiện tại | Δ |
|----|----------|----------|---|
{{#okrs}}
| {{kr}} | {{start_pct}}% | {{now_pct}}% | +{{delta_pct}}% |
{{/okrs}}

## Reflection

**Wins**

{{#wins}}
- {{.}}
{{/wins}}

**Friction**

{{#friction}}
- {{.}}
{{/friction}}

**Next week focus:** {{next_week_focus}}
