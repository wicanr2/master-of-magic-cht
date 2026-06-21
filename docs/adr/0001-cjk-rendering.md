# ADR 0001 — CJK 文字渲染路線

- 狀態:**Accepted (路線 B 已驗證,作為先行渲染管線);路線 A 點陣保留為美術升級選項**
- 日期:2026-06-21
- 相關規則:`81-retro-cjk-hires-canvas`
- 驗證:見 [`docs/phase1-cjk-prototype.md`](../phase1-cjk-prototype.md)、patch `patches/0001-cjk-font-injection.patch`

## 背景

kazzmir/master-of-magic 引擎 (Go + Ebiten) 的字型系統 (`lib/font/font.go`) 以 `int(c)-32` 索引
glyph,只支援 ASCII 32–127;非 ASCII rune 在繪製迴圈被靜默丟棄。遊戲內部畫布 320×200,Ebiten
以 `ScaleAmount=3.0` 放大顯示。需求:讓中文 (繁體) 清晰可讀,且不破壞 pixel-art 美術。

字串迴圈本就走 `for _, c := range text` (rune),因此 CJK 注入點明確:在 glyph 查找處加分流。
爭點在「中文字形怎麼來」。

## 候選

### 路線 A — 24×24 點陣 atlas (烘製)
- `build_cjk_font.py` (docker uv venv) 從自由授權 TTF (AR PL UMing TW) 子集烘出 24×24 點陣。
- `font.go` glyph 查找:碼點 ≥ 0x80 → 查 atlas。
- 優點:與原版點陣風格一致、可控、沿用銀河霸主既有 pipeline。
- 缺點:需烘字、字集子集管理;尺寸固定 (需另烘 16/14 應付密集面板)。

### 路線 B — TTF 即時 rasterize
- 用引擎已有的 `golang.org/x/image/font` / `go-text/typesetting` 在顯示解析度直接畫 TTF 中文。
- 優點:開發快、任意字皆可、字清晰。
- 缺點:與 pixel-art 底圖風格略不一致;需處理與既有 glyph cache 的整合。

## 決策

**先採路線 B (TTF 即時渲染) 為主線渲染管線**,讓翻譯與版面工作立即可推進;
**路線 A (24×24 點陣) 降為後續的美術一致性升級選項**,非阻塞項。

理由 (Phase 1 prototype 後更新):
- B 已實作進引擎 `lib/font` (注入 `doPrint` / `MeasureTextWidth` + 新增 `cjk.go`) 並編譯通過、headless 截圖驗證 (見 phase1 文件)。注入點小、風險低。
- A 需離線烘字管線;且實測發現 AR PL UMing 的 `.ttc` 是 CFF/舊式,`golang.org/x/image/sfnt` 解析失敗 —— 引擎內即時 rasterize 走 Noto Sans CJK TC,點陣 UMing 若要用得另走 freetype 烘字。先不擋翻譯進度。
- 破版觀察待「套真實版權 lbx 跑一畫面」後回填。

## 後果

- `lib/font/font.go` / `read.go` 需改 (glyph 查找分流、`MeasureTextWidth` CJK 寬度、`splitText` 逐字斷行)。
- 可能需要對 UI widget 座標做 `mapX/mapY` 比例映射 (Phase 2,實測後定)。
- 不論 A/B,版權 LBX 不動;CJK 字形是自製 (點陣) 或自由授權 (TTF 子集) 資產。
