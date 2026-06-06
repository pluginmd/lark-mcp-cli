// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package im

import (
	"strings"
)

// optimizeMarkdownStyleV2 is the card-specific wrapper around the
// shared optimizeMarkdownStyle in helpers.go. The shared helper
// (designed for the +messages-send --markdown → v1 `post` format)
// covers H1→H4 demotion, table spacing, and invalid-image stripping.
// The Lark fork's openclaw-lark builder adds one more transform on
// top for schema-2.0 cards: pad fenced code blocks with <br> on
// both sides so they render with proper spacing instead of touching
// the surrounding paragraph.
//
// We keep this wrapper card-only — touching the shared helper would
// silently change rendering of every existing +messages-send caller.
func optimizeMarkdownStyleV2(text string) string {
	if text == "" {
		return text
	}
	r := optimizeMarkdownStyle(text)
	r = padFencedCodeBlocksWithBR(r)
	return r
}

// padFencedCodeBlocksWithBR walks the lines and wraps each fenced
// code block in "\n<br>\n" sentinels (unless one is already there).
// We use a manual line scanner rather than a regex because Go's RE2
// has no backreferences and an open fence's closing fence must match
// in length.
func padFencedCodeBlocksWithBR(s string) string {
	if !strings.Contains(s, "```") {
		return s
	}
	lines := strings.Split(s, "\n")
	out := make([]string, 0, len(lines)+4)
	i := 0
	for i < len(lines) {
		line := lines[i]
		fence := leadingBacktickFenceLen(line)
		if fence < 3 {
			out = append(out, line)
			i++
			continue
		}
		// Find the matching closer.
		start := i
		j := i + 1
		for j < len(lines) && leadingBacktickFenceLen(lines[j]) != fence {
			j++
		}
		if j >= len(lines) {
			// Unterminated fence — leave alone (mirrors the v1 helper).
			out = append(out, lines[start])
			i = start + 1
			continue
		}
		// Insert leading <br> if the previous emitted line isn't one already.
		if !endsWithBR(out) {
			out = append(out, "<br>")
		}
		// Emit the fenced block verbatim.
		out = append(out, lines[start:j+1]...)
		// Insert trailing <br> unless one is already coming.
		if j+1 >= len(lines) || lines[j+1] != "<br>" {
			out = append(out, "<br>")
		}
		i = j + 1
	}
	return strings.Join(out, "\n")
}

// endsWithBR returns true when the last non-empty line of `lines`
// is the literal "<br>" sentinel. Empty trailing lines are skipped
// so "code\n\n<br>\n" still counts as already-padded.
func endsWithBR(lines []string) bool {
	for k := len(lines) - 1; k >= 0; k-- {
		if lines[k] == "" {
			continue
		}
		return lines[k] == "<br>"
	}
	return false
}

// leadingBacktickFenceLen returns the count of leading backticks on a
// line if the line is a valid fence (≥3 backticks, no other backticks
// before the end of line). Otherwise returns 0.
func leadingBacktickFenceLen(line string) int {
	n := 0
	for n < len(line) && line[n] == '`' {
		n++
	}
	if n < 3 {
		return 0
	}
	tail := line[n:]
	if strings.Contains(tail, "`") {
		return 0
	}
	return n
}

// renderMd is the single chokepoint every card-emit site goes through
// when it needs to turn user-supplied markdown into the bytes that
// will appear in the compiled card. Centralising the call here means
// adding a new optimiser step (or an opt-out) only requires touching
// one function, and guarantees that md / md.i18n / panel.title /
// div.text all behave the same way.
func renderMd(content string, noStyle bool) string {
	if noStyle {
		return content
	}
	return optimizeMarkdownStyleV2(content)
}
