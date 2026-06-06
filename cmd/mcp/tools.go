// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// tool defines one MCP tool exposed by `lark-cli mcp serve`. Each tool
// owns its JSON Schema (so the host model knows what to pass) and a
// builder that translates MCP arguments into a lark-cli argv slice.
//
// We intentionally expose ~15 high-value shortcuts rather than every
// subcommand. Power users can fall through to `lark_api` (generic
// passthrough) or run lark-cli directly outside the MCP session.
type tool struct {
	Name        string
	Description string
	Schema      string // raw JSON Schema as a Go string literal
	Build       func(args map[string]interface{}) ([]string, error)
}

// allTools returns the curated tool catalogue in stable order. Stable
// order matters: Claude Desktop renders tools in the order received.
func allTools() []tool {
	return []tool{
		toolIMSend(),
		toolIMCardSend(),
		toolIMSearch(),
		toolMailSend(),
		toolMailDraftCreate(),
		toolCalendarAgenda(),
		toolCalendarCreate(),
		toolDocCreate(),
		toolDocSearch(),
		toolDocFetch(),
		toolBaseSearch(),
		toolContactSearch(),
		toolTaskMy(),
		toolTaskCreate(),
		toolDriveUpload(),
		toolSheetsRead(),
		toolSheetsAppend(),
		toolVCSearch(),
		toolMinutesSearch(),
		toolOKRCycleList(),
		toolGenericAPI(),
	}
}

// descriptors converts the registry into the over-the-wire MCP shape.
func descriptors(tools []tool) ([]toolDescriptor, error) {
	out := make([]toolDescriptor, 0, len(tools))
	for _, t := range tools {
		out = append(out, toolDescriptor{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: json.RawMessage(t.Schema),
		})
	}
	// Sanity check: ensure schemas are valid JSON so we fail loudly at
	// boot rather than mid-session when a host first asks for the list.
	for _, d := range out {
		var probe interface{}
		if err := json.Unmarshal(d.InputSchema, &probe); err != nil {
			return nil, fmt.Errorf("invalid schema for tool %q: %w", d.Name, err)
		}
	}
	return out, nil
}

// dispatch maps a tools/call request onto the matching builder.
//
// Kept for backwards compatibility with any caller that doesn't need
// the raw execResult. The server.processToolsCall path uses
// dispatchWithRaw so audit can record exit code, byte counts, etc.
func dispatch(ctx context.Context, r *runner, name string, args map[string]interface{}) (toolsCallResult, error) {
	out, _, err := dispatchWithRaw(ctx, r, name, args)
	return out, err
}

// dispatchHookForTests, when non-nil, replaces the production
// dispatchWithRaw body. This lets tests inject a fake subprocess
// runner without refactoring the production type into an interface.
//
// Production code must never assign to this — set/unset is the
// responsibility of test fixtures only. Guarded by go test's
// single-process model + t.Cleanup restoration.
var dispatchHookForTests func(ctx context.Context, r *runner, name string, args map[string]interface{}) (toolsCallResult, execResult, error)

// dispatchWithRaw is dispatch + returns the raw execResult so the
// caller (audit logger) can record exit code, stdout/stderr byte
// counts, etc. The MCP response shape itself is in the first return.
//
// All error paths — Build validation, unknown tool, exec failure,
// shortcut non-zero exit — flow through formatResult so the host gets
// a uniform classified error envelope (see errors.go). Audit always
// sees ExitCode != 0 for failure paths.
func dispatchWithRaw(ctx context.Context, r *runner, name string, args map[string]interface{}) (toolsCallResult, execResult, error) {
	if dispatchHookForTests != nil {
		return dispatchHookForTests(ctx, r, name, args)
	}
	for _, t := range allTools() {
		if t.Name == name {
			argv, err := t.Build(args)
			if err != nil {
				// Build-layer error: synthesize the lark-cli envelope
				// shape so classifyError tags it as "validation". This
				// keeps the wire contract uniform with shortcut-layer
				// validation errors.
				synth := synthErrorEnvelope("validation", err.Error())
				res := execResult{ExitCode: -1, Stderr: synth}
				return formatResult(res), res, nil
			}
			res, err := r.run(ctx, argv)
			if err != nil {
				// Process-level exec failure (binary missing, OS error).
				synth := synthErrorEnvelope("network", fmt.Sprintf("exec failed: %v", err))
				res := execResult{ExitCode: -1, Stderr: synth}
				return formatResult(res), res, nil
			}
			return formatResult(res), res, nil
		}
	}
	// Unknown tool — synthesize a validation-flavoured error so the
	// agent knows to inspect the tool name rather than retry.
	synth := synthErrorEnvelope("validation", fmt.Sprintf("unknown tool: %s; call tools/list to see available tools", name))
	res := execResult{ExitCode: -1, Stderr: synth}
	return formatResult(res), res, nil
}

// synthErrorEnvelope produces a lark-cli-shaped error envelope from an
// MCP-layer error so classifyError() treats it uniformly with shortcut
// errors. Used only for errors that originate inside the bridge (Build
// validation, unknown tool, OS exec failure) — shortcut errors flow
// through as-is.
func synthErrorEnvelope(errType, message string) string {
	env := larkCLIErrorEnvelope{OK: false}
	env.Error.Type = errType
	env.Error.Message = message
	buf, err := json.Marshal(&env)
	if err != nil {
		// Marshal failure on a tiny struct is impossible; return the
		// message verbatim so callers still have something to log.
		return message
	}
	return string(buf)
}

// formatResult shapes a subprocess outcome into an MCP tool result.
// Successful runs return stdout as the primary block; non-zero exits
// run through classifyError() to produce a structured error envelope
// the host model can switch on (error_type + retry_advice) rather than
// grep'ing raw stderr.
//
// Wire format for errors: the content block carries a JSON document
// matching `errorClassification` (see errors.go). Agents parse this to
// decide whether to retry, escalate, or abort. Raw stderr is preserved
// in the `raw` field for human/diagnostic consumption.
func formatResult(res execResult) toolsCallResult {
	if res.ExitCode != 0 {
		class := classifyError(res.Stderr, res.ExitCode)
		if class.Raw == "" && strings.TrimSpace(res.Stdout) != "" {
			// Pathological case: non-zero exit with empty stderr but
			// stdout text. Fall back to stdout so the agent gets context.
			class.Raw = strings.TrimSpace(res.Stdout)
			if class.Message == "" {
				class.Message = firstLine(class.Raw)
			}
		}
		// Serialize the classification as the content text. json.Marshal
		// is deterministic for this struct shape.
		buf, err := json.Marshal(&class)
		if err != nil {
			// Marshal failure on a struct of plain fields is unexpected;
			// fall back to raw stderr text so the agent still gets info.
			return toolsCallResult{
				Content: []contentBlock{{Type: "text", Text: class.Raw}},
				IsError: true,
			}
		}
		return toolsCallResult{
			Content: []contentBlock{{Type: "text", Text: string(buf)}},
			IsError: true,
		}
	}
	text := strings.TrimSpace(res.Stdout)
	if text == "" {
		text = strings.TrimSpace(res.Stderr)
	}
	if text == "" {
		text = "(empty response)"
	}
	return toolsCallResult{Content: []contentBlock{{Type: "text", Text: text}}}
}

func errorResult(msg string) toolsCallResult {
	return toolsCallResult{
		Content: []contentBlock{{Type: "text", Text: msg}},
		IsError: true,
	}
}

