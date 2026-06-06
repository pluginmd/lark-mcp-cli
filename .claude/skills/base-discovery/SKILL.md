---
name: base-discovery
description: Phase 0 của base-deploy — phỏng vấn/chốt yêu cầu vận hành thành spec (entity, KPI, role, nguồn dữ liệu) ghi vào build-plan.json. Triggers "phase discovery", "chốt yêu cầu base", "base-discovery".
version: 1.0.0
last_updated: 2026-06-06
---

# base-discovery (Phase 0 · 1 agent · sequential)

"Hiểu yêu cầu, xác định phạm vi, dữ liệu & KPI." Biến mô tả thô của user
thành phần `meta` + danh sách entity thô trong `build-plan.json`. KHÔNG
thiết kế field/quan hệ (đó là DESIGN), KHÔNG ghi Base.

Đầu ra ghi vào `build-plan.json`: `meta{base_title, team_size,
owner_open_id, goal, kpis[]}`, `imports[].source` (nguồn dữ liệu thô), và
danh sách entity (→ `tables[].key`+`name`, chưa có field). Schema:
[build-plan-schema.md](../base-deploy/references/build-plan-schema.md).

## Workflow

1. Đọc memory orchestrator (`team-roster.md`, `templates.md`). Nếu yêu
   cầu khớp một template → đề xuất dùng làm điểm xuất phát.
2. **Chốt 5 trục** (hỏi gọn, gộp 1 lượt, đừng tra tấn từng câu):
   - **Mục tiêu**: vận hành cái gì? (dự án / CRM / tuyển dụng / tài sản…)
   - **Entity**: những "bảng" nào? ai/cái gì được theo dõi? (→ tables)
   - **Người dùng**: bao nhiêu người, vai trò nào (PM/member/manager)?
   - **KPI/câu hỏi**: dashboard cần trả lời câu gì? (→ `meta.kpis`)
   - **Nguồn dữ liệu**: có Excel/CSV/Base cũ để import không? đường dẫn?
3. Resolve owner + người chủ chốt → `open_id` qua `lark-contact` (lưu
   `team-roster.md`). Tên người, đừng để ID trần trong spec.
4. **Reference-first**: nếu user có Base/bảng mẫu đang chạy → xin link,
   gợi ý snapshot làm fixture cho BUILD khớp theo (CLAUDE.md G-11).
5. Ghi `build-plan.json` (chỉ phần Phase 0). In lại spec gọn cho user xác
   nhận trước khi sang DESIGN.

## Hard rules

1. Không ghi Base, không tạo field. Chỉ điền meta + entity thô + nguồn.
2. KPI phải đo được từ dữ liệu (count/sum/%/so deadline) — không KPI mơ
   hồ kiểu "hiệu quả hơn".
3. Mỗi entity phải có lý do tồn tại; nghi ngờ thừa bảng → hỏi gộp.
4. Nguồn dữ liệu: ghi đường dẫn file thật + định dạng; chưa có thì đánh
   dấu `imports: []` và báo IMPORT sẽ skip.
5. Tên người → `open_id` qua lark-contact, không bịa.

## Edge cases

- User mô tả mơ hồ "quản lý công việc" → đề xuất template chuẩn (Dự án +
  Task + Member) rồi để user trừ bớt.
- >6 entity → cảnh báo phức tạp, đề xuất giai đoạn 1 chỉ 3 bảng lõi.
- Không có nguồn dữ liệu → vẫn hợp lệ, Base rỗng nhập tay sau.

## Memory

Đọc/cập nhật `team-roster.md` (name→open_id), `templates.md`. Ghi mới
`base-conventions.md` nếu user nêu quy ước đặt tên.
