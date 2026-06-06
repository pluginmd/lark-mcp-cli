# 04 — Đăng nhập Lark & phân quyền

> Làm 1 lần. Quyết định Claude được làm gì *thay mặt bạn*.

## Hai danh tính: User vs Bot

| Danh tính | Là gì | Dùng khi |
|---|---|---|
| **User** (bạn) | Thao tác *thay mặt bạn* — đọc mail/lịch/task của chính bạn | Hầu hết việc cá nhân (digest, inbox, 1:1…) |
| **Bot** (ứng dụng) | Thao tác dưới danh nghĩa app/bot | Gửi thông báo vào group, automation |

Phần lớn skill Cowork cần **danh tính User** → bạn phải `auth login`.

## Đăng nhập

```bash
lark-cli auth login          # mở trình duyệt, đăng nhập, cấp quyền
lark-cli auth status         # kiểm tra: user = ready
```

Kết quả `auth status` mẫu:

```
bot  → ready          (sẵn sàng)
user → missing/ready  (missing = chưa login, hãy chạy auth login)
```

Token lưu trong **keychain** của hệ điều hành (macOS Keychain / Windows Credential Manager) — không nằm ở file thường.

## Chọn danh tính cho một thao tác

Mặc định `auto`. Có thể ép:

```bash
lark-cli <lệnh> --as user     # chạy như chính bạn
lark-cli <lệnh> --as bot      # chạy như bot
```

Trong các tool MCP, tham số `as` (`user`/`bot`) làm điều tương tự.

## Quyền (scopes)

`lark-cli` xin các quyền tương ứng với việc nó làm (đọc mail, đọc lịch, gửi tin nhắn…). Nếu một thao tác báo lỗi **thiếu scope**:

```bash
lark-cli doctor              # kiểm tra cấu hình, auth, kết nối
lark-cli auth login --help   # xem cách cấp thêm quyền
```

> Quyền cấp ở bước đăng nhập trên trình duyệt. Nếu tổ chức kiểm soát chặt, admin Lark có thể cần phê duyệt scope cho ứng dụng.

## Dùng ứng dụng Lark riêng của tổ chức (nâng cao)

Mặc định dùng một app Lark công khai — tiện cho cá nhân. Với **doanh nghiệp**, nên tạo **app Lark riêng** (App ID/Secret của tổ chức) và cấu hình scope chuẩn để:

- Kiểm soát đúng phạm vi quyền.
- Quản lý tập trung, thu hồi được.
- Đặt redirect URL riêng (mặc định `http://localhost:3000/callback`).

Việc này do admin/IT làm; xem `lark-cli config --help` và [07 — Bảo mật](07-bao-mat-quyen-rieng-tu.md).

## Đăng xuất / đổi tài khoản

```bash
lark-cli auth logout
lark-cli auth login          # đăng nhập tài khoản khác
```
