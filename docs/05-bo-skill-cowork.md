# 05 — Bộ skill Cowork: hỏi gì, làm được gì

> 23 "công thức" sẵn cho việc thường ngày. Bạn gõ câu kích hoạt (tiếng Việt), Claude tự chạy chuỗi công cụ Lark.
> Vị trí: [`.claude/skills/`](../.claude/skills/).

## Cách dùng

Chỉ cần **nói tự nhiên**. Mỗi skill có "câu kích hoạt" gợi ý bên dưới — bạn không phải nhớ chính xác, nói gần đúng là được.

> Mẹo: bắt đầu ngày bằng **"morning"** để có bản tóm tắt đầu ngày.

## Nhóm "Đầu ngày / Tổng hợp"

| Skill | Hỏi gì | Làm gì |
|---|---|---|
| **morning-brief** | "morning", "sáng nay có gì" | Bản tóm đầu ngày: gộp mail + chat + approval + task ưu tiên (≤15 dòng). Có 5 biến thể: default/ic/exec/pm/sales |
| **daily-digest** | "tổng kết hôm nay", "wrap up" | Tóm cuối ngày: họp đã dự + task đã xong + điểm nổi bật trong inbox |
| **weekly-review** | "weekly review", "báo cáo tuần" | Tổng hợp lịch + task + OKR thành một mạch tường thuật |
| **overwhelm-triage** | "tôi quá tải" | Hỏi 1 câu để biết bạn ngợp ở đâu (mail/chat/task/họp) rồi điều hướng đúng skill |

## Nhóm "Mail & Chat"

| Skill | Hỏi gì | Làm gì |
|---|---|---|
| **inbox-zero** | "clear inbox", "xử lý hết mail tồn" | Phân loại mail urgent/important/fyi/noise, hướng tới inbox trống |
| **im-digest** | "các group có gì", "tóm tắt chat" | Phân loại N tin mới mỗi group thành cần-action / cần-biết / bỏ qua |
| **client-followup** | "khách im lặng", "follow-up khách" | Phát hiện liên hệ CRM lâu chưa động (>21 ngày), **soạn nháp** mail tái kết nối (chỉ nháp, không tự gửi) |

## Nhóm "Họp & Lịch"

| Skill | Hỏi gì | Làm gì |
|---|---|---|
| **meeting-prep** | "chuẩn bị họp X", "action items từ meeting" | Trước họp: gom context. Sau họp: trích action item thành task |
| **calendar-optimizer** | "tôi họp quá nhiều" | Soi 30 ngày họp, gợi ý decline/gộp/chuyển-async |
| **focus-mode** | "focus 2 tiếng", "DND", "deep work" | Chặn lịch + bật DND + báo team |
| **one-on-one-prep** | "1:1 prep với <người>" | Brief 1:1: OKR, task gần đây, ghi chú cũ, câu hỏi gợi ý |
| **contact-360** | "tôi sắp gặp <tên>" | Hồ sơ 360°: IM + mail + họp + doc + task của người đó |

## Nhóm "Task & Approval"

| Skill | Hỏi gì | Làm gì |
|---|---|---|
| **task-prioritizer** | "việc nào quan trọng", "top 5 today" | Xếp hạng task theo deadline × rủi ro × OKR × người giao; nêu lý do |
| **approval-triage** | "có gì cần duyệt", "queue duyệt" | Đọc approval pending, gợi ý APPROVE/CHECK/REJECT kèm trích policy |

## Nhóm "Tài liệu & Tri thức"

| Skill | Hỏi gì | Làm gì |
|---|---|---|
| **doc-from-template** | "tạo doc theo template", "weekly report doc" | Soạn doc/wiki/sheet từ template có tên |
| **doc-restructure** | "wiki bị bừa", "wiki cleanup" | Soi Wiki: trang cũ/mồ côi/trùng → đề xuất archive/gộp/đổi cha |
| **decision-logger** | "log lại decision", "chốt cái này" | Phát hiện quyết định trong IM/minutes → ghi vào bảng Base |
| **permission-audit** | "permission audit", "quét quyền", "PII" | Quét Drive/Doc/Wiki/Base tìm quyền rủi ro (public/ngoài tổ chức/PII) — chỉ đọc |

## Nhóm "Sales / Vận hành"

| Skill | Hỏi gì | Làm gì |
|---|---|---|
| **deal-update** | "cập nhật deal sau gọi" | Lấy minutes → trích pain/budget/timeline → cập nhật Base → soạn nháp follow-up |
| **pipeline-review** | "pipeline review", "tổng quan deal" | Quét pipeline theo stage, deal kẹt, sắp chốt, xu hướng win-rate |
| **incident-retro** | "postmortem cho SEV-X" | Dựng postmortem blameless từ timeline IM on-call |
| **sprint-retro** | "sprint retro", "/retro" | Bản nháp retro cuối sprint: ticket đã đóng, velocity, blocker |

## Nhóm "Kỹ thuật" (cho người mở rộng)

| Skill | Dùng khi |
|---|---|
| **lark-cli-mcp** | Thêm/sửa/gỡ công cụ MCP, vá lỗi Claude Desktop mất kết nối, đổi schema tool |

## Lệnh nhanh `/mcp-*` (trong Claude Code/Cowork)

Trong [`.claude/commands/`](../.claude/commands/): `/mcp-tools` (liệt kê tool), `/mcp-test` (smoke-test), `/mcp-call` (gọi 1 tool), `/mcp-doctor` (báo cáo sức khoẻ), `/mcp-add` (thêm tool mới), `/mcp-rebuild` (dựng lại binary).

---

> An toàn: các skill có thao tác **gửi** (mail, tin nhắn) luôn **soạn nháp / xem trước** trước, không tự gửi. Bạn duyệt rồi mới thật. Xem [07](07-bao-mat-quyen-rieng-tu.md).
