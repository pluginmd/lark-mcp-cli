# 30-Slide Deck — Lark CLI → MCP cho Claude Desktop & Web (Business User)

Bộ 30 prompt sinh ảnh slide, bám sát bộ tài liệu business-user [`docs/`](README.md) (01–09, **không** gồm changelog). Mỗi slide là 1 prompt tạo ảnh 16:9, nhiều text/label tiếng Việt để người xem hiểu rõ cách biến `lark-cli` thành **MCP cho Lark trên Claude Desktop (Cowork)** và **claude.ai web**.

**Style guide (áp dụng mọi slide):** Vercel × Linear × Apple keynote, isometric 3D 30°, octane-render; nền navy `#0B1220` + blueprint grid + particle dust + volumetric god-rays từ top-left; **warm amber-gold = trí tuệ/quyết định** (Claude brain, planning), **cool cyan = thực thi** (lark-cli, transport, Lark Cloud); glass morphism, rim lighting; footer `TRANSFORM GROUP` thin uppercase white, wide letter-spacing; tiếng Việt đúng dấu, không gibberish.

**Logo Lark reference (paste khi prompt yêu cầu):** `https://framerusercontent.com/images/yjUmyrg0qV5d2Glkjn9sDqzacY.svg`

---

## 📑 Mục lục 30 slide

| # | Slide | Doc nguồn |
|---|---|---|
| 1 | Cover hero | README/01 |
| 2 | Vấn đề: CLI trần không nối được Claude | 01 |
| 3 | Giải pháp 1 dòng: MCP bridge = tay chân của AI | 01 |
| 4 | MCP là gì | 01 |
| 5 | Bản đồ 2 con đường: Desktop vs Web | README |
| 6 | Kiến trúc Desktop (stdio) | 02 |
| 7 | Kiến trúc Web (HTTP + tunnel) | 03 |
| 8 | Bảng so sánh Desktop vs Web | README |
| 9–12 | Desktop 4 bước | 02 |
| 13–16 | Web 4 bước | 03 |
| 17 | Danh tính User vs Bot | 04 |
| 18 | Quyền/scope + app Lark riêng | 04 |
| 19 | Bộ skill Cowork là gì | 05 |
| 20 | 23 skill catalog | 05 |
| 21 | Skill: Đầu ngày + Mail/Chat | 05 |
| 22 | Skill: Họp/Task/Sales | 05 |
| 23 | 21 công cụ MCP theo domain | 06 |
| 24 | An toàn từng tool (dry-run) | 06 |
| 25 | Data flow — data đi đâu | 07 |
| 26 | Security stack 4 lớp | 07 |
| 27 | Mở cổng web an toàn — checklist | 07 |
| 28 | Xử lý sự cố (gồm server treo) | 08 |
| 29 | Thói quen hàng ngày + bảo trì | 05/09 |
| 30 | CTA close | README |

---

## SLIDE 1 — Cover Hero

```
Tạo ảnh 16:9 widescreen (1920×1080) cho slide thuyết trình.

Cinematic isometric 3D illustration của một workspace nơi Claude (Desktop + web
browser) điều khiển toàn bộ Lark/Feishu ecosystem qua một cây cầu phát sáng
"MCP bridge", góc nhìn 30°, ultra-detailed, octane-render quality, premium
developer poster.

Logo Lark trong scene phải dùng ĐÚNG logo từ reference (không vẽ lại, không
modify): https://framerusercontent.com/images/yjUmyrg0qV5d2Glkjn9sDqzacY.svg

Bố cục trái → phải:
- LEFT (1/5): business user (không phải dev) ngồi laptop, calm, speech bubble
  tiếng Việt "Sáng nay tôi có gì?" — neon cyan input glow
- UPPER-CENTER (1/3): hai màn hình AI nổi cạnh nhau — "Claude Desktop (Cowork)"
  và "claude.ai (web browser)" — cùng phát warm amber-gold glow
- CENTER: cây cầu kính phát sáng khắc chữ "MCP BRIDGE" nối 2 màn hình AI sang
  phải, JSON-RPC streams chảy qua cầu — warm→cool gradient trên cầu
- LOWER-CENTER (1/3): metallic chip "lark-cli mcp serve" với 2 cổng kết nối
  "stdio" và "http", cool cyan glow
- RIGHT (1/5): Lark Open Platform cloud với logo Lark chuẩn từ reference, các
  domain icon nhỏ (IM, Mail, Calendar, Docs, Base, Task, Drive, Sheets,
  Meeting, OKR, Contact) bao quanh

Background: deep navy #0B1220, blueprint grid, particle dust dense, volumetric
god-rays from top-left, warm/cool gradient seam ở giữa, glass morphism trên mọi panel.

Text hiển thị trong ảnh:
- Title lớn: "LARK × CLAUDE"
- Subtitle: "Biến lark-cli thành MCP cho Lark — Desktop & Web"
- Tagline: "21 CÔNG CỤ · 23 SKILL COWORK · KHÔNG CẦN CLAUDE CODE"
- Stats badges: "Desktop (stdio)" / "Web (HTTPS + tunnel)" / "Dry-run an toàn"
- Tiếng Việt đúng dấu

Footer: "TRANSFORM GROUP" thin uppercase white sans-serif, wide letter-spacing.

Aspect ratio: 16:9 widescreen, dùng cho slide deck.
```

---

## SLIDE 2 — Vấn đề: CLI trần không nối được Claude

```
Tạo ảnh 16:9 widescreen (1920×1080) cho slide thuyết trình.

Cinematic isometric 3D illustration của một khoảng GAP/vực ngăn cách giữa Claude
AI và công cụ lark-cli — AI muốn dùng nhưng không "nói chuyện" được với CLI, góc
nhìn 30°, ultra-detailed, octane-render quality.

Bố cục trái → phải:
- LEFT (1/5): business user bối rối, speech bubble "Tôi đâu biết gõ lệnh CLI..."
  — expression mệt
- UPPER-CENTER (1/3): floating Claude brain với speech bubble "Tôi muốn thao tác
  Lark nhưng không gọi được CLI" — warm amber-gold nhưng dimmed
- CENTER: một vực sâu/crack lớn label "KHÔNG CÓ NGÔN NGỮ CHUNG", 2 bên bờ không
  có cầu nối
- LOWER-CENTER (1/3): metallic chip "lark-cli" (terminal-only) đứng đơn độc bên
  kia vực, chỉ có dấu nhắc dòng lệnh `$ _` — cool cyan nhưng isolated
- RIGHT (1/5): Lark Cloud xa, label "200+ lệnh không tự gọi được"

Background: deep navy #0B1220, blueprint grid, particle dust, god-rays,
warm/cool gradient seam.

Text hiển thị:
- Header lớn: "VẤN ĐỀ"
- 3 pain point boxes:
  "❌ lark-cli mạnh nhưng chỉ DEV gõ được"
  "❌ Claude Desktop & web KHÔNG tự gọi CLI"
  "❌ Business user bị chặn ngoài cửa"
- Tiếng Việt đúng dấu, không gibberish

Footer: "TRANSFORM GROUP" thin uppercase white sans-serif, wide letter-spacing.

Aspect ratio: 16:9 widescreen, dùng cho slide deck.
```

---

## SLIDE 3 — Giải pháp 1 dòng: MCP bridge

