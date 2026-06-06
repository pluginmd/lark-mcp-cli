#!/usr/bin/env bash
# setup-mcp.sh — one-shot bootstrap for the lark-cli MCP bridge.
#
# What it does:
#   1. Checks toolchain (go, optionally jq) and prints a clear error if missing.
#   2. Builds `lark-cli` from source.
#   3. Installs it to a user-chosen path (default: ~/bin/lark-cli).
#   4. Verifies the binary by running `lark-cli mcp tools`.
#   5. Prints next-step instructions for wiring an MCP host.
#
# What it does NOT do:
#   - Authenticate to Lark (you run `lark-cli auth login` yourself).
#   - Edit any MCP host config. Use the snippets under examples/mcp-hosts/
#     and copy them into your host's config file manually.
#
# Usage:
#   ./scripts/setup-mcp.sh                       # default install to ~/bin
#   ./scripts/setup-mcp.sh /usr/local/bin        # custom prefix (sudo may be required)
#   ./scripts/setup-mcp.sh --help

set -euo pipefail

# ---------- helpers ----------
bold()   { printf "\033[1m%s\033[0m\n" "$*"; }
green()  { printf "\033[32m%s\033[0m\n" "$*"; }
red()    { printf "\033[31m%s\033[0m\n" "$*" >&2; }
yellow() { printf "\033[33m%s\033[0m\n" "$*"; }
die()    { red "✗ $*"; exit 1; }

# Disable colour entirely if NO_COLOR is set (for CI / pipes).
if [ -n "${NO_COLOR:-}" ]; then
  bold() { printf "%s\n" "$*"; }
  green() { printf "%s\n" "$*"; }
  red() { printf "%s\n" "$*" >&2; }
  yellow() { printf "%s\n" "$*"; }
fi

usage() {
  cat <<EOF
Usage: $0 [INSTALL_DIR]

Build and install the lark-cli binary so MCP hosts (Claude Desktop,
Claude Code, Cursor, Zed, OpenClaw, …) can spawn it.

Arguments:
  INSTALL_DIR    Directory to install the binary into. Default: \$HOME/bin
                 Recommended alternatives: /usr/local/bin (may need sudo),
                 \$HOME/.local/bin (Linux convention).

Examples:
  $0                     # ~/bin/lark-cli
  $0 ~/bin               # same
  $0 /usr/local/bin      # system-wide (sudo required if not writable)

Env:
  NO_COLOR=1             # disable ANSI colours

After install, see examples/mcp-hosts/ for ready-to-paste config
snippets for each supported MCP host.
EOF
}

# ---------- parse args ----------
case "${1:-}" in
  -h|--help) usage; exit 0 ;;
  -*)        usage; die "Unknown flag: $1" ;;
esac

INSTALL_DIR="${1:-$HOME/bin}"
INSTALL_DIR="${INSTALL_DIR/#\~/$HOME}"   # expand leading ~

# ---------- locate repo root ----------
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$REPO_ROOT"

if [ ! -f "$REPO_ROOT/go.mod" ] || [ ! -d "$REPO_ROOT/cmd/mcp" ]; then
  die "This script must run from inside the lark-cli repo. Looked at: $REPO_ROOT"
fi

bold "▶ lark-cli MCP bridge — setup"
echo "  Repo root  : $REPO_ROOT"
echo "  Install to : $INSTALL_DIR"
echo

# ---------- check toolchain ----------
command -v go >/dev/null 2>&1 || die "Go is not installed. Install Go 1.23+ from https://go.dev/dl/"

GO_VERSION="$(go version | awk '{print $3}' | sed 's/^go//')"
GO_MAJOR="$(printf "%s" "$GO_VERSION" | cut -d. -f1)"
GO_MINOR="$(printf "%s" "$GO_VERSION" | cut -d. -f2)"
if [ "${GO_MAJOR}" -lt 1 ] || { [ "${GO_MAJOR}" -eq 1 ] && [ "${GO_MINOR}" -lt 23 ]; }; then
  die "Go 1.23+ required (found ${GO_VERSION})."
fi
green "✓ Go ${GO_VERSION}"

# ---------- build ----------
echo
bold "▶ Building lark-cli"
TMP_BIN="$(mktemp -d)/lark-cli"
go build -o "$TMP_BIN" .
green "✓ Built $(ls -la "$TMP_BIN" | awk '{print $5}') bytes → $TMP_BIN"

# ---------- install ----------
echo
bold "▶ Installing to $INSTALL_DIR"
mkdir -p "$INSTALL_DIR"

if [ -w "$INSTALL_DIR" ]; then
  cp "$TMP_BIN" "$INSTALL_DIR/lark-cli"
else
  yellow "  $INSTALL_DIR is not writable, trying with sudo…"
  sudo cp "$TMP_BIN" "$INSTALL_DIR/lark-cli"
fi
INSTALLED="$INSTALL_DIR/lark-cli"
green "✓ Installed $INSTALLED"

# ---------- verify ----------
echo
bold "▶ Verifying"
VERSION_OUT="$("$INSTALLED" --version 2>&1 || true)"
echo "  $VERSION_OUT"

TOOLS_OUT="$("$INSTALLED" mcp tools 2>&1 | tail -1 || true)"
echo "  $TOOLS_OUT"

if printf "%s" "$TOOLS_OUT" | grep -q "tools total"; then
  green "✓ MCP bridge operational"
else
  red "✗ MCP bridge check failed. Run: $INSTALLED mcp tools"
  exit 1
fi

# ---------- config init (first-time bootstrap) ----------
# `lark-cli auth login` requires a config dir to exist. Run config init
# now so the user doesn't hit "no profile" errors on first login. This
# is idempotent — a second run is a no-op.
if [ ! -d "$HOME/.config/lark-cli" ] && [ ! -d "${LARKSUITE_CLI_CONFIG_DIR:-}" ]; then
  echo
  bold "▶ Initialising config"
  if "$INSTALLED" config init >/dev/null 2>&1; then
    green "✓ Config initialised at \$HOME/.config/lark-cli"
  else
    yellow "⚠ Could not auto-init config. Run manually: $INSTALLED config init"
  fi
fi

# ---------- next steps ----------
echo
bold "▶ Next steps"
cat <<EOF
  1. Authenticate (one-time):
       $INSTALLED auth login

  2. Wire your MCP host. Pick the matching snippet:
       examples/mcp-hosts/claude-desktop.json
       examples/mcp-hosts/claude-code.json
       examples/mcp-hosts/openclaw.json
       examples/mcp-hosts/cursor.json
       examples/mcp-hosts/zed.json
       examples/mcp-hosts/cline.json

     Replace <BINARY-PATH> with:  $INSTALLED

  3. Restart the host (full quit + reopen for Claude Desktop / Zed).

  4. Confirm tools appear in the host. For Claude Desktop, look under
     Settings → Developer → Local MCP servers.

  Trouble? See cmd/mcp/README.md → Troubleshooting.
EOF

# ---------- PATH hint ----------
case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *) yellow "
  ⚠  $INSTALL_DIR is not on your PATH. Add this to your shell rc:
       export PATH=\"$INSTALL_DIR:\$PATH\"
" ;;
esac

green "✓ Done."