// ---------------------------------------------------------------------
// Helpers used by every tool builder.
// ---------------------------------------------------------------------

func argString(args map[string]interface{}, key string) (string, bool) {
	v, ok := args[key]
	if !ok || v == nil {
		return "", false
	}
	s, ok := v.(string)
	if !ok {
		return fmt.Sprintf("%v", v), true
	}
	return s, s != ""
}

func argInt(args map[string]interface{}, key string) (int64, bool) {
	v, ok := args[key]
	if !ok || v == nil {
		return 0, false
	}
	switch n := v.(type) {
	case float64:
		return int64(n), true
	case int:
		return int64(n), true
	case int64:
		return n, true
	case string:
		if n == "" {
			return 0, false
		}
		// fall through: let the CLI parse; we just signal "present"
		return 0, true
	}
	return 0, false
}

// appendFlag appends `--flag value` when value is non-empty. Cleanly
// omitted flags let lark-cli fall back to its own defaults.
func appendFlag(argv []string, flag, value string) []string {
	if value == "" {
		return argv
	}
	return append(argv, "--"+flag, value)
}

// appendBoolFlag appends `--flag` when the arg is a literal true.
// lark-cli boolean flags often reject `--flag=false`; we follow the
// "omit to disable" convention.
func appendBoolFlag(argv []string, args map[string]interface{}, key, flag string) []string {
	if v, ok := args[key]; ok && v != nil {
		if b, ok := v.(bool); ok && b {
			return append(argv, "--"+flag)
		}
	}
	return argv
}

// appendIdentity appends `--as <identity>` when the caller provided one.
// Accepts "user" or "bot" (the only values lark-cli understands).
func appendIdentity(argv []string, args map[string]interface{}) []string {
	if id, ok := argString(args, "as"); ok {
		id = strings.ToLower(id)
		if id == "user" || id == "bot" {
			argv = append(argv, "--as", id)
		}
	}
	return argv
}

// appendFormat is the same idea for --format. Default JSON is fine for
// machine consumption, so we only override when explicitly requested.
func appendFormat(argv []string, args map[string]interface{}) []string {
	if v, ok := argString(args, "format"); ok {
		argv = append(argv, "--format", v)
	}
	return argv
}

// appendJq appends `--jq <expr>` when the caller passes a jq expression.
// --jq is registered universally on every shortcut via the common runner
// (shortcuts/common/runner.go), so this works on any tool wrapping a
// shortcut. Note: --jq is mutually exclusive with --format / --output;
// the shortcut layer enforces this via output.ValidateJqFlags.
//
// Exposing jq at the MCP layer is the single biggest token-reduction
// lever for read-type tools — a full doc fetch can drop from 25k tokens
// to ~500 with a focused projection like `.content | .[0:2000]`.
func appendJq(argv []string, args map[string]interface{}) []string {
	if v, ok := argString(args, "jq"); ok {
		return append(argv, "--jq", v)
	}
	return argv
}

// requireOneOf returns an error if none of the listed keys are present.
// Used to enforce mutually-exclusive-or-required-pair shortcuts in the
// schema (which JSON Schema can express but Claude clients often skip).
func requireOneOf(args map[string]interface{}, keys ...string) error {
	for _, k := range keys {
		if s, ok := argString(args, k); ok && s != "" {
			return nil
		}
	}
	return fmt.Errorf("one of %s is required", strings.Join(keys, ", "))
}

// ---------------------------------------------------------------------
// Tool definitions.
//
// Every flag name below is verified against the Go shortcut definition
// in shortcuts/<domain>/*.go. When you add a tool, run
// `grep '{Name:' shortcuts/<domain>/<file>.go` and copy the flag
// names verbatim — never paraphrase. The MCP bridge has no flag
// translation layer; mismatches cause "unknown flag" failures.
// ---------------------------------------------------------------------

func toolIMSend() tool {
	return tool{
		Name:        "lark_im_send",
		Description: "Send a plain-text or markdown IM message to one chat (chat_id) or one user (user_id) — exactly one required. Use when the user asks to send a quick message, ping someone, or notify a chat. ALWAYS preview with dry_run=true first to verify recipient + body before commit. For rich layouts (sections, action buttons, status pills, deploy notifications, bilingual cards), use lark_im_card_send instead — markdown here only supports inline formatting, not interactive elements.",
		Schema: `{
  "type": "object",
  "properties": {
    "chat_id":  {"type": "string", "description": "Target chat id (oc_xxx). Mutually exclusive with user_id."},
    "user_id":  {"type": "string", "description": "Target user open_id (ou_xxx). Mutually exclusive with chat_id."},
    "text":     {"type": "string", "description": "Plain-text message body."},
    "markdown": {"type": "string", "description": "Markdown body (auto-converted to post format)."},
    "dry_run":  {"type": "boolean", "description": "Preview the would-be IM send without delivering. Returns the planned request as JSON. Recommended before any commit."},
    "as":       {"type": "string", "enum": ["user", "bot"], "description": "Identity to send as."}
  },
  "anyOf": [
    {"required": ["chat_id"]},
    {"required": ["user_id"]}
  ]
}`,
		Build: func(args map[string]interface{}) ([]string, error) {
			if err := requireOneOf(args, "chat_id", "user_id"); err != nil {
				return nil, err
			}
			if err := requireOneOf(args, "text", "markdown"); err != nil {
				return nil, err
			}
			argv := []string{"im", "+messages-send"}
			if v, _ := argString(args, "chat_id"); v != "" {
				argv = appendFlag(argv, "chat-id", v)
			}
			if v, _ := argString(args, "user_id"); v != "" {
				argv = appendFlag(argv, "user-id", v)
			}
			if v, _ := argString(args, "markdown"); v != "" {
				argv = appendFlag(argv, "markdown", v)
			} else if v, _ := argString(args, "text"); v != "" {
				argv = appendFlag(argv, "text", v)
			}
			argv = appendBoolFlag(argv, args, "dry_run", "dry-run")
			argv = appendIdentity(argv, args)
			return argv, nil
		},
	}
}

