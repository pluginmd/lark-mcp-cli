# 09 — Cập nhật & bảo trì

> Đối tượng: admin / người triển khai.

## Cập nhật binary

```bash
lark-cli update              # tự cập nhật lên bản mới (nếu cài qua kênh phát hành)
```

Nếu dựng từ mã nguồn: kéo code mới rồi dựng & cài lại:

```bash
git pull
./scripts/setup-mcp.sh       # build + cài lại ~/bin/lark-cli
# rồi Quit & mở lại Claude Desktop để nạp binary mới
```

> Lệnh nhanh `/mcp-rebuild` (trong Cowork) làm bước dựng-lại + nhắc cài đặt.

## Đồng bộ skills

Binary và bộ skill phải **cùng phiên bản**. Khi thấy `_notice.skills` trong output JSON nghĩa là skills lệch:

```bash
lark-cli update              # đồng bộ lại
```

## Sau mỗi lần đổi binary — BẮT BUỘC

1. **Quit hẳn** Claude Desktop (không chỉ đóng cửa sổ).
2. Mở lại → kiểm tra connector `lark-cli` *connected*, đủ 21 tool.
3. (macOS) Nếu vừa thay binary mà bị treo → xem [08 §"Server treo"](08-xu-ly-su-co.md) (gỡ quarantine + ký ad-hoc).

> Lý do: tiến trình `mcp serve` đang chạy giữ binary cũ trong bộ nhớ; phải khởi động lại mới nạp bản mới.

## Sao lưu & hoàn tác

Trước khi thay binary, nên giữ bản cũ để hoàn tác:

```bash
cp ~/bin/lark-cli ~/bin/lark-cli.bak-$(date +%Y%m%d)
# hoàn tác:
cp ~/bin/lark-cli.bak-YYYYMMDD ~/bin/lark-cli
```

## Cổng web — giữ sống lâu dài

Quick tunnel tắt khi đóng máy & đổi URL. Cho vận hành ổn định:

- **Cloudflare Named Tunnel** → URL cố định.
- Dịch vụ nền: **launchd** (macOS) / **systemd** (Linux) để `mcp serve` + tunnel tự chạy lại sau reboot.
- Theo dõi `--audit-log`; cân nhắc xoay log định kỳ.

## Kiểm tra chất lượng trước khi phát hành (cho người build)

```bash
make unit-test               # test có -race (bắt buộc trước PR)
go vet ./...
gofmt -l .                   # phải rỗng
go mod tidy                  # không được đổi go.mod/go.sum
```

## Gỡ cài đặt

```bash
rm ~/bin/lark-cli            # xoá binary
lark-cli auth logout         # (chạy trước khi xoá) thu hồi token
# xoá khối "lark-cli" trong claude_desktop_config.json
```
