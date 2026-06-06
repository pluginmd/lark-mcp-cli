<!--
template: project-kickoff
target: docx (v2)
vars:
  - project: project name
  - owner: resolved name (DRI)
  - sponsor: resolved name — exec sponsor
  - stakeholders: list of {name, role}
  - problem: short problem statement (≤2 sentences)
  - goal: outcome statement (≤2 sentences)
  - non_goals: list of strings
  - timeline: list of {milestone, date, owner}
  - success_metrics: list of {metric, target}
  - risks: list of {risk, mitigation}
  - resources: list of {type, link}
-->

# {{project}}

> **DRI:** {{owner}} · **Sponsor:** {{sponsor}} · **Status:** Kickoff

## Problem

{{problem}}

## Goal

{{goal}}

## Non-goals

{{#non_goals}}
- {{.}}
{{/non_goals}}

## Stakeholders

| Name | Role |
|------|------|
{{#stakeholders}}
| {{name}} | {{role}} |
{{/stakeholders}}

## Timeline

| Milestone | Target date | Owner |
|-----------|-------------|-------|
{{#timeline}}
| {{milestone}} | {{date}} | {{owner}} |
{{/timeline}}

## Success metrics

| Metric | Target |
|--------|--------|
{{#success_metrics}}
| {{metric}} | {{target}} |
{{/success_metrics}}

## Risks & mitigations

{{#risks}}
- **{{risk}}** → {{mitigation}}
{{/risks}}

## Resources

{{#resources}}
- {{type}}: [{{link}}]({{link}})
{{/resources}}

## Open questions

_Track unresolved decisions here. Move to "Decisions" once settled._

- [ ]

## Decisions

- [ ]