// toolIMCardSend exposes `im +card-send` — a YAML-driven Feishu
// interactive card. Use this instead of lark_im_send when the message
// would benefit from a header, collapsible sections, action buttons,
// list items with side-by-side button, or a notification preview line.
// The YAML spec is documented in skills/lark-im-card/references/spec-reference.md.
func toolIMCardSend() tool {
	return tool{
		Name:        "lark_im_card_send",
		Description: "Send a Feishu interactive card to a chat or user, or as a reply, by passing a YAML card spec. Use for any reply richer than plain markdown: reports with sections, lists with action buttons, status briefs, decision prompts, deploy/CI notifications, bilingual announcements. ALWAYS preview with dry_run=true (verifies the actual send request) or print_json=true (validates the spec offline) before commit. For plain text or markdown without interactive elements, use lark_im_send instead — it's lighter and doesn't need a YAML spec. Element kinds: md, hr, note, div (icon + text — the only way to render inline-iconed status pills), actions (button row, equal or inline layout), item (text + side button), select (dropdown), panel (collapsible section, recursive), columns (side-by-side), image, raw (escape hatch). Header supports template (colour), icon, title_color, subtitle. Markdown content is auto-optimised: H1/H2/H3 demoted to H4/H5 (Feishu otherwise renders them gigantic), invalid image refs stripped (prevents CardKit error 200570), fenced code blocks padded with <br>. Bilingual cards: set *_i18n maps on header.title/subtitle, summary, md.i18n, div.text_i18n, panel.title_i18n — config.locales auto-derives from the keys used. Always preview with print_json=true before sending. Full grammar and templates in skills/lark-im-card.",
		Schema: `{
  "type": "object",
  "properties": {
    "chat_id":   {"type": "string", "description": "Target chat id (oc_xxx). Mutually exclusive with user_id and reply_to."},
    "user_id":   {"type": "string", "description": "Target user open_id (ou_xxx). Mutually exclusive with chat_id and reply_to."},
    "reply_to":  {"type": "string", "description": "Reply to this message id (om_xxx). Mutually exclusive with chat_id and user_id."},
    "in_thread": {"type": "boolean", "description": "With reply_to: post the reply in the message's thread stream instead of the main chat."},
    "spec":      {"type": "string", "description": "YAML card spec. Top-level keys: header (title, subtitle, template, icon, *_i18n, *_color), summary, summary_i18n, locales (auto-derived from i18n usage), elements. Element kinds: md, hr, note, div, actions, item, select, panel, columns, image, raw. See skills/lark-im-card/references/spec-reference.md for the full grammar."},
    "idempotency_key": {"type": "string", "description": "Idempotency key — set to a stable value to prevent duplicate sends on retry."},
    "print_json": {"type": "boolean", "description": "Compile and return the Feishu card JSON without sending. Useful to preview a card structure. Different from dry_run: print_json validates + serialises the spec but is offline; dry_run reports what the actual send request would look like."},
    "dry_run":   {"type": "boolean", "description": "Preview the would-be card-send request without delivering. Returns the planned API call as JSON. Recommended before any commit."},
    "as":        {"type": "string", "enum": ["user", "bot"], "description": "Identity to send as. Defaults to bot."}
  },
  "required": ["spec"],
  "anyOf": [
    {"required": ["chat_id"]},
    {"required": ["user_id"]},
    {"required": ["reply_to"]},
    {"required": ["print_json"]}
  ]
}`,
		Build: func(args map[string]interface{}) ([]string, error) {
			if err := requireOneOf(args, "chat_id", "user_id", "reply_to", "print_json"); err != nil {
				return nil, err
			}
			spec, ok := argString(args, "spec")
			if !ok || spec == "" {
				return nil, fmt.Errorf("spec is required")
			}
			argv := []string{"im", "+card-send", "--spec", spec}
			if v, _ := argString(args, "chat_id"); v != "" {
				argv = appendFlag(argv, "chat-id", v)
			}
			if v, _ := argString(args, "user_id"); v != "" {
				argv = appendFlag(argv, "user-id", v)
			}
			if v, _ := argString(args, "reply_to"); v != "" {
				argv = appendFlag(argv, "reply-to", v)
			}
			argv = appendBoolFlag(argv, args, "in_thread", "in-thread")
			if v, _ := argString(args, "idempotency_key"); v != "" {
				argv = appendFlag(argv, "idempotency-key", v)
			}
			argv = appendBoolFlag(argv, args, "print_json", "print-json")
			argv = appendBoolFlag(argv, args, "dry_run", "dry-run")
			argv = appendIdentity(argv, args)
			return argv, nil
		},
	}
}

func toolIMSearch() tool {
	return tool{
		Name:        "lark_im_search",
		Description: "Search IM messages by keyword across the user's conversations. Use when the user asks to find what was said in chat, recall a decision made in IM, locate a shared link, or check whether something was discussed. Returns paginated message hits — pair with the jq arg to project only id/sender/snippet and avoid full payload bloat.",
		Schema: `{
  "type": "object",
  "properties": {
    "query":     {"type": "string", "description": "Search keyword."},
    "page_size": {"type": "integer", "description": "Results per page (default 20)."},
    "jq":        {"type": "string", "description": "Optional jq expression to project/filter the JSON output (e.g. '.items[] | {id, chat_id, body}'). Cuts token cost of large result sets. Mutually exclusive with --format."},
    "as":        {"type": "string", "enum": ["user", "bot"]}
  },
  "required": ["query"]
}`,
		Build: func(args map[string]interface{}) ([]string, error) {
			q, ok := argString(args, "query")
			if !ok {
				return nil, fmt.Errorf("query is required")
			}
			argv := []string{"im", "+messages-search", "--query", q}
			if v, _ := argString(args, "page_size"); v != "" {
				argv = appendFlag(argv, "page-size", v)
			}
			argv = appendJq(argv, args)
			argv = appendIdentity(argv, args)
			return argv, nil
		},
	}
}

// toolMailSend exposes `mail +send`. NEW in v0.3 — previously only
// `lark_api` could send mail via raw API call.
//
// Safety: the hook layer (pre-send.sh) intercepts every `mail +send`
// invocation and prints a preview to stderr before the command runs.
// We deliberately do NOT pass --confirm-send by default; the caller
// must opt in. Without --confirm-send, mail +send creates a draft.
func toolMailSend() tool {
	return tool{
		Name:        "lark_mail_send",
		Description: "Send an email via Lark Mail. Use when the user has explicitly confirmed sending — set confirm_send=true to commit. Without confirm_send (default) this saves to drafts instead, never reaching a recipient. Use lark_mail_draft_create if you want to compose a draft that the user reviews in the Lark Mail UI before sending. The pre-send hook prints a preview to stderr before execution.",
		Schema: `{
  "type": "object",
  "properties": {
    "to":           {"type": "string", "description": "Recipient email address(es), comma-separated."},
    "subject":      {"type": "string", "description": "Email subject."},
    "body":         {"type": "string", "description": "Email body. HTML or plain text — auto-detected."},
    "cc":           {"type": "string", "description": "CC recipients, comma-separated."},
    "bcc":          {"type": "string", "description": "BCC recipients, comma-separated."},
    "from":         {"type": "string", "description": "Sender address (e.g. an alias)."},
    "mailbox":      {"type": "string", "description": "Mailbox that owns the message."},
    "attach":       {"type": "string", "description": "Attachment file path(s), comma-separated, relative paths only."},
    "plain_text":   {"type": "boolean", "description": "Force plain-text mode, ignoring HTML auto-detection."},
    "confirm_send": {"type": "boolean", "description": "If true, send immediately. Default is to save as draft."},
    "as":           {"type": "string", "enum": ["user", "bot"]}
  },
  "required": ["to", "subject", "body"]
}`,
		Build: func(args map[string]interface{}) ([]string, error) {
			to, ok := argString(args, "to")
			if !ok {
				return nil, fmt.Errorf("to is required")
			}
			subject, ok := argString(args, "subject")
			if !ok {
				return nil, fmt.Errorf("subject is required")
			}
			body, ok := argString(args, "body")
			if !ok {
				return nil, fmt.Errorf("body is required")
			}
			argv := []string{"mail", "+send",
				"--to", to,
				"--subject", subject,
				"--body", body,
			}
			argv = appendFlag(argv, "cc", mustString(args, "cc"))
			argv = appendFlag(argv, "bcc", mustString(args, "bcc"))
			argv = appendFlag(argv, "from", mustString(args, "from"))
			argv = appendFlag(argv, "mailbox", mustString(args, "mailbox"))
			argv = appendFlag(argv, "attach", mustString(args, "attach"))
			argv = appendBoolFlag(argv, args, "plain_text", "plain-text")
			argv = appendBoolFlag(argv, args, "confirm_send", "confirm-send")
			argv = appendIdentity(argv, args)
			return argv, nil
		},
	}
}