```
Tạo ảnh 16:9 widescreen (1920×1080) cho slide thuyết trình.

Cinematic minimal poster: một cây cầu kính phát sáng "MCP BRIDGE" bắc qua vực ở
slide trước, nối Claude sang lark-cli, lit by warm amber-gold spotlight từ trên.

Logo Lark khi xuất hiện phải dùng đúng logo từ reference:
https://framerusercontent.com/images/yjUmyrg0qV5d2Glkjn9sDqzacY.svg

Bố cục:
- LEFT (1/5): silhouette Claude brain với "bàn tay" mới mọc ra vươn qua cầu
- UPPER-CENTER (1/3): large bold Vietnamese typography "MCP bridge — TAY CHÂN
  của Claude trong Lark" — warm amber-gold glow
- CENTER: cây cầu kính khắc "lark-cli mcp serve", JSON-RPC chảy qua
- LOWER-CENTER (1/3): 3 glass chip nổi: "21 CÔNG CỤ" / "2 TRANSPORT (stdio·http)"
  / "DÙNG TÀI KHOẢN CỦA BẠN" — cool cyan
- RIGHT (1/5): Lark Open Platform cloud với logo Lark chuẩn, nay đã được nối

Background: deep navy #0B1220 gradient, minimal blueprint grid, god-rays from
top-left.

Style: Apple keynote minimal, premium typography poster, glass morphism.

Text hiển thị:
- Câu chính lớn: "MCP bridge — TAY CHÂN của Claude trong Lark"
- 3 chip labels: "21 công cụ · 2 transport · auth của bạn"
- Tiếng Việt đúng dấu

Footer: "TRANSFORM GROUP" thin uppercase white sans-serif, wide letter-spacing.

Aspect ratio: 16:9 widescreen, dùng cho slide deck.
```

---

## SLIDE 4 — MCP là gì

```
Tạo ảnh 16:9 widescreen (1920×1080) cho slide thuyết trình.

Cinematic isometric 3D illustration giải thích Model Context Protocol (MCP) như
một "ổ cắm tiêu chuẩn" để AI cắm vào công cụ ngoài, góc nhìn 30°, ultra-detailed,
octane-render quality.

Bố cục trái → phải:
- LEFT (1/5): label "MCP = Model Context Protocol" + icon ổ cắm chuẩn 🔌
- UPPER-CENTER (1/3): AI host (Claude) với 2 hành động cốt lõi hiển thị dạng
  glass card:
  "① tools/list — Claude HỎI: có công cụ gì?"
  "② tools/call — Claude GỌI: chạy công cụ này với tham số X"
  — warm amber-gold
- LOWER-CENTER (1/3): MCP server "lark-cli mcp serve" trả lời, mỗi tool call →
  spawn subprocess `lark-cli <lệnh>` → tái dùng auth — cool cyan
- RIGHT (1/5): callout "Chuẩn MỞ — Claude Desktop & claude.ai đều hỗ trợ" +
  icon JSON-RPC 2.0

Background: deep navy #0B1220, blueprint grid, particle dust, god-rays.

Style: Vercel × Linear × Apple keynote, glass morphism, rim lighting.

Text hiển thị:
- Header: "MCP LÀ GÌ?"
- 2 cơ chế: "tools/list · tools/call"
- Subtitle: "Giao thức chuẩn để AI dùng công cụ ngoài, qua JSON-RPC"
- Tiếng Việt đúng dấu

Footer: "TRANSFORM GROUP" thin uppercase white sans-serif, wide letter-spacing.

Aspect ratio: 16:9 widescreen, dùng cho slide deck.
```

---

## SLIDE 5 — Bản đồ 2 con đường: Desktop vs Web

```
Tạo ảnh 16:9 widescreen (1920×1080) cho slide thuyết trình.

Cinematic isometric 3D illustration của 2 con đường song song nối Claude tới
lark-cli: đường Desktop (stdio, local) và đường Web (HTTP qua tunnel), góc nhìn
30°, ultra-detailed, octane-render quality.

Logo Lark trên Lark Cloud đích phải dùng đúng logo từ reference:
https://framerusercontent.com/images/yjUmyrg0qV5d2Glkjn9sDqzacY.svg

Bố cục — 2 lane ngang song song, cùng đổ về Lark Cloud bên phải:
- LANE TRÊN (warm amber-gold) "CON ĐƯỜNG 1 · DESKTOP":
  "Claude Desktop" → mũi tên "stdio (local)" → "lark-cli mcp serve" trên máy →
  Lark Cloud. Badge "⭐ Dễ · cài 1 lần · không cần internet công khai"
- LANE DƯỚI (cool cyan) "CON ĐƯỜNG 2 · WEB":
  "claude.ai (browser)" → "HTTPS" → "Cloudflare Tunnel" → "lark-cli mcp serve
  --transport http" trên máy → Lark Cloud. Badge "⭐⭐⭐ Khó hơn · cần cổng HTTPS
  công khai + bảo mật"
- CENTER node chung: chip "lark-cli" với 2 cổng "stdio" và "http"
- RIGHT (1/5): Lark Open Platform cloud với logo Lark chuẩn

Background: deep navy #0B1220, blueprint grid, particle dust, god-rays,
warm/cool gradient ngang giữa 2 lane.

Style: Vercel × Linear × Apple keynote, glass morphism, rim lighting.

Text hiển thị:
- Header: "2 CON ĐƯỜNG NỐI CLAUDE TỚI LARK"
- 2 lane labels + badges readable
- Tiếng Việt đúng dấu

Footer: "TRANSFORM GROUP" thin uppercase white sans-serif, wide letter-spacing.

Aspect ratio: 16:9 widescreen, dùng cho slide deck.
```

---

## SLIDE 6 — Kiến trúc Desktop (stdio)

```
Tạo ảnh 16:9 widescreen (1920×1080) cho slide thuyết trình.

Cinematic isometric 3D illustration zoom vào CON ĐƯỜNG DESKTOP — Claude Desktop
spawn lark-cli như tiến trình con, nói chuyện qua stdio JSON-RPC, tất cả chạy
local, góc nhìn 30°, ultra-detailed, octane-render quality.

Logo Lark trên Lark Cloud phải dùng đúng logo từ reference:
https://framerusercontent.com/images/yjUmyrg0qV5d2Glkjn9sDqzacY.svg

Bố cục trái → phải (trong một "máy tính của bạn" boundary box lớn):
- LEFT (1/5): icon 💻 "MÁY CỦA BẠN" + label "CON ĐƯỜNG DESKTOP"
- UPPER-CENTER (1/3): app window "Claude Desktop (Cowork)" đọc file cấu hình
  glass card hiển thị:
  `command: /Users/ban/bin/lark-cli`
  `args: ["mcp","serve"]`
  `env: NO_COLOR=1`
  — warm amber-gold
- CENTER: ống nối "stdio · newline-delimited JSON-RPC" giữa app và binary
- LOWER-CENTER (1/3): chip "lark-cli mcp serve" → spawn subprocess
  `lark-cli <verb>` → đọc token từ OS Keychain — cool cyan
- RIGHT (1/5): mũi tên HTTPS rời "máy của bạn" đi tới Lark Cloud với logo Lark
  chuẩn

Background: deep navy #0B1220, blueprint grid, particle dust, god-rays.

Style: Vercel × Linear × Apple keynote, glass morphism, rim lighting.

Text hiển thị:
- Header: "KIẾN TRÚC DESKTOP · STDIO"
- Config snippet readable
- Labels: "stdio · subprocess · Keychain"
- Callout: "Tất cả LOCAL — không cần internet công khai"
- Tiếng Việt đúng dấu

Footer: "TRANSFORM GROUP" thin uppercase white sans-serif, wide letter-spacing.

Aspect ratio: 16:9 widescreen, dùng cho slide deck.
```

