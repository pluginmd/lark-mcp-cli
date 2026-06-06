// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/larksuite/cli/internal/build"
	"github.com/larksuite/cli/internal/cmdutil"
)

// NewCmdMCPServe creates the `mcp serve` subcommand. It runs a blocking
// MCP server suitable for any MCP-aware host. Transport is selected
// with --transport (default: stdio, for local hosts like Claude
// Desktop; alternative: http, for containerized hosts like OpenClaw
// running in Docker that reach the binary over host.docker.internal).
func NewCmdMCPServe(f *cmdutil.Factory) *cobra.Command {
	var (
		transport      string
		addr           string
		auditLogPath   string
		auditLogArgs   bool
		maxConcurrency int
	)
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Run an MCP server exposing lark-cli tools (stdio or HTTP)",
		Long: `Run an MCP server exposing lark-cli tools.

Transports:
  --transport stdio  (default) newline-delimited JSON-RPC on stdin/stdout.
                     Suits local hosts that spawn lark-cli as a subprocess
                     (Claude Desktop, Claude Code, Cursor, Zed, …).

  --transport http   streamable-http JSON-RPC over a single HTTP endpoint.
                     Suits containerized MCP hosts that reach the binary
                     over the host network (OpenClaw in Docker, remote
                     agents, etc.).

Stdio wire-up example (Claude Desktop):

  {
    "mcpServers": {
      "lark-cli": {
        "command": "lark-cli",
        "args": ["mcp", "serve"]
      }
    }
  }

HTTP wire-up example (OpenClaw in Docker on macOS host):

  # On host:
  lark-cli mcp serve --transport http --addr 127.0.0.1:3000

  # OpenClaw config (inside container reaches host via host.docker.internal):
  {
    "mcp": {
      "servers": {
        "lark-cli": {
          "url": "http://host.docker.internal:3000",
          "transport": "streamable-http"
        }
      }
    }
  }

Security (HTTP transport):
  The HTTP endpoint acts as the logged-in Lark account. Before exposing it
  beyond localhost (e.g. via a Cloudflare tunnel), set a shared secret:

    LARK_MCP_BEARER_TOKEN=$(openssl rand -hex 32) \
      lark-cli mcp serve --transport http --addr 127.0.0.1:3000 --audit-log ~/.lark-mcp-audit.ndjson

  Clients must then send  Authorization: Bearer <token>  on every request.
  /health stays open for liveness probes. Without the env var the endpoint
  is UNAUTHENTICATED and must stay bound to 127.0.0.1.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			audit, err := newAuditLogger(auditLogPath, auditLogArgs)
			if err != nil {
				return err
			}
			defer audit.Close()
			switch transport {
			case "stdio", "":
				return runServeStdio(cmd.Context(), f, audit, maxConcurrency)
			case "http", "streamable-http":
				bearer := strings.TrimSpace(os.Getenv("LARK_MCP_BEARER_TOKEN"))
				return runServeHTTP(cmd.Context(), f, addr, audit, bearer)
			default:
				return fmt.Errorf("unknown --transport %q (use stdio or http)", transport)
			}
		},
	}
	cmdutil.DisableAuthCheck(cmd)
	cmd.Flags().StringVar(&transport, "transport", "stdio", "transport: stdio | http")
	cmd.Flags().StringVar(&addr, "addr", "127.0.0.1:3000", "HTTP listen address (only with --transport http)")
	cmd.Flags().StringVar(&auditLogPath, "audit-log", "",
		"path to write NDJSON audit log (one line per tools/call). Empty = disabled.")
	cmd.Flags().BoolVar(&auditLogArgs, "audit-log-args", false,
		"include raw tool arguments in audit log entries. Default hashes args only to protect PII.")
	cmd.Flags().IntVar(&maxConcurrency, "max-concurrency", 4,
		"max in-flight tool dispatches on stdio (0 or 1 = serial). HTTP transport is always concurrent.")
	return cmd
}

// NewCmdMCPTools creates `mcp tools` — a human-readable listing of the
// tools the stdio server would advertise. Handy for sanity checks
// outside an MCP host.
func NewCmdMCPTools(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tools",
		Short: "List the MCP tools exposed by `lark-cli mcp serve`",
		RunE: func(cmd *cobra.Command, args []string) error {
			tools := allTools()
			descs, err := descriptors(tools)
			if err != nil {
				return err
			}
			out := f.IOStreams.Out
			for _, d := range descs {
				fmt.Fprintf(out, "%-22s  %s\n", d.Name, d.Description)
			}
			fmt.Fprintf(out, "\n%d tools total. Sorted names: %s\n",
				len(descs), strings.Join(sortedToolNames(), ", "))
			return nil
		},
	}
	cmdutil.DisableAuthCheck(cmd)
	return cmd
}

// newServer assembles a *server with the runner, tools, IO streams and
// version baked in. Both the stdio and HTTP transports build their
// server with this helper so request handling stays identical.
//
// audit may be nil — that signals audit-log is disabled. Log() and
// Close() handle nil receivers gracefully.
func newServer(ctx context.Context, f *cmdutil.Factory, audit *auditLogger) (*server, error) {
	r, err := newRunner()
	if err != nil {
		return nil, err
	}
	descs, err := descriptors(allTools())
	if err != nil {
		return nil, err
	}
	return &server{
		ctx:     ctx,
		runner:  r,
		tools:   descs,
		stdin:   f.IOStreams.In,
		stdout:  f.IOStreams.Out,
		stderr:  f.IOStreams.ErrOut,
		version: build.Version,
		audit:   audit,
	}, nil
}

func runServeStdio(ctx context.Context, f *cmdutil.Factory, audit *auditLogger, maxConcurrency int) error {
	srv, err := newServer(ctx, f, audit)
	if err != nil {
		return err
	}
	srv.maxConcurrency = maxConcurrency
	return srv.runStdio()
}

type server struct {
	ctx     context.Context
	runner  *runner
	tools   []toolDescriptor
	stdin   io.Reader
	stdout  io.Writer
	stderr  io.Writer
	version string

	// writeMu serializes writes to stdout for the stdio transport. The
	// HTTP transport uses Go's http.ResponseWriter, which is single-use
	// per request, so this mutex is only relevant for stdio.
	writeMu sync.Mutex

	// audit may be nil (audit disabled). Log() is nil-safe.
	audit *auditLogger

	// maxConcurrency caps in-flight tool dispatches on stdio. 0 or 1 =
	// serial (legacy behaviour); higher values spawn a worker pool that
	// processes inbound JSON-RPC messages concurrently. JSON-RPC 2.0
	// does NOT require responses in request order — host MUST match by
	// `id` — so out-of-order replies are spec-compliant.
	maxConcurrency int
}

func (s *server) logf(format string, a ...interface{}) {
	fmt.Fprintf(s.stderr, "[mcp] "+format+"\n", a...)
}

func (s *server) runStdio() error {
	scanner := bufio.NewScanner(s.stdin)
	// MCP messages can be sizeable (tool results, base data dumps).
	// 10MB buffer matches what most MCP SDKs use and covers normal usage.
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 10*1024*1024)

	concurrency := s.maxConcurrency
	if concurrency < 1 {
		concurrency = 1
	}
	s.logf("lark-cli MCP server starting on stdio (protocol %s, version %s, max_concurrency=%d)",
		mcpProtocolVersion, s.version, concurrency)

	if concurrency <= 1 {
		// Legacy serial path — preserved for backward compat and for
		// hosts that depend on response ordering (none, per JSON-RPC 2.0,
		// but useful for debugging).
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			if err := s.handleStdioLine(line); err != nil {
				s.logf("handle line failed: %v", err)
			}
		}
		if err := scanner.Err(); err != nil && err != io.EOF {
			return err
		}
		s.logf("stdin closed, exiting")
		return nil
	}

	// Worker-pool path: spawn N workers reading from a shared lines
	// channel. The reader goroutine pushes scanned lines; workers run
	// handleStdioLine concurrently. writeMu serializes the stdout
	// writer, so frames stay well-formed even when workers finish in
	// arbitrary order.
	lines := make(chan string, concurrency*2)

	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for line := range lines {
				if err := s.handleStdioLine(line); err != nil {
					s.logf("handle line failed: %v", err)
				}
			}
		}()
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		lines <- line
	}
	close(lines)
	wg.Wait()

	if err := scanner.Err(); err != nil && err != io.EOF {
		return err
	}
	s.logf("stdin closed, exiting")
	return nil
}

// handleStdioLine parses one newline-delimited JSON-RPC frame from
// stdin, dispatches it through the shared `process` pipeline, and
// writes the response back on stdout (or nothing for notifications).
func (s *server) handleStdioLine(line string) error {
	var msg rpcMessage
	if err := json.Unmarshal([]byte(line), &msg); err != nil {
		return s.writeStdio(rpcMessage{
			JSONRPC: jsonrpcVersion,
			Error:   &rpcError{Code: codeParseError, Message: fmt.Sprintf("parse error: %v", err)},
		})
	}
	resp, ok := s.process(&msg)
	if !ok {
		return nil // notification or fire-and-forget
	}
	return s.writeStdio(resp)
}

// process is the transport-agnostic core. It accepts one inbound
// JSON-RPC message and returns the response (plus a bool indicating
// whether a response should be sent at all — false for notifications
// and silently-ignored unknown notifications). Both stdio and HTTP
// transports route through here so behaviour stays identical.
func (s *server) process(msg *rpcMessage) (rpcMessage, bool) {
	if msg.JSONRPC != jsonrpcVersion {
		return errorMessage(msg.ID, codeInvalidRequest, "jsonrpc must be \"2.0\""), true
	}
	// JSON-RPC 2.0 §4.2: a notification is a request without an `id`
	// member. `id: null` IS still a request and MUST receive a
	// response (the spec only forbids null IDs for notifications in
	// some implementations). We treat absent ID as notification and
	// `null` ID as a request that gets `id: null` in the response.
	// The previous implementation silently dropped `id: null` which
	// violates the spec.
	isNotification := len(msg.ID) == 0

	switch msg.Method {
	case "initialize":
		return s.processInitialize(*msg), true
	case "initialized", "notifications/initialized":
		s.logf("client initialized")
		return rpcMessage{}, false
	case "ping":
		return resultMessage(msg.ID, struct{}{}), true
	case "tools/list":
		return resultMessage(msg.ID, toolsListResult{Tools: s.tools}), true
	case "tools/call":
		return s.processToolsCall(*msg), true
	case "shutdown":
		return resultMessage(msg.ID, struct{}{}), true
	default:
		if isNotification {
			return rpcMessage{}, false
		}
		return errorMessage(msg.ID, codeMethodNotFound, "method not found: "+msg.Method), true
	}
}

func (s *server) processInitialize(msg rpcMessage) rpcMessage {
	var params initializeParams
	if len(msg.Params) > 0 {
		if err := json.Unmarshal(msg.Params, &params); err != nil {
			return errorMessage(msg.ID, codeInvalidParams, fmt.Sprintf("invalid initialize params: %v", err))
		}
	}
	// Negotiate protocol version. We advertise mcpProtocolVersion;
	// if the client requests an older version we still accept (so
	// long as it's at or above mcpMinProtocolVersion) and echo the
	// client's version back so they see the agreed value.
	negotiated := mcpProtocolVersion
	if params.ProtocolVersion != "" {
		if params.ProtocolVersion >= mcpMinProtocolVersion && params.ProtocolVersion <= mcpProtocolVersion {
			negotiated = params.ProtocolVersion
		}
		s.logf("client requests protocol %s — negotiated to %s",
			params.ProtocolVersion, negotiated)
	}
	result := initializeResult{
		ProtocolVersion: negotiated,
		Capabilities: serverCapabilities{
			Tools: &toolsCapability{ListChanged: false},
		},
		ServerInfo: serverInfo{
			Name:    mcpServerName,
			Version: s.version,
		},
		Instructions: "lark-cli MCP bridge. Run `lark-cli auth login` before calling tools. " +
			"Logs are on stderr; stdout is JSON-RPC only.",
	}
	return resultMessage(msg.ID, result)
}

func (s *server) processToolsCall(msg rpcMessage) rpcMessage {
	var params toolsCallParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return errorMessage(msg.ID, codeInvalidParams, fmt.Sprintf("invalid params: %v", err))
	}
	if params.Name == "" {
		return errorMessage(msg.ID, codeInvalidParams, "tool name required")
	}
	if params.Arguments == nil {
		params.Arguments = map[string]interface{}{}
	}
	start := time.Now()
	res, raw, err := dispatchWithRaw(s.ctx, s.runner, params.Name, params.Arguments)
	latency := time.Since(start)
	if err != nil {
		// Even on internal error we want a trail of what was attempted.
		s.audit.Log(params.Name, params.Arguments, execResult{ExitCode: -1, Stderr: err.Error()}, latency)
		return errorMessage(msg.ID, codeInternalError, err.Error())
	}
	s.audit.Log(params.Name, params.Arguments, raw, latency)
	return resultMessage(msg.ID, res)
}

// ---------------------------------------------------------------------
// Wire helpers.
// ---------------------------------------------------------------------

func resultMessage(id json.RawMessage, result interface{}) rpcMessage {
	return rpcMessage{JSONRPC: jsonrpcVersion, ID: id, Result: result}
}

func errorMessage(id json.RawMessage, code int, message string) rpcMessage {
	return rpcMessage{JSONRPC: jsonrpcVersion, ID: id, Error: &rpcError{Code: code, Message: message}}
}

func (s *server) writeStdio(msg rpcMessage) error {
	buf, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	if _, err := s.stdout.Write(buf); err != nil {
		return err
	}
	_, err = s.stdout.Write([]byte("\n"))
	return err
}