// toolMailDraftCreate exposes `mail +draft-create`. NEW in v0.3.
// Drafts are the safe default — they never reach a recipient until
// the user explicitly sends from Lark Mail UI.
func toolMailDraftCreate() tool {
	return tool{
		Name:        "lark_mail_draft_create",
		Description: "Create a Lark Mail draft saved to the user's Drafts folder. Use when the user wants to compose without sending, or when policy requires manual review in the Lark Mail UI before delivery. Drafts NEVER reach a recipient automatically — the user must explicitly click Send in the Mail UI. For commit-and-send (with confirm_send guard), use lark_mail_send.",
		Schema: `{
  "type": "object",
  "properties": {
    "to":         {"type": "string", "description": "Recipient email address(es), comma-separated."},
    "subject":    {"type": "string", "description": "Draft subject."},
    "body":       {"type": "string", "description": "Draft body. HTML or plain text — auto-detected."},
    "cc":         {"type": "string", "description": "CC recipients, comma-separated."},
    "bcc":        {"type": "string", "description": "BCC recipients, comma-separated."},
    "from":       {"type": "string", "description": "Sender address (e.g. an alias)."},
    "mailbox":    {"type": "string", "description": "Mailbox that owns the draft."},
    "attach":     {"type": "string", "description": "Attachment file path(s), comma-separated, relative paths only."},
    "plain_text": {"type": "boolean", "description": "Force plain-text mode."},
    "as":         {"type": "string", "enum": ["user", "bot"]}
  },
  "required": ["subject", "body"]
}`,
		Build: func(args map[string]interface{}) ([]string, error) {
			subject, ok := argString(args, "subject")
			if !ok {
				return nil, fmt.Errorf("subject is required")
			}
			body, ok := argString(args, "body")
			if !ok {
				return nil, fmt.Errorf("body is required")
			}
			argv := []string{"mail", "+draft-create",
				"--subject", subject,
				"--body", body,
			}
			argv = appendFlag(argv, "to", mustString(args, "to"))
			argv = appendFlag(argv, "cc", mustString(args, "cc"))
			argv = appendFlag(argv, "bcc", mustString(args, "bcc"))
			argv = appendFlag(argv, "from", mustString(args, "from"))
			argv = appendFlag(argv, "mailbox", mustString(args, "mailbox"))
			argv = appendFlag(argv, "attach", mustString(args, "attach"))
			argv = appendBoolFlag(argv, args, "plain_text", "plain-text")
			argv = appendIdentity(argv, args)
			return argv, nil
		},
	}
}

func toolCalendarAgenda() tool {
	return tool{
		Name:        "lark_calendar_agenda",
		Description: "List upcoming calendar events for the authenticated user (defaults to today). Use when the user asks 'what's on my calendar', 'do I have meetings today', or wants to check availability before scheduling. For past meetings + recordings, use lark_vc_search instead — this tool only reads scheduled future events.",
		Schema: `{
  "type": "object",
  "properties": {
    "start":       {"type": "string", "description": "Start time (ISO 8601). Defaults to start of today."},
    "end":         {"type": "string", "description": "End time (ISO 8601). Defaults to end of the start day."},
    "calendar_id": {"type": "string", "description": "Calendar id. Defaults to the primary calendar."},
    "jq":          {"type": "string", "description": "Optional jq projection (e.g. '.events[] | {summary, start, attendees}'). Reduces token cost vs full event payload."},
    "as":          {"type": "string", "enum": ["user", "bot"]}
  }
}`,
		Build: func(args map[string]interface{}) ([]string, error) {
			argv := []string{"calendar", "+agenda"}
			argv = appendFlag(argv, "start", mustString(args, "start"))
			argv = appendFlag(argv, "end", mustString(args, "end"))
			argv = appendFlag(argv, "calendar-id", mustString(args, "calendar_id"))
			argv = appendJq(argv, args)
			argv = appendIdentity(argv, args)
			return argv, nil
		},
	}
}

func toolCalendarCreate() tool {
	return tool{
		Name:        "lark_calendar_create",
		Description: "Create a calendar event on the user's primary (or specified) calendar. Use when the user asks to schedule a meeting, block focus time, or invite attendees. ALWAYS preview with dry_run=true first to verify start/end time, attendee resolution, and rrule recurrence; re-call with dry_run omitted to commit. Resolve attendee open_ids via lark_contact_search before passing in attendee_ids.",
		Schema: `{
  "type": "object",
  "properties": {
    "summary":      {"type": "string", "description": "Event title."},
    "start":        {"type": "string", "description": "Start time (ISO 8601)."},
    "end":          {"type": "string", "description": "End time (ISO 8601)."},
    "description":  {"type": "string"},
    "attendee_ids": {"type": "string", "description": "Comma-separated attendee IDs (open_ids ou_, chat oc_, room omm_)."},
    "calendar_id":  {"type": "string", "description": "Calendar id (default: primary)."},
    "rrule":        {"type": "string", "description": "Recurrence rule (RFC 5545)."},
    "dry_run":      {"type": "boolean", "description": "Preview the would-be event without creating it. Returns the planned request as JSON. Recommended before any commit."},
    "as":           {"type": "string", "enum": ["user", "bot"]}
  },
  "required": ["summary", "start", "end"]
}`,
		Build: func(args map[string]interface{}) ([]string, error) {
			summary, ok := argString(args, "summary")
			if !ok {
				return nil, fmt.Errorf("summary is required")
			}
			start, ok := argString(args, "start")
			if !ok {
				return nil, fmt.Errorf("start is required")
			}
			end, ok := argString(args, "end")
			if !ok {
				return nil, fmt.Errorf("end is required")
			}
			argv := []string{"calendar", "+create",
				"--summary", summary,
				"--start", start,
				"--end", end,
			}
			argv = appendFlag(argv, "description", mustString(args, "description"))
			argv = appendFlag(argv, "attendee-ids", mustString(args, "attendee_ids"))
			argv = appendFlag(argv, "calendar-id", mustString(args, "calendar_id"))
			argv = appendFlag(argv, "rrule", mustString(args, "rrule"))
			argv = appendBoolFlag(argv, args, "dry_run", "dry-run")
			argv = appendIdentity(argv, args)
			return argv, nil
		},
	}
}