---

## SLIDE 7 — Kiến trúc Web (HTTP + tunnel)

```
Tạo ảnh 16:9 widescreen (1920×1080) cho slide thuyết trình.

Cinematic isometric 3D illustration zoom vào CON ĐƯỜNG WEB — claude.ai trên cloud
Anthropic không với tới máy bạn, phải đi qua Cloudflare Tunnel tới lark-cli HTTP
server có bearer-token gate, góc nhìn 30°, ultra-detailed, octane-render quality.

Logo Lark trên Lark Cloud phải dùng đúng logo từ reference:
https://framerusercontent.com/images/yjUmyrg0qV5d2Glkjn9sDqzacY.svg

Bố cục trái → phải:
- LEFT (1/5): "claude.ai" trong browser, trên server Anthropic xa — warm
  amber-gold
- CENTER-LEFT: HTTPS arrow tới một "Cloudflare Tunnel" node phát sáng (URL
  https://...trycloudflare.com/mcp)
- CENTER: một cổng bảo vệ "🔐 BEARER TOKEN GATE" — request thiếu token bị chặn
  (đỏ 401), request đúng token đi qua (xanh 200)
- LOWER-CENTER (1/3): trong "máy của bạn" boundary: chip "lark-cli mcp serve
  --transport http --addr 127.0.0.1:3000", endpoints "POST /  POST /mcp  GET
  /health" — cool cyan
- RIGHT (1/5): mũi tên HTTPS đi Lark Cloud với logo Lark chuẩn

Background: deep navy #0B1220, blueprint grid, particle dust, god-rays,
warm/cool gradient seam.

Style: Vercel × Linear × Apple keynote, glass morphism, rim lighting.

Text hiển thị:
- Header: "KIẾN TRÚC WEB · HTTP + TUNNEL"
- Node labels: "Cloudflare Tunnel · Bearer token gate · :3000"
- Warning chip: "⚠️ Cổng công khai = PHẢI có bảo mật"
- Tiếng Việt đúng dấu

Footer: "TRANSFORM GROUP" thin uppercase white sans-serif, wide letter-spacing.

Aspect ratio: 16:9 widescreen, dùng cho slide deck.
```

---

## SLIDE 8 — Bảng so sánh Desktop vs Web

```
Tạo ảnh 16:9 widescreen (1920×1080) cho slide thuyết trình.

Cinematic isometric 3D illustration của comparison table 2-column floating glass
card so sánh Desktop vs Web, góc nhìn 30°, ultra-detailed, octane-render quality.

Bố cục trái → phải:
- LEFT (1/5): label "DESKTOP vs WEB" + 2 icon 💻 / 🌐
- CENTER (3/5): large floating comparison table 2 cột, 6 rows:
  ROW 1 "Chạy ở đâu" — Desktop "Máy bạn (local)" vs Web "Browser + cổng công khai"
  ROW 2 "Transport" — "stdio" vs "HTTP streamable + tunnel"
  ROW 3 "Độ dễ" — "⭐ Dễ, cài 1 lần" vs "⭐⭐⭐ Cần cổng HTTPS"
  ROW 4 "Bảo mật cần thêm" — "Không (local)" vs "Bearer token + Cloudflare Access"
  ROW 5 "Internet công khai" — "Không cần" vs "Bắt buộc"
  ROW 6 "Phù hợp" — "Mọi nhân viên" vs "Tổ chức có admin/IT"
  Cột Desktop warm amber-gold glow, cột Web cool cyan glow
- RIGHT (1/5): khuyến nghị "Bắt đầu DESKTOP trước → Web khi cần đội ngũ truy cập
  từ browser"

Numbered row markers ①→⑥.

Background: deep navy #0B1220, blueprint grid, particle dust, god-rays.

Style: Vercel × Linear × Apple keynote, glass morphism, table card aesthetic.

Text hiển thị:
- Header: "DESKTOP vs WEB — CHỌN ĐƯỜNG NÀO?"
- 6 row labels + values readable
- Recommendation readable
- Tiếng Việt đúng dấu

Footer: "TRANSFORM GROUP" thin uppercase white sans-serif, wide letter-spacing.

Aspect ratio: 16:9 widescreen, dùng cho slide deck.
```

---

## SLIDE 9 — Desktop Bước 1: Cài binary

```
Tạo ảnh 16:9 widescreen (1920×1080) cho slide thuyết trình.

Cinematic isometric 3D illustration của bước cài đặt lark-cli binary qua script
setup-mcp.sh, góc nhìn 30°, ultra-detailed, octane-render quality.

Bố cục trái → phải:
- LEFT (1/5): badge lớn "DESKTOP · BƯỚC 1/4" + icon ⚙️ — cool cyan
- UPPER-CENTER (1/3): terminal glass card hiển thị command:
  `./scripts/setup-mcp.sh`           (cài vào ~/bin/lark-cli)
  `./scripts/setup-mcp.sh /usr/local/bin`  (toàn máy)
  — warm amber-gold prompt glow
- LOWER-CENTER (1/3): 4 bước script tự làm dạng checklist:
  "✓ Kiểm tra toolchain"
  "✓ Build từ mã nguồn"
  "✓ Cài binary vào PATH"
  "✓ Chạy lark-cli mcp tools → xác nhận 21 tool"
  — cool cyan
- RIGHT (1/5): cảnh báo glass card "⚠️ macOS: dùng ĐƯỜNG DẪN TUYỆT ĐỐI (which
  lark-cli) — app GUI không đọc PATH terminal"

Background: deep navy #0B1220, blueprint grid, particle dust, god-rays.

Style: Vercel × Linear × Apple keynote, glass morphism, rim lighting, terminal aesthetic.

Text hiển thị:
- Header: "BƯỚC 1 · CÀI lark-cli"
- Commands readable
- 4-step checklist readable
- Warning readable
- Tiếng Việt đúng dấu

Footer: "TRANSFORM GROUP" thin uppercase white sans-serif, wide letter-spacing.

Aspect ratio: 16:9 widescreen, dùng cho slide deck.
```

---

## SLIDE 10 — Desktop Bước 2: Đăng nhập Lark

```
Tạo ảnh 16:9 widescreen (1920×1080) cho slide thuyết trình.

Cinematic isometric 3D illustration của bước đăng nhập Lark qua OAuth trình
duyệt, token lưu vào OS keychain vault, góc nhìn 30°, ultra-detailed,
octane-render quality.

Logo Lark trên trang đăng nhập phải dùng đúng logo từ reference:
https://framerusercontent.com/images/yjUmyrg0qV5d2Glkjn9sDqzacY.svg

Bố cục trái → phải:
- LEFT (1/5): badge "DESKTOP · BƯỚC 2/4" + icon 🔑 — cool cyan
- UPPER-CENTER (1/3): terminal card `lark-cli auth login` → mũi tên → browser
  window mở trang đăng nhập Lark với logo Lark chuẩn + nút "Cấp quyền" — warm
  amber-gold
- LOWER-CENTER (1/3): OS Keychain vault (macOS Keychain / Windows Credential
  Manager) đóng token an toàn vào trong, label "Token KHÔNG nằm file thường" —
  cool cyan
- RIGHT (1/5): kiểm tra `lark-cli auth status` → kết quả "user: ready · bot:
  ready" với check xanh

Background: deep navy #0B1220, blueprint grid, particle dust, god-rays.

Style: Vercel × Linear × Apple keynote, glass morphism, rim lighting.

Text hiển thị:
- Header: "BƯỚC 2 · ĐĂNG NHẬP LARK"
- Commands: "lark-cli auth login · lark-cli auth status"
- Vault label readable
- Status "user: ready"
- Tiếng Việt đúng dấu

Footer: "TRANSFORM GROUP" thin uppercase white sans-serif, wide letter-spacing.

Aspect ratio: 16:9 widescreen, dùng cho slide deck.
```

