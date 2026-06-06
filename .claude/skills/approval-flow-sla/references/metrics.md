# Approval flow — metric & SLA definitions

Tất cả timestamp từ `instances.get` là **epoch milliseconds** (string).
Parse → ms → trừ → đổi sang giờ: `(t2 - t1) / 3_600_000`.

## Nguồn dữ liệu (per instance)

`data.start_time` / `data.end_time` — đơn tạo / kết thúc.
`data.status` — `PENDING` | `APPROVED` | `REJECTED` | `CANCELED` | `DELETED`.
`data.task_list[]` — mỗi node duyệt:
- `node_name` / `node_id` — tên bước.
- `user_id` — approver tại node đó.
- `status` — `PENDING` | `APPROVED` | `REJECTED` | `TRANSFERRED` | `DONE`.
- `start_time` — lúc đơn TỚI node (vào hàng chờ của approver).
- `end_time` — lúc approver xử lý xong (rỗng nếu còn pending).
`data.timeline[]` — log thao tác (dùng cho audit trail, comment, transfer).

## Định nghĩa thời gian

| Metric | Công thức | Ý nghĩa |
|---|---|---|
| **Node dwell** | `task.end_time − task.start_time` | Đơn nằm chờ tại 1 approver. **Đây là nghẽn vận hành.** |
| **Backlog age** | `now − task.start_time` (task PENDING) | Đơn đang kẹt bao lâu. |
| **Cycle time** | `instance.end_time − instance.start_time` | Tổng từ tạo → xong (chỉ đơn đã đóng). |
| **Wait time** | Σ node dwell (các node tuần tự) | Thời gian ngồi chờ người. |
| **Touch time** | (ước lượng) thời gian thao tác thực ≈ rất nhỏ so wait | Khối lượng xử lý. |
| **WIP / queue** | đếm task PENDING của 1 approver | Tải đang gánh. |

Quy tắc: **nghẽn = wait, không phải touch.** Một approver có WIP cao
nhưng dwell thấp = bận-mà-thông; dwell cao = thật sự ngâm.

## Thống kê (mỗi node & mỗi approver)

Gom mảng dwell → tính:
- **p50** (median), **p90**, **max** — p90 phản ánh đuôi xấu, dùng để
  định SLA; max để spot đơn cá biệt.
- **breach %** = (số dwell > target) / tổng.
- **throughput** = số đơn đóng / cửa sổ.

Chỉ đưa đơn đã `APPROVED`/`REJECTED` vào phân phối cycle/dwell; đơn còn
mở → rổ backlog age riêng (tránh méo p90).

## Định vị nghẽn (ranking)

Sắp node/approver theo, ưu tiên giảm dần:
1. **breach %** vs SLA target (vỡ cam kết nhiều nhất).
2. **p90 dwell** (đuôi chậm).
3. **WIP age tổng** = Σ backlog age các đơn PENDING (lượng việc đang ứ).

Node đứng đầu cả 3 = điểm nghẽn chính. Nếu 1 approver chiếm > 50% dwell
của 1 node → nghẽn *do người*, không do thiết kế luồng.

## SLA — khung dựng & chấm

**Tier mặc định (khi `sla-targets.md` trống):** node 24h, end-to-end 72h.

**Dựng baseline từ dữ liệu:** nếu chưa có target, đề xuất
`target = p75(dwell hiện tại)` làm mốc khởi điểm — đạt được nhưng có
sức ép, rồi siết dần. LUÔN cite "baseline = p75 đo từ N đơn".

**Scorecard mỗi process:**
```
process | cycle p50 | cycle p90 | breach% (node) | breach% (e2e) | throughput | backlog
```

**Aging buckets cho backlog:** `<SLA` · `1–2× SLA` · `>2× SLA` (đỏ).

**Phân loại đạt/vỡ:**
- 🟢 p90 ≤ target.
- 🟡 target < p90 ≤ 1.5× target.
- 🔴 p90 > 1.5× target → khuyến nghị can thiệp (thêm approver / song song
  hóa node / tự-duyệt dưới ngưỡng).

## Audit đơn đã duyệt (cross-check `policies.md`)

- **Rubber-stamp**: dwell < ngưỡng tối thiểu hợp lý (vd đơn tiền duyệt
  trong <60s) → nghi duyệt không đọc.
- **Self-approval**: `task.user_id` == người phát đơn.
- **Over-threshold pass**: số tiền > hard limit §N nhưng `APPROVED`.
- **Skipped node**: chuỗi `node_name` thực tế thiếu bước so với
  `process-catalog.md`.

Mỗi flag → nêu nghi vấn + cách kiểm chứng (mở timeline, hỏi approver),
KHÔNG khẳng định sai phạm.
