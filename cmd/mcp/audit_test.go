// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestAuditLoggerNilIsNoop locks in the API contract: callers don't
// have to nil-check before calling Log/Close. This keeps the dispatch
// path clean of `if a != nil` boilerplate.
func TestAuditLoggerNilIsNoop(t *testing.T) {
	var a *auditLogger
	// Both calls below must NOT panic.
	a.Log("lark_doc_fetch", map[string]interface{}{"doc": "xyz"}, execResult{}, 5*time.Millisecond)
	if err := a.Close(); err != nil {
		t.Errorf("Close on nil returned %v, want nil", err)
	}
}

// TestAuditLoggerWritesNDJSON verifies one Log call produces exactly
// one newline-terminated JSON line with the expected shape.
func TestAuditLoggerWritesNDJSON(t *testing.T) {
	var buf bytes.Buffer
	a := newAuditLoggerWriter(&buf, false)
	a.Log(
		"lark_calendar_agenda",
		map[string]interface{}{"start": "2026-05-28"},
		execResult{Stdout: "hello", Stderr: "", ExitCode: 0},
		120*time.Millisecond,
	)

	out := buf.String()
	if !strings.HasSuffix(out, "\n") {
		t.Errorf("audit line missing trailing newline: %q", out)
	}
	if strings.Count(out, "\n") != 1 {
		t.Errorf("expected exactly 1 line, got %d", strings.Count(out, "\n"))
	}

	var got auditEntry
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &got); err != nil {
		t.Fatalf("audit line is not valid JSON: %v\nline: %s", err, out)
	}
	if got.Tool != "lark_calendar_agenda" {
		t.Errorf("tool=%q want lark_calendar_agenda", got.Tool)
	}
	if !strings.HasPrefix(got.ArgsHash, "sha256:") {
		t.Errorf("args_hash missing sha256 prefix: %q", got.ArgsHash)
	}
	if got.StdoutBytes != 5 {
		t.Errorf("stdout_bytes=%d want 5", got.StdoutBytes)
	}
	if got.LatencyMs != 120 {
		t.Errorf("latency_ms=%d want 120", got.LatencyMs)
	}
	if got.IsError {
		t.Errorf("is_error=true want false for exit 0")
	}
	if got.Session == "" {
		t.Errorf("session id should be non-empty")
	}
	if !strings.HasPrefix(got.Session, "pid-") {
		t.Errorf("session=%q expected pid- prefix", got.Session)
	}
}

// TestAuditLoggerExitCodeFlagsError covers the maker-checker safety
// signal: any non-zero exit must surface as is_error=true so log
// consumers don't have to re-derive it.
func TestAuditLoggerExitCodeFlagsError(t *testing.T) {
	var buf bytes.Buffer
	a := newAuditLoggerWriter(&buf, false)
	a.Log(
		"lark_calendar_create",
		map[string]interface{}{},
		execResult{Stderr: "denied", ExitCode: 1},
		50*time.Millisecond,
	)
	var got auditEntry
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &got); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if !got.IsError {
		t.Errorf("is_error=false; want true for exit 1")
	}
	if got.StderrBytes != 6 {
		t.Errorf("stderr_bytes=%d want 6", got.StderrBytes)
	}
}

// TestHashArgsDeterministic locks the contract: same input map →
// same hash, regardless of insertion order. Go's json.Marshal sorts
// map keys, which is what makes this work.
func TestHashArgsDeterministic(t *testing.T) {
	a := map[string]interface{}{"b": 2, "a": 1, "c": 3}
	b := map[string]interface{}{"c": 3, "a": 1, "b": 2}
	hashA, _ := hashArgs(a)
	hashB, _ := hashArgs(b)
	if hashA != hashB {
		t.Errorf("hash differs by insertion order: %s vs %s", hashA, hashB)
	}
}

// TestHashArgsEmpty ensures empty map maps to the hash of "{}".
func TestHashArgsEmpty(t *testing.T) {
	h1, _ := hashArgs(nil)
	h2, _ := hashArgs(map[string]interface{}{})
	if h1 != h2 {
		t.Errorf("nil vs empty map hash differ: %s vs %s", h1, h2)
	}
	if !strings.HasPrefix(h1, "sha256:") {
		t.Errorf("missing prefix: %s", h1)
	}
}

// TestAuditLoggerIncludesArgsOnlyWhenOptIn validates the privacy
// default: args are hash-only unless --audit-log-args is set.
func TestAuditLoggerIncludesArgsOnlyWhenOptIn(t *testing.T) {
	t.Run("default_hash_only", func(t *testing.T) {
		var buf bytes.Buffer
		a := newAuditLoggerWriter(&buf, false)
		a.Log("lark_mail_send", map[string]interface{}{"to": "leak@example.com"}, execResult{}, 0)
		if strings.Contains(buf.String(), "leak@example.com") {
			t.Errorf("PII leaked into audit log without opt-in:\n%s", buf.String())
		}
	})
	t.Run("opt_in_includes_args", func(t *testing.T) {
		var buf bytes.Buffer
		a := newAuditLoggerWriter(&buf, true)
		a.Log("lark_mail_send", map[string]interface{}{"to": "verbose@example.com"}, execResult{}, 0)
		if !strings.Contains(buf.String(), "verbose@example.com") {
			t.Errorf("opt-in did not include args in line:\n%s", buf.String())
		}
	})
}