---

## SLIDE 11 — Desktop Bước 3: claude_desktop_config.json

```
Tạo ảnh 16:9 widescreen (1920×1080) cho slide thuyết trình.

Cinematic isometric 3D illustration của bước khai báo lark-cli vào file cấu hình
Claude Desktop, hiển thị JSON snippet nổi 3D, góc nhìn 30°, ultra-detailed,
octane-render quality.

Bố cục trái → phải:
- LEFT (1/5): badge "DESKTOP · BƯỚC 3/4" + icon 📝 — cool cyan + đường dẫn file
  nhỏ: "~/Library/Application Support/Claude/claude_desktop_config.json"
- CENTER (3/5): large floating JSON code card với syntax highlight:
  {
    "mcpServers": {
      "lark-cli": {
        "command": "/Users/ban/bin/lark-cli",
        "args": ["mcp","serve"],
        "env": { "NO_COLOR": "1" }
      }
    }
  }
  — warm amber-gold trên key, cool cyan trên value
- RIGHT (1/5): 2 callout chip:
  "🎯 command = đường dẫn TUYỆT ĐỐI"
  "🎯 NO_COLOR=1 tránh hỏng JSON-RPC"
  + note "Gộp vào mcpServers, giữ server khác"

Background: deep navy #0B1220, blueprint grid, particle dust, god-rays.

Style: Vercel × Linear × Apple keynote, glass morphism, code editor aesthetic.

Text hiển thị:
- Header: "BƯỚC 3 · KHAI BÁO VÀO CLAUDE DESKTOP"
- JSON snippet readable, không sai cú pháp
- 2 callout readable
- Tiếng Việt đúng dấu

Footer: "TRANSFORM GROUP" thin uppercase white sans-serif, wide letter-spacing.

Aspect ratio: 16:9 widescreen, dùng cho slide deck.
```

---

## SLIDE 12 — Desktop Bước 4: Restart & verify Cowork

```
Tạo ảnh 16:9 widescreen (1920×1080) cho slide thuyết trình.

Cinematic isometric 3D illustration của bước khởi động lại Claude Desktop và xác
minh trong Cowork — connector lark-cli "connected" với 21 tool, góc nhìn 30°,
ultra-detailed, octane-render quality.

Logo Lark khi xuất hiện phải dùng đúng logo từ reference:
https://framerusercontent.com/images/yjUmyrg0qV5d2Glkjn9sDqzacY.svg

Bố cục trái → phải:
- LEFT (1/5): badge "DESKTOP · BƯỚC 4/4" + icon 🔄 "Quit & mở lại"
- UPPER-CENTER (1/3): Claude Desktop Connectors panel hiển thị "lark-cli ●
  connected" với danh sách tool scroll (lark_im_send, lark_calendar_agenda,
  lark_contact_search...) — warm amber-gold "connected" badge
- LOWER-CENTER (1/3): Cowork chat demo: user "Liệt kê lịch hôm nay của tôi" →
  Claude gọi tool `lark_calendar_agenda` → trả về danh sách sự kiện — cool cyan
- RIGHT (1/5): success card "🎉 21 TOOL SẴN SÀNG" + checkmark + "Hỏi bằng tiếng
  Việt, Claude tự làm"

Background: deep navy #0B1220, blueprint grid, particle dust, god-rays.

Style: Vercel × Linear × Apple keynote, glass morphism, rim lighting.

Text hiển thị:
- Header: "BƯỚC 4 · RESTART & KIỂM TRA"
- "lark-cli ● connected · 21 tools" readable
- Demo prompt readable
- Success readable
- Tiếng Việt đúng dấu

Footer: "TRANSFORM GROUP" thin uppercase white sans-serif, wide letter-spacing.

Aspect ratio: 16:9 widescreen, dùng cho slide deck.
```

---

## SLIDE 13 — Web Bước 1: HTTP server + bearer token

```
Tạo ảnh 16:9 widescreen (1920×1080) cho slide thuyết trình.

Cinematic isometric 3D illustration của bước chạy lark-cli ở chế độ HTTP kèm
bearer token bí mật, góc nhìn 30°, ultra-detailed, octane-render quality.

Bố cục trái → phải:
- LEFT (1/5): badge "WEB · BƯỚC 1/4" + icon 🌐 — cool cyan
- UPPER-CENTER (1/3): terminal glass card:
  `export LARK_MCP_BEARER_TOKEN=$(openssl rand -hex 32)`
  `lark-cli mcp serve --transport http --addr 127.0.0.1:3000 \`
  `  --audit-log ~/.lark-mcp-audit.ndjson`
  — warm amber-gold
- LOWER-CENTER (1/3): server boot log card:
  "✓ bearer-token auth ENABLED"
  "endpoints: POST /  POST /mcp  GET /health"
  + token chip "🔑 sinh ngẫu nhiên 32 byte — lưu lại để khai báo connector" —
  cool cyan
- RIGHT (1/5): cảnh báo "Không set token → log UNAUTHENTICATED → CHỈ dùng khi
  còn bound 127.0.0.1, KHÔNG mở tunnel"

Background: deep navy #0B1220, blueprint grid, particle dust, god-rays.

Style: Vercel × Linear × Apple keynote, glass morphism, terminal aesthetic.

Text hiển thị:
- Header: "WEB BƯỚC 1 · HTTP SERVER + TOKEN"
- Commands readable
- Boot log "bearer-token auth ENABLED" readable
- Warning readable
- Tiếng Việt đúng dấu

Footer: "TRANSFORM GROUP" thin uppercase white sans-serif, wide letter-spacing.

Aspect ratio: 16:9 widescreen, dùng cho slide deck.
```

---

## SLIDE 14 — Web Bước 2: Cloudflare Tunnel

```
Tạo ảnh 16:9 widescreen (1920×1080) cho slide thuyết trình.

Cinematic isometric 3D illustration của Cloudflare Tunnel mở một URL HTTPS công
khai trỏ về server local :3000, góc nhìn 30°, ultra-detailed, octane-render quality.

Bố cục trái → phải:
- LEFT (1/5): badge "WEB · BƯỚC 2/4" + Cloudflare-style cloud icon (không trùng
  thương hiệu, generic orange cloud) — cool cyan
- UPPER-CENTER (1/3): terminal card `cloudflared tunnel --url
  http://127.0.0.1:3000` → output URL phát sáng `https://<ngẫu-nhiên>.
  trycloudflare.com` — warm amber-gold
- LOWER-CENTER (1/3): đường hầm kính 3D nối "Internet công khai" (trái) xuyên
  về "máy của bạn :3000" (phải), packets chảy qua — cool cyan
- RIGHT (1/5): note "Quick tunnel: URL đổi mỗi lần chạy, tắt khi đóng máy" +
  "Muốn cố định → Named Tunnel"

Background: deep navy #0B1220, blueprint grid, particle dust, god-rays.

Style: Vercel × Linear × Apple keynote, glass morphism, tunnel aesthetic.

Text hiển thị:
- Header: "WEB BƯỚC 2 · CLOUDFLARE TUNNEL"
- Command + URL readable
- Note readable
- Tiếng Việt đúng dấu

Footer: "TRANSFORM GROUP" thin uppercase white sans-serif, wide letter-spacing.

Aspect ratio: 16:9 widescreen, dùng cho slide deck.
```

