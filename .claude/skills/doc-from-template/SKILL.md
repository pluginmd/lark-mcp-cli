---
name: doc-from-template
description: Author a Lark doc/wiki/sheet by filling a named template. Triggers "tạo doc theo template", "weekly report doc", "1-on-1 doc", "meeting notes template".
version: 1.0.0
last_updated: 2026-05-11
---

# doc-from-template

Produce structured Lark docs/wiki/sheets without re-thinking the
shape every time. Templates encode the WHAT; user provides the WHY.

## When to activate

- "Tạo weekly report doc"
- "Viết 1-on-1 với <person>"
- "Meeting notes cho <meeting>"
- "Project kickoff doc"
- "Postmortem template"
- Any "tạo doc theo mẫu X"

If the user describes shape inline ("tạo doc gồm A B C"), use
`lark-doc-author` directly without this skill — no template needed.

## Built-in templates

All templates live in `./templates/` as `<purpose>.md` files using
**Mustache-style** placeholders (`{{var}}`, `{{#list}}…{{/list}}`,
`{{^var}}…{{/var}}` for negation). The skill renders to markdown,
then lark-cli converts to DocxXML for v2 docs.

| Template               | File                              | Output target |
| ---------------------- | --------------------------------- | ------------- |
| `weekly-report`        | `templates/weekly-report.md`      | docx          |
| `one-on-one`           | `templates/one-on-one.md`         | docx          |
| `meeting-notes`        | `templates/meeting-notes.md`      | docx          |
| `project-kickoff`      | `templates/project-kickoff.md`    | docx          |
| `postmortem`           | `templates/postmortem.md`         | docx          |
| `decision-log-row`     | `templates/decision-log-row.md`   | sheet append  |
| `wiki-runbook`         | `templates/wiki-runbook.md`       | wiki node     |

Each template's HTML comment header at the top documents required vars
and (where relevant) the sheet schema.

User can also point at a template they've authored: `--template
./path/to/my.md`.

## Workflow

Verbs (`.claude/SHORTCUTS.generated.md`): `docs +create`,
`wiki +node-create`, `sheets +append`. Atomic shapes: `.claude/RECIPES.md`.

1. **Identify template** — infer from request; ambiguous → list + ask.
2. **Collect variables** — ask for each required var with a default.
   Front-loaded request ("weekly report tuần này") → pull from the
   same sources `weekly-review` uses.
3. **Render** locally, show preview (first 500 chars) for confirmation.
4. **Target:** drive doc → `docs +create --api-version v2`; wiki →
   `wiki +node-create`; sheet → `sheets +append`.
5. **Write** via `lark-doc-author` subagent. Return URL + token.
6. **Cross-link** — relates to a meeting/task → offer to attach.

## Hard rules

- **Confirm before write**, always. Show the rendered preview first.
- **Variables are validated**:
  - Dates: ISO format. Reject "Thursday" without resolving to date.
  - People: must be resolved via `lark-contact` first. No raw
    open_ids in template output.
- **No template injection**: if user-provided variable contains
  what looks like template syntax (`{{...}}`, `<title>`, raw HTML),
  escape it.
- **Pass `--api-version v2`** for docs.
- **Default DocxXML** for v2 docs unless user said markdown.

## Adding new templates

Drop a file into `./templates/<name>.md`. Use Mustache-style
placeholders. Document required vars in an HTML comment at the top:

```markdown
<!--
template: my-template
target: docx (v2) | sheet append | wiki node
vars:
  - foo: description
  - bar: list of {a, b}
-->

# Title — {{foo}}

{{#bar}}
- {{a}} → {{b}}
{{/bar}}
```

The skill auto-discovers templates by file name. No registration
needed. The comment header is stripped before rendering.

## Why a composition skill

`lark-doc` (atomic) writes whatever string you give it. This skill
adds: opinionated shapes, variable validation, multi-target output
(doc vs wiki vs sheet), preview-confirm flow, and cross-linking.