// TestAuditLoggerConcurrentWrites pounds the mutex with 100 goroutines
// each writing 10 lines. All 1000 lines must be present and parseable.
// Without the mutex, lines would interleave at byte boundaries.
func TestAuditLoggerConcurrentWrites(t *testing.T) {
	var buf bytes.Buffer
	a := newAuditLoggerWriter(&buf, false)

	var wg sync.WaitGroup
	const goroutines = 100
	const perGoroutine = 10
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < perGoroutine; j++ {
				a.Log(
					"lark_doc_fetch",
					map[string]interface{}{"doc": fmt.Sprintf("g%d-i%d", id, j)},
					execResult{Stdout: "ok"},
					time.Millisecond,
				)
			}
		}(i)
	}
	wg.Wait()

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != goroutines*perGoroutine {
		t.Errorf("got %d lines, want %d", len(lines), goroutines*perGoroutine)
	}
	for i, line := range lines {
		var probe auditEntry
		if err := json.Unmarshal([]byte(line), &probe); err != nil {
			t.Errorf("line %d not valid JSON (likely interleaved write): %v\n  line: %s", i, err, line)
			return
		}
	}
}

// TestNewAuditLoggerOpensFileInAppendMode verifies the file-path
// constructor: subsequent runs append rather than truncate. Operators
// running mcp serve repeatedly against the same path keep history.
func TestNewAuditLoggerOpensFileInAppendMode(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.ndjson")

	// First session.
	a, err := newAuditLogger(path, false)
	if err != nil {
		t.Fatalf("newAuditLogger: %v", err)
	}
	a.Log("lark_im_search", map[string]interface{}{"query": "first"}, execResult{}, 0)
	if err := a.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	// Second session — must append, not clobber.
	a2, err := newAuditLogger(path, false)
	if err != nil {
		t.Fatalf("re-open: %v", err)
	}
	a2.Log("lark_im_search", map[string]interface{}{"query": "second"}, execResult{}, 0)
	if err := a2.Close(); err != nil {
		t.Fatalf("Close (2): %v", err)
	}

	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(contents)), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines after two sessions, got %d:\n%s", len(lines), string(contents))
	}
}

// TestNewAuditLoggerEmptyPathReturnsNil verifies the convention:
// empty path = audit disabled = nil logger.
func TestNewAuditLoggerEmptyPathReturnsNil(t *testing.T) {
	a, err := newAuditLogger("", false)
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if a != nil {
		t.Errorf("expected nil logger for empty path")
	}
}

// TestNewAuditLoggerBadPathErrors confirms unwritable paths surface
// an error at startup rather than silently dropping audit lines later.
func TestNewAuditLoggerBadPathErrors(t *testing.T) {
	_, err := newAuditLogger("/nonexistent-dir/audit.ndjson", false)
	if err == nil {
		t.Errorf("expected open error for unwritable path")
	}
}

// TestDispatchAuditPathPopulatesExitCode is an integration-style test
// for the dispatchWithRaw → audit chain. We simulate a tool with an
// invalid arg (forces the Build closure to error) and check audit
// sees ExitCode -1, IsError true.
func TestDispatchAuditPathPopulatesExitCode(t *testing.T) {
	// lark_doc_fetch requires `doc` arg; omit it to force Build error.
	r := &runner{exe: "/nonexistent"}
	_, raw, err := dispatchWithRaw(context.Background(), r, "lark_doc_fetch", map[string]interface{}{})
	if err != nil {
		t.Fatalf("dispatchWithRaw returned err: %v", err)
	}
	if raw.ExitCode != -1 {
		t.Errorf("ExitCode=%d for Build error; want -1", raw.ExitCode)
	}
	if raw.Stderr == "" {
		t.Errorf("Stderr empty; expected error message")
	}
}

// TestDispatchAuditPathUnknownTool similarly checks the unknown-tool
// path emits a -1 exit code so audit logs surface it as is_error.
func TestDispatchAuditPathUnknownTool(t *testing.T) {
	r := &runner{exe: "/nonexistent"}
	_, raw, err := dispatchWithRaw(context.Background(), r, "lark_does_not_exist", map[string]interface{}{})
	if err != nil {
		t.Fatalf("dispatchWithRaw: %v", err)
	}
	if raw.ExitCode != -1 {
		t.Errorf("ExitCode=%d for unknown tool; want -1", raw.ExitCode)
	}
	if !strings.Contains(raw.Stderr, "unknown tool") {
		t.Errorf("Stderr=%q expected to mention 'unknown tool'", raw.Stderr)
	}
}