---

## SLIDE 15 — Web Bước 3: Kiểm chứng bảo mật (curl 401/200)

```
Tạo ảnh 16:9 widescreen (1920×1080) cho slide thuyết trình.

Cinematic isometric 3D illustration của bước tự kiểm chứng bearer token: request
thiếu token bị 401, request đúng token trả 200 + danh sách tool, góc nhìn 30°,
ultra-detailed, octane-render quality.

Bố cục trái → phải:
- LEFT (1/5): badge "WEB · BƯỚC 3/4" + icon 🛡️ "BẮT BUỘC trước khi mở"
- UPPER-CENTER (1/3): split test panel:
  TOP (đỏ) "THIẾU TOKEN": curl POST /mcp không header → kết quả lớn "401
  Unauthorized" ✗
  BOTTOM (xanh) "ĐÚNG TOKEN": curl POST /mcp với `Authorization: Bearer <token>`
  → "200 OK · danh sách 21 tool" ✓
  — warm amber-gold trên nhãn test
- LOWER-CENTER (1/3): note kỹ thuật "So sánh token constant-time (crypto/subtle)
  → không lộ token qua timing. /health luôn mở cho liveness." — cool cyan
- RIGHT (1/5): rule card "❗ Nếu thiếu-token KHÔNG trả 401 → DỪNG, đừng mở tunnel"

Background: deep navy #0B1220, blueprint grid, particle dust, god-rays,
warm/cool gradient ngang.

Style: Vercel × Linear × Apple keynote, glass morphism, rim lighting.

Text hiển thị:
- Header: "WEB BƯỚC 3 · KIỂM CHỨNG BẢO MẬT"
- "401 thiếu token" vs "200 đúng token" readable
- Technical note readable
- Rule readable
- Tiếng Việt đúng dấu

Footer: "TRANSFORM GROUP" thin uppercase white sans-serif, wide letter-spacing.

Aspect ratio: 16:9 widescreen, dùng cho slide deck.
```

---

## SLIDE 16 — Web Bước 4: Custom connector claude.ai

```
Tạo ảnh 16:9 widescreen (1920×1080) cho slide thuyết trình.

Cinematic isometric 3D illustration của bước khai báo Custom Connector trên
claude.ai và gọi thử một tool, góc nhìn 30°, ultra-detailed, octane-render quality.

Logo Lark khi xuất hiện phải dùng đúng logo từ reference:
https://framerusercontent.com/images/yjUmyrg0qV5d2Glkjn9sDqzacY.svg

Bố cục trái → phải:
- LEFT (1/5): badge "WEB · BƯỚC 4/4" + icon 🔗
- UPPER-CENTER (1/3): claude.ai Settings → Connectors → "Add custom connector"
  form glass card:
  "URL: https://<...>.trycloudflare.com/mcp"
  "Auth header (nếu form hỗ trợ): Authorization: Bearer <token>"
  — warm amber-gold
- LOWER-CENTER (1/3): browser claude.ai chat: user "Tìm doc tên Kế hoạch Q3" →
  tool `lark_doc_search` chạy → trả kết quả — cool cyan
- RIGHT (1/5): note tương thích "⚠️ Nếu UI không cho header tĩnh → dùng
  Cloudflare Access làm lớp chặn, đừng mở no-auth ra tunnel trần"

Background: deep navy #0B1220, blueprint grid, particle dust, god-rays.

Style: Vercel × Linear × Apple keynote, glass morphism, rim lighting.

Text hiển thị:
- Header: "WEB BƯỚC 4 · CUSTOM CONNECTOR"
- URL + auth header readable
- Demo prompt readable
- Compatibility note readable
- Tiếng Việt đúng dấu

Footer: "TRANSFORM GROUP" thin uppercase white sans-serif, wide letter-spacing.

Aspect ratio: 16:9 widescreen, dùng cho slide deck.
```

---

## SLIDE 17 — Danh tính User vs Bot

```
Tạo ảnh 16:9 widescreen (1920×1080) cho slide thuyết trình.

Cinematic isometric 3D illustration của 2 danh tính thao tác Lark — User (thay
mặt bạn) và Bot (danh nghĩa app), góc nhìn 30°, ultra-detailed, octane-render
quality.

Bố cục trái → phải:
- LEFT (1/5): label "2 DANH TÍNH" + icon 👤 / 🤖
- UPPER-CENTER (1/3): 2 glass panel cạnh nhau:
  "👤 USER — thao tác THAY MẶT BẠN · đọc mail/lịch/task của chính bạn · hầu hết
  skill cá nhân" (warm amber-gold)
  "🤖 BOT — danh nghĩa app/bot · gửi thông báo group, automation" (cool cyan)
- LOWER-CENTER (1/3): cách chọn: `--as user` / `--as bot` (CLI) hoặc tham số
  `as` trong tool MCP; mặc định `auto`
- RIGHT (1/5): callout "Hầu hết skill Cowork cần danh tính USER → nhớ auth login"

Background: deep navy #0B1220, blueprint grid, particle dust, god-rays,
warm/cool gradient.

Style: Vercel × Linear × Apple keynote, glass morphism, rim lighting.

Text hiển thị:
- Header: "USER vs BOT"
- 2 panel mô tả readable
- "--as user / --as bot" readable
- Callout readable
- Tiếng Việt đúng dấu

Footer: "TRANSFORM GROUP" thin uppercase white sans-serif, wide letter-spacing.

Aspect ratio: 16:9 widescreen, dùng cho slide deck.
```

---

## SLIDE 18 — Quyền/scope + app Lark riêng

```
Tạo ảnh 16:9 widescreen (1920×1080) cho slide thuyết trình.

Cinematic isometric 3D illustration của hệ thống quyền (scopes) và lựa chọn app
Lark công khai vs app riêng của tổ chức, góc nhìn 30°, ultra-detailed,
octane-render quality.

Bố cục trái → phải:
- LEFT (1/5): label "QUYỀN & SCOPE" + icon 🛡️
- UPPER-CENTER (1/3): scope grant flow: trình duyệt đăng nhập → cấp các quyền
  (đọc mail, đọc lịch, gửi tin...) → nếu thiếu scope → `lark-cli doctor` chẩn
  đoán — warm amber-gold
- LOWER-CENTER (1/3): 2 lựa chọn app dạng glass card:
  "🌐 App công khai — tiện cho cá nhân"
  "🏢 App Lark RIÊNG của tổ chức — App ID/Secret riêng, scope tối thiểu, quản lý
  tập trung, thu hồi được" (highlighted cho enterprise)
  — cool cyan
- RIGHT (1/5): callout "Enterprise → app riêng + least privilege" + redirect URL
  note "localhost:3000/callback"

Background: deep navy #0B1220, blueprint grid, particle dust, god-rays.

Style: Vercel × Linear × Apple keynote, glass morphism, rim lighting.

Text hiển thị:
- Header: "QUYỀN & APP LARK RIÊNG"
- Scope flow readable
- 2 app options readable
- Enterprise callout readable
- Tiếng Việt đúng dấu

Footer: "TRANSFORM GROUP" thin uppercase white sans-serif, wide letter-spacing.

Aspect ratio: 16:9 widescreen, dùng cho slide deck.
```

---

## SLIDE 19 — Bộ skill Cowork là gì

