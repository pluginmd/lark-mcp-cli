// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package mcp

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// auditEntry is one NDJSON line written per tools/call dispatch.
//
// Privacy contract: arguments are hashed by default (args_hash). The
// raw arg map is recorded ONLY when the operator opts in via
// --audit-log-args, since args can legitimately contain PII (recipient
// email, IM body, mail subject, etc.).
//
// Schema is intentionally flat + stable so the file can be tailed and
// parsed by simple tools (jq, awk, grep). Field order in the JSON
// reflects what a human grep'ing the file most wants to see first
// (tool name, ts) — Go's json package preserves struct field order.
type auditEntry struct {
	TS          string          `json:"ts"`
	Session     string          `json:"session"`
	Tool        string          `json:"tool"`
	ArgsHash    string          `json:"args_hash"`
	Args        json.RawMessage `json:"args,omitempty"` // populated only with --audit-log-args
	ExitCode    int             `json:"exit_code"`
	LatencyMs   int64           `json:"latency_ms"`
	StdoutBytes int             `json:"stdout_bytes"`
	StderrBytes int             `json:"stderr_bytes"`
	IsError     bool            `json:"is_error"`
}

// auditLogger writes one NDJSON line per dispatched tools/call.
//
// nil receiver is a deliberate sentinel meaning "no audit configured" —
// callers can call Log on a nil *auditLogger and it's a no-op. This
// keeps the dispatch path clean of conditional checks.
type auditLogger struct {
	mu      sync.Mutex
	w       io.Writer
	closer  io.Closer // optional; nil for non-file writers (tests)
	session string    // stable for the life of the process

	// includeArgs controls whether raw arguments are written alongside
	// the hash. Opt-in only — default is hash-only to protect PII.
	includeArgs bool
}

// newAuditLogger opens (or creates) path in append mode and returns a
// logger whose Log() method is safe for concurrent use. Returns nil
// when path is empty (audit disabled) — Log() handles nil receiver.
//
// Append mode is important: if the operator runs `mcp serve` multiple
// times against the same file, we keep history rather than clobbering.
// The trade-off: no automatic rotation. Operators wanting rotation
// should pipe through logrotate or similar.
func newAuditLogger(path string, includeArgs bool) (*auditLogger, error) {
	if path == "" {
		return nil, nil
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open audit log %q: %w", path, err)
	}
	return &auditLogger{
		w:           f,
		closer:      f,
		session:     newSessionID(),
		includeArgs: includeArgs,
	}, nil
}

// newAuditLoggerWriter is the in-memory variant used by tests. No file
// is opened; entries go to the supplied writer.
func newAuditLoggerWriter(w io.Writer, includeArgs bool) *auditLogger {
	return &auditLogger{
		w:           w,
		session:     newSessionID(),
		includeArgs: includeArgs,
	}
}

// Close flushes and closes the underlying file (no-op for in-memory
// writers). Safe to call on nil.
func (a *auditLogger) Close() error {
	if a == nil || a.closer == nil {
		return nil
	}
	return a.closer.Close()
}

// Log emits one NDJSON line. Errors are silently swallowed because
// audit logging must never break tool dispatch — if disk is full, the
// tool call should still complete and return its result to the host.
// The operator finds out about the audit problem via missing lines
// rather than a tool failure. This trade-off is intentional.
//
// Safe on nil receiver: a nil *auditLogger means audit is disabled.
func (a *auditLogger) Log(toolName string, args map[string]interface{}, res execResult, latency time.Duration) {
	if a == nil {
		return
	}

	hash, raw := hashArgs(args)
	entry := auditEntry{
		TS:          time.Now().UTC().Format(time.RFC3339Nano),
		Session:     a.session,
		Tool:        toolName,
		ArgsHash:    hash,
		ExitCode:    res.ExitCode,
		LatencyMs:   latency.Milliseconds(),
		StdoutBytes: len(res.Stdout),
		StderrBytes: len(res.Stderr),
		IsError:     res.ExitCode != 0,
	}
	if a.includeArgs {
		entry.Args = raw
	}

	buf, err := json.Marshal(&entry)
	if err != nil {
		// Marshal failure on a struct of plain types means the args raw
		// JSON was malformed — drop the line silently rather than
		// crashing the server.
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	_, _ = a.w.Write(buf)
	_, _ = a.w.Write([]byte("\n"))
}

// hashArgs returns sha256 hex of the canonical-JSON form of args, plus
// the raw JSON bytes (used only when --audit-log-args is set).
//
// json.Marshal is deterministic for map[string]interface{} keys (Go
// sorts them), so two calls with semantically-equal args produce the
// same hash. The hash is prefixed `sha256:` for forward compatibility
// (future versions could emit different algorithms with different
// prefixes; consumers can dispatch on prefix).
func hashArgs(args map[string]interface{}) (string, json.RawMessage) {
	if len(args) == 0 {
		empty := json.RawMessage("{}")
		sum := sha256.Sum256(empty)
		return "sha256:" + hex.EncodeToString(sum[:]), empty
	}
	buf, err := json.Marshal(args)
	if err != nil {
		// Marshal failure is unexpected on a JSON-RPC-derived map; emit
		// a sentinel hash so the audit line is still meaningful.
		return "sha256:marshal-error", nil
	}
	sum := sha256.Sum256(buf)
	return "sha256:" + hex.EncodeToString(sum[:]), buf
}

// newSessionID returns a process-stable identifier — pid plus boot
// epoch — useful for grouping all calls from one `mcp serve` invocation
// when multiple instances tail the same audit file.
func newSessionID() string {
	return fmt.Sprintf("pid-%d-%d", os.Getpid(), time.Now().Unix())
}