func toolDocCreate() tool {
	return tool{
		Name:        "lark_doc_create",
		Description: "Create a Lark Doc from a title and optional markdown body. Always uses v2 API. Use when the user asks to draft a new document, capture meeting notes, generate a report, or open a writing surface. ALWAYS preview with dry_run=true first to verify the destination + body before commit. Provide folder_token (Drive folder) OR wiki_node + wiki_space (Wiki location); omit both for the user's My Drive root. For appending to an existing doc, use lark_api with the docx block-create endpoint instead.",
		Schema: `{
  "type": "object",
  "properties": {
    "title":        {"type": "string", "description": "Document title."},
    "markdown":     {"type": "string", "description": "Markdown body for the initial content."},
    "folder_token": {"type": "string", "description": "Target folder token. Omit for root."},
    "wiki_node":    {"type": "string", "description": "Target wiki node token."},
    "wiki_space":   {"type": "string", "description": "Wiki space id (use 'my_library' for personal library)."},
    "dry_run":      {"type": "boolean", "description": "Preview the would-be doc creation request without persisting. Returns the planned API call as JSON. Recommended before any commit."},
    "as":           {"type": "string", "enum": ["user", "bot"]}
  },
  "required": ["title"]
}`,
		Build: func(args map[string]interface{}) ([]string, error) {
			title, ok := argString(args, "title")
			if !ok {
				return nil, fmt.Errorf("title is required")
			}
			// docs v2 is mandatory per the lark-doc skill contract;
			// the raw flag defaults to v1 which produces a deprecated
			// doc type.
			argv := []string{"docs", "+create",
				"--api-version", "v2",
				"--title", title,
			}
			argv = appendFlag(argv, "markdown", mustString(args, "markdown"))
			argv = appendFlag(argv, "folder-token", mustString(args, "folder_token"))
			argv = appendFlag(argv, "wiki-node", mustString(args, "wiki_node"))
			argv = appendFlag(argv, "wiki-space", mustString(args, "wiki_space"))
			argv = appendBoolFlag(argv, args, "dry_run", "dry-run")
			argv = appendIdentity(argv, args)
			return argv, nil
		},
	}
}

func toolDocSearch() tool {
	return tool{
		Name:        "lark_doc_search",
		Description: "Search Lark Docs in the user's drive by keyword. Use when the user asks to find a doc by name/topic, list recent docs on a subject, or check whether a doc already exists. Returns title + token for each hit; pair with lark_doc_fetch to read full content of a specific result. Use the jq arg to project only the fields you need when searches return many candidates.",
		Schema: `{
  "type": "object",
  "properties": {
    "query":     {"type": "string"},
    "filter":    {"type": "string", "description": "Optional filter conditions as a JSON object."},
    "page_size": {"type": "integer", "description": "Results per page (default 15, max 20)."},
    "jq":        {"type": "string", "description": "Optional jq projection (e.g. '.items[] | {title, token, url}'). Cuts token cost vs full search response."},
    "as":        {"type": "string", "enum": ["user", "bot"]}
  },
  "required": ["query"]
}`,
		Build: func(args map[string]interface{}) ([]string, error) {
			q, ok := argString(args, "query")
			if !ok {
				return nil, fmt.Errorf("query is required")
			}
			argv := []string{"docs", "+search", "--query", q}
			argv = appendFlag(argv, "filter", mustString(args, "filter"))
			argv = appendFlag(argv, "page-size", mustString(args, "page_size"))
			argv = appendJq(argv, args)
			argv = appendIdentity(argv, args)
			return argv, nil
		},
	}
}

func toolDocFetch() tool {
	return tool{
		Name:        "lark_doc_fetch",
		Description: "Fetch a Lark Doc's full contents as markdown. Always uses v2 API. Use when you need to read, summarize, translate, or quote from a specific doc whose token/URL you already have. For large docs (>20 pages), use the jq arg to project a slice (e.g. '.content[:5000]' for first 5k chars, or '.title' for just metadata) to control token cost. To find docs first when you don't know the token, use lark_doc_search.",
		Schema: `{
  "type": "object",
  "properties": {
    "doc": {"type": "string", "description": "Document URL or token (the id segment in the doc URL)."},
    "jq":  {"type": "string", "description": "Optional jq projection. A full doc can be 25k+ tokens; use jq to slice (e.g. '.content[:5000]' for first 5k chars, or '.title' for just metadata)."},
    "as":  {"type": "string", "enum": ["user", "bot"]}
  },
  "required": ["doc"]
}`,
		Build: func(args map[string]interface{}) ([]string, error) {
			id, ok := argString(args, "doc")
			if !ok {
				// Backward compat for the prior schema (doc_id).
				id, ok = argString(args, "doc_id")
				if !ok {
					return nil, fmt.Errorf("doc is required (the document URL or token)")
				}
			}
			argv := []string{"docs", "+fetch",
				"--api-version", "v2",
				"--doc", id,
			}
			argv = appendJq(argv, args)
			argv = appendIdentity(argv, args)
			return argv, nil
		},
	}
}

func toolBaseSearch() tool {
	return tool{
		Name: "lark_base_search",
		Description: "Search records in a Lark Base (Bitable) table by keyword. " +
			"Use when the user asks to find records matching a name/keyword in a known Base table, " +
			"check whether a record exists, or list records meeting a simple filter. " +
			"The `query` is wrapped into the JSON shape the shortcut requires; " +
			"for full control over search_fields / select_fields / view_id / limit, " +
			"pass `search_json` as a complete JSON object instead. " +
			"For counts, sums, or aggregations across many records, use lark_api with the " +
			"base data-query endpoint — fetching raw records to count client-side wastes tokens.",
		Schema: `{
  "type": "object",
  "properties": {
    "base_token":     {"type": "string", "description": "Base (Bitable) app token."},
    "table_id":       {"type": "string", "description": "Table id (tblXXX) or display name."},
    "query":          {"type": "string", "description": "Search keyword. Wrapped into the required JSON shape."},
    "search_fields":  {"type": "string", "description": "Optional. Comma-separated field names/ids to search in."},
    "select_fields":  {"type": "string", "description": "Optional. Comma-separated field names/ids to return."},
    "view_id":        {"type": "string", "description": "Optional. Scope to a specific view."},
    "limit":          {"type": "integer", "description": "Optional. Records per page (1-200, default 10)."},
    "search_json":    {"type": "object", "description": "Advanced: pass the full search JSON object directly. Overrides query/search_fields/etc when present."},
    "jq":             {"type": "string", "description": "Optional jq projection (e.g. '.records[] | {id, fields}'). Base records can be wide; jq reduces token cost."},
    "as":             {"type": "string", "enum": ["user", "bot"]}
  },
  "required": ["base_token", "table_id"]
}`,
		Build: func(args map[string]interface{}) ([]string, error) {
			baseToken, ok := argString(args, "base_token")
			if !ok {
				return nil, fmt.Errorf("base_token is required")
			}
			tableID, ok := argString(args, "table_id")
			if !ok {
				return nil, fmt.Errorf("table_id is required")
			}

			// Build the --json payload the shortcut requires.
			var jsonStr string
			if raw, ok := args["search_json"]; ok && raw != nil {
				buf, err := json.Marshal(raw)
				if err != nil {
					return nil, fmt.Errorf("search_json must be a JSON object: %w", err)
				}
				jsonStr = string(buf)
			} else {
				q, ok := argString(args, "query")
				if !ok {
					return nil, fmt.Errorf("query or search_json is required")
				}
				payload := map[string]interface{}{"keyword": q}
				if v, _ := argString(args, "search_fields"); v != "" {
					payload["search_fields"] = strings.Split(v, ",")
				}
				if v, _ := argString(args, "select_fields"); v != "" {
					payload["select_fields"] = strings.Split(v, ",")
				}
				if v, _ := argString(args, "view_id"); v != "" {
					payload["view_id"] = v
				}
				if n, ok := argInt(args, "limit"); ok && n > 0 {
					payload["limit"] = n
				}
				buf, _ := json.Marshal(payload)
				jsonStr = string(buf)
			}

			argv := []string{"base", "+record-search",
				"--base-token", baseToken,
				"--table-id", tableID,
				"--json", jsonStr,
			}
			argv = appendJq(argv, args)
			argv = appendIdentity(argv, args)
			return argv, nil
		},
	}
}