```
Tạo ảnh 16:9 widescreen (1920×1080) cho slide thuyết trình.

Cinematic isometric 3D illustration của khái niệm "skill = công thức" — một cuốn
sổ tay hướng dẫn Claude tự chạy chuỗi công cụ Lark từ một câu kích hoạt tiếng
Việt, góc nhìn 30°, ultra-detailed, octane-render quality.

Bố cục trái → phải:
- LEFT (1/5): business user nói "morning" — speech bubble cyan
- UPPER-CENTER (1/3): một cuốn "SKILL: morning-brief" mở ra, Claude brain đọc,
  bên trong là pipeline "gọi mail-triage + im-digest + approval-triage +
  task-prioritizer SONG SONG" — warm amber-gold
- LOWER-CENTER (1/3): output gộp ≤15 dòng: "📧 3 mail gấp · 💬 2 group cần action
  · ✅ 5 approval chờ · 🎯 top 5 task" — cool cyan
- RIGHT (1/5): callout "Skill = VĂN BẢN công thức, KHÔNG phải code · bạn chỉ nói
  tự nhiên"

Background: deep navy #0B1220, blueprint grid, particle dust, god-rays.

Style: Vercel × Linear × Apple keynote, glass morphism, rim lighting.

Text hiển thị:
- Header: "BỘ SKILL COWORK · CÔNG THỨC SẴN"
- Pipeline labels readable
- Output sample readable
- Callout readable
- Tiếng Việt đúng dấu

Footer: "TRANSFORM GROUP" thin uppercase white sans-serif, wide letter-spacing.

Aspect ratio: 16:9 widescreen, dùng cho slide deck.
```

---

## SLIDE 20 — 23 skill catalog

```
Tạo ảnh 16:9 widescreen (1920×1080) cho slide thuyết trình.

Cinematic isometric 3D illustration của 23 skill cards arranged như app launcher
grid, góc nhìn 30°, ultra-detailed, octane-render quality.

Bố cục trái → phải:
- LEFT (1/5): label "23 SKILL COWORK" lớn + subtitle "Nói tiếng Việt là chạy"
- CENTER (3/5): grid card (mỗi card tên skill + icon đặc trưng), đan xen warm
  amber-gold và cool cyan border:
  morning-brief · daily-digest · weekly-review · overwhelm-triage ·
  inbox-zero · im-digest · client-followup · meeting-prep ·
  calendar-optimizer · focus-mode · one-on-one-prep · contact-360 ·
  task-prioritizer · approval-triage · doc-from-template · doc-restructure ·
  decision-logger · permission-audit · deal-update · pipeline-review ·
  incident-retro · sprint-retro · lark-cli-mcp
- RIGHT (1/5): stats panel:
  "📦 23 skill"
  "🗣️ trigger tiếng Việt"
  "⚡ chạy chuỗi tool tự động"

Background: deep navy #0B1220, blueprint grid, particle dust, god-rays.

Style: Vercel × Linear × Apple keynote, glass morphism, card grid aesthetic.

Text hiển thị:
- Header: "23 SKILL COWORK · CATALOG"
- 23 skill names readable trên cards
- Stats readable
- Tiếng Việt đúng dấu

Footer: "TRANSFORM GROUP" thin uppercase white sans-serif, wide letter-spacing.

Aspect ratio: 16:9 widescreen, dùng cho slide deck.
```

---

## SLIDE 21 — Skill: Đầu ngày + Mail/Chat

```
Tạo ảnh 16:9 widescreen (1920×1080) cho slide thuyết trình.

Cinematic isometric 3D illustration của nhóm skill "Đầu ngày" và "Mail/Chat" với
ví dụ kích hoạt thật, góc nhìn 30°, ultra-detailed, octane-render quality.

Bố cục trái → phải:
- LEFT (1/5): label "NHÓM 1 · ĐẦU NGÀY + MAIL/CHAT"
- UPPER-CENTER (1/3): "ĐẦU NGÀY / TỔNG HỢP" (warm amber-gold):
  "morning-brief → 'morning' · daily-digest → 'tổng kết hôm nay' · weekly-review
  → 'báo cáo tuần' · overwhelm-triage → 'tôi quá tải'"
- LOWER-CENTER (1/3): "MAIL & CHAT" (cool cyan):
  "inbox-zero → 'clear inbox' · im-digest → '47 group có gì' · client-followup →
  'khách im lặng' (chỉ soạn nháp)"
- RIGHT (1/5): demo card user "morning" → output brief ≤15 dòng

Background: deep navy #0B1220, blueprint grid, particle dust, god-rays,
warm/cool gradient.

Style: Vercel × Linear × Apple keynote, glass morphism, rim lighting.

Text hiển thị:
- Header: "SKILL · ĐẦU NGÀY + MAIL/CHAT"
- Skill names + trigger tiếng Việt readable
- Demo readable
- Tiếng Việt đúng dấu

Footer: "TRANSFORM GROUP" thin uppercase white sans-serif, wide letter-spacing.

Aspect ratio: 16:9 widescreen, dùng cho slide deck.
```

---

## SLIDE 22 — Skill: Họp/Task/Sales

```
Tạo ảnh 16:9 widescreen (1920×1080) cho slide thuyết trình.

Cinematic isometric 3D illustration của nhóm skill "Họp & Lịch", "Task &
Approval", "Sales/Vận hành", góc nhìn 30°, ultra-detailed, octane-render quality.

Bố cục — 3 cột glass card:
- LEFT (1/5): label "NHÓM 2 · HỌP · TASK · SALES"
- CỘT 1 "HỌP & LỊCH" (warm amber-gold): "meeting-prep · calendar-optimizer ·
  focus-mode · one-on-one-prep · contact-360"
- CỘT 2 "TASK & APPROVAL" (cyan): "task-prioritizer → 'top 5 today' ·
  approval-triage → 'có gì cần duyệt'"
- CỘT 3 "SALES/VẬN HÀNH" (warm amber-gold): "deal-update · pipeline-review ·
  incident-retro · sprint-retro · decision-logger · permission-audit ·
  doc-from-template · doc-restructure"
- RIGHT (1/5): demo "chuẩn bị họp với khách ABC" → contact-360 + meeting-prep
  brief

Background: deep navy #0B1220, blueprint grid, particle dust, god-rays.

Style: Vercel × Linear × Apple keynote, glass morphism, rim lighting,
3-column aesthetic.

Text hiển thị:
- Header: "SKILL · HỌP · TASK · SALES"
- 3 cột skill names readable
- Demo readable
- Tiếng Việt đúng dấu

Footer: "TRANSFORM GROUP" thin uppercase white sans-serif, wide letter-spacing.

Aspect ratio: 16:9 widescreen, dùng cho slide deck.
```

---

## SLIDE 23 — 21 công cụ MCP theo domain

```
Tạo ảnh 16:9 widescreen (1920×1080) cho slide thuyết trình.

Cinematic isometric 3D illustration của 21 công cụ MCP nhóm theo domain Lark,
dạng radial map quanh chip lark-cli, góc nhìn 30°, ultra-detailed, octane-render
quality.

Logo Lark trên cụm trung tâm phải dùng đúng logo từ reference:
https://framerusercontent.com/images/yjUmyrg0qV5d2Glkjn9sDqzacY.svg

Bố cục — radial map, chip "lark-cli mcp" trung tâm với logo Lark, các cụm domain
toả ra:
- "IM (3): lark_im_send · lark_im_card_send · lark_im_search"
- "MAIL (2): lark_mail_send · lark_mail_draft_create"
- "CALENDAR (2): lark_calendar_agenda · lark_calendar_create"
- "DOCS/DRIVE/SHEETS (6): doc_create/search/fetch · drive_upload ·
  sheets_read/append"
- "BASE (1): lark_base_search"
- "CONTACT (1): lark_contact_search"
- "TASK (2): lark_task_my · lark_task_create"
- "MEETINGS (2): lark_vc_search · lark_minutes_search"
- "OKR (1): lark_okr_cycle_list"
- "GENERIC (1): lark_api — cửa thoát hiểm"
Cụm đọc (read-only) màu cool cyan, cụm ghi (write) viền warm amber-gold.

Background: deep navy #0B1220, blueprint grid, particle dust, god-rays.

Style: Vercel × Linear × Apple keynote, glass morphism, radial map aesthetic.

Text hiển thị:
- Header: "21 CÔNG CỤ MCP · THEO DOMAIN"
- Tên tool readable theo cụm
- Tổng "21 tools"
- Tiếng Việt đúng dấu

Footer: "TRANSFORM GROUP" thin uppercase white sans-serif, wide letter-spacing.

Aspect ratio: 16:9 widescreen, dùng cho slide deck.
```

