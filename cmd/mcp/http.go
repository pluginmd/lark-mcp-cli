// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package mcp

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/larksuite/cli/internal/cmdutil"
)

// runServeHTTP starts the streamable-http transport on addr.
//
// We implement just enough of the MCP streamable-http profile to talk
// to OpenClaw and other containerized hosts that reach the binary over
// host.docker.internal:
//
//   - One endpoint accepts POST requests carrying a JSON-RPC message.
//   - Notifications (no id) return 202 Accepted with no body.
//   - Requests return 200 OK + application/json + the response message.
//   - Errors at the HTTP layer use 4xx; protocol errors come back as
//     JSON-RPC error objects with HTTP 200.
//
// Session management (Mcp-Session-Id), GET-for-SSE, and server-pushed
// notifications are intentionally NOT implemented — every tool call is
// stateless and request-scoped, which matches how the subprocess
// runner already behaves on stdio. If you ever need server-initiated
// streaming, add a separate GET handler and a per-session SSE writer
// keyed by Mcp-Session-Id.
func runServeHTTP(ctx context.Context, f *cmdutil.Factory, addr string, audit *auditLogger, bearerToken string) error {
	srv, err := newServer(ctx, f, audit)
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", srv.handleHTTPRoot)
	mux.HandleFunc("/mcp", srv.handleHTTPRoot)
	mux.HandleFunc("/health", srv.handleHTTPHealth)

	// The HTTP transport spawns lark-cli subprocesses that read the
	// caller's Lark auth token from the OS keychain — i.e. anyone who can
	// POST to this endpoint acts as the logged-in account. NEVER expose it
	// to a public tunnel without an auth gate. When LARK_MCP_BEARER_TOKEN
	// is set we require `Authorization: Bearer <token>` on every JSON-RPC
	// path; /health stays open for liveness probes.
	var handler http.Handler = mux
	if bearerToken != "" {
		handler = bearerAuthMiddleware(bearerToken, mux)
		srv.logf("bearer-token auth ENABLED — POST / and /mcp require Authorization: Bearer <token> (/health open)")
	} else {
		srv.logf("WARNING: LARK_MCP_BEARER_TOKEN is empty — HTTP endpoint is UNAUTHENTICATED. Do NOT expose this to a public tunnel.")
	}

	httpSrv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		// No write timeout: tool calls can legitimately take many
		// seconds (uploads, large search results). The MCP host's own
		// timeout governs liveness.
	}

	srv.logf("lark-cli MCP server starting on http://%s (protocol %s, version %s)",
		addr, mcpProtocolVersion, srv.version)
	srv.logf("endpoints: POST /  POST /mcp  GET /health")

	errCh := make(chan error, 1)
	go func() { errCh <- httpSrv.ListenAndServe() }()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = httpSrv.Shutdown(shutdownCtx)
		return ctx.Err()
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

// bearerAuthMiddleware gates every path except /health behind a static
// bearer token. The comparison is constant-time so a caller cannot probe
// the token byte-by-byte via response timing. A missing or wrong header
// returns 401 with a WWW-Authenticate challenge.
func bearerAuthMiddleware(token string, next http.Handler) http.Handler {
	want := []byte("Bearer " + token)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}
		got := []byte(r.Header.Get("Authorization"))
		// ConstantTimeCompare returns 0 when lengths differ, so a short
		// or empty header is rejected without leaking the expected length.
		if subtle.ConstantTimeCompare(got, want) != 1 {
			w.Header().Set("WWW-Authenticate", `Bearer realm="lark-cli-mcp"`)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// handleHTTPHealth is a tiny liveness check, useful when the server
// runs in a docker compose stack and another container wants to wait
// for it before issuing the first tool call.
func (s *server) handleHTTPHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	// Use json.Marshal — version strings can theoretically contain
	// characters that would break a string-concatenated JSON literal
	// (e.g. a hyphen-prefixed git SHA "1.2.3-dev+abc" is fine, but
	// future builds that embed a quote in the version metadata would
	// silently corrupt the response).
	body, err := json.Marshal(map[string]interface{}{
		"ok":      true,
		"name":    mcpServerName,
		"version": s.version,
	})
	if err != nil {
		http.Error(w, "health marshal failed", http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(body)
}

// handleHTTPRoot handles the MCP JSON-RPC POST. One request, one
// response — no batching, no SSE.
func (s *server) handleHTTPRoot(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// MCP spec optionally allows GET to open an SSE stream for
		// server-initiated notifications. We don't emit any today, so
		// return a clean 405 instead of leaving the client hanging.
		w.Header().Set("Allow", "POST")
		http.Error(w, "GET not supported; this server is request-response only", http.StatusMethodNotAllowed)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ct := r.Header.Get("Content-Type")
	if ct != "" && !strings.HasPrefix(ct, "application/json") {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 16*1024*1024))
	if err != nil {
		http.Error(w, "read body: "+err.Error(), http.StatusBadRequest)
		return
	}
	_ = r.Body.Close()

	body = trimUTF8BOM(body)
	if len(body) == 0 {
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	// Reject batched arrays explicitly. The current MCP draft removed
	// JSON-RPC batching, and supporting it would change the response
	// shape. Encourage clients to send one message at a time.
	if first := firstNonSpace(body); first == '[' {
		http.Error(w, "JSON-RPC batching is not supported; send one request per POST", http.StatusBadRequest)
		return
	}

	var msg rpcMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		writeJSON(w, http.StatusOK, errorMessage(nil, codeParseError, fmt.Sprintf("parse error: %v", err)))
		return
	}

	resp, ok := s.process(&msg)
	if !ok {
		// Notification — no response body, per MCP HTTP profile.
		w.WriteHeader(http.StatusAccepted)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// writeJSON marshals msg and writes it as the response body. Failures
// here are logged but cannot be surfaced to the client (response is
// already partially written), so keep marshaling defensive.
func writeJSON(w http.ResponseWriter, status int, msg rpcMessage) {
	buf, err := json.Marshal(msg)
	if err != nil {
		http.Error(w, "encode response: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(buf)
}

func trimUTF8BOM(b []byte) []byte {
	if len(b) >= 3 && b[0] == 0xEF && b[1] == 0xBB && b[2] == 0xBF {
		return b[3:]
	}
	return b
}

func firstNonSpace(b []byte) byte {
	for _, c := range b {
		switch c {
		case ' ', '\t', '\r', '\n':
			continue
		default:
			return c
		}
	}
	return 0
}
