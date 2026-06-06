# 01 — Tổng quan & giá trị

> Đối tượng: người quyết định mua/triển khai + mọi user. Đọc trong 8 phút.

## Một câu

`lark-cli + MCP bridge` cho phép **AI (Claude) tự thao tác Lark/Feishu hộ bạn** — bằng prompt tiếng Việt, ngay trong Claude Desktop hoặc claude.ai. Không cần mở Lark, không copy-paste, không cần biết lập trình.

## Bài toán nó giải

Một ngày của leader/nhân viên trên Lark:

| Việc | Thời gian/ngày | Pain |
|---|---|---|
| Đọc & phân loại mail | 60–90 phút | Quá nhiều noise |
| Lướt hàng chục group chat | 30–45 phút | Sợ bỏ sót VIP/khách |
| Tóm các cuộc họp để follow-up | ~45 phút | Mở từng minutes, copy action |
| Duyệt approval pending | ~25 phút | Mở app, đọc policy, click |
| Sắp xếp task | ~20 phút | Không biết bắt đầu từ đâu |
| Chuẩn bị 1:1 / gặp khách | ~40 phút | Lục lại lịch sử IM/mail/task |

→ Claude làm hộ phần **đọc – phân loại – tóm – dự thảo**. Bạn chỉ **quyết định & bấm gửi**.

## Tại sao cần "MCP bridge" (không chỉ CLI)?

`lark-cli` gốc là công cụ dòng lệnh — rất mạnh nhưng **chỉ lập trình viên dùng được**, và **Claude không tự gọi được** nó. MCP bridge (`lark-cli mcp serve`) là lớp dịch giúp Claude:

- **Thấy** danh sách việc làm được (21 công cụ: gửi mail, tạo task, tìm doc…).
- **Tự gọi** đúng công cụ với đúng tham số khi bạn ra lệnh bằng lời.

> MCP = Model Context Protocol — chuẩn mở để AI kết nối công cụ ngoài. Claude Desktop và claude.ai đều hỗ trợ.

## Cách hoạt động (sơ đồ)

```
Bạn ──prompt tiếng Việt──▶  Claude (Desktop/Web)
                                  │ chọn công cụ + tham số
                                  ▼
                        lark-cli mcp serve   (chạy trên máy bạn)
                                  │ gọi lệnh lark-cli, dùng tài khoản của BẠN
                                  ▼
                        open.larksuite.com / open.feishu.cn  (API Lark)
```

Toàn bộ chạy **local bằng credential của bạn**. Không có server trung gian nào giữ data (trừ trường hợp dùng cổng web — xem [03](03-ket-noi-web-claude-ai.md)).

## Bạn nhận được gì

1. **21 công cụ MCP** phủ IM, Mail, Calendar, Docs, Base, Contact, Task, Drive, Sheets, Meetings, OKR ([danh sách](06-cong-cu-mcp.md)).
2. **23 bộ skill Cowork** — "công thức" sẵn cho việc thường ngày: morning-brief, inbox-zero, meeting-prep, task-prioritizer… ([chi tiết](05-bo-skill-cowork.md)).
3. **An toàn mặc định** — xem trước mọi thao tác ghi, log đầy đủ ([bảo mật](07-bao-mat-quyen-rieng-tu.md)).

## Giá trị đo được

- Tiết kiệm thực tế ước tính **60–90 phút/người/ngày** ở team có workload Lark cao.
- Chi phí hạ tầng **$0** (chạy local) + chi phí AI theo gói Claude bạn đang dùng.
- Mã nguồn mở, không vendor lock-in.

> Lưu ý trung thực: ROI phụ thuộc workload Lark thật và mức độ dùng đều. Workload thấp → giá trị thấp. Nên chạy **pilot 5 người × 2 tuần** trước khi mở rộng.
