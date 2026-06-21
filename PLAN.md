# PLAN — 工作魔法大帝 繁體中文化 (Master of Magic CHT)

> 把 1994 年的奇幻 4X 經典《Master of Magic》(MicroProse) 做成**全程繁體中文**可玩,
> 跑在 Go 重製引擎 [kazzmir/master-of-magic](https://github.com/kazzmir/master-of-magic) 上。
> 本 repo 為 **patch-only / 譯文資產 repo**:不 vendor 引擎本體、不散布任何版權遊戲檔。

最後更新:2026-06-21。本檔是專案的單一真實計畫來源 (single source of truth),隨進度回填。

---

## 1. 為什麼這次跟銀河霸主不一樣

先前《銀河霸主》(MOO1) 中文化跑在 **1oom (C 引擎)**,文字渲染要自己改 `lbxfont.c`、自建 LBX 覆蓋層。
這次的引擎是 **Go + Ebiten** 重製版,差異直接決定技術路線:

| 面向 | 1oom (MOO1) | kazzmir/master-of-magic (本專案) |
|---|---|---|
| 語言 | C (GPLv2) | Go 1.25 + Ebiten v2.9 |
| 字型渲染 | `lib/lbxfont` 點陣,改 C source | `lib/font/font.go`,glyph 索引 `int(c)-32`,只認 ASCII 32–127 |
| 文字迭代 | byte 迴圈 | `for _, c := range text` 已是 **rune (UTF-8) 迭代**,但非 ASCII 被靜默丟棄 |
| 高解析渲染 | 需自建 2× 合成層 | Ebiten 已做顯示縮放 (`ScaleAmount=3.0`),`font.Print` 已吃 scale 乘數 |
| 字型工具 | 無 | 內建 `util/fontviewer` / `fonteditor` / `lbxdump` / `lbxviewer` / `make-lbx` |
| TTF 能力 | 無 | `golang.org/x/image/font` + `go-text/typesetting` 已是依賴 |

**結論**:Go 引擎讓 CJK 渲染門檻顯著降低 — 字串迴圈本來就走 rune,我們只要在 glyph 查找處加一條
「碼點 ≥ 0x80 → 改用 CJK 字形來源」的支線,不必像 1oom 那樣重寫整條點陣管線。

---

## 2. 引擎架構盤點 (中文化要動的點)

| 子系統 | 檔案 (engine 內路徑) | 角色 | 中文化動作 |
|---|---|---|---|
| 字型載入/繪製 | `lib/font/font.go` | glyph 查找 `int(c)-32`、`doPrint()`、`Print()`、`MeasureTextWidth()` | **P0** 加 CJK glyph 分流 + 寬度量測 |
| 字型讀取 | `lib/font/read.go` | `GlyphForRune()`、`Glyph{Data,Width,Height}` | **P0** 擴充 Unicode glyph 來源 |
| LBX 讀取 | `lib/lbx/lbx.go` | `readStringsSection()` null-terminated ASCII | **P1** 確認可吃 UTF-8 位元組 |
| 解析度/縮放 | `game/magic/data/data.go` (320×200)、`game/magic/scale/scale.go` (`ScaleAmount`) | 內部畫布常數 | **P2** 視破版情況決定是否拉高 |
| 換行 | `lib/font/font.go` `splitText()` | 以空白斷字 | **P2** CJK 無空白,需逐字斷行 |
| 單位名 | `game/magic/units/unit.go` | hardcode 英文字串 (~100+) | **P1** 改為查表 / 覆蓋 |

### 字串來源分佈 (翻譯戰場)

| 類別 | 來源 | 出處 |
|---|---|---|
| 法術名/描述 | LBX (`spelldat.lbx` / `desc.lbx`) | 引擎已從 LBX 讀,無需改碼 |
| 物品/神器名 | LBX (`itemdata.lbx`,250 筆) | `artifact/item.go` |
| 物品能力/附魔 | LBX (`itempow.lbx`,64 筆) | 同上 |
| Help 文字 | LBX (`help.lbx`) | `help/help.go` |
| 建築描述 | LBX (`buildesc.lbx`) | `building/description.go` |
| 城市名/英雄名 | LBX (`cityname.lbx` / `names.lbx`) | |
| **單位名** | **hardcode Go source** | `units/unit.go` — 需改碼 |

---

## 3. 文字渲染路線 (待 ADR 0001 定案)

兩條候選,皆遵守全域規則 `81-retro-cjk-hires-canvas`(不縮小中文硬塞低解析):

- **路線 A — 烘 24×24 點陣 atlas**:用 `build_cjk_font.py` (docker uv venv) 從 TTF (AR PL UMing TW) 子集烘出
  CJK 點陣,在 `font.go` glyph 查找加分流。風格與原版點陣一致,可控。沿用銀河霸主 pipeline。
- **路線 B — 直接走 TTF 即時 rasterize**:利用引擎已有的 `golang.org/x/image/font` / `go-text`,
  在顯示解析度直接畫 TTF 中文。開發快、字清晰,但與 pixel-art 底圖風格略不一致。

**初步傾向 A**(維持點陣美術一致性 + 與 MOO 經驗共用工具),B 作為快速 prototype 驗證渲染管線。
最終決策寫入 `docs/adr/0001-cjk-rendering.md`。

關鍵注意:Ebiten 顯示已 3× 放大,logical 8px 的中文不可讀 → 必須讓 CJK 字形以**較大 logical 尺寸**
(或獨立高解析層) 繪製;UI widget 座標可能需 `mapX/mapY` 比例映射避免破版 (P2,實測後定)。

---

## 4. 階段規劃

### Phase 0 — 盤點與骨架 (本 session,進行中)
- [x] 解壓 DOS 原版,盤點 LBX 資產 (142 檔)
- [x] Clone 引擎,盤點字型/LBX/渲染架構
- [x] 確認文字渲染路線候選 (A/B)
- [x] 建 repo 骨架 + CONTEXT.md + PLAN.md + README + `.gitignore`
- [x] 抽出物品字串,翻譯第一批 (item powers 64 筆)
- [ ] `scripts/fetch-engine.sh` 固定引擎 commit

### Phase 1 — 渲染管線 (CJK glyph) — 核心已完成
- [x] ADR 0001 定案:採路線 B (TTF) 為主線,A 降為美術升級選項
- [x] prototype 驗證「一行中文上畫面」(`prototype/cjk-hello/`,docker+Xvfb 截圖 `docs/img/phase1-cjk-hello.png`)
- [x] `font.go` `doPrint` / `MeasureTextWidth` 加 CJK 分流 + 寬度量測;新增 `lib/font/cjk.go` (patch `patches/0001-cjk-font-injection.patch`,`go build ./lib/font` 通過)
- [ ] 套真實版權 lbx 跑一畫面,觀察破版,回填 ADR
- [ ] 多字級 (16/14) 與原版字高對齊;CJK 字型納入打包
- [ ] (選項 A) `build_cjk_font.py` 烘 24×24 atlas、`util/fontviewer` 擴充顯示 CJK

### Phase 2 — 字串翻譯與注入
- [ ] LBX 字串覆蓋機制 (載入後 override,**不改版權 LBX**)
- [ ] 翻譯表逐類完成:item powers → 神器名 → 法術 → 建築 → help → 單位名
- [ ] 單位名 hardcode (`units/unit.go`) 改查表

### Phase 3 — 版面與打包
- [ ] 破版修正 (`mapX/mapY`、密集面板自動縮字)
- [ ] 跨平台打包,順序見 [`docs/porting-difficulty.md`](docs/porting-difficulty.md):Linux/Windows (🟢) → Web WASM (🟢,上游已通) → macOS (🟡,GitHub Actions runner) → Android (🟠,觸控 UX)
- [ ] README 升級成「給玩家的信」(套 rule `80-retro-cht-readme-polish`)

---

## 5. 翻譯資產 (docs/strings/)

| 檔案 | 內容 | 狀態 |
|---|---|---|
| `item-powers.tsv` | 64 物品能力/附魔 (`itempow.lbx`) | ✅ 第一批已譯 |
| `artifacts.tsv` | 250 預設神器名 (`itemdata.lbx`) | 🔲 scaffold,待譯 |
| (待建) `spells.tsv` | 法術名/描述 | 🔲 |
| (待建) `buildings.tsv` | 建築名/描述 | 🔲 |
| (待建) `units.tsv` | 單位名 | 🔲 |

格式:TSV，欄位 `source(原文，含前導空白以利精確比對)` / `zh(繁中譯文)` / `note`。
譯文表是本專案的**衍生資產**,版權 LBX 分毫未動。

---

## 6. 版權與散布原則 [HARD]

- **不 vendor 引擎**:`scripts/fetch-engine.sh` 取得 kazzmir/master-of-magic 後套 patch。
- **不散布版權檔**:`original_game/`、`extracted/`、引擎 clone 全列入 `.gitignore`,絕不入 git。
- 本 repo 只放:PLAN/文件、英文→繁中譯文表、patch、字型烘製腳本、CJK 字型 (自製點陣或自由授權 TTF 子集)。
- 玩家自備正版 Master of Magic 資料檔,套用本專案即可。

---

## 7. 參考

- 引擎:<https://github.com/kazzmir/master-of-magic>
- 銀河霸主中文化前例:`~/master-of-orion` (1oom,patch-only,24×24 CJK 管線)
- 全域規則:`81-retro-cjk-hires-canvas` (拉高畫布不縮字)、`80-retro-cht-readme-polish` (README 雜誌風)、`50-ubiquitous-language` (CONTEXT.md)
