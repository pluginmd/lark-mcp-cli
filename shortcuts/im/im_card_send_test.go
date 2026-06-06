// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package im

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/larksuite/cli/shortcuts/common"
	"github.com/spf13/cobra"
)

// newCardSendRuntime wires up the same RuntimeContext shape ImCardSend
// receives when invoked via cobra at runtime. Uses newBotShortcutRuntime
// (the canonical pattern in helpers_network_test.go) so SDK/auth/IOStreams
// match production. Returns the runtime plus the stdout buffer for
// downstream assertions.
func newCardSendRuntime(t *testing.T, stringFlags map[string]string, boolFlags map[string]bool, rt http.RoundTripper) (*common.RuntimeContext, *bytes.Buffer) {
	t.Helper()
	if rt == nil {
		// Fail-loud default: any unintended network call during a
		// --print-json or validation test should surface, not hang.
		rt = shortcutRoundTripFunc(func(req *http.Request) (*http.Response, error) {
			t.Fatalf("unexpected network call to %s — print-json/validate must not hit the network", req.URL)
			return nil, nil
		})
	}
	runtime := newBotShortcutRuntime(t, rt)

	cmd := &cobra.Command{Use: "test"}
	for _, name := range []string{"chat-id", "user-id", "reply-to", "spec", "idempotency-key"} {
		cmd.Flags().String(name, "", "")
	}
	for _, name := range []string{"in-thread", "print-json", "dry-run"} {
		cmd.Flags().Bool(name, false, "")
	}
	if err := cmd.ParseFlags(nil); err != nil {
		t.Fatalf("ParseFlags() error = %v", err)
	}
	for name, val := range stringFlags {
		if err := cmd.Flags().Set(name, val); err != nil {
			t.Fatalf("Flags().Set(%q) error = %v", name, err)
		}
	}
	for name, val := range boolFlags {
		boolStr := "false"
		if val {
			boolStr = "true"
		}
		if err := cmd.Flags().Set(name, boolStr); err != nil {
			t.Fatalf("Flags().Set(%q) error = %v", name, err)
		}
	}
	runtime.Cmd = cmd

	outBuf, _ := runtime.Factory.IOStreams.Out.(*bytes.Buffer)
	return runtime, outBuf
}

// TestImCardSend_PrintJSONIntegration drives ImCardSend through Validate
// then Execute (with --print-json) and asserts the emitted JSON shape.
// This is the canonical AGENTS.md-mandated shortcut-level test: real
// RuntimeContext, real Cobra flag plumbing, no network. If anything in
// the shortcut → Compile pipeline drifts, this test catches it.
func TestImCardSend_PrintJSONIntegration(t *testing.T) {
	spec := `header:
  title: "Integration test"
  template: blue
  title_i18n: { zh_cn: "集成测试" }
summary: "Pipeline check"
elements:
  - md: "# Should demote"
  - div:
      text: "Build green"
      text_color: green
      icon: standard:check-circle_filled
      icon_color: green
  - panel:
      title: "Section"
      border: green
      expanded: true
      elements:
        - md: "panel body"
`
	runtime, out := newCardSendRuntime(t, map[string]string{
		"spec": spec,
	}, map[string]bool{
		"print-json": true,
	}, nil)

	if err := ImCardSend.Validate(context.Background(), runtime); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if err := ImCardSend.Execute(context.Background(), runtime); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// stdout carries the wrapped envelope; assert structure end-to-end.
	var env struct {
		OK   bool                   `json:"ok"`
		Data map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\n%s", err, out.String())
	}
	if !env.OK {
		t.Fatalf("envelope ok=false: %s", out.String())
	}
	card := env.Data
	if card["schema"] != "2.0" {
		t.Errorf("schema = %v, want 2.0", card["schema"])
	}

	cfg := card["config"].(map[string]interface{})
	locales, ok := cfg["locales"].([]interface{})
	if !ok || len(locales) != 1 || locales[0] != "zh_cn" {
		t.Errorf("auto-derived locales = %#v, want [zh_cn]", cfg["locales"])
	}

	body := card["body"].(map[string]interface{})
	elements := body["elements"].([]interface{})
	if len(elements) != 3 {
		t.Fatalf("element count = %d, want 3", len(elements))
	}
	mdContent := elements[0].(map[string]interface{})["content"].(string)
	if !strings.HasPrefix(mdContent, "#### Should demote") {
		t.Errorf("md H1 not demoted, got: %q", mdContent)
	}
	div := elements[1].(map[string]interface{})
	if div["tag"] != "div" {
		t.Errorf("element[1].tag = %v, want div", div["tag"])
	}
	panel := elements[2].(map[string]interface{})
	if panel["tag"] != "collapsible_panel" {
		t.Errorf("element[2].tag = %v, want collapsible_panel", panel["tag"])
	}
}

