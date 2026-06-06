// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package mcp

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestClassifyErrorParsesLarkCLIEnvelope verifies the happy path: a
// well-formed lark-cli error JSON on stderr produces a typed classification.
//
// Sample envelope captured 2026-05-28 from `lark-cli docs +fetch --doc fake`.
func TestClassifyErrorParsesLarkCLIEnvelope(t *testing.T) {
	stderr := `{
  "ok": false,
  "identity": "user",
  "error": {
    "type": "api_error",
    "code": 1,
    "message": "Internal error. Please retry.",
    "detail": {"log_id": "202605282240122A30D3BDA7B93D7A23BC"}
  }
}`
	c := classifyError(stderr, 1)
	if c.Type != "api_error" {
		t.Errorf("Type=%q want api_error", c.Type)
	}
	if c.Message != "Internal error. Please retry." {
		t.Errorf("Message=%q want %q", c.Message, "Internal error. Please retry.")
	}
	if c.Code != 1 {
		t.Errorf("Code=%d want 1", c.Code)
	}
	if c.LogID != "202605282240122A30D3BDA7B93D7A23BC" {
		t.Errorf("LogID=%q (expected the log_id from detail)", c.LogID)
	}
	if !strings.Contains(c.RetryAdvice, "log_id") {
		t.Errorf("RetryAdvice should mention log_id since one was captured; got %q", c.RetryAdvice)
	}
}

// TestClassifyErrorValidationType — `validation` errors from the
// shortcut layer indicate user-fixable arg problems. Agent must NOT
// retry blindly.
func TestClassifyErrorValidationType(t *testing.T) {
	stderr := `{"ok":false,"identity":"user","error":{"type":"validation","message":"specify at least one of --query, --user-ids"}}`
	c := classifyError(stderr, 1)
	if c.Type != "validation" {
		t.Errorf("Type=%q want validation", c.Type)
	}
	if !strings.Contains(strings.ToLower(c.RetryAdvice), "do not retry") {
		t.Errorf("RetryAdvice should warn against blind retry; got %q", c.RetryAdvice)
	}
}

// TestClassifyErrorHeuristics covers each pattern-based bucket. Each
// row uses a synthesised message that triggers a specific branch.
func TestClassifyErrorHeuristics(t *testing.T) {
	cases := []struct {
		name     string
		message  string
		wantType string
		wantHint string // substring expected in retry_advice
	}{
		{"token_expired", "access_token expired", "auth_expired", "auth login"},
		{"unauthorized_401", "401 unauthorized", "auth_expired", "auth login"},
		{"forbidden_scope", "permission denied: missing scope im:message", "permission_denied", "scope"},
		{"forbidden_403", "403 forbidden", "permission_denied", "scope"},
		{"rate_429", "429 too many requests", "rate_limit", "wait"},
		{"rate_throttled", "request throttled, slow down", "rate_limit", "wait"},
		{"network_timeout", "connection timeout reaching open.feishu.cn", "network", "retry"},
		{"network_refused", "connection refused", "network", "retry"},
		{"generic_api_no_logid", "something else went wrong", "api_error", "generic"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stderr := `{"ok":false,"error":{"type":"api_error","message":"` + tc.message + `"}}`
			c := classifyError(stderr, 1)
			if c.Type != tc.wantType {
				t.Errorf("Type=%q want %q (msg=%q)", c.Type, tc.wantType, tc.message)
			}
			if !strings.Contains(strings.ToLower(c.RetryAdvice), tc.wantHint) {
				t.Errorf("RetryAdvice=%q expected to contain %q", c.RetryAdvice, tc.wantHint)
			}
		})
	}
}

// TestClassifyErrorNonJSONStderr — if stderr is plain text (legacy
// shortcuts, panics, OS errors), classifier degrades gracefully to
// "unknown" with raw preserved.
func TestClassifyErrorNonJSONStderr(t *testing.T) {
	c := classifyError("fatal: out of memory\nstack trace ...", 137)
	if c.Type != "unknown" {
		t.Errorf("Type=%q want unknown for non-JSON stderr", c.Type)
	}
	if c.Message != "fatal: out of memory" {
		t.Errorf("Message=%q want first line of raw", c.Message)
	}
	if c.Raw == "" {
		t.Errorf("Raw should preserve original stderr")
	}
}

// TestClassifyErrorEmptyStderr — non-zero exit with empty stderr is a
// degenerate case (process killed, OOM, etc). Classifier provides a
// sensible default.
func TestClassifyErrorEmptyStderr(t *testing.T) {
	c := classifyError("", -1)
	if c.Type != "unknown" {
		t.Errorf("Type=%q want unknown", c.Type)
	}
	if c.Message == "" {
		t.Errorf("Message should have a fallback explanation")
	}
}

// TestFormatResultEmitsClassifiedJSON locks the wire contract: error
// tool results carry classification JSON in content[0].text. Agents
// switch on error_type without grep'ing stderr.
func TestFormatResultEmitsClassifiedJSON(t *testing.T) {
	res := execResult{
		Stderr:   `{"ok":false,"error":{"type":"validation","message":"missing --query"}}`,
		ExitCode: 1,
	}
	out := formatResult(res)
	if !out.IsError {
		t.Fatalf("IsError=false; want true for non-zero exit")
	}
	if len(out.Content) == 0 {
		t.Fatal("Content empty")
	}
	var got errorClassification
	if err := json.Unmarshal([]byte(out.Content[0].Text), &got); err != nil {
		t.Fatalf("content[0].text should be JSON classification; got: %q\n  err: %v", out.Content[0].Text, err)
	}
	if got.Type != "validation" {
		t.Errorf("Type=%q want validation", got.Type)
	}
	if !strings.Contains(got.RetryAdvice, "do not retry") {
		t.Errorf("RetryAdvice should warn against blind retry on validation; got %q", got.RetryAdvice)
	}
	if got.Raw == "" {
		t.Errorf("Raw stderr should be preserved")
	}
}

// TestFormatResultSuccessUnchanged guards that successful (exit 0)
// tool results are NOT wrapped — the content block is still raw stdout.
// Agents reading happy-path output don't pay a JSON-parse cost.
func TestFormatResultSuccessUnchanged(t *testing.T) {
	res := execResult{Stdout: `{"events": [{"id": "evt_1"}]}`, ExitCode: 0}
	out := formatResult(res)
	if out.IsError {
		t.Errorf("IsError=true for exit 0")
	}
	if out.Content[0].Text != `{"events": [{"id": "evt_1"}]}` {
		t.Errorf("Success path mutated stdout: %q", out.Content[0].Text)
	}
}

// TestFormatResultFallbackToStdoutWhenStderrEmpty — pathological case:
// non-zero exit, stderr empty, stdout has text. Classifier should pull
// from stdout so the agent still sees something useful.
func TestFormatResultFallbackToStdoutWhenStderrEmpty(t *testing.T) {
	res := execResult{Stdout: "oops something printed to stdout instead", Stderr: "", ExitCode: 2}
	out := formatResult(res)
	if !out.IsError {
		t.Errorf("IsError=false; want true")
	}
	if !strings.Contains(out.Content[0].Text, "oops") {
		t.Errorf("content should contain stdout fallback; got %q", out.Content[0].Text)
	}
}
