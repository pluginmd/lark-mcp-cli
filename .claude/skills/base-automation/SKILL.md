---
name: base-automation
description: Phase 6 của base-deploy — dựng Base workflow tự động + bot nhắc việc/thông báo theo build-plan.json. Triggers "phase automation", "tự động hoá base", "bot nhắc việc base", "base-automation".
version: 1.0.0
last_updated: 2026-06-06
---

# base-automation (Phase 6 · 1 agent · sequential)

"Thiết lập automation, bot notification, nhắc việc & tự động hoá." Hiện
thực hoá `automation[]` trong `build-plan.json`: Base workflow (trigger →
action) + thông báo qua bot/IM. Verb + `steps` JSON route về `lark-base`
(`lark-base-workflow-guide.md`, `lark-base-workflow-schema.md`); thông
báo route về `lark-im`.

## Workflow

1. **Đọc `automation[]`.** Mỗi mục có `trigger` (record tạo/sửa, field
   đổi, tới hạn) + `action` (nhắc người, cập nhật field, thông báo group).
2. **Dựng workflow.** `base +workflow-create` với `steps` JSON theo
   SSOT `lark-base-workflow-schema.md` (BẮT BUỘC đọc trước — steps là
   phần phức tạp nhất). Lưu `workflow_id` vào plan. Mẫu hay dùng:
   - Task quá hạn → nhắc PM (`user` field) qua bot.
   - Record mới gán người → thông báo người được gán.
   - Trạng thái → "Hoàn thành" → cập nhật ngày hoàn thành / báo group.
3. **Bot notification.** Kênh thông báo: ưu tiên Base workflow gửi tin
   nội bộ; cần gửi vào group chat cụ thể thì cấu hình target qua `lark-im`
   (resolve chat_id của group dự án; gửi card/text). Người nhận resolve
   qua `team-roster`/lark-contact.
4. **Bật + test.** `+workflow-enable`. Test bằng cách tạo/sửa 1 record
   mẫu khớp trigger, xác nhận action chạy (tin nhắn tới, field cập nhật).
   KHÔNG coi `+workflow-create` trả `code:0` là "đã chạy".
5. Report: danh sách workflow (trigger→action, trạng thái enable) + kết
   quả test thật từng cái.

## Hard rules

1. `steps` JSON bắt buộc theo `lark-base-workflow-schema.md` — không đoán
   cấu trúc. list/get/enable chỉ cần `workflow_id` + trạng thái.
2. Mỗi automation phải TEST bằng trigger thật, không tin code:0.
3. Người/nhóm nhận thông báo resolve qua roster/contact, không bịa
   id/chat_id.
4. Automation gửi tin tới người thật = hành động hướng ngoại — xác nhận
   nội dung + người nhận TRƯỚC khi enable; tránh spam (gộp nhắc, đặt
   ngưỡng tần suất).
5. Tránh vòng lặp: workflow sửa field lại kích chính trigger của nó →
   thiết kế điều kiện thoát.
6. Mặc định `--as user`; bot gửi thông báo cần đúng scope IM → lỗi scope
   route `lark-shared`.

## Edge cases

- Trigger "tới hạn/định kỳ" cần field date chuẩn → xác nhận WIRE-UP đã có
  field "Hạn"; thiếu thì quay lại bổ sung.
- Gửi vào group chưa tồn tại → tạo/﻿xác nhận group qua lark-im trước, hoặc
  gửi DM người phụ trách.
- Nhiều automation cùng trigger → gộp 1 workflow nhiều nhánh nếu được,
  giảm trùng thông báo.
- Test trên Base rỗng (IMPORT skip) → tạo record mẫu test rồi xoá.

## Memory

Đọc `team-roster.md` (người nhận/PM), `base-conventions.md`. Không ghi
state (ở `build-plan.json`).
