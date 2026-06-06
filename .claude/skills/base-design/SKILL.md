---
name: base-design
description: Phase 1 của base-deploy — thiết kế schema (table, field, view, quan hệ, rule, option, role) thành build-plan.json hoàn chỉnh để BUILD thực thi. Triggers "phase design", "thiết kế schema base", "base-design".
version: 1.0.0
last_updated: 2026-06-06
---

# base-design (Phase 1 · 1 agent · sequential)

"Thiết kế cấu trúc 3 table, quan hệ, trường dữ liệu và luồng xử lý." Lấy
`meta`+entity từ DISCOVERY, sản xuất `build-plan.json` ĐẦY ĐỦ:
`fields[]` mỗi table, `views[]`, `relations[]`, `roles[]`, khung
`dashboard`/`automation`. Đây là SSOT mà toàn bộ phase sau thực thi —
chất lượng phase này quyết định cả pipeline. Vẫn KHÔNG ghi Base.

Schema + bảng map type→CLI:
[build-plan-schema.md](../base-deploy/references/build-plan-schema.md).
Chi tiết field JSON route về skill `lark-base`
(`lark-base-field-json.md`, `formula-field-guide.md`,
`lookup-field-guide.md`) — đừng tự chế JSON ở đây, chỉ mô tả logic.

## Workflow

1. **Field mỗi table.** Mỗi entity: chọn primary field (text định danh),
   liệt kê storage field + type. Trạng thái/phân loại → `single_select`
   kèm `options[]`. Người phụ trách → `user`. Mốc thời gian → `date`.
2. **Quan hệ.** Xác định `relations[]` giữa các table (many_to_one /
   many_to_many). Mỗi quan hệ thêm một `link` field (với `link_table` =
   key bảng đích) — BUILD tạo field, WIRE-UP resolve id.
3. **Field phái sinh.** Chỉ thêm `formula`/`lookup`/`rollup` khi cần
   HIỂN THỊ lâu dài trong bảng (vd "ngày trễ", "tổng task"); phân tích
   một lần để `data-query` ở VIZ, không nhồi field.
4. **View.** Mỗi table 1–3 view phục vụ công việc thật: Kanban theo
   trạng thái, Grid lọc "quá hạn", view cá nhân theo `user`. Mô tả
   `filter` bằng ngôn ngữ logic; JSON filter để WIRE-UP dựng.
5. **Role.** Map vai trò DISCOVERY → `roles[]` với scope (member sửa
   record mình, manager sửa hết). Để WIRE-UP dựng `+role`/`+advperm`.
6. **Khung dashboard + automation** (chi tiết do VIZ/AUTOMATION làm):
   liệt kê block KPI/chart/filter ứng mỗi `meta.kpis`, và automation cần
   (nhắc quá hạn, thông báo gán việc).
7. Ghi `build-plan.json` đầy đủ. In sơ đồ ERD gọn (table + quan hệ) cho
   user duyệt trước BUILD. Nếu orchestrator `--dry-run` → dừng ở đây.

## Hard rules

1. Mỗi `meta.kpis[i]` phải truy được tới ≥1 field/quan hệ dựng được —
   nếu không, thiếu field, bổ sung.
2. Primary field mỗi table là text định danh duy nhất, không phải select.
3. `link.link_table` phải trỏ tới một `tables[].key` có thật trong plan.
4. Chỉ field storage do người nhập mới để user ghi; formula/lookup là
   read-only — đánh dấu đúng để IMPORT không map nhầm vào đó.
5. Không over-engineer: bám 3 bảng lõi trước, field "nice-to-have" để
   giai đoạn 2. Cảnh báo nếu >15 field/bảng.
6. select option phải hữu hạn, đặt tên nhất quán theo `base-conventions`.

## Edge cases

- Quan hệ many-to-many → hai-chiều link; lưu ý lookup ngược.
- KPI cần dữ liệu chưa có field → quay lại bổ sung field, đừng để VIZ
  phát hiện trễ.
- Bảng tham chiếu tĩnh (danh mục) → vẫn là 1 table + link, không hardcode.

## Memory

Đọc `base-conventions.md` (đặt tên), `templates.md`. Đề xuất lưu schema
tốt làm template mới nếu user xác nhận tái dùng.
