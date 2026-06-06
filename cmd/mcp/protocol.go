// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package mcp

import "encoding/json"

// MCP / JSON-RPC 2.0 wire types. We hand-roll these instead of pulling
// in a third-party MCP library so go.mod stays lean and the dependency
// footprint of an MCP tool call matches that of a normal lark-cli call.

const (
	jsonrpcVersion = "2.0"
	// mcpProtocolVersion is the version we ADVERTISE. Hosts may
	// request a newer version; we negotiate down in initialize.
	// Bumped from 2024-11-05 to track the 2025-06-18 draft.
	mcpProtocolVersion = "2025-06-18"
	// mcpMinProtocolVersion is the oldest client version we still
	// accept (older clients silently downgrade to this string).
	mcpMinProtocolVersion = "2024-11-05"
	mcpServerName         = "lark-cli"
	mcpServerDescription  = "Lark/Feishu CLI bridge — IM, Calendar, Docs, Base, Contact, Task, Drive, Mail, Sheets, VC, Minutes, OKR, API"
)

// JSON-RPC error codes (subset).
const (
	codeParseError     = -32700
	codeInvalidRequest = -32600
	codeMethodNotFound = -32601
	codeInvalidParams  = -32602
	codeInternalError  = -32603
)

type rpcMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  interface{}     `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// MCP-specific message payloads.

type initializeParams struct {
	ProtocolVersion string          `json:"protocolVersion"`
	Capabilities    json.RawMessage `json:"capabilities"`
	ClientInfo      clientInfo      `json:"clientInfo"`
}

type clientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type initializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    serverCapabilities `json:"capabilities"`
	ServerInfo      serverInfo         `json:"serverInfo"`
	Instructions    string             `json:"instructions,omitempty"`
}

type serverCapabilities struct {
	Tools *toolsCapability `json:"tools,omitempty"`
}

type toolsCapability struct {
	ListChanged bool `json:"listChanged"`
}

type serverInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// tools/list
type toolsListResult struct {
	Tools []toolDescriptor `json:"tools"`
}

type toolDescriptor struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

// tools/call
type toolsCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

type toolsCallResult struct {
	Content []contentBlock `json:"content"`
	IsError bool           `json:"isError,omitempty"`
}

type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}
