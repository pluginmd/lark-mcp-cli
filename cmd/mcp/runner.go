// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package mcp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

// runner executes lark-cli as a subprocess so MCP tool calls reuse the
// existing shortcut machinery (auth, profile resolution, output format)
// without sharing in-process state with the long-running mcp serve loop.
type runner struct {
	exe string // absolute path to the lark-cli binary handling this process
}

func newRunner() (*runner, error) {
	exe, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("locate lark-cli executable: %w", err)
	}
	return &runner{exe: exe}, nil
}

// execResult captures both stdout/stderr so the MCP response can surface
// structured failure detail (lark-cli prints machine-readable JSON errors
// to stderr in most failure paths) back to the host model.
type execResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// run invokes `lark-cli <args...>` and returns its captured output.
// A non-zero exit is returned as a normal result with ExitCode set rather
// than an error so callers can decide whether to surface it as an MCP
// tool error or as informational text.
func (r *runner) run(ctx context.Context, args []string) (execResult, error) {
	cmd := exec.CommandContext(ctx, r.exe, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	// Inherit the host environment so token storage, profile selection,
	// and locale settings line up with what the user has configured.
	cmd.Env = os.Environ()

	err := cmd.Run()
	res := execResult{Stdout: stdout.String(), Stderr: stderr.String()}
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			res.ExitCode = exitErr.ExitCode()
			return res, nil
		}
		return res, fmt.Errorf("exec lark-cli: %w", err)
	}
	return res, nil
}
