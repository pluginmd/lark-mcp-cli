---
name: base-build
description: Phase 2 của base-deploy — tạo Base + table + field + view từ build-plan.json, fan-out 1 sub-agent/table song song, điền ID thật lại vào plan. Triggers "phase build", "dựng table base", "base-build".
version: 1.0.0
last_updated: 2026-06-06
---

# base-build (Phase 2 · 6 sub-agent · PARALLEL ⚡)

"6 sub-agent xây dựng đồng thời: table, field, view, quan hệ, rule,
option." Đọc `build-plan.json` đã DESIGN, hiện thực hoá lên Lark: tạo
Base, rồi mỗi table một sub-agent dựng table+field+view của nó. Điền
`base_token`/`table_id`/`field_id`/`view_id` **thật** ngược lại plan.
Chi tiết verb route về skill `lark-base`.

## Workflow

1. **Tạo Base (tuần tự, trước fan-out).** `base +base-create` với
   `meta.base_title`. Lưu `base_token` vào plan. Nếu bot tạo → chú ý
   `permission_grant`, báo user có mở được Base không (lark-base rule).
   Tuỳ chọn: `+base-block-create` folder gom table.
2. **Fan-out: 1 sub-agent / table, trong MỘT message** (nhiều Task call
   song song — đây là điểm "parallel ⚡" của phase). Mỗi sub-agent nhận
   `base_token` + một `tables[i]`, làm tuần tự NỘI BỘ table đó:
   - `+table-create` → lưu `table_id`.
   - `+field-create --json` mỗi field **storage** (xem map type→JSON
     trong `lark-base`: `lark-base-field-json.md`). select kèm
     `options[]`. Bỏ qua field `link`/`formula`/`lookup` ở bước này —
     để WIRE-UP (cần `table_id` bảng khác / guide riêng). Lưu `field_id`.
   - `+view-create` các view; `filter` đơn giản set luôn, filter phức
     tạp đánh dấu để WIRE-UP (`+view-set-filter`).
   - Trả về mảnh plan đã điền id của table đó.
3. **Barrier + gom.** Chờ cả N sub-agent; merge id vào `build-plan.json`.
   Kiểm: mọi table có `table_id`, mọi storage field có `field_id`.
4. Report: Base link + bảng table/field đã tạo. Báo các field
   link/formula/lookup CÒN HOÃN sang WIRE-UP.

## Hard rules

1. Tạo Base TRƯỚC, fan-out SAU (sub-agent cần `base_token`). Trong một
   table thì các lệnh ghi chạy TUẦN TỰ (lark-base: nối nhau, gặp
   `1254291` đợi ngắn rồi retry).
2. Field `link`/`formula`/`lookup` KHÔNG tạo ở đây — chuyển WIRE-UP. Lý
   do: link cần `table_id` đích, formula/lookup cần đọc guide +
   `--i-have-read-guide`.
3. Chỉ tạo field có trong plan; không tự thêm field "cho đẹp".
4. Mọi id ghi lại plan phải từ output thật của CLI, không bịa.
5. Reference-first: nếu DISCOVERY có fixture Base mẫu → tạo field khớp
   field-by-field theo fixture, không sáng tác schema (CLAUDE.md G-11).
6. select option đúng tên trong plan; đừng để ghi record sau tự sinh
   option lạ.

## Edge cases

- Primary field đã tự sinh khi `+table-create` → cập nhật tên/loại thay
  vì tạo trùng.
- `1254104` (batch >200) không xảy ra ở tạo field; nếu tạo record mẫu thì
  chia lô.
- Table phụ thuộc lẫn nhau (link 2 chiều) → cả hai vẫn tạo độc lập ở
  BUILD, link nối ở WIRE-UP.
- Fan-out lỗi 1 table → giữ id các table thành công, báo table hỏng, cho
  retry riêng table đó (idempotent theo `key`).

## Memory

Đọc `base-conventions.md` để đặt tên đúng chuẩn tổ chức. Không ghi memory
(state nằm ở `build-plan.json`).
