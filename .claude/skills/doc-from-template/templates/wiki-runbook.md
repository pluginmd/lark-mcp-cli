<!--
template: wiki-runbook
target: wiki node (create or update under a space)
vars:
  - service: service / system name
  - owner_team: team name
  - owner_oncall: oncall rotation handle (e.g. "@platform-oncall")
  - description: 1-paragraph service summary
  - architecture_url: optional link to architecture doc
  - dashboards: list of {name, url}
  - alerts: list of {name, severity, runbook_section}
  - escalation: list of {level, who, when}
  - common_issues: list of {symptom, diagnosis, fix}
  - dependencies: list of {service, type, contact}
-->

# Runbook: {{service}}

> **Owner team:** {{owner_team}} · **On-call:** {{owner_oncall}}
> _Living document. Keep it accurate or delete it._

## What it is

{{description}}

{{#architecture_url}}
**Architecture:** [diagram & design]({{architecture_url}})
{{/architecture_url}}

## Where to look

### Dashboards

{{#dashboards}}
- [{{name}}]({{url}})
{{/dashboards}}

### Alerts

| Alert | Severity | Goes to |
|-------|----------|---------|
{{#alerts}}
| {{name}} | {{severity}} | §{{runbook_section}} |
{{/alerts}}

## Escalation path

{{#escalation}}
- **{{level}}** — {{who}} (when: {{when}})
{{/escalation}}

## Common issues

{{#common_issues}}
### {{symptom}}

**Diagnose:** {{diagnosis}}

**Fix:** {{fix}}

---
{{/common_issues}}

## Dependencies

| Service | Type | Contact |
|---------|------|---------|
{{#dependencies}}
| {{service}} | {{type}} | {{contact}} |
{{/dependencies}}

## Maintenance log

_Track non-incident changes here. Append, never delete._

| Date | Who | What | Why |
|------|-----|------|-----|
|      |     |      |     |
