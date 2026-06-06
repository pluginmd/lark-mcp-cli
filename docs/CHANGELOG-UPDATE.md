# Nhật ký bản dựng — MCP bridge + Cowork kit

> Tài liệu kỹ thuật: chính xác những gì đã thêm/sửa để biến bản clone `larksuite/cli` (v1.0.48, base tháng 6) thành bản **có MCP cho Lark desktop + web**. Dùng để tái lập & kiểm toán.

## Bối cảnh

- **Base:** upstream `larksuite/cli`, commit `8c3cba1`, **v1.0.48** (2026-06).
- **Nguồn port:** một bản fork nội bộ (base ~v1.0.41, tháng 5) đã xây sẵn MCP bridge + Cowork kit.
- **Việc đã làm:** port lớp MCP + card + Cowork kit sang base mới, xử lý drift, verify.

## Delta (theo git)

```
 M cmd/build.go                  (+2)  đăng ký lệnh mcp
 M shortcuts/im/shortcuts.go     (+1)  đăng ký shortcut +card-send
 M shortcuts/im/helpers_test.go  (+1)  cập nhật danh sách test
?? cmd/mcp/                      MCP bridge (14 file, ~3.627 LOC Go gồm test)
?? shortcuts/im/card_spec.go            }
?? shortcuts/im/card_markdown.go        } shortcut +card-send (~1.313 LOC nguồn)
?? shortcuts/im/im_card_send.go         }
?? shortcuts/im/card_spec_test.go       } test kèm
?? shortcuts/im/im_card_send_test.go    }
?? .claude/                      Cowork kit: 23 skill + 6 lệnh /mcp-*
?? examples/mcp-hosts/           6 host config (claude-desktop/code, cursor, zed, cline, openclaw)
?? scripts/setup-mcp.sh          installer build + cài + verify
?? MCP_QUICKSTART.md             quickstart đa-host
?? docs/                         bộ tài liệu business-user (file này + 01..09)
```

## Chi tiết các nhóm

### 1. MCP bridge — `cmd/mcp/` (mới)
Lệnh `lark-cli mcp serve` (transport **stdio** + **http**) và `lark-cli mcp tools`.
- `mcp.go` đăng ký lệnh; `serve.go` chọn transport; `http.go` streamable-HTTP (mặc định `127.0.0.1:3000`);
  `protocol.go` JSON-RPC 2.0; `runner.go` spawn subprocess `lark-cli <verb>` (tái dùng auth);
  `tools.go` **21 tool** curated; `errors.go`; `audit.go` (cờ `--audit-log`).
- Test: `audit_test, concurrency_test, descriptions_test, errors_test, tools_test`.
- Cơ chế: mỗi tool call → spawn lại `lark-cli` → dùng chung credential/keychain sẵn có.

### 2. Shortcut card — `shortcuts/im/card_*.go` (mới)
`lark-cli im +card-send`: compile **YAML card spec → Feishu interactive card JSON**.
- Đây là **drift duy nhất** giữa base tháng 5→6: fork dùng `im +card-send`, base tháng 6 chưa có → đã port trọn shortcut (compiler + markdown optimizer + sender) và đăng ký vào `shortcuts/im/shortcuts.go`.
- Phụ thuộc đều có sẵn ở base tháng 6 (`common.NewDryRunAPI`, `RuntimeContext.DoAPIJSON`, `yaml.v3`) → compile sạch.

### 3. Cowork kit — `.claude/` (mới)
- 23 skill workflow (đầy đủ ở [05](05-bo-skill-cowork.md)) + skill kỹ thuật `lark-cli-mcp`.
- 6 lệnh `/mcp-*`: `mcp-add, mcp-call, mcp-doctor, mcp-rebuild, mcp-test, mcp-tools`.
- **Không** port agents/hooks/memory dev-only của fork.

### 4. Host config + installer + quickstart
- `examples/mcp-hosts/*.json` (6 host) + `examples/README.md`.
- `scripts/setup-mcp.sh`, `MCP_QUICKSTART.md`, `cmd/mcp/README.md`.
- Đã sửa URL `pluginmd/lark-cli` → `larksuite/cli`.

### 5. Wiring vào core (3 dòng, low-risk)
- `cmd/build.go`: `import cmdmcp` + `rootCmd.AddCommand(cmdmcp.NewCmdMCP(f))`.
- `shortcuts/im/shortcuts.go`: thêm `ImCardSend`.
- `shortcuts/im/helpers_test.go`: thêm `+card-send` vào expected list.

## Thay đổi ngoài repo (môi trường máy)

- Đã thay `~/bin/lark-cli` = binary mới (v1.0.48 + MCP). **Backup:** `~/bin/lark-cli.bak-fork-v1.0.41`.
- `claude_desktop_config.json` **đã có sẵn** entry `lark-cli` → `~/bin/lark-cli mcp serve` (không sửa).
- `cloudflared` đã cài sẵn.

## Đã verify (bằng chứng)

- ✅ `go build ./...`, `go vet ./...`, `gofmt -l` sạch.
- ✅ `go test ./cmd/mcp ./shortcuts/im` → PASS.
- ✅ `lark-cli mcp tools` → **21 tool**.
- ✅ JSON-RPC `initialize` + `tools/list` → đúng framing (chạy từ thư mục repo).
- ✅ `im +card-send --print-json` → compile YAML→JSON đúng (offline).
- ✅ `go mod tidy` không đổi `go.mod`/`go.sum`.

## Chưa xong / rủi ro mở (PHẢI giải trước khi tuyên bố "chạy được cho user")

1. ⛔ **Blocker desktop:** binary **treo khi cwd ≠ repo** (kể cả `mcp tools`). Claude Desktop spawn từ cwd khác → có thể treo. Nghi Gatekeeper/keychain — khắc phục & chi tiết ở [08 §"Server treo"](08-xu-ly-su-co.md). **Chưa xác minh 1 tool call thật trong Cowork.**
2. ⏳ **Auth user = missing** → cần `lark-cli auth login` để chạy tool định danh user.
3. ⏳ Chưa chạy `make unit-test` (-race) toàn bộ + `golangci-lint`.
4. ⏳ **Web/claude.ai chưa triển khai** (tunnel + lớp bảo mật + connector) — hướng dẫn ở [03](03-ket-noi-web-claude-ai.md).

## Lệnh tái lập nhanh

```bash
make build && ./lark-cli mcp tools           # 21 tool
go vet ./... && gofmt -l cmd/mcp shortcuts/im # sạch
go test ./cmd/mcp ./shortcuts/im             # PASS
```
