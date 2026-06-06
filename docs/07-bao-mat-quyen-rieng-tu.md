# 07 — Bảo mật & quyền riêng tư

> Đối tượng: admin / IT / security + user quan tâm. Đọc trước khi mở cổng web.

## Data đi đâu?

```
Claude (Desktop/Web) ──▶ lark-cli (máy bạn) ──HTTPS──▶ open.larksuite.com / open.feishu.cn
```

- Khi dùng **Desktop (stdio)**: mọi thứ chạy **local**, data chỉ đi tới **máy chủ Lark**. Không bên thứ ba.
- Khi dùng **Web qua cổng HTTP**: data đi qua **cổng MCP bạn dựng** (tunnel/server). Cổng này thấy nội dung request → **phải bảo vệ** (mục bên dưới).
- Token đăng nhập nằm trong **keychain** hệ điều hành, không ở file thường.

## Cơ chế an toàn có sẵn

| Cơ chế | Bảo vệ điều gì |
|---|---|
| **Dry-run** trên mọi thao tác ghi | Xem trước nội dung trước khi gửi/tạo thật |
| **Mail mặc định lưu nháp** | Email không tự bay đi (cần `confirm_send=true`) |
| **Audit log** (`--audit-log <file>`) | Ghi lại mọi tool call (ai/khi/làm gì) để truy vết |
| **Bearer token** (`LARK_MCP_BEARER_TOKEN`) | Chặn truy cập HTTP trái phép — mọi request thiếu/sai token bị 401 |
| **Phân tách danh tính** user/bot | Giới hạn Claude thao tác đúng vai |
| **Keychain** | Token không lưu plaintext |

Bật bearer token + audit khi chạy server HTTP:

```bash
export LARK_MCP_BEARER_TOKEN=$(openssl rand -hex 32)
lark-cli mcp serve --transport http --addr 127.0.0.1:3000 \
  --audit-log ~/.lark-mcp-audit.ndjson
```

- Có token → log in `bearer-token auth ENABLED`; mọi POST `/` và `/mcp` cần header `Authorization: Bearer <token>`. Riêng `GET /health` luôn mở (liveness probe).
- Không token → log in cảnh báo `UNAUTHENTICATED`; chỉ chấp nhận khi server còn bound `127.0.0.1`.
- So sánh token dùng constant-time (`crypto/subtle`) → không lộ token qua timing.

## ⚠️ Khi mở cổng web — bắt buộc làm

Cổng HTTP (`--transport http`) **phơi toàn quyền tài khoản Lark của bạn**. Ai chạm được URL = thao tác được Lark của bạn. Trước khi mở ra internet:

1. **Đặt lớp xác thực trước cổng** — nên dùng cả hai:
   - **Bearer token built-in** qua `LARK_MCP_BEARER_TOKEN` (xem mục trên) — tự kiểm bằng `curl` thiếu token phải nhận `401` trước khi mở tunnel.
   - **Cloudflare Access** trên hostname tunnel (email/OTP) — lớp 2, nhất là với Named Tunnel/URL cố định.
2. **Không bao giờ** để URL tunnel trần, public, không token.
3. Bật `--audit-log` để có dấu vết.
4. Chỉ mở cổng khi cần; tắt khi xong (quick tunnel đổi URL mỗi lần — tốt cho tạm thời).

## Khuyến nghị governance (tổ chức)

- **Allowlist/denylist tool:** giới hạn nhóm công cụ AI được gọi; chặn verb nguy hiểm (gửi ra ngoài, xoá, đổi quyền). *(Lớp policy — cần admin cấu hình; xem `lark-cli config` và `cmd/mcp/README.md`.)*
- **App Lark riêng** với scope tối thiểu cần thiết (least privilege) thay vì app công khai — xem [04](04-dang-nhap-va-quyen.md).
- **Pilot có kiểm soát** trước khi mở rộng; rà `--audit-log` định kỳ.
- **permission-audit skill**: quét Drive/Doc/Wiki/Base tìm chia sẻ public/ngoài tổ chức/PII — chạy định kỳ.

## Triển khai đa người dùng (lưu ý)

Mỗi nhân viên cần danh tính Lark **riêng**. Cổng web đơn giản (1 binary, 1 tài khoản) **không** phù hợp cho nhiều người dùng chung. Đa-tenant cần kiến trúc **sidecar/gateway** tách credential từng người — việc của admin/IT, ngoài phạm vi cài cá nhân.

## Câu hỏi kiểm toán nhanh (checklist)

- [ ] Token nằm trong keychain? (mặc định: có)
- [ ] Cổng web đã có lớp xác thực? (`LARK_MCP_BEARER_TOKEN` đã set + `curl` thiếu token nhận 401)
- [ ] `--audit-log` đã bật cho cổng dùng chung?
- [ ] Dùng app Lark riêng + scope tối thiểu cho enterprise?
- [ ] Đã chạy `permission-audit` rà quyền rủi ro?