func toolContactSearch() tool {
	return tool{
		Name:        "lark_contact_search",
		Description: "Search the organization directory by name/email/phone, or look up users by open_id list. Use when the user mentions someone by name and the next action needs an open_id (sending mail, scheduling a meeting, creating a task, mentioning in a card). Pass `query` for keyword search; pass `user_ids` (CSV, supports 'me' for caller) to fetch attributes of known open_ids. Always resolve via this tool before hand-typing an open_id elsewhere — open_ids are opaque and cannot be guessed.",
		Schema: `{
  "type": "object",
  "properties": {
    "query":     {"type": "string", "description": "Keyword (≤50 chars)."},
    "user_ids":  {"type": "string", "description": "open_ids to look up or restrict --query against (CSV; 'me' = caller)."},
    "page_size": {"type": "integer", "description": "Rows per request, 1-30 (default 20)."},
    "jq":        {"type": "string", "description": "Optional jq projection (e.g. '.users[] | {open_id, name, email, department}'). Avoids fetching full directory metadata into context."},
    "as":        {"type": "string", "enum": ["user", "bot"]}
  },
  "anyOf": [
    {"required": ["query"]},
    {"required": ["user_ids"]}
  ]
}`,
		Build: func(args map[string]interface{}) ([]string, error) {
			if err := requireOneOf(args, "query", "user_ids"); err != nil {
				return nil, err
			}
			argv := []string{"contact", "+search-user"}
			argv = appendFlag(argv, "query", mustString(args, "query"))
			argv = appendFlag(argv, "user-ids", mustString(args, "user_ids"))
			argv = appendFlag(argv, "page-size", mustString(args, "page_size"))
			argv = appendJq(argv, args)
			argv = appendIdentity(argv, args)
			return argv, nil
		},
	}
}

// toolTaskMy exposes `task +get-my-tasks`. NEW in v0.3 — previously
// only `lark_task_create` was exposed, so the bridge could create but
// never read tasks.
func toolTaskMy() tool {
	return tool{
		Name:        "lark_task_my",
		Description: "List the authenticated user's Lark Tasks (todos), optionally filtered by completion status, due time, or creation time. Use when the user asks 'what are my tasks', 'what's overdue', 'what's due this week', or wants a status snapshot. Combine with the jq arg to project id/summary/due/completed only and avoid full task payload bloat when many tasks match. For tasks assigned to others, use lark_api with the task list endpoint.",
		Schema: `{
  "type": "object",
  "properties": {
    "query":      {"type": "string", "description": "Filter by summary keyword."},
    "complete":   {"type": "boolean", "description": "true=completed only, false=incomplete only, omit=both."},
    "created_at": {"type": "string", "description": "Tasks created after this time (ISO 8601 / +Nd / ms)."},
    "due_start":  {"type": "string", "description": "Tasks with due date after this time."},
    "due_end":    {"type": "string", "description": "Tasks with due date before this time."},
    "page_limit": {"type": "integer", "description": "Max items per page (default 20, max 40 with page_all=true)."},
    "page_all":   {"type": "boolean", "description": "If true, automatically paginate through all pages."},
    "jq":         {"type": "string", "description": "Optional jq projection (e.g. '.tasks[] | {id, summary, due, completed}'). Reduces token cost for large task lists."},
    "as":         {"type": "string", "enum": ["user", "bot"]}
  }
}`,
		Build: func(args map[string]interface{}) ([]string, error) {
			argv := []string{"task", "+get-my-tasks"}
			argv = appendFlag(argv, "query", mustString(args, "query"))
			argv = appendFlag(argv, "created_at", mustString(args, "created_at"))
			argv = appendFlag(argv, "due-start", mustString(args, "due_start"))
			argv = appendFlag(argv, "due-end", mustString(args, "due_end"))
			argv = appendFlag(argv, "page-limit", mustString(args, "page_limit"))
			argv = appendBoolFlag(argv, args, "complete", "complete")
			argv = appendBoolFlag(argv, args, "page_all", "page-all")
			argv = appendJq(argv, args)
			argv = appendIdentity(argv, args)
			return argv, nil
		},
	}
}

func toolTaskCreate() tool {
	return tool{
		Name:        "lark_task_create",
		Description: "Create a Lark Task (todo) with a single assignee per call (not a list). Use when the user asks to add a todo, capture a follow-up action item from a meeting, schedule a reminder, or assign work. Resolve the assignee open_id via lark_contact_search first — open_ids cannot be hand-typed. For batch creation, loop one call per task. ALWAYS preview with dry_run=true first to verify summary + assignee + due before commit.",
		Schema: `{
  "type": "object",
  "properties": {
    "summary":     {"type": "string", "description": "Task summary/title."},
    "description": {"type": "string"},
    "due":         {"type": "string", "description": "Due time (ISO 8601 / date:YYYY-MM-DD / relative:+2d / ms)."},
    "assignee":    {"type": "string", "description": "Assignee open_id (ou_xxx) or app id (cli_xxx). Single value, not a list."},
    "follower":    {"type": "string", "description": "Follower open_id (single value)."},
    "tasklist_id": {"type": "string", "description": "Tasklist id to attach the task to."},
    "dry_run":     {"type": "boolean", "description": "Preview the would-be task creation request without persisting. Returns the planned API call as JSON. Recommended before any commit."},
    "as":          {"type": "string", "enum": ["user", "bot"]}
  },
  "required": ["summary"]
}`,
		Build: func(args map[string]interface{}) ([]string, error) {
			summary, ok := argString(args, "summary")
			if !ok {
				return nil, fmt.Errorf("summary is required")
			}
			argv := []string{"task", "+create", "--summary", summary}
			argv = appendFlag(argv, "description", mustString(args, "description"))
			argv = appendFlag(argv, "due", mustString(args, "due"))
			argv = appendFlag(argv, "assignee", mustString(args, "assignee"))
			argv = appendFlag(argv, "follower", mustString(args, "follower"))
			argv = appendFlag(argv, "tasklist-id", mustString(args, "tasklist_id"))
			argv = appendBoolFlag(argv, args, "dry_run", "dry-run")
			argv = appendIdentity(argv, args)
			return argv, nil
		},
	}
}

func toolDriveUpload() tool {
	return tool{
		Name:        "lark_drive_upload",
		Description: "Upload a local file to Lark Drive. Use when the user asks to share a file, attach a document to a chat, save a local artifact to the cloud, or migrate a file from disk to Drive. Provide folder_token (default location) OR wiki_token (mutually exclusive); omit both for the user's Drive root. Files > 20MB use multipart upload automatically. Set dry_run=true to preview the upload plan (destination + file size) without transferring bytes. For uploads paired with a doc-creation flow, prefer lark_doc_create with a markdown body when the source is text-based.",
		Schema: `{
  "type": "object",
  "properties": {
    "file":         {"type": "string", "description": "Local file path."},
    "folder_token": {"type": "string", "description": "Target folder token (default: root). Mutually exclusive with wiki_token."},
    "wiki_token":   {"type": "string", "description": "Target wiki node token. Mutually exclusive with folder_token."},
    "name":         {"type": "string", "description": "Uploaded file name (default: local file name)."},
    "dry_run":      {"type": "boolean", "description": "Preview the upload request without transferring bytes. Returns planned destination + file size as JSON."},
    "as":           {"type": "string", "enum": ["user", "bot"]}
  },
  "required": ["file"]
}`,
		Build: func(args map[string]interface{}) ([]string, error) {
			path, ok := argString(args, "file")
			if !ok {
				// Backward compat for the prior schema (file_path).
				path, ok = argString(args, "file_path")
				if !ok {
					return nil, fmt.Errorf("file is required")
				}
			}
			argv := []string{"drive", "+upload", "--file", path}
			argv = appendFlag(argv, "folder-token", mustString(args, "folder_token"))
			argv = appendFlag(argv, "wiki-token", mustString(args, "wiki_token"))
			argv = appendFlag(argv, "name", mustString(args, "name"))
			argv = appendBoolFlag(argv, args, "dry_run", "dry-run")
			argv = appendIdentity(argv, args)
			return argv, nil
		},
	}
}