---

## SLIDE 24 — An toàn từng tool (dry-run)

```
Tạo ảnh 16:9 widescreen (1920×1080) cho slide thuyết trình.

Cinematic isometric 3D illustration của cơ chế an toàn mặc định: dry-run xem
trước mọi thao tác ghi, mail mặc định lưu nháp, góc nhìn 30°, ultra-detailed,
octane-render quality.

Bố cục trái → phải:
- LEFT (1/5): label "AN TOÀN MẶC ĐỊNH"
- UPPER-CENTER (1/3): split "DRY-RUN":
  tool ghi (vd lark_task_create) chạy `dry_run=true` → preview "Sẽ tạo task: ...
  assignee ... due ..." KHÔNG ghi thật → user duyệt → chạy lại commit
  — warm amber-gold
- LOWER-CENTER (1/3): "MAIL = NHÁP MẶC ĐỊNH":
  lark_mail_send không `confirm_send` → lưu Drafts, KHÔNG bay đi; cần
  `confirm_send=true` mới gửi — cool cyan
- RIGHT (1/5): 3 chip "✅ Xem trước · ✅ Không tự gửi · ✅ Bạn quyết định"

Background: deep navy #0B1220, blueprint grid, particle dust, god-rays,
warm/cool gradient ngang.

Style: Vercel × Linear × Apple keynote, glass morphism, rim lighting.

Text hiển thị:
- Header: "AN TOÀN TỪNG TOOL"
- "dry_run" + "confirm_send" readable
- 3 chip readable
- Tiếng Việt đúng dấu

Footer: "TRANSFORM GROUP" thin uppercase white sans-serif, wide letter-spacing.

Aspect ratio: 16:9 widescreen, dùng cho slide deck.
```

---

## SLIDE 25 — Data flow: data đi đâu

```
Tạo ảnh 16:9 widescreen (1920×1080) cho slide thuyết trình.

Cinematic isometric 3D illustration của luồng dữ liệu end-to-end, nhấn mạnh data
chỉ đi tới máy chủ Lark, góc nhìn 30°, ultra-detailed, octane-render quality.

Logo Lark trên đích phải dùng đúng logo từ reference:
https://framerusercontent.com/images/yjUmyrg0qV5d2Glkjn9sDqzacY.svg

Bố cục trái → phải (1 dòng chảy):
- "Claude (Desktop/Web)" → "lark-cli (máy bạn)" → "HTTPS" → "open.larksuite.com
  / open.feishu.cn" với logo Lark chuẩn
- Desktop lane (warm amber-gold): annotation "Tất cả LOCAL, không bên thứ ba"
- Web lane (cool cyan): annotation "Đi qua CỔNG MCP bạn dựng → phải bảo vệ"
- Token vault icon dọc đường: "Token trong OS Keychain, không file thường"

Background: deep navy #0B1220, blueprint grid, particle dust, god-rays.

Style: Vercel × Linear × Apple keynote, glass morphism, data-flow aesthetic.

Text hiển thị:
- Header: "DATA ĐI ĐÂU?"
- Flow labels readable
- 2 annotation (desktop/web) readable
- Keychain note readable
- Tiếng Việt đúng dấu

Footer: "TRANSFORM GROUP" thin uppercase white sans-serif, wide letter-spacing.

Aspect ratio: 16:9 widescreen, dùng cho slide deck.
```

---

## SLIDE 26 — Security stack 4 lớp

```
Tạo ảnh 16:9 widescreen (1920×1080) cho slide thuyết trình.

Cinematic isometric 3D illustration của 4 lớp bảo vệ bọc quanh lark-cli, góc nhìn
30°, ultra-detailed, octane-render quality.

Bố cục trái → phải:
- LEFT (1/5): threat actor bị chặn bởi lớp ngoài
- CENTER (3/5): 4 concentric glass shields, mỗi shield label:
  "① 🔐 OS KEYCHAIN — token không plaintext"
  "② 🔑 BEARER TOKEN — chặn HTTP trái phép (401), so sánh constant-time"
  "③ 📋 AUDIT LOG — ghi mọi tool call (--audit-log)"
  "④ 🧪 DRY-RUN — xem trước mọi thao tác ghi"
  — warm amber-gold trên shields, chip lark-cli cool cyan ở tâm
- RIGHT (1/5): 3 outcome "✅ Token không leak · ✅ Truy vết được · ✅ Đảo ngược lỗi"

Numbered markers ①②③④.

Background: deep navy #0B1220, blueprint grid, particle dust, god-rays.

Style: Vercel × Linear × Apple keynote, glass morphism, rim lighting.

Text hiển thị:
- Header: "SECURITY · 4 LỚP BẢO VỆ"
- 4 shield labels readable
- 3 outcome readable
- Tiếng Việt đúng dấu

Footer: "TRANSFORM GROUP" thin uppercase white sans-serif, wide letter-spacing.

Aspect ratio: 16:9 widescreen, dùng cho slide deck.
```

---

## SLIDE 27 — Mở cổng web an toàn — checklist

```
Tạo ảnh 16:9 widescreen (1920×1080) cho slide thuyết trình.

Cinematic isometric 3D illustration của checklist bắt buộc trước khi expose cổng
web ra internet, dạng các bậc thang khoá an toàn, góc nhìn 30°, ultra-detailed,
octane-render quality.

Bố cục trái → phải:
- LEFT (1/5): cảnh báo lớn "⚠️ CỔNG HTTP = PHƠI TOÀN QUYỀN LARK"
- CENTER (3/5): checklist 4 mục dạng glass card có ô tick:
  "☑ Bearer token (LARK_MCP_BEARER_TOKEN) — curl thiếu token PHẢI nhận 401"
  "☑ Cloudflare Access trên hostname (email/OTP) — lớp 2, nhất là Named Tunnel"
  "☑ Bật --audit-log để có dấu vết"
  "☑ Chỉ mở khi cần, tắt khi xong"
  — warm amber-gold
- LOWER-CENTER: note "Đa người dùng → cần sidecar/gateway đa-tenant (tách
  credential từng người) — việc của admin/IT" — cool cyan
- RIGHT (1/5): rule "❗ Tuyệt đối KHÔNG mở URL tunnel trần, no-auth"

Background: deep navy #0B1220, blueprint grid, particle dust, god-rays.

Style: Vercel × Linear × Apple keynote, glass morphism, rim lighting,
checklist aesthetic.

Text hiển thị:
- Header: "MỞ CỔNG WEB AN TOÀN · CHECKLIST"
- 4 checklist items readable
- Multi-tenant note readable
- Rule readable
- Tiếng Việt đúng dấu

Footer: "TRANSFORM GROUP" thin uppercase white sans-serif, wide letter-spacing.

Aspect ratio: 16:9 widescreen, dùng cho slide deck.
```

