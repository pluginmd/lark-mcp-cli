// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

// Package mcp implements a Model Context Protocol (MCP) stdio server
// that exposes a curated set of lark-cli shortcuts as MCP tools. It is
// intended to be wired into Claude Desktop (or any MCP-aware host) via:
//
//	{
//	  "mcpServers": {
//	    "lark-cli": { "command": "lark-cli", "args": ["mcp", "serve"] }
//	  }
//	}
//
// The server spawns lark-cli subprocesses to execute each tool call so the
// existing auth, profile, and shortcut machinery is reused as-is. Run
// `lark-cli auth login` once before connecting from the MCP host.
package mcp

import (
	"github.com/spf13/cobra"

	"github.com/larksuite/cli/internal/cmdutil"
)

// NewCmdMCP creates the `mcp` parent command.
func NewCmdMCP(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Model Context Protocol (MCP) server for Claude Desktop and other MCP hosts",
		Long: `Run lark-cli as a Model Context Protocol (MCP) server.

The server speaks MCP over stdio (newline-delimited JSON-RPC 2.0) and
exposes a curated set of Lark/Feishu tools (IM, Calendar, Docs, Base,
Contact, Task, Drive, plus a generic API passthrough). Wire it into
Claude Desktop or any MCP-aware host.

Example claude_desktop_config.json:

  {
    "mcpServers": {
      "lark-cli": {
        "command": "lark-cli",
        "args": ["mcp", "serve"]
      }
    }
  }

Run ` + "`lark-cli auth login`" + ` first so the bridge has credentials.`,
	}
	cmdutil.DisableAuthCheck(cmd)
	cmdutil.SetTips(cmd, []string{
		"Run `lark-cli auth login` before connecting an MCP host so tool calls succeed.",
		"Logs go to stderr; stdout is reserved for MCP JSON-RPC framing — do not mix.",
	})

	cmd.AddCommand(NewCmdMCPServe(f))
	cmd.AddCommand(NewCmdMCPTools(f))
	return cmd
}
