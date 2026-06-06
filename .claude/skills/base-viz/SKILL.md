---
name: base-viz
description: Phase 5 của base-deploy — dựng dashboard KPI/chart/filter, fan-out 3 sub-agent soạn data_config song song rồi tạo block tuần tự. Triggers "phase viz", "dashboard base", "biểu đồ KPI base", "base-viz".
version: 1.0.0
last_updated: 2026-06-06
---

# base-viz (Phase 5 · 3 sub-agent · PARALLEL ⚡)

"3 sub-agent xây dashboard KPI, chart, filter đồng thời." Hiện thực hoá
`dashboard` trong `build-plan.json`: tạo dashboard + các block (KPI số,
chart, filter). Soạn `data_config` song song nhưng **tạo block tuần tự**
(lark-base rule: dashboard block phải create nối tiếp). Verb +
`data_config` route về `lark-base` (`lark-base-dashboard.md`,
`dashboard-block-data-config.md`).

## Workflow

1. **Tạo dashboard.** `base +dashboard-create` trong `base_token`. Lưu
   `dashboard_id` vào plan. (Hoặc `+base-block-list` xác nhận dashboard
   có sẵn.)
2. **Fan-out 3 sub-agent SOẠN data_config song song** (một message,
   nhiều Task — đây là điểm parallel ⚡). Mỗi nhóm block đọc
   `dashboard-block-data-config.md` + cấu trúc table thật:
   - **KPI**: block số đơn (count/sum/% theo `metric`+`filter`), vd
     "task quá hạn", "tiến độ %".
   - **Chart**: bar/line/pie theo `group_by`+`metric`, vd "task theo dự
     án", "tải mỗi PM".
   - **Filter**: filter control theo `field` (PM, trạng thái, kỳ).
   Mỗi agent TRẢ VỀ JSON `data_config` đã build (không tự tạo block).
3. **Barrier + tạo block TUẦN TỰ.** Gom 3 nhóm `data_config`,
   `+dashboard-block-create` lần lượt (không song song ở tầng CLI).
   `+dashboard-arrange` chỉ chạy nếu user muốn auto layout/đẹp.
4. **Verify số liệu**: `+dashboard-block-get-data` đọc kết quả tính của
   từng block, đối chiếu `+data-query` cùng điều kiện → số phải khớp.
   Lệch = `data_config` sai, sửa lại (đừng tin `code:0` là đúng).
5. Report: dashboard link + danh sách block + số mẫu mỗi KPI đã verify.

## Hard rules

1. Soạn `data_config` song song được; TẠO block phải tuần tự (lark-base).
2. `data_config` mọi field/source phải khớp table+field THẬT — GET cấu
   trúc trước, không đoán tên field.
3. KPI/chart phải truy được tới `meta.kpis`; thừa block không phục vụ
   KPI nào → bỏ.
4. Verify bằng `+dashboard-block-get-data` vs `+data-query`; `code:0`
   KHÔNG phải success (CLAUDE.md). Số sai → sửa trước khi report.
5. `+dashboard-arrange` chỉ khi user yêu cầu — nó ghi đè layout.
6. `+dashboard-block-get-data` không trả metadata block; cần tên/loại
   thì `+dashboard-block-get`.

## Edge cases

- KPI cần field chưa tồn tại → quay lại DESIGN/WIRE-UP bổ sung
  formula/lookup, đừng fake trong data_config.
- Dashboard chưa có dữ liệu (IMPORT skip) → block vẫn dựng, báo "0 record,
  số sẽ lên sau khi nạp dữ liệu".
- Chart group theo link field → group theo field người-đọc, không theo
  `record_id`.
- Block tạo lỗi giữa chừng → giữ block đã tạo, retry block lỗi, không
  tạo trùng.

## Memory

Đọc `base-conventions.md`. Không ghi state (ở `build-plan.json`).
