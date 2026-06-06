// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package mcp

import (
	"strings"
	"testing"
)

// TestToolDescriptionsHaveSelectionGuidance enforces the ACI contract:
// every tool's Description must give the host model a clear answer to
// "when should I pick this tool?" Either via an explicit "Use when"
// clause OR by being substantively detailed enough that the answer is
// unambiguous (lark_im_card_send is the canonical example — its long
// description IS the guidance).
//
// Rationale: tool description is the single biggest lever for
// tool-selection accuracy. Vague descriptions like "Search documents"
// produce wrong-tool picks in 10-30% of similar tasks; explicit
// "Use when" cuts this to <5%. See Anthropic's tool-use docs and
// AUDIT.md Pattern A for the empirical breakdown.
//
// The rule: each description must either
//
//	(a) contain a "Use when" / "USE WHEN" / "Use this" phrase, OR
//	(b) be ≥ 400 chars (substantively self-describing).
//
// AND every description ≥ 80 chars (so we catch one-liners).
func TestToolDescriptionsHaveSelectionGuidance(t *testing.T) {
	const minLen = 80
	const longEnoughToSelfDescribe = 400

	for _, tl := range allTools() {
		t.Run(tl.Name, func(t *testing.T) {
			desc := tl.Description
			if len(desc) < minLen {
				t.Errorf("description too short: %d chars (min %d). Need substantive guidance.\n  desc: %q",
					len(desc), minLen, desc)
				return
			}
			lower := strings.ToLower(desc)
			hasUseWhen := strings.Contains(lower, "use when") ||
				strings.Contains(lower, "use this when") ||
				strings.Contains(lower, "use for")
			isLongEnough := len(desc) >= longEnoughToSelfDescribe
			if !hasUseWhen && !isLongEnough {
				t.Errorf("description lacks selection guidance: no \"Use when\" / \"Use this when\" / \"Use for\" phrase and only %d chars (need ≥%d for self-describing). Add explicit trigger guidance.\n  desc: %q",
					len(desc), longEnoughToSelfDescribe, desc)
			}
		})
	}
}

// TestConfusableToolPairsCrossReference enforces that tools likely to
// be confused name each other in their description. The host model
// reads tool descriptions to pick among similar tools; if neither
// description mentions the other, the model picks at random.
//
// Pairs locked:
//   - lark_im_send (plain text/markdown) ↔ lark_im_card_send (rich card)
//   - lark_mail_send (commits send when confirm_send=true) ↔
//     lark_mail_draft_create (always draft, never sends)
//   - lark_calendar_agenda (read upcoming) ↔ lark_vc_search (past meetings)
//
// Each tool in a pair must reference the other by exact name.
func TestConfusableToolPairsCrossReference(t *testing.T) {
	pairs := []struct {
		a, b string
	}{
		{"lark_im_send", "lark_im_card_send"},
		{"lark_mail_send", "lark_mail_draft_create"},
		{"lark_calendar_agenda", "lark_vc_search"},
	}

	byName := map[string]tool{}
	for _, tl := range allTools() {
		byName[tl.Name] = tl
	}

	for _, p := range pairs {
		t.Run(p.a+"_mentions_"+p.b, func(t *testing.T) {
			tl, ok := byName[p.a]
			if !ok {
				t.Fatalf("tool %q not found", p.a)
			}
			if !strings.Contains(tl.Description, p.b) {
				t.Errorf("description of %q must mention %q to disambiguate; got:\n  %s",
					p.a, p.b, tl.Description)
			}
		})
		t.Run(p.b+"_mentions_"+p.a, func(t *testing.T) {
			tl, ok := byName[p.b]
			if !ok {
				t.Fatalf("tool %q not found", p.b)
			}
			if !strings.Contains(tl.Description, p.a) {
				t.Errorf("description of %q must mention %q to disambiguate; got:\n  %s",
					p.b, p.a, tl.Description)
			}
		})
	}
}

// TestMutatingToolDescriptionsMentionSafety enforces that tools which
// have side effects (send, create, append, upload) name their safety
// mechanism in the description — so the model surfaces it to the user
// rather than blindly committing.
//
// Safety mechanism varies per tool:
//   - mail_send: confirm_send flag (primary safety)
//   - mail_draft_create: "never reach a recipient"
//   - all other mutating tools: dry_run flag — now uniformly first-class
//     after T5-shortcut-PR shipped 2026-05-28 (all underlying shortcuts
//     implement DryRun; MCP layer exposes dry_run schema arg).
func TestMutatingToolDescriptionsMentionSafety(t *testing.T) {
	expectations := map[string]string{
		"lark_mail_send":         "confirm_send",
		"lark_mail_draft_create": "draft",
		"lark_calendar_create":   "dry_run",
		"lark_drive_upload":      "dry_run",
		"lark_api":               "dry_run",
		"lark_task_create":       "dry_run",
		"lark_doc_create":        "dry_run",
		"lark_sheets_append":     "dry_run",
		"lark_im_send":           "dry_run",
		"lark_im_card_send":      "dry_run",
	}

	byName := map[string]tool{}
	for _, tl := range allTools() {
		byName[tl.Name] = tl
	}

	for name, keyword := range expectations {
		t.Run(name, func(t *testing.T) {
			tl, ok := byName[name]
			if !ok {
				t.Fatalf("tool %q not found", name)
			}
			if !strings.Contains(strings.ToLower(tl.Description), strings.ToLower(keyword)) {
				t.Errorf("mutating tool %q description must mention safety mechanism %q; got:\n  %s",
					name, keyword, tl.Description)
			}
		})
	}
}
