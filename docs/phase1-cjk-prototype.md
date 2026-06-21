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

## 真實引擎驗證 (Phase 1 收尾)

不只獨立 prototype,進一步在**真實引擎字型管線**上驗證:新增引擎 `test/cjk-render`
(基於 `test/font`),用引擎真正的 `VaultFonts` + `lib/font` 的 `Print` / `PrintDropShadow` /
`PrintOutline` 畫中英混排,套真實 `fonts.lbx` (1.31 資料),docker + Xvfb 跑、ImageMagick 抓 root window。

![真實引擎 CJK](img/phase1-engine-cjk.png)

**結果:CJK 在真實引擎三條繪字路徑全部渲染** — 標題/神器名/物品能力中英混排、drop shadow、
shader outline 都正確套到中文 glyph。英文續用引擎原本的金色花體點陣字,中文用 Noto。

**這一畫面也暴露兩個必修項 (跑真畫面的價值):**

1. **`PrintOutline` 原本漏 patch** → 中文在 outline 路徑被丟棄 (「神聖復仇者」一開始不顯示)。
   **已修**:`PrintOutline` 的 glyph 迴圈補上同樣的 CJK 分流 (patch 已含)。三路徑現一致。
2. **CJK 字級未對齊引擎字型高度** → CJK (固定基準 16px) 比引擎的小 ASCII 字 (如 ResourceFont) 大一截、
   基線不齊,行距一擠就**重疊破版**。這是 Phase 2 的核心精修:`cjkGlyphImage` 需依呼叫端字型的
   `internalFont.Height` 決定 CJK 渲染尺寸 (並以 (rune, size) 為快取 key),讓中英基線與行高一致。

## Phase 2 進度 — 字級對齊 (已完成)

把 `cjkGlyphImage` 改為 size-aware:依呼叫端字型的 `GlyphHeight` rasterize、cell 框在字高內、
以 (rune, height) 為快取 key、各字高一份 face。三條繪字路徑都傳入字高。

![字級對齊後](img/phase2-cjk-aligned.png)

對比 `phase1-engine-cjk.png`:中文不再溢出行高、與英文基線一致、各行不再重疊。
trade-off:最小字型 (如 ResourceFont) 下 CJK 偏小偏粗;MoM 最小字多放數字 (ASCII),中文少用,
密集面板若需更大可讀中文,日後走 hi-res canvas (拉高內部畫布) 或路線 A 固定尺寸點陣。

## 待辦 (Phase 2 接續)

- [x] **多字級對齊**:CJK 尺寸跟著各引擎字型高度走,修基線/行高重疊。
- [ ] 把 CJK 字型納入打包 (自由授權 Noto / 或自製點陣),而非依賴系統字型路徑。
- [ ] 字串覆蓋層 (Phase 2):讓真實遊戲畫面的英文字串顯示為中文 (本測試是直接餵中文字串驗證渲染)。
- [ ] (選項 A) 若要點陣美術一致,離線烘 24×24 atlas;B 已足以推進翻譯與版面。