// TestImCardSend_ValidateRejectsBadSpec confirms YAML-strict mode bites
// at Validate time: an unknown top-level field (typo) errors before
// Execute runs. This is the regression guard for the strict-mode change
// — without it a `templat:` typo would silently render a default-coloured
// header at Feishu's end.
func TestImCardSend_ValidateRejectsBadSpec(t *testing.T) {
	tests := []struct {
		name      string
		spec      string
		wantInErr string
	}{
		{
			name: "unknown top-level field (typo)",
			spec: `templat: blue
elements: [{ md: x }]
`,
			wantInErr: "templat",
		},
		{
			name: "unknown header field (typo)",
			spec: `header:
  titl: "Hello"
elements: [{ md: x }]
`,
			wantInErr: "titl",
		},
		{
			name: "invalid image key prefix",
			spec: `elements:
  - image: { key: "not-an-img-key", alt: "x" }
`,
			wantInErr: "must start with 'img_'",
		},
		{
			name: "invalid icon spec — no recognised prefix",
			spec: `header:
  title: H
  icon: "bell"
elements: [{ md: x }]
`,
			wantInErr: "no recognised prefix",
		},
		{
			name: "custom icon without img_ prefix",
			spec: `header:
  title: H
  icon: "custom:abc123"
elements: [{ md: x }]
`,
			wantInErr: "must start with 'img_'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runtime, _ := newCardSendRuntime(t, map[string]string{"spec": tt.spec}, map[string]bool{"print-json": true}, nil)
			err := ImCardSend.Validate(context.Background(), runtime)
			if err == nil {
				t.Fatalf("Validate() error = nil, want error containing %q", tt.wantInErr)
			}
			if !strings.Contains(err.Error(), tt.wantInErr) {
				t.Errorf("Validate() error = %v, want substring %q", err, tt.wantInErr)
			}
		})
	}
}

// TestImCardSend_DryRunReplyPath ensures the DryRun hook produces the
// reply API shape (different endpoint than the create path) and that
// --in-thread sets reply_in_thread on the body.
func TestImCardSend_DryRunReplyPath(t *testing.T) {
	spec := `elements:
  - md: "thread reply"
`
	runtime, _ := newCardSendRuntime(t, map[string]string{
		"spec":     spec,
		"reply-to": "om_abc123",
	}, map[string]bool{
		"in-thread": true,
	}, nil)

	if err := ImCardSend.Validate(context.Background(), runtime); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	dryRun := ImCardSend.DryRun(context.Background(), runtime)
	if dryRun == nil {
		t.Fatal("DryRun returned nil")
	}
	got := mustMarshalDryRun(t, dryRun)
	if !strings.Contains(got, "/open-apis/im/v1/messages/om_abc123/reply") {
		t.Errorf("dry-run URL missing reply endpoint: %s", got)
	}
	if !strings.Contains(got, `"reply_in_thread":true`) {
		t.Errorf("dry-run body missing reply_in_thread=true: %s", got)
	}
	if !strings.Contains(got, `"msg_type":"interactive"`) {
		t.Errorf("dry-run body missing interactive msg_type: %s", got)
	}
}

// TestImCardSend_DryRunUserIDPath covers the open_id branch of DryRun
// (--user-id instead of --chat-id) so receive_id_type=open_id is
// exercised — previously only chat_id path was tested.
func TestImCardSend_DryRunUserIDPath(t *testing.T) {
	spec := `elements:
  - md: "DM body"
`
	runtime, _ := newCardSendRuntime(t, map[string]string{
		"spec":    spec,
		"user-id": "ou_aabbccddeeff00112233445566778899",
	}, nil, nil)

	if err := ImCardSend.Validate(context.Background(), runtime); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	got := mustMarshalDryRun(t, ImCardSend.DryRun(context.Background(), runtime))
	if !strings.Contains(got, `"receive_id_type":"open_id"`) {
		t.Errorf("dry-run missing receive_id_type=open_id: %s", got)
	}
	if !strings.Contains(got, `"receive_id":"ou_aabbccddeeff00112233445566778899"`) {
		t.Errorf("dry-run missing user open_id in receive_id: %s", got)
	}
}

