---
name: base-deploy
description: Triển khai Lark Base end-to-end cho team — orchestrator 8 phase (Discovery→Handover), spawn sub-agent song song ở BUILD và VIZ. Triggers "triển khai base", "dựng base cho team", "build lark base end to end", "deploy base 30 người", "/base-deploy".
version: 1.0.0
last_updated: 2026-06-06
---

# base-deploy (orchestrator · USE CASE 4)

Biến 1 câu yêu cầu ("dựng Base cho team 30 người: 3 table + dashboard +
automation + import Excel") thành một Base chạy được, đã verify, đã bàn
giao. Điều phối 8 phase, fan-out sub-agent song song ở BUILD (6) và VIZ
(3). **Không gọi `lark-cli` trực tiếp** — mỗi phase-skill sở hữu verb
mapping của nó; mọi chi tiết Base verb route về skill `lark-base`.

Artifact xuyên suốt: **`build-plan.json`** (SSOT). Schema + invariant:
[references/build-plan-schema.md](references/build-plan-schema.md).

## Bản đồ phase

| # | Phase | Skill | Agent | Đầu vào → đầu ra |
|---|---|---|---|---|
| 0 | DISCOVERY | `base-discovery` | 1 · seq | yêu cầu thô → `meta`, KPI, nguồn dữ liệu |
| 1 | DESIGN | `base-design` | 1 · seq | meta → `tables/fields/views/relations/roles` |
| 2 | BUILD | `base-build` | **6 · ∥** | plan → `base_token` + `table_id`/field/view thật |
| 3 | WIRE-UP | `base-wireup` | 1 · seq | resolve `link_table`→id, set view filter, role |
| 4 | IMPORT | `base-import` | 1 · seq | Excel → record (clean + mapping) |
| 5 | VIZ | `base-viz` | **3 · ∥** | plan → dashboard KPI/chart/filter |
| 6 | AUTOMATION | `base-automation` | 1 · seq | workflow + bot nhắc việc |
| 7 | HANDOVER | `base-handover` | 1 · seq | **verify UI** + tài liệu + bàn giao |

## Workflow điều phối

1. **Khởi tạo.** Đọc memory (`base-conventions.md`, `team-roster.md`,
   `templates.md`). Xác định path `build-plan.json` (arg hoặc
   `./build-plan.json`). Nếu file đã tồn tại → hỏi resume từ phase nào
   thay vì làm lại từ đầu.
2. **Chạy tuần tự phase 0→7**, mỗi phase là một Task (sub-agent) nhận
   path `build-plan.json` + đọc/ghi nó. Phase sau chỉ chạy khi phase
   trước đã ghi đủ field SSOT (xem invariant).
3. **Phase 2 BUILD — fan-out 6 sub-agent trong MỘT message** (nhiều
   Task call song song): mỗi table 1 agent dựng table+field+view của
   table đó; agent thứ 6 lo `base-create` + folder/`base-block`. Có
   barrier: chờ cả 6 xong, gom `table_id` vào plan, rồi mới sang WIRE-UP.
   (Số sub-agent = số table, hình minh hoạ là 3 table → mở rộng tới 6.)
4. **Phase 5 VIZ — fan-out 3 sub-agent song song**: KPI block, chart
   block, filter block. Dashboard block phải tạo **tuần tự ở tầng CLI**
   (lark-base rule) → 3 agent soạn `data_config`, một bước gom tạo nối
   tiếp.
5. **Gate sau mỗi phase**: nếu phase trả lỗi/thiếu invariant → DỪNG, báo
   user, không chạy phase sau trên dữ liệu hỏng.
6. **HANDOVER là cổng nghiệm thu** — không tuyên bố "xong" cho tới khi
   `base-handover` verify được UI (mở Base/`+diff-canonical`). `code:0`
   từ save KHÔNG phải success.

## Chế độ chạy

- `full` (default): chạy 0→7.
- `--from <phase>` / `--to <phase>`: chạy một đoạn (vd `--from import`
  trên Base có sẵn). Cần `build-plan.json` đã có `base_token`+`table_id`.
- `--dry-run`: chạy 0→1, in plan để user duyệt, KHÔNG ghi lên Base.
- Phase lẻ cũng gọi trực tiếp được (vd chỉ `base-viz` cho Base có sẵn).

## Hard rules

1. SSOT là `build-plan.json`. Mọi `*_id` chỉ điền bằng giá trị trả về
   thật — không bịa token/id.
2. Reference-first: trước khi để BUILD/VIZ generate, nếu user có Base
   mẫu → snapshot làm fixture, build khớp field-by-field (CLAUDE.md G-11).
3. Luôn `--dry-run` (in plan) cho user duyệt **trước** khi ghi Base, trừ
   khi user bảo cứ chạy.
4. Identity mặc định `--as user`; lỗi scope → route `lark-shared`, không
   tự hạ `--as bot`.
5. Phase BUILD/VIZ fan-out trong MỘT message; các phase seq chạy lần lượt.
6. Ghi Base là hành động khó đảo ngược — xác nhận trước khi tạo Base mới
   hoặc import hàng loạt; báo rõ cái gì sắp được tạo.
7. Không declare done khi HANDOVER chưa verify UI.

## Memory

`base-conventions.md` (quy ước đặt tên table/field/view của tổ chức),
`team-roster.md` (name→open_id, ai là PM/manager để gán user field +
phân quyền), `templates.md` (mẫu Base hay dùng: CRM, dự án, tuyển dụng…).
Trống → vẫn chạy, DISCOVERY sẽ hỏi bù.

## Test scenarios

- "Dựng Base dự án cho 30 người, 3 table + dashboard + automation +
  import Excel" → full pipeline, kết thúc bằng HANDOVER report đã verify.
- User đã có Base, chỉ muốn thêm dashboard → `base-viz` trực tiếp.
- Thiếu file Excel khi tới IMPORT → DỪNG ở phase 4, hỏi nguồn dữ liệu.
- `--dry-run` → in plan, không tạo gì.
