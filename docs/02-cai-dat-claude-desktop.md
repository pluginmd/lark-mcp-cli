# 02 — Cài đặt trên Claude Desktop (Cowork)

> Kết quả: trong Claude Desktop, bạn gõ "Sáng nay tôi có gì?" và Claude tự đọc Lark trả lời.
> Thời gian: ~10–15 phút. Cần làm 1 lần.

## Bạn cần gì

- macOS hoặc Windows, đã cài **Claude Desktop**.
- Tài khoản **Lark/Feishu** (để đăng nhập).
- (Cách A) Không cần gì thêm — dùng binary dựng sẵn.
- (Cách B) Nếu tự dựng từ mã nguồn: cần **Go ≥ 1.23** và **Python 3**.

---

## Bước 1 — Cài đặt `lark-cli` (có lệnh `mcp`)

### Cách A — Dùng script dựng sẵn (khuyến nghị)

Trong thư mục mã nguồn, chạy:

```bash
./scripts/setup-mcp.sh            # cài vào ~/bin/lark-cli
# hoặc cài toàn máy:
./scripts/setup-mcp.sh /usr/local/bin
```

Script sẽ: kiểm tra toolchain → build → cài binary → chạy `lark-cli mcp tools` để xác nhận.

### Cách B — Build tay

```bash
make build                        # tạo ./lark-cli
cp ./lark-cli ~/bin/lark-cli      # hoặc /usr/local/bin
```

### Kiểm tra đã cài đúng

```bash
~/bin/lark-cli mcp tools          # phải in ra 21 tool
```

> ⚠️ **macOS — đường dẫn tuyệt đối.** Claude Desktop là app giao diện, **không** đọc PATH của terminal. Luôn dùng đường dẫn đầy đủ (ví dụ `/Users/<bạn>/bin/lark-cli`) trong cấu hình ở Bước 3. Lấy đường dẫn bằng `which lark-cli`.

---

## Bước 2 — Đăng nhập Lark

```bash
~/bin/lark-cli auth login
```

Trình duyệt mở ra → đăng nhập Lark → cấp quyền. Xong, token lưu an toàn trong **keychain** máy bạn. Chi tiết quyền/scope: [04 — Đăng nhập & quyền](04-dang-nhap-va-quyen.md).

Kiểm tra:

```bash
~/bin/lark-cli auth status        # user: ready
```

---

## Bước 3 — Khai báo vào Claude Desktop

Mở file cấu hình:

- **macOS:** `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows:** `%APPDATA%\Claude\claude_desktop_config.json`

Gộp khối `lark-cli` vào `mcpServers` (giữ nguyên các server khác nếu có):

```json
{
  "mcpServers": {
    "lark-cli": {
      "command": "/Users/<BẠN>/bin/lark-cli",
      "args": ["mcp", "serve"],
      "env": { "NO_COLOR": "1" }
    }
  }
}
```

- `command`: **đường dẫn tuyệt đối** tới binary (Bước 1).
- `env.NO_COLOR=1`: bắt buộc — tránh ký tự màu làm hỏng giao thức.

> Mẫu sẵn: [`examples/mcp-hosts/claude-desktop.json`](../examples/mcp-hosts/claude-desktop.json).

---

## Bước 4 — Khởi động lại & kiểm tra

1. **Thoát hẳn** Claude Desktop (Quit) rồi mở lại.
2. Vào phần kết nối/Connectors → thấy **`lark-cli`** trạng thái *connected*, liệt kê 21 tool.
3. Thử trong Cowork: *"Liệt kê lịch hôm nay của tôi"* hoặc *"Tìm liên hệ tên Nguyễn Văn A"*.

---

## (Tuỳ chọn) Bộ skill Cowork

Các "công thức" như `morning-brief`, `inbox-zero`, `meeting-prep` nằm trong [`.claude/skills/`](../.claude/skills/). Đặt chúng vào không gian Cowork để gõ lệnh ngắn (ví dụ "morning") thay vì mô tả dài. Xem [05 — Bộ skill Cowork](05-bo-skill-cowork.md).

---

## Sự cố thường gặp

| Triệu chứng | Cách xử lý |
|---|---|
| `command not found` trong Claude | Dùng đường dẫn tuyệt đối (`which lark-cli`) ở Bước 3 |
| Tool trả lỗi `isError` | Chạy lệnh `lark-cli ...` tương đương ở terminal để xem lý do thật |
| Output lỗi / Claude mất kết nối | Đảm bảo `NO_COLOR=1`; không bọc `mcp serve` trong script in ra stdout |
| Claude báo `lark-cli` không phản hồi (treo) | Xem [08 — Xử lý sự cố §"Server treo khi khởi động"](08-xu-ly-su-co.md) |

Đầy đủ: [08 — Xử lý sự cố](08-xu-ly-su-co.md).
