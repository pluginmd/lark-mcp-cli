---
name: base-wireup
description: Phase 3 của base-deploy — nối liên table (link/lookup/formula), set view filter, dựng role/quyền truy cập từ build-plan.json. Triggers "phase wire-up", "nối liên kết base", "phân quyền base", "base-wireup".
version: 1.0.0
last_updated: 2026-06-06
---

# base-wireup (Phase 3 · 1 agent · sequential)

"Kết nối liên table, thiết lập view, quy tắc & quyền truy cập." Hoàn tất
những gì BUILD cố tình hoãn: field quan hệ (`link`), field phái sinh
(`lookup`/`formula`), view filter phức tạp, và role/advanced permission.
Đọc/ghi `build-plan.json`; verb route về `lark-base`.

## Workflow

1. **Link field.** Mỗi field `link`: resolve `link_table` (key) →
   `table_id` thật từ plan, rồi `+field-create --json` kiểu
   one_way/two_way_link. Lưu `field_id`. Cập nhật `relations[]` đã nối.
2. **Lookup / rollup.** Sau khi link tồn tại, tạo field `lookup`/`rollup`
   tham chiếu qua link đó — BẮT BUỘC đọc `lookup-field-guide.md` (skill
   lark-base) rồi thêm `--i-have-read-guide`.
3. **Formula.** Field `formula` (vd "ngày trễ", "tổng/đếm"): đọc
   `formula-field-guide.md` rồi tạo với `--i-have-read-guide`.
4. **View filter.** Filter phức tạp BUILD hoãn → `+view-set-filter` theo
   `lark-base-view-set-filter.md`. sort/group/visible-fields: GET hiện
   trạng trước, giữ field không đổi, chỉ sửa phần cần.
5. **Role & quyền.** Dựng `roles[]`: `+role-create` (chỉ custom role) +
   `+advperm-*` cho scope (member sửa record mình qua điều kiện field
   user; manager sửa hết). Đọc `lark-base-role-guide.md` + `role-config.md`
   trước khi viết JSON quyền. Gán người theo `team-roster`.
6. Kiểm invariant: không còn field/quan hệ "pending"; mọi `link.link_table`
   đã resolve. Report các liên kết + role đã dựng.

## Hard rules

1. Link tạo TRƯỚC lookup/rollup (lookup cần link tồn tại).
2. formula/lookup bắt buộc đọc guide tương ứng + `--i-have-read-guide` —
   không đoán cú pháp.
3. `+role-create` chỉ tạo custom role; system role không sửa/xoá.
   `+role-update` là delta-merge — đọc hiện trạng trước.
4. Bật advanced permission có thể ảnh hưởng role hiện hữu — xác nhận
   trước, báo phạm vi ảnh hưởng (lark-base rule).
5. Mọi `record_id` trong link chỉ là khoá nối, không phải giá trị người
   đọc; đừng nhầm khi kiểm.
6. Tên người → `open_id` qua `team-roster`/lark-contact, không bịa.

## Edge cases

- Two-way link tạo tự động field ngược ở bảng kia → đừng tạo trùng; ghi
  nhận `field_id` field ngược vào plan.
- Lookup ngược cần field nguồn read-only → đánh dấu trong plan để IMPORT
  bỏ qua.
- Member-scope cần điều kiện "record.user == current_user" → mô hình hoá
  qua advanced permission theo `role-config.md`, không hardcode từng người.
- Đóng advanced permission làm mất custom role → cảnh báo, không tự đóng.

## Memory

Đọc `team-roster.md` (gán role), `base-conventions.md`. Không ghi state
(ở `build-plan.json`).
