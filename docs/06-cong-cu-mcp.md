# 06 — Danh sách công cụ MCP (21 tool)

> Đây là những việc Claude **tự gọi được** khi bạn ra lệnh. Bạn không cần nhớ — Claude tự chọn. Bảng này để tham chiếu / kiểm toán.
> Xem trực tiếp danh sách đang chạy: `lark-cli mcp tools`.

## Nguyên tắc an toàn chung

- Tool **ghi** (gửi/tạo) đều có `dry_run` để **xem trước** — Claude sẽ preview rồi mới thực thi.
- `lark_mail_send` mặc định **lưu nháp**; chỉ gửi thật khi `confirm_send=true`.
- Tham số `as` = `user` (mặc định cho việc cá nhân) hoặc `bot`.

## IM (Tin nhắn)

| Tool | Việc | Ghi chú an toàn |
|---|---|---|
| `lark_im_send` | Gửi tin text/markdown tới 1 chat hoặc 1 người | dry_run xem trước |
| `lark_im_card_send` | Gửi **card tương tác** (header, panel, nút) bằng YAML | print_json/dry_run trước |
| `lark_im_search` | Tìm tin nhắn theo từ khoá | chỉ đọc |

## Mail

| Tool | Việc | Ghi chú |
|---|---|---|
| `lark_mail_send` | Gửi email | **Mặc định lưu nháp**; gửi thật cần `confirm_send=true` |
| `lark_mail_draft_create` | Tạo nháp trong Drafts | Không bao giờ tự gửi |

## Calendar (Lịch)

| Tool | Việc | Ghi chú |
|---|---|---|
| `lark_calendar_agenda` | Liệt kê lịch sắp tới (mặc định hôm nay) | chỉ đọc |
| `lark_calendar_create` | Tạo sự kiện/mời người | dry_run trước; giải mã người mời qua `lark_contact_search` |

## Docs / Drive / Sheets

| Tool | Việc | Ghi chú |
|---|---|---|
| `lark_doc_create` | Tạo Lark Doc từ tiêu đề + markdown | dry_run trước |
| `lark_doc_search` | Tìm doc theo từ khoá | chỉ đọc |
| `lark_doc_fetch` | Đọc toàn bộ nội dung doc (markdown) | chỉ đọc |
| `lark_drive_upload` | Tải file lên Drive | dry_run xem kế hoạch |
| `lark_sheets_read` | Đọc vùng ô của Sheet | chỉ đọc |
| `lark_sheets_append` | Thêm dòng vào Sheet | dry_run trước |

## Base (Bitable)

| Tool | Việc | Ghi chú |
|---|---|---|
| `lark_base_search` | Tìm bản ghi trong bảng Base theo từ khoá | chỉ đọc |

## Contact (Danh bạ)

| Tool | Việc | Ghi chú |
|---|---|---|
| `lark_contact_search` | Tra cứu người theo tên/email/điện thoại hoặc open_id | Dùng để lấy ID trước khi mời/giao việc |

## Task

| Tool | Việc | Ghi chú |
|---|---|---|
| `lark_task_my` | Liệt kê task của bạn (lọc theo trạng thái/hạn) | chỉ đọc |
| `lark_task_create` | Tạo task (1 người nhận/lần) | dry_run trước; lấy open_id qua contact_search |

## Meetings (VC / Minutes)

| Tool | Việc | Ghi chú |
|---|---|---|
| `lark_vc_search` | Tìm cuộc họp đã qua (theo từ khoá/thời gian/người) | chỉ đọc |
| `lark_minutes_search` | Tìm bản ghi/biên bản (transcript) cuộc họp | chỉ đọc |

## OKR

| Tool | Việc | Ghi chú |
|---|---|---|
| `lark_okr_cycle_list` | Liệt kê chu kỳ OKR ('me' hoặc người khác) | chỉ đọc |

## Cửa thoát hiểm

| Tool | Việc | Ghi chú |
|---|---|---|
| `lark_api` | Gọi **bất kỳ** Open API nào của Lark khi không có tool chuyên dụng | Hỗ trợ dry_run; dành cho nhu cầu nâng cao |

---

> Muốn **thêm tool mới**? Dùng skill `lark-cli-mcp` + lệnh `/mcp-add`. Kiến trúc trong [`cmd/mcp/README.md`](../cmd/mcp/README.md).