// toolSheetsRead exposes `sheets +read`. NEW in v0.3 — sheets was
// entirely unreachable through MCP before.
func toolSheetsRead() tool {
	return tool{
		Name:        "lark_sheets_read",
		Description: "Read a cell range from a Lark Sheet. Use when the user references a specific sheet by URL or token and asks to inspect, summarize, or analyze cell values. Provide spreadsheet_token OR url, plus a range like 'Sheet1!A1:C10' or 'A1:C10' (with sheet_id). For wide sheets (many columns), project specific columns via the jq arg (e.g. '.values[][0:3]' for first 3 columns) to control token cost. For appending rows, use lark_sheets_append.",
		Schema: `{
  "type": "object",
  "properties": {
    "spreadsheet_token":  {"type": "string", "description": "Spreadsheet token. Mutually exclusive with url."},
    "url":                {"type": "string", "description": "Spreadsheet URL. Mutually exclusive with spreadsheet_token."},
    "range":              {"type": "string", "description": "Cell range. Examples: '<sheetId>!A1:D10', 'A1:D10' (with sheet_id), 'C2'."},
    "sheet_id":           {"type": "string", "description": "Sheet (tab) id. Required if range omits the sheetId prefix."},
    "value_render_option": {"type": "string", "enum": ["ToString", "FormattedValue", "Formula", "UnformattedValue"], "description": "How cell values are rendered."},
    "jq":                 {"type": "string", "description": "Optional jq projection (e.g. '.values[][0:3]' to keep only first 3 columns). Cuts token cost for wide sheets."},
    "as":                 {"type": "string", "enum": ["user", "bot"]}
  },
  "anyOf": [
    {"required": ["spreadsheet_token"]},
    {"required": ["url"]}
  ]
}`,
		Build: func(args map[string]interface{}) ([]string, error) {
			if err := requireOneOf(args, "spreadsheet_token", "url"); err != nil {
				return nil, err
			}
			argv := []string{"sheets", "+read"}
			argv = appendFlag(argv, "spreadsheet-token", mustString(args, "spreadsheet_token"))
			argv = appendFlag(argv, "url", mustString(args, "url"))
			argv = appendFlag(argv, "range", mustString(args, "range"))
			argv = appendFlag(argv, "sheet-id", mustString(args, "sheet_id"))
			argv = appendFlag(argv, "value-render-option", mustString(args, "value_render_option"))
			argv = appendJq(argv, args)
			argv = appendIdentity(argv, args)
			return argv, nil
		},
	}
}

// toolSheetsAppend exposes `sheets +append`. NEW in v0.3.
// Append is a mutating verb — pre-mutate.sh logs it; the host's
// permissions.ask still asks for confirmation per call.
func toolSheetsAppend() tool {
	return tool{
		Name:        "lark_sheets_append",
		Description: "Append rows to a Lark Sheet range. Use when the user asks to log new data into a tracking sheet, add records to a list, or accumulate audit/decision entries. `values` is a 2D-array JSON string (rows × columns) — example: '[[\"Alice\",30],[\"Bob\",25]]'. ALWAYS preview with dry_run=true first to verify the destination + values before commit. For reading existing data first, use lark_sheets_read.",
		Schema: `{
  "type": "object",
  "properties": {
    "spreadsheet_token": {"type": "string", "description": "Spreadsheet token. Mutually exclusive with url."},
    "url":               {"type": "string", "description": "Spreadsheet URL. Mutually exclusive with spreadsheet_token."},
    "range":             {"type": "string", "description": "Append range. Examples: '<sheetId>!A1:D10', 'A1:D10' (with sheet_id)."},
    "sheet_id":          {"type": "string", "description": "Sheet (tab) id. Required if range omits the sheetId prefix."},
    "values":            {"type": "string", "description": "2D-array JSON. Example: '[[\"Alice\",30],[\"Bob\",25]]'. Required."},
    "dry_run":           {"type": "boolean", "description": "Preview the would-be append request without writing. Returns the planned API call as JSON. Recommended before any commit."},
    "as":                {"type": "string", "enum": ["user", "bot"]}
  },
  "required": ["values"],
  "anyOf": [
    {"required": ["spreadsheet_token"]},
    {"required": ["url"]}
  ]
}`,
		Build: func(args map[string]interface{}) ([]string, error) {
			if err := requireOneOf(args, "spreadsheet_token", "url"); err != nil {
				return nil, err
			}
			values, ok := argString(args, "values")
			if !ok {
				return nil, fmt.Errorf("values is required (2D-array JSON string)")
			}
			argv := []string{"sheets", "+append", "--values", values}
			argv = appendFlag(argv, "spreadsheet-token", mustString(args, "spreadsheet_token"))
			argv = appendFlag(argv, "url", mustString(args, "url"))
			argv = appendFlag(argv, "range", mustString(args, "range"))
			argv = appendFlag(argv, "sheet-id", mustString(args, "sheet_id"))
			argv = appendBoolFlag(argv, args, "dry_run", "dry-run")
			argv = appendIdentity(argv, args)
			return argv, nil
		},
	}
}

// toolVCSearch exposes `vc +search`. NEW in v0.3 — required by
// meeting-prep, deal-update, contact-360 skills.
func toolVCSearch() tool {
	return tool{
		Name:        "lark_vc_search",
		Description: "Search past Lark video meetings by keyword, time range, organizer, participant, or room. Use when the user asks 'when did we last meet about X', 'who attended the Q3 review', or wants meeting metadata (start/end/duration/attendees) for already-concluded meetings. For upcoming meetings on the calendar, use lark_calendar_agenda instead. For transcripts/recording content from those meetings, use lark_minutes_search.",
		Schema: `{
  "type": "object",
  "properties": {
    "query":           {"type": "string", "description": "Search keyword."},
    "start":           {"type": "string", "description": "Start time (ISO 8601 or YYYY-MM-DD)."},
    "end":             {"type": "string", "description": "End time (ISO 8601 or YYYY-MM-DD)."},
    "organizer_ids":   {"type": "string", "description": "Organizer open_id list, comma-separated."},
    "participant_ids": {"type": "string", "description": "Participant open_id list, comma-separated."},
    "room_ids":        {"type": "string", "description": "Room id list, comma-separated."},
    "page_size":       {"type": "integer", "description": "Page size, 1-30 (default 15)."},
    "jq":              {"type": "string", "description": "Optional jq projection (e.g. '.meetings[] | {id, topic, start, organizer}'). Reduces token cost vs full meeting payload."},
    "as":              {"type": "string", "enum": ["user", "bot"]}
  }
}`,
		Build: func(args map[string]interface{}) ([]string, error) {
			argv := []string{"vc", "+search"}
			argv = appendFlag(argv, "query", mustString(args, "query"))
			argv = appendFlag(argv, "start", mustString(args, "start"))
			argv = appendFlag(argv, "end", mustString(args, "end"))
			argv = appendFlag(argv, "organizer-ids", mustString(args, "organizer_ids"))
			argv = appendFlag(argv, "participant-ids", mustString(args, "participant_ids"))
			argv = appendFlag(argv, "room-ids", mustString(args, "room_ids"))
			argv = appendFlag(argv, "page-size", mustString(args, "page_size"))
			argv = appendJq(argv, args)
			argv = appendIdentity(argv, args)
			return argv, nil
		},
	}
}

