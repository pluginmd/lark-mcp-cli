// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package mcp

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestAllToolsHaveValidJSONSchema verifies every tool's Schema parses.
// Mirrors the runtime check in descriptors(); having it as a unit test
// catches bad schemas before the host model ever sees them.
func TestAllToolsHaveValidJSONSchema(t *testing.T) {
	for _, tl := range allTools() {
		t.Run(tl.Name, func(t *testing.T) {
			var probe interface{}
			if err := json.Unmarshal([]byte(tl.Schema), &probe); err != nil {
				t.Fatalf("schema does not parse: %v", err)
			}
		})
	}
}

// TestReadToolsExposeJq locks in the contract: every read-type tool
// declares a `jq` property in its schema AND its Build closure honors
// the arg by appending `--jq <expr>` to the argv. This is the central
// token-reduction lever; regressing on it would silently bloat agent
// context costs.
func TestReadToolsExposeJq(t *testing.T) {
	readToolNames := []string{
		"lark_im_search",
		"lark_calendar_agenda",
		"lark_doc_search",
		"lark_doc_fetch",
		"lark_base_search",
		"lark_contact_search",
		"lark_task_my",
		"lark_sheets_read",
		"lark_vc_search",
		"lark_minutes_search",
		"lark_okr_cycle_list",
		"lark_api",
	}

	byName := map[string]tool{}
	for _, tl := range allTools() {
		byName[tl.Name] = tl
	}

	for _, name := range readToolNames {
		t.Run(name+"/schema_has_jq", func(t *testing.T) {
			tl, ok := byName[name]
			if !ok {
				t.Fatalf("tool %q not in allTools()", name)
			}
			// Parse schema and check `properties.jq` exists.
			var schema struct {
				Properties map[string]json.RawMessage `json:"properties"`
			}
			if err := json.Unmarshal([]byte(tl.Schema), &schema); err != nil {
				t.Fatalf("schema parse: %v", err)
			}
			if _, present := schema.Properties["jq"]; !present {
				t.Errorf("tool %q schema missing `jq` property", name)
			}
		})

		t.Run(name+"/build_passes_jq_flag", func(t *testing.T) {
			tl := byName[name]
			args := minimalArgsFor(name)
			args["jq"] = ".items[] | {id}"
			argv, err := tl.Build(args)
			if err != nil {
				t.Fatalf("Build returned err: %v", err)
			}
			if !containsPair(argv, "--jq", ".items[] | {id}") {
				t.Errorf("tool %q Build did not append --jq; got argv=%v", name, argv)
			}
		})

		t.Run(name+"/build_omits_jq_when_absent", func(t *testing.T) {
			tl := byName[name]
			args := minimalArgsFor(name)
			argv, err := tl.Build(args)
			if err != nil {
				t.Fatalf("Build returned err: %v", err)
			}
			if containsFlag(argv, "--jq") {
				t.Errorf("tool %q Build appended --jq without input; got argv=%v", name, argv)
			}
		})
	}
}

// TestMutatingToolsExposeDryRun checks the safety pattern: tools that
// wrap shortcuts with DryRun() implemented MUST expose dry_run in their
// schema and pass --dry-run when set. This is the maker-checker
// enforcement at the MCP layer.
//
// Scope: all 8 mutating MCP tools whose underlying shortcut implements
// DryRun. Verified empirically 2026-05-28 by running `--dry-run` on
// each underlying CLI verb.
//
// lark_mail_send uses confirm_send=false→draft as its primary safety
// pattern (predates dry_run) and is excluded from this check, though
// the mail +send shortcut also supports --dry-run separately.
func TestMutatingToolsExposeDryRun(t *testing.T) {
	mutatingWithDryRun := []string{
		"lark_calendar_create",
		"lark_drive_upload",
		"lark_api",
		"lark_task_create",
		"lark_doc_create",
		"lark_sheets_append",
		"lark_im_send",
		"lark_im_card_send",
	}

	byName := map[string]tool{}
	for _, tl := range allTools() {
		byName[tl.Name] = tl
	}

	for _, name := range mutatingWithDryRun {
		t.Run(name+"/schema_has_dry_run", func(t *testing.T) {
			tl, ok := byName[name]
			if !ok {
				t.Fatalf("tool %q not in allTools()", name)
			}
			var schema struct {
				Properties map[string]json.RawMessage `json:"properties"`
			}
			if err := json.Unmarshal([]byte(tl.Schema), &schema); err != nil {
				t.Fatalf("schema parse: %v", err)
			}
			if _, present := schema.Properties["dry_run"]; !present {
				t.Errorf("tool %q schema missing `dry_run` property", name)
			}
		})

		t.Run(name+"/build_passes_dry_run_flag", func(t *testing.T) {
			tl := byName[name]
			args := minimalArgsFor(name)
			args["dry_run"] = true
			argv, err := tl.Build(args)
			if err != nil {
				t.Fatalf("Build returned err: %v", err)
			}
			if !containsFlag(argv, "--dry-run") {
				t.Errorf("tool %q Build did not append --dry-run; got argv=%v", name, argv)
			}
		})

		t.Run(name+"/build_omits_dry_run_when_false", func(t *testing.T) {
			tl := byName[name]
			args := minimalArgsFor(name)
			args["dry_run"] = false
			argv, err := tl.Build(args)
			if err != nil {
				t.Fatalf("Build returned err: %v", err)
			}
			if containsFlag(argv, "--dry-run") {
				t.Errorf("tool %q Build appended --dry-run for false value; argv=%v", name, argv)
			}
		})
	}
}

