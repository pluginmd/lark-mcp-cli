# 08 — Xử lý sự cố

> Mẹo vàng: khi một tool trong Claude lỗi, chạy **lệnh `lark-cli` tương đương ở terminal** — output sẽ cho biết lý do thật.

## Chẩn đoán nhanh

```bash
lark-cli doctor              # kiểm tra config + auth + kết nối
lark-cli auth status         # user/bot đã sẵn sàng?
lark-cli mcp tools           # server liệt kê đủ 21 tool?
```

Smoke-test giao thức MCP (phải in JSON có `"jsonrpc"`):

```bash
printf '%s\n' \
 '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"t","version":"0"}}}' \
 '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}' \
 | lark-cli mcp serve
```

---

## Sự cố thường gặp

### "command not found" trong Claude Desktop (macOS)
App giao diện không đọc PATH của terminal. → Trong `claude_desktop_config.json` dùng **đường dẫn tuyệt đối** (`which lark-cli`).

### Tool trả `isError: true`
Chạy lệnh `lark-cli ...` tương đương ở terminal để xem nguyên nhân (thường: thiếu scope, sai tham số, hoặc chưa `auth login`).

### Output lỗi / Claude mất kết nối giữa chừng
Có thứ gì đó in ra **stdout** không phải JSON-RPC. → Đặt `"env": {"NO_COLOR": "1"}` trong cấu hình host. **Không** bọc `mcp serve` trong script in ra stdout.

### Lỗi thiếu quyền (scope)
```bash
lark-cli auth login --help   # cấp thêm quyền
```
Nếu tổ chức kiểm soát chặt, admin Lark cần phê duyệt scope.

### `user: missing` khi `auth status`
Chưa đăng nhập danh tính user → `lark-cli auth login`. Nhiều skill cần danh tính user.

---

## ⚠️ Sự cố đã biết: "Server treo khi khởi động"

**Triệu chứng:** chạy `lark-cli mcp tools` hoặc `mcp serve` **không in gì** (treo) khi thư mục làm việc (cwd) **khác** thư mục mã nguồn; nhưng chạy **từ trong thư mục repo** thì ra ngay 21 tool. Vì Claude Desktop khởi chạy server từ một cwd khác, lỗi này có thể khiến connector treo.

**Đang điều tra** — nghi 1 trong 2:
- **Quyền chạy (macOS Gatekeeper):** binary mới bị "kiểm tra nhà phát triển" lần đầu.
- **Keychain prompt:** binary mới đổi chữ ký → macOS hỏi quyền truy cập keychain (hộp thoại nền chặn tiến trình).

**Cách khắc phục thử (theo thứ tự):**

```bash
# 1) Gỡ cờ kiểm dịch Gatekeeper
xattr -dr com.apple.quarantine ~/bin/lark-cli

# 2) Ký ad-hoc để ổn định chữ ký (giúp keychain không hỏi lại)
codesign -s - --force ~/bin/lark-cli

# 3) Chạy thử 1 lần ở Terminal để bấm "Always Allow" nếu có hộp thoại keychain
~/bin/lark-cli mcp tools
```

> Lưu ý: lệnh `lark-cli --version` có thể "treo" do trình kiểm tra cập nhật gọi mạng — **không** dùng nó để test. Dùng `mcp tools` (đặt `LARKSUITE_CLI_NO_UPDATE_NOTIFIER=1` nếu cần tắt thông báo cập nhật).

Sau khi hết treo: **Quit & mở lại Claude Desktop** để nạp binary mới.

---

## Tắt thông báo phiền (script tự động hoá)

| Biến môi trường | Tác dụng |
|---|---|
| `LARKSUITE_CLI_NO_UPDATE_NOTIFIER=1` | Tắt nhắc cập nhật binary |
| `LARKSUITE_CLI_NO_SKILLS_NOTIFIER=1` | Tắt nhắc đồng bộ skills |
| `NO_COLOR=1` | Tắt màu (bắt buộc cho MCP) |

---

## Sự cố cổng Web (claude.ai)

| Triệu chứng | Xử lý |
|---|---|
| Không kết nối | URL có đuôi `/mcp`? tunnel còn chạy? server `--transport http` còn sống? |
| Kết nối nhưng tool lỗi | Xem `--audit-log`; chạy lệnh tương đương ở terminal |
| Lo lộ tài khoản | Chưa bật xác thực trước cổng → tắt cổng, bật Access/token ([07](07-bao-mat-quyen-rieng-tu.md)) |

Báo cáo sức khoẻ 1 trang: lệnh `/mcp-doctor` (trong Cowork/Claude Code).
