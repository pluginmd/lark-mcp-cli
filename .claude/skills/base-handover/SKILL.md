---
name: base-handover
description: Phase 7 của base-deploy — nghiệm thu UI (verify thật, không tin code:0), viết tài liệu hướng dẫn + bàn giao Base cho team. Triggers "phase handover", "nghiệm thu base", "bàn giao base", "tài liệu base", "base-handover".
version: 1.0.0
last_updated: 2026-06-06
---

# base-handover (Phase 7 · 1 agent · sequential)

"Kiểm thử, bàn giao, training nhanh & tài liệu hướng dẫn sử dụng." Cổng
NGHIỆM THU của pipeline. Không tuyên bố "Base xong" cho tới khi đã verify
bằng mắt/`+diff-canonical` — `code:0` từ các bước save KHÔNG phải success
(CLAUDE.md hard rule). Tổng kết `build-plan.json` thành báo cáo + tài liệu.

## Workflow

1. **Verify invariant `build-plan.json`** (xem
   [build-plan-schema.md](../base-deploy/references/build-plan-schema.md)
   §Invariant): mọi `*_id`/`base_token` ≠ null; mọi `link.link_table`
   đã resolve; `imports.mapping` khớp field thật; `dashboard.blocks.source`
   khớp table; `automation.workflow_id` ≠ null.
2. **Verify thật từng tầng** (không tin code:0):
   - **Schema**: `+table-list` + `+field-list` mỗi table khớp plan
     (đếm field, đúng type, select đủ option).
   - **Quan hệ**: `+record-list` 1 record có link → record đích đọc được.
   - **Dữ liệu**: `+data-query` đếm record vs kỳ vọng IMPORT.
   - **Dashboard**: `+dashboard-block-get-data` mỗi block ra số hợp lý
     (≠ lỗi/0 ngoài ý muốn); nếu có fixture tham chiếu →
     `+diff-canonical`. Lý tưởng: mở Base trên UI xác nhận bằng mắt.
   - **Automation**: xác nhận workflow `enable` + đã test ở phase 6.
   - **Quyền**: `+role-*`/`+advperm-*` đúng scope member/manager.
3. **Tài liệu hướng dẫn** (lark-doc/lark-wiki): tạo doc "Hướng dẫn dùng
   Base <tên>" gồm: sơ đồ table+quan hệ, ý nghĩa mỗi field, cách dùng
   từng view, cách đọc dashboard, automation nào tự chạy, ai quyền gì,
   FAQ thường gặp. Đính link Base + dashboard.
4. **Bàn giao**: chia sẻ Base/doc cho team (qua drive permission /
   wiki), gửi tin bàn giao + link cho owner/group qua lark-im. Đề xuất
   buổi training nhanh (lịch nếu user muốn).
5. **Report nghiệm thu** (xem Output) + nêu việc còn mở.

## Output — báo cáo nghiệm thu

- ✅ **CHECKLIST**: schema / quan hệ / dữ liệu (X record) / dashboard (N
  block verified) / automation (M workflow tested) / quyền — mỗi dòng
  PASS/FAIL kèm bằng chứng (số, không "có vẻ ok").
- 🔗 **LINK**: Base, dashboard, doc hướng dẫn.
- 📦 **BÀN GIAO**: đã share cho ai, tin đã gửi.
- ⚠️ **CÒN MỞ**: dòng "cần xử lý tay" từ IMPORT, field hoãn, câu hỏi.

## Hard rules

1. KHÔNG tuyên bố done khi chưa verify thật. `code:0` ≠ success. Mỗi mục
   checklist phải có bằng chứng (số/đối chiếu/diff), không khẳng định suông.
2. Verify đối chiếu hai nguồn khi có thể (`get-data` vs `data-query`).
3. FAIL bất kỳ tầng nào → báo rõ, KHÔNG ghi PASS tổng; đề xuất quay lại
   phase tương ứng sửa.
4. Chia sẻ/bàn giao = hướng ngoại — xác nhận phạm vi share (nội bộ team,
   không public) trước khi mở quyền; tránh lộ PII (đối chiếu
   `permission-audit` nếu có dữ liệu nhạy cảm).
5. Tài liệu phải khớp Base THẬT (tên table/field/view từ get, không từ
   plan nếu lệch) — verify trước khi viết.

## Edge cases

- Verify phát hiện lệch plan↔Base → Base là sự thật; cập nhật plan/tài
  liệu hoặc quay lại sửa, đừng viết tài liệu theo plan sai.
- Base rỗng (IMPORT skip) → checklist dữ liệu = "0 record (chủ ý)", vẫn
  bàn giao được, ghi rõ nhập tay sau.
- Không mở được UI để verify mắt + không có fixture → nêu rõ giới hạn
  verify, không ngầm coi là đã verify hình ảnh.

## Memory

Đọc `team-roster.md` (bàn giao cho ai), `base-conventions.md`. Sau bàn
giao, gợi ý lưu schema thành `templates.md` để tái dùng; lưu link Base +
doc vào project memory nếu là hệ thống vận hành lâu dài.
