<!--
template: decision-log-row
target: sheet append (one row)
vars:
  - decision: short title (1 line)
  - date: ISO date
  - by: resolved name (decision-maker)
  - context: 1-2 sentence rationale
  - alternatives: list of strings — options considered
  - why: 1-2 sentence justification
  - reversible: yes | no | partial
  - revisit_by: optional ISO date — review this decision by

Sheet schema expected (8 columns, A1):
  A: date  B: decision  C: by  D: context  E: alternatives
  F: chosen_why  G: reversible  H: revisit_by

The skill renders this template as a single appended row, NOT a doc.
Comma-join lists. Truncate cells to ~300 chars.
-->

| date | decision | by | context | alternatives | chosen_why | reversible | revisit_by |
|------|----------|----|---------|--------------|------------|------------|------------|
| {{date}} | {{decision}} | {{by}} | {{context}} | {{#alternatives}}{{.}}; {{/alternatives}} | {{why}} | {{reversible}} | {{revisit_by}} |