// toolMinutesSearch exposes `minutes +search`. NEW in v0.3 — required
// by meeting-prep Phase 2 (post-meeting) and deal-update skills.
func toolMinutesSearch() tool {
	return tool{
		Name:        "lark_minutes_search",
		Description: "Search Lark Minutes (meeting recordings + transcripts) by keyword, owner, participant, or time range. Use when the user asks to find a recording, retrieve a transcript excerpt, recall what was said in a specific meeting, or surface action items from past meeting content. For meeting metadata only (attendees, room, times) without content, use lark_vc_search — it's lighter.",
		Schema: `{
  "type": "object",
  "properties": {
    "query":           {"type": "string", "description": "Search keyword."},
    "owner_ids":       {"type": "string", "description": "Owner open_id list, comma-separated (use 'me' for current user)."},
    "participant_ids": {"type": "string", "description": "Participant open_id list, comma-separated (use 'me' for current user)."},
    "start":           {"type": "string", "description": "Time lower bound (ISO 8601 or YYYY-MM-DD)."},
    "end":             {"type": "string", "description": "Time upper bound (ISO 8601 or YYYY-MM-DD)."},
    "page_size":       {"type": "integer", "description": "Page size, 1-30 (default 15)."},
    "jq":              {"type": "string", "description": "Optional jq projection (e.g. '.minutes[] | {token, topic, owner, created_at}'). Skips full transcript download into context."},
    "as":              {"type": "string", "enum": ["user", "bot"]}
  }
}`,
		Build: func(args map[string]interface{}) ([]string, error) {
			argv := []string{"minutes", "+search"}
			argv = appendFlag(argv, "query", mustString(args, "query"))
			argv = appendFlag(argv, "owner-ids", mustString(args, "owner_ids"))
			argv = appendFlag(argv, "participant-ids", mustString(args, "participant_ids"))
			argv = appendFlag(argv, "start", mustString(args, "start"))
			argv = appendFlag(argv, "end", mustString(args, "end"))
			argv = appendFlag(argv, "page-size", mustString(args, "page_size"))
			argv = appendJq(argv, args)
			argv = appendIdentity(argv, args)
			return argv, nil
		},
	}
}

// toolOKRCycleList exposes `okr +cycle-list`. NEW in v0.3 — required
// by weekly-review, one-on-one-prep, task-prioritizer skills.
func toolOKRCycleList() tool {
	return tool{
		Name:        "lark_okr_cycle_list",
		Description: "List OKR cycles for a user — 'me' for the authenticated user, or an open_id/union_id/user_id for someone else. Use when the user asks about current/past quarter OKRs, wants to review goal progress, or prepares for a 1:1 or weekly review. Filter with time_range='YYYY-MM--YYYY-MM' to scope by quarter. Pair with the jq arg to project name/status/period only when listing many cycles.",
		Schema: `{
  "type": "object",
  "properties": {
    "user_id":      {"type": "string", "description": "User id ('me' for caller, or open_id/union_id/user_id)."},
    "user_id_type": {"type": "string", "enum": ["open_id", "union_id", "user_id"], "description": "Type of user_id (default open_id)."},
    "time_range":   {"type": "string", "description": "Time range in format 'YYYY-MM--YYYY-MM'. Omit for all cycles."},
    "jq":           {"type": "string", "description": "Optional jq projection (e.g. '.cycles[] | {id, name, status}'). Reduces token cost when listing many cycles."},
    "as":           {"type": "string", "enum": ["user", "bot"]}
  },
  "required": ["user_id"]
}`,
		Build: func(args map[string]interface{}) ([]string, error) {
			userID, ok := argString(args, "user_id")
			if !ok {
				return nil, fmt.Errorf("user_id is required")
			}
			argv := []string{"okr", "+cycle-list", "--user-id", userID}
			argv = appendFlag(argv, "user-id-type", mustString(args, "user_id_type"))
			argv = appendFlag(argv, "time-range", mustString(args, "time_range"))
			argv = appendJq(argv, args)
			argv = appendIdentity(argv, args)
			return argv, nil
		},
	}
}

func toolGenericAPI() tool {
	return tool{
		Name: "lark_api",
		Description: "Generic Lark Open API passthrough. Use when no dedicated tool fits. " +
			"Equivalent to `lark-cli api <METHOD> <PATH> --params <json> --data <json>`. " +
			"Supports jq projection and dry_run for safe verification before commit.",
		Schema: `{
  "type": "object",
  "properties": {
    "method":  {"type": "string", "enum": ["GET", "POST", "PUT", "PATCH", "DELETE"], "description": "HTTP method."},
    "path":    {"type": "string", "description": "API path, e.g. /open-apis/im/v1/messages."},
    "params":  {"type": "object", "description": "URL/query parameters as JSON."},
    "data":    {"type": "object", "description": "Request body as JSON."},
    "jq":      {"type": "string", "description": "Optional jq projection on the response. Mutually exclusive with format (unless format=json)."},
    "dry_run": {"type": "boolean", "description": "Print the request without executing. Useful before mutating verbs (POST/PUT/PATCH/DELETE)."},
    "as":      {"type": "string", "enum": ["user", "bot"]}
  },
  "required": ["method", "path"]
}`,
		Build: func(args map[string]interface{}) ([]string, error) {
			method, ok := argString(args, "method")
			if !ok {
				return nil, fmt.Errorf("method is required")
			}
			path, ok := argString(args, "path")
			if !ok {
				return nil, fmt.Errorf("path is required")
			}
			argv := []string{"api", strings.ToUpper(method), path}
			if raw, ok := args["params"]; ok && raw != nil {
				if buf, err := json.Marshal(raw); err == nil && string(buf) != "null" {
					argv = append(argv, "--params", string(buf))
				}
			}
			if raw, ok := args["data"]; ok && raw != nil {
				if buf, err := json.Marshal(raw); err == nil && string(buf) != "null" {
					argv = append(argv, "--data", string(buf))
				}
			}
			argv = appendJq(argv, args)
			argv = appendBoolFlag(argv, args, "dry_run", "dry-run")
			argv = appendIdentity(argv, args)
			argv = appendFormat(argv, args)
			return argv, nil
		},
	}
}

// mustString is a small ergonomic helper for builders that want a
// "string or empty" without the second-return-value dance. Returns
// "" when the key is missing, nil, or not a string.
func mustString(args map[string]interface{}, key string) string {
	v, _ := argString(args, key)
	return v
}

// sortedToolNames is exposed for `lark-cli mcp tools` (human-readable
// listing) so the CLI surface mirrors what an MCP client would see.
func sortedToolNames() []string {
	tools := allTools()
	names := make([]string, 0, len(tools))
	for _, t := range tools {
		names = append(names, t.Name)
	}
	sort.Strings(names)
	return names
}
