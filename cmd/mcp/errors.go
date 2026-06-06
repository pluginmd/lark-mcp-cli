// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package mcp

import (
	"encoding/json"
	"strings"
)

// errorClassification is the structured representation of a tool failure
// surfaced to the MCP host. The host model switches on `error_type` to
// decide whether to retry, escalate, or abort — without grep'ing raw
// stderr text.
//
// Shape contract (stable):
//
//	{
//	  "error_type": "validation|auth_expired|rate_limit|permission_denied|api_error|network|unknown",
//	  "message":    "<human-readable summary>",
//	  "retry_advice": "<actionable next step for the agent>",
//	  "code":       <int, optional>,
//	  "log_id":     "<string, optional>",
//	  "raw":        "<original stderr>"
//	}
type errorClassification struct {
	Type        string `json:"error_type"`
	Message     string `json:"message"`
	RetryAdvice string `json:"retry_advice"`
	Code        int    `json:"code,omitempty"`
	LogID       string `json:"log_id,omitempty"`
	Raw         string `json:"raw"`
}

// larkCLIErrorEnvelope matches the JSON shape that lark-cli emits on
// stderr in failure paths (verified via empirical inspection 2026-05-28
// of `docs +fetch`, `contact +search-user`, `api GET` etc).
type larkCLIErrorEnvelope struct {
	OK       bool   `json:"ok"`
	Identity string `json:"identity,omitempty"`
	Error    struct {
		Type    string                 `json:"type"`
		Code    int                    `json:"code,omitempty"`
		Message string                 `json:"message"`
		Detail  map[string]interface{} `json:"detail,omitempty"`
	} `json:"error"`
}

// classifyError takes a tool's stderr (whatever shortcuts wrote there)
// plus its exit code and returns a structured classification.
//
// If stderr parses as the lark-cli error envelope, fields are extracted
// and we use heuristics on Error.Type + Error.Message to assign one of
// the well-known classification types. Otherwise the type defaults to
// "unknown" and Raw carries the unparsed text.
//
// The classifier is intentionally conservative: when in doubt, default
// to "api_error" with a generic retry advice. False positives in the
// classifier are worse than no classification, because the agent will
// act on the wrong advice.
func classifyError(stderr string, exitCode int) errorClassification {
	c := errorClassification{
		Type: "unknown",
		Raw:  stderr,
	}

	trimmed := strings.TrimSpace(stderr)
	if trimmed == "" {
		c.Message = "tool exited with non-zero status but produced no stderr"
		c.RetryAdvice = "inspect the tool name and arguments; if correct, retry once"
		return c
	}

	var env larkCLIErrorEnvelope
	if err := json.Unmarshal([]byte(trimmed), &env); err != nil || env.Error.Type == "" {
		// Not the lark-cli envelope — leave as unknown but populate
		// message from the raw text so the agent has something to read.
		c.Message = firstLine(trimmed)
		c.RetryAdvice = "stderr was not structured JSON; read raw for details"
		return c
	}

	c.Message = env.Error.Message
	c.Code = env.Error.Code
	if logID, ok := env.Error.Detail["log_id"].(string); ok {
		c.LogID = logID
	}

	// Classification heuristics. Order matters — more specific tests
	// run before generic api_error fallthrough.
	lowerMsg := strings.ToLower(env.Error.Message)

	switch {
	case env.Error.Type == "validation":
		c.Type = "validation"
		c.RetryAdvice = "fix the argument(s) called out in the message; do not retry with the same args"

	case containsAny(lowerMsg, "token", "unauthorized", "401", "access_token", "auth"):
		c.Type = "auth_expired"
		c.RetryAdvice = "run `lark-cli auth login` and ask the user to re-authenticate before retrying"

	case containsAny(lowerMsg, "permission", "forbidden", "403", "scope"):
		c.Type = "permission_denied"
		c.RetryAdvice = "check OAuth scopes; try the call with as=\"bot\" if the user identity lacks the scope"

	case containsAny(lowerMsg, "rate", "429", "throttl", "too many"):
		c.Type = "rate_limit"
		c.RetryAdvice = "wait at least 5 seconds before retry; for repeated 429s, batch fewer calls per second"

	case containsAny(lowerMsg, "network", "timeout", "connection refused", "no such host"):
		c.Type = "network"
		c.RetryAdvice = "retry once after a brief delay; if persistent, surface to user (Lark API may be down)"

	default:
		c.Type = "api_error"
		if c.LogID != "" {
			c.RetryAdvice = "Lark-side error (log_id captured); retry once with backoff, escalate to user with log_id if it recurs"
		} else {
			c.RetryAdvice = "generic API failure; retry once, then surface message to user"
		}
	}

	return c
}

// containsAny is a small ergonomic helper for the classifier's string
// substring checks — keeps the switch statement readable.
func containsAny(haystack string, needles ...string) bool {
	for _, n := range needles {
		if strings.Contains(haystack, n) {
			return true
		}
	}
	return false
}

// firstLine returns the first newline-delimited line of s, used as a
// fallback message when the stderr was not structured.
func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}
