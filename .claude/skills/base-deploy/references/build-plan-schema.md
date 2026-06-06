# build-plan.json — SSOT artifact của pipeline base-deploy

Một file JSON duy nhất chảy qua cả 8 phase. DISCOVERY/DESIGN **ghi**,
BUILD→AUTOMATION **đọc + cập nhật ID thật trả về**, HANDOVER **verify**.
Mặc định lưu `./build-plan.json` (hoặc path truyền qua arg). Mọi `*_id`
khởi tạo `null`, chỉ điền bằng giá trị **trả về thật** từ `lark-cli`.

```jsonc
{
  "meta": {
    "base_title": "Vận hành dự án — Team 30",
    "team_size": 30,
    "owner_open_id": "ou_xxx",          // resolve qua lark-contact
    "goal": "1 câu mô tả mục tiêu vận hành",
    "kpis": ["số task quá hạn", "tiến độ dự án %", "tải mỗi người"]
  },
  "base_token": null,                    // BUILD điền sau +base-create
  "tables": [
    {
      "key": "projects",                 // khóa logic, ổn định, dùng để nối phase
      "name": "Dự án",
      "table_id": null,                  // BUILD điền
      "primary_field": "Tên dự án",
      "fields": [
        { "name": "Tên dự án", "type": "text", "primary": true },
        { "name": "Trạng thái", "type": "single_select",
          "options": ["Mới", "Đang chạy", "Tạm dừng", "Hoàn thành"] },
        { "name": "PM", "type": "user" },
        { "name": "Bắt đầu", "type": "date" },
        { "name": "Hạn", "type": "date" },
        { "name": "Tasks", "type": "link", "link_table": "tasks" }   // điền ở WIRE-UP
      ],
      "views": [
        { "name": "Kanban trạng thái", "type": "kanban", "group_by": "Trạng thái" },
        { "name": "Quá hạn", "type": "grid",
          "filter": "Hạn < today AND Trạng thái != Hoàn thành" }
      ]
    }
  ],
  "relations": [
    { "from": "tasks.Dự án", "to": "projects", "type": "many_to_one" }
  ],
  "imports": [
    { "source": "projects.xlsx", "into": "projects",
      "mapping": { "Cột Excel": "Tên field Base" },
      "dedupe_key": "Tên dự án" }
  ],
  "dashboard": {
    "name": "KPI Vận hành",
    "dashboard_id": null,
    "blocks": [
      { "type": "kpi",   "title": "Task quá hạn",  "source": "tasks",
        "metric": "count", "filter": "Hạn < today AND !Hoàn thành" },
      { "type": "chart", "title": "Tiến độ theo dự án", "chart": "bar",
        "source": "tasks", "group_by": "Dự án", "metric": "count" },
      { "type": "filter", "title": "Lọc theo PM", "source": "tasks", "field": "PM" }
    ]
  },
  "automation": [
    { "key": "notify-overdue", "trigger": "task quá hạn",
      "action": "bot nhắc PM qua IM", "workflow_id": null }
  ],
  "roles": [
    { "name": "Thành viên",  "scope": "sửa record mình phụ trách" },
    { "name": "Quản lý",     "scope": "sửa toàn bộ + xem dashboard" }
  ]
}
```

## Quy ước type → CLI

`build-plan` chỉ mô tả **thiết kế logic**. BUILD dịch sang `+field-create
--json` theo SSOT field JSON trong skill `lark-base`
(`references/lark-base-field-json.md`, `formula-field-guide.md`,
`lookup-field-guide.md`). Đừng tự chế JSON field ở đây — luôn route qua
`lark-base`.

| build-plan type | lark-base field type | ghi chú |
|---|---|---|
| text | text | primary field là field đầu |
| single_select / multi_select | single_select / multi_select | `options[]` |
| user | user | giá trị `[{ "id": "ou_xxx" }]` |
| date / datetime | date | định dạng `YYYY-MM-DD HH:mm:ss` |
| number / currency / percent | number | |
| checkbox | checkbox | |
| link | one_way_link / two_way_link | `link_table` → resolve `table_id` ở WIRE-UP |
| lookup / rollup | lookup | đọc `lookup-field-guide.md` |
| formula | formula | đọc `formula-field-guide.md` |
| attachment | attachment | ghi/đọc qua `+record-*-attachment` |

## Invariant (HANDOVER kiểm)

1. Mọi `*_id`/`base_token` ≠ null sau phase tương ứng.
2. Mỗi `link` field có `link_table` trỏ tới một `tables[].key` tồn tại.
3. `imports[].mapping` mọi value khớp một `fields[].name` thật.
4. Mỗi `dashboard.blocks[].source` khớp `tables[].key`.
5. `automation[].workflow_id` ≠ null nếu phase AUTOMATION chạy.
