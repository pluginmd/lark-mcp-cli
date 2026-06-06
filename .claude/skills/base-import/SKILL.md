---
name: base-import
description: Phase 4 của base-deploy — import dữ liệu Excel/CSV vào Base, làm sạch & map field chuẩn hoá theo build-plan.json. Triggers "phase import", "import excel vào base", "nạp dữ liệu base", "base-import".
version: 1.0.0
last_updated: 2026-06-06
---

# base-import (Phase 4 · 1 agent · sequential)

"Import dữ liệu từ Excel, làm sạch & mapping chuẩn hoá." Nạp các nguồn
trong `imports[]` vào table đích đã BUILD/WIRE-UP. Hai đường: import
nhanh tạo bảng mới (lark-drive) hoặc ghi record vào bảng đã thiết kế
(lark-base — khuyến nghị, giữ schema chuẩn). Verb route về `lark-base`
(record) và `lark-drive` (file import).

## Workflow

1. **Kiểm nguồn.** Mỗi `imports[i]`: file tồn tại? đọc header. Nếu
   `imports: []` → skip phase, báo "Base rỗng, nhập tay sau".
2. **Chọn đường:**
   - Bảng đích đã có schema chuẩn (mặc định) → đọc file local (csv/xlsx),
     map cột→field theo `mapping`, ghi bằng `base +record-batch-create`
     (≤200/lô, tuần tự, gặp `1254291` đợi ngắn retry).
   - User chỉ cần bê nguyên file thành bảng mới, chưa thiết kế →
     `drive +import --type bitable` rồi quay lại chuẩn hoá. KHÔNG dùng
     đường này nếu bảng đích đã thiết kế kỹ (sẽ lệch schema).
3. **Làm sạch trước khi ghi** (theo `mapping` + field type plan):
   - Trim khoảng trắng; chuẩn ngày → `YYYY-MM-DD HH:mm:ss`; số bỏ ký
     tự tiền tệ; bool→checkbox.
   - select: giá trị phải khớp `options[]`; lệch → map về option gần
     nhất HOẶC gom "cần xử lý tay", KHÔNG để CLI tự sinh option rác.
   - user field: tên người → `open_id` qua lark-contact/`team-roster`;
     không resolve được → để trống + log.
   - link field: match khoá tự nhiên (vd tên dự án) → `record_id` bảng
     đích; phải nạp bảng cha TRƯỚC bảng con.
   - Bỏ qua field read-only (`formula`/`lookup`/system) — không map vào.
4. **Dedupe** theo `imports[].dedupe_key`: trùng → `+record-upsert`
   thay vì tạo mới.
5. **Ghi theo thứ tự phụ thuộc**: bảng cha (được link tới) trước, bảng
   con (chứa link) sau, để resolve được `record_id`.
6. **Verify nạp**: `+data-query` đếm record mỗi bảng so kỳ vọng; sample
   `+record-list` vài dòng kiểm cell đúng kiểu. Report: nạp X/Y dòng,
   Z dòng "cần xử lý tay" + lý do.

## Hard rules

1. Map vào đúng storage field; không ghi formula/lookup/system.
2. select option phải có sẵn — không để ghi sinh option mới ngoài ý muốn
   (`+field-search-options` xác nhận trước nếu nghi ngờ).
3. Batch ≤200; lỗi `1254104` → chia lô; `1254015` (sai kiểu) → quay lại
   làm sạch theo `lark-base-cell-value.md`, đừng ép ghi.
4. Bảng cha nạp trước bảng con (link cần record_id đích).
5. Dòng bẩn không đoán bừa — gom "cần xử lý tay" + lý do, báo user.
6. Import là ghi hàng loạt khó đảo — xác nhận số dòng + bảng đích TRƯỚC
   khi chạy; với file lớn đề xuất nạp thử 5 dòng kiểm mapping rồi mới full.

## Edge cases

- Header Excel lệch `mapping` → dừng, in diff header vs mapping, hỏi
  user, không tự đoán cột.
- File >vài nghìn dòng → chia lô + báo tiến độ; cân nhắc
  `drive +import` rồi reshape.
- Encoding/ô gộp/nhiều sheet → nêu rõ giới hạn, xử lý sheet chỉ định.
- Người trùng tên khi resolve open_id → để trống + flag, không gán nhầm.

## Memory

Đọc `team-roster.md` (resolve user), `base-conventions.md`. Không ghi
state (ở `build-plan.json`); có thể lưu `mapping` ổn định vào plan để
lần import sau tái dùng.