// TestAppendJqHelper exercises the helper directly.
func TestAppendJqHelper(t *testing.T) {
	cases := []struct {
		name     string
		args     map[string]interface{}
		wantFlag bool
		wantExpr string
	}{
		{"absent", map[string]interface{}{}, false, ""},
		{"empty string", map[string]interface{}{"jq": ""}, false, ""},
		{"nil value", map[string]interface{}{"jq": nil}, false, ""},
		{"valid expr", map[string]interface{}{"jq": ".items[]"}, true, ".items[]"},
		{"complex expr", map[string]interface{}{"jq": ".a | .b[0:5] | {c, d}"}, true, ".a | .b[0:5] | {c, d}"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			argv := []string{"some", "+cmd"}
			argv = appendJq(argv, tc.args)
			has := containsFlag(argv, "--jq")
			if has != tc.wantFlag {
				t.Errorf("--jq present=%v, want %v; argv=%v", has, tc.wantFlag, argv)
			}
			if tc.wantFlag {
				if !containsPair(argv, "--jq", tc.wantExpr) {
					t.Errorf("--jq value mismatch; argv=%v want value=%q", argv, tc.wantExpr)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------
// Test helpers.
// ---------------------------------------------------------------------

// minimalArgsFor returns the smallest arg map that lets each tool's
// Build closure run without a "required" error. We're testing arg
// pass-through, not Lark business logic.
func minimalArgsFor(toolName string) map[string]interface{} {
	switch toolName {
	case "lark_im_send":
		return map[string]interface{}{"chat_id": "oc_test", "text": "hi"}
	case "lark_im_card_send":
		return map[string]interface{}{"chat_id": "oc_test", "spec": "header:\n  title: t"}
	case "lark_im_search":
		return map[string]interface{}{"query": "test"}
	case "lark_mail_send":
		return map[string]interface{}{"to": "a@b.com", "subject": "s", "body": "b"}
	case "lark_mail_draft_create":
		return map[string]interface{}{"subject": "s", "body": "b"}
	case "lark_calendar_agenda":
		return map[string]interface{}{}
	case "lark_calendar_create":
		return map[string]interface{}{"summary": "m", "start": "2026-01-01T09:00:00Z", "end": "2026-01-01T10:00:00Z"}
	case "lark_doc_create":
		return map[string]interface{}{"title": "t"}
	case "lark_doc_search":
		return map[string]interface{}{"query": "test"}
	case "lark_doc_fetch":
		return map[string]interface{}{"doc": "doccnXXX"}
	case "lark_base_search":
		return map[string]interface{}{"base_token": "bascnXXX", "table_id": "tblXXX", "query": "test"}
	case "lark_contact_search":
		return map[string]interface{}{"query": "alice"}
	case "lark_task_my":
		return map[string]interface{}{}
	case "lark_task_create":
		return map[string]interface{}{"summary": "t"}
	case "lark_drive_upload":
		return map[string]interface{}{"file": "/tmp/x.txt"}
	case "lark_sheets_read":
		return map[string]interface{}{"spreadsheet_token": "shtXXX", "range": "Sheet1!A1:B2"}
	case "lark_sheets_append":
		return map[string]interface{}{"spreadsheet_token": "shtXXX", "values": "[[\"a\",1]]"}
	case "lark_vc_search":
		return map[string]interface{}{}
	case "lark_minutes_search":
		return map[string]interface{}{}
	case "lark_okr_cycle_list":
		return map[string]interface{}{"user_id": "me"}
	case "lark_api":
		return map[string]interface{}{"method": "GET", "path": "/open-apis/test"}
	default:
		return map[string]interface{}{}
	}
}

// containsFlag reports whether argv contains the literal flag token.
func containsFlag(argv []string, flag string) bool {
	for _, s := range argv {
		if s == flag {
			return true
		}
	}
	return false
}

// containsPair reports whether argv contains `flag value` as adjacent tokens.
// More strict than containsFlag — catches the case where a flag is present
// but the value got dropped or paired with the wrong arg.
func containsPair(argv []string, flag, value string) bool {
	for i := 0; i < len(argv)-1; i++ {
		if argv[i] == flag && argv[i+1] == value {
			return true
		}
	}
	return false
}

// Lock the spec: appendJq must NOT mutate the original slice header.
// Future refactor could reach for `argv = append(argv, ...)` and break
// this if argv has spare capacity from upstream — keep the contract loud.
func TestAppendJqDoesNotMutateOriginalSlice(t *testing.T) {
	// Build a slice with extra capacity so an in-place append could shadow
	// the original backing array.
	original := make([]string, 2, 10)
	original[0] = "im"
	original[1] = "+messages-send"

	// Call appendJq with input. The returned slice MAY share backing array
	// (Go convention), but the original's len must remain 2.
	_ = appendJq(original, map[string]interface{}{"jq": ".x"})

	if len(original) != 2 {
		t.Errorf("appendJq mutated original slice length: got %d, want 2", len(original))
	}
	if strings.Join(original, " ") != "im +messages-send" {
		t.Errorf("appendJq mutated original slice contents: %v", original)
	}
}
