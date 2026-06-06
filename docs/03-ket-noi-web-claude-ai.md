# 03 — Kết nối Lark MCP với web claude.ai

> Kết quả: dùng được các công cụ Lark ngay trong **claude.ai trên trình duyệt** (không cần app desktop).
> Độ khó: ⭐⭐⭐ — cần một **cổng MCP HTTPS công khai**. Nên có admin/IT hỗ trợ.

## Vì sao web khó hơn desktop?

- Claude Desktop **chạy binary local** trên máy bạn (stdio) → đơn giản.
- claude.ai **chạy trên server của Anthropic** → không với tới binary trên máy bạn. Nó chỉ kết nối được tới **một địa chỉ HTTPS công khai** (Custom Connector).

→ Phải cho `lark-cli mcp serve` chạy ở chế độ **HTTP** rồi **mở ra internet** qua một đường hầm (tunnel) hoặc server có domain.

```
claude.ai ──HTTPS──▶  Cloudflare Tunnel  ──▶  lark-cli mcp serve --transport http  (máy bạn)
```

---

## Cách 1 — Cloudflare Tunnel (nhanh, miễn phí, cho cá nhân)

### Bước 1 — Chạy server HTTP (kèm bearer token)

Sinh một token bí mật rồi truyền qua biến môi trường `LARK_MCP_BEARER_TOKEN`. Khi biến này có giá trị, mọi request tới `/` và `/mcp` **bắt buộc** kèm header `Authorization: Bearer <token>` (riêng `/health` để mở cho liveness probe):

```bash
export LARK_MCP_BEARER_TOKEN=$(openssl rand -hex 32)
echo "Token (lưu lại để khai báo connector): $LARK_MCP_BEARER_TOKEN"

~/bin/lark-cli mcp serve --transport http --addr 127.0.0.1:3000 \
  --audit-log ~/lark-mcp-audit.ndjson
```

Khi khởi động, log sẽ in `bearer-token auth ENABLED`. Nếu **không** đặt biến này, log in cảnh báo `UNAUTHENTICATED` — chỉ được dùng khi server còn bound `127.0.0.1`, **tuyệt đối không** mở tunnel ở trạng thái đó.

### Bước 2 — Mở đường hầm công khai

```bash
cloudflared tunnel --url http://127.0.0.1:3000
```

Lệnh in ra một URL dạng `https://<ngẫu-nhiên>.trycloudflare.com`.

### Bước 3 — ⚠️ Thêm lớp bảo mật (BẮT BUỘC)

Cổng này **mở toàn quyền tài khoản Lark của bạn ra internet**. Tuyệt đối không để trần. Có 2 lớp (nên dùng cả hai):

- **Bearer token built-in** — đã bật ở Bước 1 qua `LARK_MCP_BEARER_TOKEN`. Tự kiểm chứng trước khi mở tunnel:
  ```bash
  # không token -> phải trả 401
  curl -s -o /dev/null -w "%{http_code}\n" -X POST http://127.0.0.1:3000/mcp \
    -H 'Content-Type: application/json' -d '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}'
  # đúng token -> 200 + danh sách tool
  curl -s -X POST http://127.0.0.1:3000/mcp -H 'Content-Type: application/json' \
    -H "Authorization: Bearer $LARK_MCP_BEARER_TOKEN" \
    -d '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}' | head -c 200
  ```
- **Cloudflare Access** đặt trước hostname tunnel (yêu cầu email/OTP) — khuyến nghị thêm cho lớp 2, đặc biệt khi dùng Named Tunnel.

Xem chi tiết: [07 — Bảo mật & quyền riêng tư](07-bao-mat-quyen-rieng-tu.md).

### Bước 4 — Khai báo trên claude.ai

1. Vào **Settings → Connectors → Add custom connector**.
2. Dán URL: `https://<...>.trycloudflare.com/mcp`.
3. Nếu form connector có ô **Authentication / custom header**: thêm `Authorization: Bearer <token>` (token ở Bước 1).
4. Lưu → claude.ai kết nối → liệt kê 21 tool.
5. Thử: *"Tìm doc tên Kế hoạch Q3"*.

> ⚠️ **Lưu ý tương thích:** một số phiên bản UI custom connector của claude.ai **không** cho thêm header tĩnh (chỉ hỗ trợ OAuth hoặc no-auth). Nếu rơi vào trường hợp đó, **đừng** mở connector no-auth ra tunnel trần — hãy dùng **Cloudflare Access** trên hostname tunnel làm lớp chặn (Bước 3), hoặc dựng một OAuth shim phía trước. Bearer token built-in vẫn luôn bảo vệ được ở tầng `curl`/script và các host MCP cho phép set header (Claude Code, Cursor…).

---

## Cách 2 — Server có domain + TLS (bền vững, cho tổ chức)

Dành cho triển khai lâu dài, nhiều người dùng:

1. Đưa `lark-cli mcp serve --transport http` lên một server (VM/VPS).
2. Đặt **Caddy/nginx** phía trước để có **HTTPS + domain cố định**.
3. Bật xác thực (Cloudflare Access / OAuth proxy / bearer token).
4. Khai báo URL domain đó làm Custom Connector trên claude.ai.

> Triển khai **đa người dùng** (mỗi nhân viên một danh tính Lark riêng) cần kiến trúc **sidecar/gateway đa-tenant** — không nằm trong phạm vi cài cá nhân. Liên hệ admin; tham khảo [07 — Bảo mật](07-bao-mat-quyen-rieng-tu.md).

---

## Giữ cho cổng luôn sống (tuỳ chọn)

Tunnel quick (Cách 1) **đổi URL mỗi lần chạy** và tắt khi đóng máy. Muốn ổn định:

- Dùng **Named Tunnel** của Cloudflare (URL cố định).
- Tạo dịch vụ nền (**launchd** trên macOS / **systemd** trên Linux) để `mcp serve` + tunnel tự chạy lại.

---

## So sánh nhanh

| | Cloudflare Tunnel | Server + domain |
|---|---|---|
| Chi phí | Miễn phí | VPS + domain |
| URL cố định | Không (trừ Named Tunnel) | Có |
| Phù hợp | 1 người, thử nghiệm | Tổ chức, lâu dài |
| Bảo mật | Cần tự thêm | Đầy đủ hơn |

---

## Sự cố thường gặp

| Triệu chứng | Cách xử lý |
|---|---|
| claude.ai không kết nối được | Kiểm tra URL có đuôi `/mcp`, tunnel còn chạy, server `--transport http` còn sống |
| Kết nối nhưng gọi tool lỗi | Xem `--audit-log`; chạy lệnh tương đương ở terminal để biết lý do |
| Lo ngại lộ tài khoản | Chưa bật lớp bảo mật ở Bước 3 → **tắt cổng ngay**, bật Access/token rồi mới mở lại |
