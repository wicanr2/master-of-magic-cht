# Phase 1 — CJK 渲染 prototype 結果

日期:2026-06-21。結論:**注入點驗證成功,路線 B (TTF) 可行,已實作進引擎 `lib/font` 並編譯通過。**

## 做了什麼

1. **獨立 prototype** (`prototype/cjk-hello/`):用「逐 rune 預渲染 glyph image + 快取」機制
   (對應引擎 `lib/font/font.go` 的 `getGlyphImage` / `doPrint`) 畫出繁體中文,docker + Xvfb headless 截圖。
2. **引擎注入** (`patches/0001-cjk-font-injection.patch`):新增 `lib/font/cjk.go` + 改 `doPrint` /
   `MeasureTextWidth`,在 glyphIndex 超出 ASCII 範圍且碼點為 CJK 時改走 TTF glyph。
3. **編譯驗證**:docker (golang:1.25) `go build ./lib/font` 通過。

## 截圖

![Phase 1 CJK prototype](img/phase1-cjk-hello.png)

畫面內容:標題「工作魔法大帝」(2× scale) + 第一批 item 譯文 (烈焰 / 吸血 / 神聖復仇者 /
+3 攻擊 / 魔法免疫 / 隱形),全部經由 CJK glyph 支線渲染。字型用 Noto Sans CJK TC。

## 注入點 (確認)

| 位置 | 原行為 | 注入後 |
|---|---|---|
| `font.go` `doPrint()` 繪字迴圈 | `glyphIndex = int(c)-32`,超範圍 `continue` (丟棄非 ASCII) | 超範圍時呼叫 `cjkGlyphImage(c)`,有中文 glyph 就 blit 並前進 x;含 drop shadow |
| `font.go` `MeasureTextWidth()` | 同樣丟棄非 ASCII | 中文 glyph 寬度計入,讓置中/置右/換行正確 |
| `cjk.go` (新增) | — | TTF 載入 (env `MOM_CHT_FONT` 或系統路徑)、rune→glyph rasterize + 快取;字型不可用時回傳 nil 退回原行為,**不影響英文** |

## 關鍵發現 / 校正

- **Ebiten Linux 桌面走 CGo + glfw**,build 需 X11 dev headers (`libx11-dev` 等);purego 免 CGo 是
  **macOS / Windows** 路徑。已回填 `porting-difficulty.md`。
- **AR PL UMing (`uming.ttc`) 是 CFF/舊式 collection,`golang.org/x/image/sfnt` 解析失敗** (`invalid table offset`)。
  改用 **Noto Sans CJK TC** (`.ttc`,`ParseCollection` + `Font(0)`) 正常。→ 點陣路線 (A) 若要用 UMing,
  需走離線烘字 (`build_cjk_font.py`,freetype) 而非引擎內即時 sfnt。
- `.ttc` 一律走 `opentype.ParseCollection`,不是 `Parse`。

## 待辦 (Phase 1 收尾 / Phase 2 接續)

- [ ] 字級:目前 CJK 固定基準 16px,密集面板需多尺寸 (16/14) 與原版字高對齊 (避免破版)。
- [ ] 路線 A 決策:若要點陣美術一致,離線烘 24×24 atlas;B 已足以推進翻譯與版面測試。
- [ ] 把 CJK 字型納入打包 (自由授權 Noto / 或自製點陣),而非依賴系統字型路徑。
- [ ] 用真實遊戲 (套版權 lbx) 跑一畫面,觀察破版,回填 ADR 0001 定案。