---

## SLIDE 28 — Xử lý sự cố (gồm server treo)

```
Tạo ảnh 16:9 widescreen (1920×1080) cho slide thuyết trình.

Cinematic isometric 3D illustration của bảng xử lý sự cố thường gặp + ô đặc biệt
"server treo khi khởi động", góc nhìn 30°, ultra-detailed, octane-render quality.

Bố cục trái → phải:
- LEFT (1/5): label "XỬ LÝ SỰ CỐ" + icon 🛠️ + "Mẹo: chạy lệnh lark-cli tương
  đương ở terminal để xem lý do thật"
- CENTER (3/5): bảng 2 cột "Triệu chứng → Khắc phục" (warm amber-gold):
  "command not found (macOS) → dùng đường dẫn TUYỆT ĐỐI (which lark-cli)"
  "Tool trả isError → chạy lệnh tương đương ở terminal"
  "Output lỗi / mất kết nối → đặt NO_COLOR=1, không in stdout lạ"
  "user: missing → lark-cli auth login"
- LOWER-CENTER: ô đỏ nổi bật "⚠️ SERVER TREO KHI KHỞI ĐỘNG (cwd ≠ repo)":
  "1) xattr -dr com.apple.quarantine ~/bin/lark-cli"
  "2) codesign -s - --force ~/bin/lark-cli"
  "3) chạy thử 1 lần, bấm Always Allow nếu keychain hỏi"
  "(lark-cli --version có thể treo do update-notifier — dùng mcp tools để test)"
  — cool cyan
- RIGHT (1/5): "/mcp-doctor → báo cáo sức khoẻ 1 trang"

Background: deep navy #0B1220, blueprint grid, particle dust, god-rays.

Style: Vercel × Linear × Apple keynote, glass morphism, table aesthetic.

Text hiển thị:
- Header: "XỬ LÝ SỰ CỐ"
- Bảng triệu chứng/khắc phục readable
- Ô "server treo" + 3 lệnh readable
- Tiếng Việt đúng dấu

Footer: "TRANSFORM GROUP" thin uppercase white sans-serif, wide letter-spacing.

Aspect ratio: 16:9 widescreen, dùng cho slide deck.
```

---

## SLIDE 29 — Thói quen hàng ngày + bảo trì

```
Tạo ảnh 16:9 widescreen (1920×1080) cho slide thuyết trình.

Cinematic isometric 3D illustration của một ngày dùng Lark CLI (5 touchpoint) kết
hợp lưu ý bảo trì, dạng daily timeline horizontal, góc nhìn 30°, ultra-detailed,
octane-render quality.

Bố cục trái → phải:
- LEFT (1/5): morning sun icon + label "1 NGÀY + BẢO TRÌ"
- CENTER (3/5): day timeline 5 touchpoint glass card:
  "🌅 SÁNG — 'morning' → brief đầu ngày" (warm)
  "📋 TRƯỚC HỌP — 'chuẩn bị họp X'" (cyan)
  "✅ SAU HỌP — 'action items từ meeting → tạo task'" (warm)
  "📊 CUỐI TUẦN — 'báo cáo tuần'" (cyan)
  "📅 KHI CẦN — 'lên lịch họp tránh giờ trưa'" (warm)
- LOWER-CENTER: dải bảo trì "BẢO TRÌ" (cool cyan):
  "lark-cli update (binary + skill) · Sau đổi binary: QUIT & mở lại Claude
  Desktop · Backup binary trước khi thay · Web: Named Tunnel + launchd để sống
  lâu"
- RIGHT (1/5): ROI panel "⏱️ tiết kiệm ~10h/tuần · 🎯 0 deadline miss"

Background: deep navy #0B1220, blueprint grid, particle dust, god-rays.

Style: Vercel × Linear × Apple keynote, glass morphism, rim lighting,
daily journey aesthetic.

Text hiển thị:
- Header: "THÓI QUEN HÀNG NGÀY + BẢO TRÌ"
- 5 touchpoint readable
- Dải bảo trì readable
- ROI readable
- Tiếng Việt đúng dấu

Footer: "TRANSFORM GROUP" thin uppercase white sans-serif, wide letter-spacing.

Aspect ratio: 16:9 widescreen, dùng cho slide deck.
```

---

## SLIDE 30 — CTA Close

```
Tạo ảnh 16:9 widescreen (1920×1080) cho slide thuyết trình.

Cinematic isometric 3D illustration của closing scene — business user như nhạc
trưởng chỉ huy Claude + lark-cli + Lark ecosystem, góc nhìn 30°, ultra-detailed,
octane-render quality.

Logo Lark trên Lark ecosystem cloud phải dùng ĐÚNG logo từ reference:
https://framerusercontent.com/images/yjUmyrg0qV5d2Glkjn9sDqzacY.svg

Bố cục trái → phải:
- LEFT (1/5): user figure đứng như nhạc trưởng, baton in hand — warm amber-gold
  spotlight
- UPPER-CENTER (1/3): tagline lớn "BẠN NÓI · CLAUDE LÀM · LARK CHẠY" — warm
  amber-gold glow
- LOWER-CENTER (1/3): Claude brain center + cây cầu "MCP bridge" + chip lark-cli
  → Lark Cloud với logo Lark chuẩn — cool cyan
- RIGHT (1/5): 3 CTA stones diagonal lên:
  "① CÀI: ./scripts/setup-mcp.sh"
  "② ĐĂNG NHẬP: lark-cli auth login"
  "③ HỎI: 'Sáng nay tôi có gì?'"
  + final glow "🚀 DÙNG NGAY HÔM NAY · KHÔNG CẦN CLAUDE CODE"

Background: deep navy #0B1220, blueprint grid, particle dust dense, god-rays
intense from top-left, warm/cool gradient cinematic.

Style: Vercel × Linear × Apple keynote, glass morphism, rim lighting,
cinematic closing poster.

Text hiển thị:
- Tagline lớn: "BẠN NÓI · CLAUDE LÀM · LARK CHẠY"
- 3 CTA steps readable
- "DÙNG NGAY HÔM NAY · KHÔNG CẦN CLAUDE CODE"
- Tiếng Việt đúng dấu

Footer: "TRANSFORM GROUP" thin uppercase white sans-serif, wide letter-spacing.

Aspect ratio: 16:9 widescreen, dùng cho slide deck.
```

---

## ✅ Checklist khi render 30 slide

1. **Kích thước** 16:9 widescreen 1920×1080 trên cả 30 slide.
2. **Logo Lark từ reference URL** đúng (không tự vẽ lại) trên slide: 1, 3, 5, 6, 7, 10, 12, 16, 23, 25, 30.
3. **Color discipline:** warm amber-gold = trí tuệ/quyết định (Claude, planning, skill); cool cyan = thực thi (lark-cli, transport, Lark Cloud, security mechanics).
4. **Nội dung chính xác sản phẩm:** 21 công cụ MCP · 23 skill Cowork · 2 transport (stdio/http) · bearer token `LARK_MCP_BEARER_TOKEN` · endpoints `/ /mcp /health`.
5. **Tiếng Việt đúng dấu**, không gibberish; command/snippet đọc được, không sai cú pháp.
6. **Footer** `TRANSFORM GROUP` thin uppercase, wide letter-spacing trên mọi slide.

> Nguồn nội dung: [docs/01–09](README.md). Đã **loại** changelog kỹ thuật khỏi deck theo yêu cầu.
