# Tài liệu lark-cli + MCP — dành cho Business User

> Mục tiêu: biến **Lark/Feishu** thành thứ mà **Claude tự thao tác được** — đọc mail, tóm họp, dọn task, gửi tin nhắn, tạo doc, duyệt approval — bằng **prompt tiếng Việt**, ngay trong **Claude Desktop** hoặc **claude.ai (web)**. Bạn **không cần là lập trình viên**, không cần Claude Code.

## Bắt đầu nhanh (3 bước)

1. **Cài** lark-cli vào máy → xem [02 — Cài đặt Claude Desktop](02-cai-dat-claude-desktop.md)
2. **Đăng nhập** Lark một lần → xem [04 — Đăng nhập & quyền](04-dang-nhap-va-quyen.md)
3. **Hỏi Claude** "Sáng nay tôi có gì?" → xem [05 — Bộ skill Cowork](05-bo-skill-cowork.md)

## Mục lục

| # | Tài liệu | Cho ai |
|---|---|---|
| 01 | [Tổng quan & giá trị](01-tong-quan.md) | Người quyết định, mọi user |
| 02 | [Cài đặt trên Claude Desktop (Cowork)](02-cai-dat-claude-desktop.md) | User desktop |
| 03 | [Kết nối Lark MCP với web claude.ai](03-ket-noi-web-claude-ai.md) | User web / admin |
| 04 | [Đăng nhập Lark & phân quyền](04-dang-nhap-va-quyen.md) | Mọi user |
| 05 | [Bộ skill Cowork — hỏi gì, làm được gì](05-bo-skill-cowork.md) | Mọi user |
| 06 | [Danh sách công cụ MCP (21 tool)](06-cong-cu-mcp.md) | User nâng cao / admin |
| 07 | [Bảo mật & quyền riêng tư](07-bao-mat-quyen-rieng-tu.md) | Admin / IT / Security |
| 08 | [Xử lý sự cố](08-xu-ly-su-co.md) | Mọi user |
| 09 | [Cập nhật & bảo trì](09-cap-nhat-bao-tri.md) | Admin |
| — | [Nhật ký bản dựng (kỹ thuật)](CHANGELOG-UPDATE.md) | Người triển khai |

## 2 cách dùng

| | **Claude Desktop (Cowork)** | **claude.ai (web)** |
|---|---|---|
| Chạy ở đâu | Máy bạn (local) | Trình duyệt + server MCP công khai |
| Độ dễ | ⭐ Dễ — cài 1 lần | ⭐⭐⭐ Khó hơn — cần cổng MCP HTTPS |
| Phù hợp | Cá nhân, mọi nhân viên | Tổ chức có admin/IT vận hành cổng |
| Hướng dẫn | [02](02-cai-dat-claude-desktop.md) | [03](03-ket-noi-web-claude-ai.md) |

## Quan trọng — đọc trước

- Mọi thao tác **ghi** (gửi mail, tạo task, gửi tin nhắn) đều **xem trước (dry-run)** rồi mới thực thi. Claude sẽ hỏi/cho bạn duyệt trước khi gửi thật. Xem [07 — Bảo mật](07-bao-mat-quyen-rieng-tu.md).
- Data của bạn **chỉ đi tới máy chủ Lark** (open.larksuite.com / open.feishu.cn), **không** qua bên thứ ba.
- Mã nguồn mở (MIT) — kiểm chứng & tuỳ biến được.