// TestImCardSend_ExecuteSendSuccess drives Execute() through the real
// HTTP code path with a mocked transport. Covers the chat-id create
// branch and the runtime.Out() response shaping. Previously
// uncovered (the integration test exercised only --print-json).
func TestImCardSend_ExecuteSendSuccess(t *testing.T) {
	var capturedPath string
	var capturedBody map[string]any
	rt := shortcutRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		capturedPath = req.URL.Path
		raw, _ := io.ReadAll(req.Body)
		_ = json.Unmarshal(raw, &capturedBody)
		return shortcutJSONResponse(200, map[string]any{
			"code": 0,
			"data": map[string]any{
				"message_id":  "om_response_id",
				"chat_id":     "oc_test_chat",
				"create_time": "1716800000",
			},
		}), nil
	})

	spec := `elements:
  - md: "hello from execute"
`
	runtime, out := newCardSendRuntime(t, map[string]string{
		"spec":            spec,
		"chat-id":         "oc_test_chat",
		"idempotency-key": "exec-test-uuid",
	}, nil, rt)

	if err := ImCardSend.Validate(context.Background(), runtime); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if err := ImCardSend.Execute(context.Background(), runtime); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if capturedPath != "/open-apis/im/v1/messages" {
		t.Errorf("captured POST path = %q, want /open-apis/im/v1/messages", capturedPath)
	}
	if capturedBody["msg_type"] != "interactive" {
		t.Errorf("captured msg_type = %v, want interactive", capturedBody["msg_type"])
	}
	if capturedBody["uuid"] != "exec-test-uuid" {
		t.Errorf("idempotency-key did not propagate to body.uuid: %v", capturedBody["uuid"])
	}
	if capturedBody["receive_id"] != "oc_test_chat" {
		t.Errorf("receive_id = %v, want oc_test_chat", capturedBody["receive_id"])
	}
	// The compiled card JSON is embedded as the `content` string —
	// re-parse it to confirm the schema 2.0 shape made it through.
	contentJSON, _ := capturedBody["content"].(string)
	var inner map[string]any
	if err := json.Unmarshal([]byte(contentJSON), &inner); err != nil {
		t.Fatalf("body.content is not valid JSON: %v\n%s", err, contentJSON)
	}
	if inner["schema"] != "2.0" {
		t.Errorf("posted card not schema 2.0: %#v", inner)
	}

	// stdout envelope carries the success summary.
	var env struct {
		OK   bool                   `json:"ok"`
		Data map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("stdout not JSON: %v\n%s", err, out.String())
	}
	if env.Data["message_id"] != "om_response_id" {
		t.Errorf("stdout missing response message_id: %#v", env.Data)
	}
}

// TestImCardSend_ExecuteReplyInThread covers the --reply-to + --in-thread
// branch of Execute against a mocked transport. Confirms reply_in_thread
// is set in the body and the URL is the reply endpoint.
func TestImCardSend_ExecuteReplyInThread(t *testing.T) {
	var capturedPath string
	var capturedBody map[string]any
	rt := shortcutRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		capturedPath = req.URL.Path
		raw, _ := io.ReadAll(req.Body)
		_ = json.Unmarshal(raw, &capturedBody)
		return shortcutJSONResponse(200, map[string]any{
			"code": 0,
			"data": map[string]any{
				"message_id":  "om_reply_id",
				"chat_id":     "oc_test_chat",
				"create_time": "1716800001",
			},
		}), nil
	})

	spec := `elements:
  - md: "thread reply body"
`
	runtime, _ := newCardSendRuntime(t, map[string]string{
		"spec":     spec,
		"reply-to": "om_original_msg",
	}, map[string]bool{
		"in-thread": true,
	}, rt)

	if err := ImCardSend.Validate(context.Background(), runtime); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if err := ImCardSend.Execute(context.Background(), runtime); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if capturedPath != "/open-apis/im/v1/messages/om_original_msg/reply" {
		t.Errorf("captured path = %q, want reply endpoint", capturedPath)
	}
	if capturedBody["reply_in_thread"] != true {
		t.Errorf("body.reply_in_thread = %v, want true", capturedBody["reply_in_thread"])
	}
	if _, has := capturedBody["receive_id"]; has {
		t.Errorf("reply path should NOT set receive_id: %#v", capturedBody)
	}
}

// TestImCardSend_ExecutePropagatesAPIError ensures network errors
// surface to the caller rather than being silently swallowed.
func TestImCardSend_ExecutePropagatesAPIError(t *testing.T) {
	rt := shortcutRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		return shortcutJSONResponse(500, map[string]any{
			"code": 99999,
			"msg":  "internal server error",
		}), nil
	})
	spec := `elements:
  - md: "trigger error"
`
	runtime, _ := newCardSendRuntime(t, map[string]string{
		"spec":    spec,
		"chat-id": "oc_test_chat",
	}, nil, rt)

	if err := ImCardSend.Validate(context.Background(), runtime); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	err := ImCardSend.Execute(context.Background(), runtime)
	if err == nil {
		t.Fatal("Execute() error = nil, want error from API 500")
	}
}

// TestImCardSend_DryRunSendPath covers the create-message branch of
// DryRun (no --reply-to). Asserts the API path, params, and that the
// idempotency-key flows through as `uuid` in the body — previously
// untested.
func TestImCardSend_DryRunSendPath(t *testing.T) {
	spec := `elements:
  - md: "fresh message"
`
	runtime, _ := newCardSendRuntime(t, map[string]string{
		"spec":            spec,
		"chat-id":         "oc_111111111111111111111111111111",
		"idempotency-key": "stable-uuid-2026",
	}, nil, nil)

	if err := ImCardSend.Validate(context.Background(), runtime); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	got := mustMarshalDryRun(t, ImCardSend.DryRun(context.Background(), runtime))
	if !strings.Contains(got, "/open-apis/im/v1/messages") {
		t.Errorf("dry-run missing create endpoint: %s", got)
	}
	if !strings.Contains(got, `"receive_id_type":"chat_id"`) {
		t.Errorf("dry-run missing receive_id_type param: %s", got)
	}
	if !strings.Contains(got, `"uuid":"stable-uuid-2026"`) {
		t.Errorf("idempotency-key did not flow through to body.uuid: %s", got)
	}
}
