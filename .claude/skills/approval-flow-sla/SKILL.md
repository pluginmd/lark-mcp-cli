---
name: approval-flow-sla
description: Đo luồng vận hành approval — phát hiện nghẽn ở node/người duyệt nào, audit toàn bộ dữ liệu đã duyệt, dựng & chấm SLA. Triggers "nghẽn duyệt ở đâu", "ai đang ngâm duyệt", "SLA duyệt", "phân tích luồng approval", "approval bottleneck".
version: 1.0.0
last_updated: 2026-06-06
---

# approval-flow-sla

"Đơn duyệt chậm mà không biết kẹt chỗ nào" → "node X / người Y ngâm
p90 = 18h, vỡ SLA 6h, đây là 3 đơn cần thúc."
Process-owner view (cả luồng), KHÁC `approval-triage` (per-item decide).
Agent đo + chỉ nghẽn; user quyết hành động.

**Verbs**: raw `approval` → `instances initiated` (list instance_code
theo người + time range), `instances get --instance-code <c>` (timeline
+ `task_list[]` mỗi node có `start_time`/`end_time`/`status`/`node_name`
→ nguồn tính dwell), `tasks query` (pending theo người), `tasks remind`
(thúc approver), `tasks transfer` (gỡ nghẽn). MUST `lark-cli schema
approval.<res>.<method>` trước khi gọi raw — đừng đoán field. Atomic
shapes + token flags: `.claude/RECIPES.md`. Công thức metric + SLA:
`references/metrics.md`.

## Workflow

1. **Read memory.** `sla-targets.md` (target mỗi node + end-to-end mỗi
   process), `process-catalog.md` (definition_code → tên + chuỗi node kỳ
   vọng), `team-roster.md` (approver → manager, để escalate),
   `policies.md` (audit đơn đã duyệt). Thiếu `sla-targets.md` → dùng
   default (node 24h / end-to-end 72h) + cảnh báo chưa cấu hình.
2. **Chốt scope.** process (definition_code) + cửa sổ thời gian
   (default 30d). Liệt kê instance_code: `approval instances initiated
   --user-id <id> --start <t-30d> --end <today>` gom theo `team-roster`
   người phát đơn, hoặc nhận list dán sẵn. Giới hạn: enumerate chỉ
   theo người-phát; KHÔNG có list org-wide trong CLI → state rõ độ phủ.
3. **Pull timeline mỗi instance** (parallel ~5): `approval instances get
   --instance-code <c> --jq '.data|{status,start_time,end_time,task_list}'`.
4. **Tính metric** (xem `references/metrics.md`):
   - **Node dwell** = `end_time − start_time` mỗi task → thời gian đơn
     NẰM CHỜ ở approver đó (≠ thời gian xử lý thực).
   - **Cycle time** = instance `end_time − start_time` (đơn đã xong) /
     `now − start_time` (đơn còn mở, = backlog age).
   - **Per-approver**: p50 / p90 / max time-to-act, số đơn đang ngâm (WIP).
   - **Per-node**: p50/p90 dwell, breach % vs `sla-targets`.
5. **Định vị nghẽn.** Node/approver có p90 dwell cao nhất HOẶC WIP cao
   nhất HOẶC breach % cao nhất = nghẽn. Tách *wait time* (chờ người)
   khỏi *touch time* — nghẽn vận hành là wait, không phải khối lượng.
6. **Audit đơn đã duyệt** (cross-check `policies.md`): cờ
   rubber-stamping (duyệt < ngưỡng tối thiểu, vd <60s cho đơn tiền),
   self-approval, vượt threshold mà vẫn pass, node bị skip so với
   chuỗi kỳ vọng trong `process-catalog.md`.
7. **Chấm SLA + report.** SCORECARD (mỗi process: cycle p50/p90,
   breach %, throughput, backlog) · NGHẼN (node/người + số liệu +
   đơn cụ thể) · AUDIT FLAGS · SLA ĐỀ XUẤT (nếu chưa có target, propose
   từ p75 hiện tại làm baseline) · ACTION (thúc/chuyển đơn nào).
8. **On confirm** — `tasks remind` đơn quá hạn / đề xuất `tasks transfer`
   khi approver vắng (check `team-roster` cover). ≤10 thao tác / confirm.

## Hard rules

1. Read-only mặc định — KHÔNG remind/transfer/approve khi chưa confirm.
2. Mọi kết luận nghẽn phải kèm số (p90 dwell, WIP, breach %) + đơn cụ
   thể — không "có vẻ chậm" suông.
3. Phân biệt rõ wait time vs touch time; đừng đổ lỗi approver cho thời
   gian đơn kẹt ở node trước.
4. SLA đề xuất cite baseline (p75 đo được hoặc §N trong `sla-targets.md`),
   không bịa con số.
5. Audit flag chỉ nêu *nghi vấn* + cách kiểm chứng, không khẳng định gian
   lận.
6. n < 10 đơn trong cửa sổ → "mẫu quá nhỏ, số liệu chỉ tham khảo".
7. Không escalate vượt thẩm quyền user (check `team-roster.md`).

## Edge cases

- Đơn còn mở khi đo → tính backlog age, loại khỏi cycle-time p50/p90
  (chỉ dùng đơn đã `APPROVED`/`REJECTED`) để khỏi méo phân phối.
- Node song song (加签/会签) → dwell lấy max nhánh, ghi chú đã gộp.
- Approver đã `transfer` → quy dwell cho người NHẬN, ghi lại chuỗi.
- Thiếu `process-catalog.md` → bỏ qua audit "skip node", phần còn vẫn chạy.
- Lệch múi giờ: `start_time`/`end_time` là epoch ms → chuẩn hóa trước khi trừ.

## Memory

Required: `sla-targets.md`, `process-catalog.md`.
Recommended: `team-roster.md` (escalate/cover), `policies.md` (audit).
Trống → chạy ở chế độ "đo + đề xuất baseline", emit setup warning.

## Cadence

Weekly ops review (thứ Hai): scorecard + top nghẽn. Daily nếu backlog
> ngưỡng: chỉ phần aging + đơn vỡ SLA. Pairs với `morning-brief` exec.
