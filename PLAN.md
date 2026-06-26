# PLAN — 工作魔法大帝 繁體中文化 (Master of Magic CHT)

> 把 1994 年的奇幻 4X 經典《Master of Magic》(MicroProse) 做成**全程繁體中文**可玩,
> 跑在 Go 重製引擎 [kazzmir/master-of-magic](https://github.com/kazzmir/master-of-magic) 上。
> 本 repo 為 **patch-only / 譯文資產 repo**:不 vendor 引擎本體、不散布任何版權遊戲檔。

最後更新:2026-06-25。本檔是專案的單一真實計畫來源 (single source of truth),隨進度回填。

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
- [x] `font.go` 三條繪字路徑 (`doPrint` / `PrintOutline` / `MeasureTextWidth`) 加 CJK 分流;新增 `lib/font/cjk.go` (patch `patches/0001-cjk-font-injection.patch`)
- [x] 套真實 lbx 在**真實引擎字型管線**跑一畫面 (`test/cjk-render`),CJK 三路徑全渲染 (`docs/img/phase1-engine-cjk.png`);發現並修正 `PrintOutline` 漏 patch
- [x] 觀察破版:CJK 字級未對齊引擎字型高度 → 行距重疊 (Phase 2 首要精修)
- [x] (Phase 2) 多字級對齊原版字高
- [x] **CJK 字型 + 譯文表 go:embed 內嵌**:`assets/cht-subset.otf` (126KB 子集,fonttools) + `cht_strings/*.tsv` 內建進引擎,standalone 出包不依賴 env/系統字型 (env 仍可覆寫);`scripts/prepare-embed.sh` 組裝。已驗證無 env 純內嵌渲染中文
- [ ] (選項 A) `build_cjk_font.py` 烘 24×24 atlas、`util/fontviewer` 擴充顯示 CJK

### Phase 2 — 字級對齊與字串翻譯注入
- [x] **CJK 字級對齊字高**:`cjkGlyphImage(rune, height)` 依呼叫端字型 `GlyphHeight` 渲染、以 (rune,height) 為快取 key、多 face 快取;三路徑傳入字高。修掉行距重疊破版 (對比 `docs/img/phase1-engine-cjk.png` → `docs/img/phase2-cjk-aligned.png`)。trade-off:最小字型下 CJK 偏小,密集面板日後可走 hi-res canvas / 點陣固定尺寸。
- [x] **字串覆蓋層 (顯示層)** — 第一個 slice 完成:在 `lib/font` 三繪字點翻譯,不碰邏輯字串 (對照 `create-artifact.go:198` 英文當 key,決定走顯示層,見 [ADR 0003](docs/adr/0003-string-override-layer.md));TSV 從 `MOM_CHT_STRINGS` 載入、TrimSpace 精確比對。真實引擎驗證英文 power 名自動轉中文 (`docs/img/phase2-override.png`)
- [x] **版本對齊** (見 [ADR 0002](docs/adr/0002-target-game-version.md)):已升 1.60 並 diff,名稱類 0 差異
- [x] **散文類覆蓋 + CJK 逐字斷行**:`CreateWrappedText` 進入點整段翻譯 + 重寫 `splitText` 支援中文逐字斷行。基於 **CP 1.60** 資料驗證長中文描述自動換行 (`docs/img/phase2-prose-wrap-1.60.png`)
- [ ] 翻譯表逐類完成:item powers ✅ → 神器名 ✅ (250 筆,`artifacts.tsv`,真實引擎驗證) → 法術 → 建築 → help → 單位名
- [ ] 單位名 hardcode (`units/unit.go`) 改查表

### Phase 3 — 版面與打包(已完成主體)
- [x] 破版修正、密集面板字級對齊
- [x] 跨平台打包:Linux AppImage / Windows(CGO=0 免 DLL)/ macOS(GitHub Actions arm64)三平台皆已建,
      公開 GitHub Release v0.1 提供 data-free 包(玩家自備遊戲檔)。
- [x] README 升級成「給玩家的信」

### Phase 4 — 重現經典 + 設定補完(2026-06,已完成主體)
專案目標從「純中文化」擴展為「**也修對遊戲、補回原版設定**」。詳見 [`docs/classic-rules-plan.md`](docs/classic-rules-plan.md)。
- [x] 玩家回報 10 個 issue:6 修(水上秒敗 / 神器成本存讀檔 / 海上尋路 / 創角被踢 / 召喚第七英雄 / 裝備傳送)
      + 4 證偽(英雄復活 / 三色書他系法術,以資料層 + 手冊 oracle 佐證)。每項 `cht_*_test.go`。
- [x] 原版 Settings 18 項補回 16(`data.Settings` 全域 runtime + 雙欄設定畫面),含「自動建議」
      (規則由 5+ 份攻略歸納,見 [`docs/mom-strategy-notes.md`](docs/mom-strategy-notes.md))。跳過 Event Music、Expanding Help。
- [x] 閃退記錄(crash.log + panic 攔截)。
- [x] 紀律:修正用旗標守護、預設維持原版行為;單元測試 + headless 截圖雙驗。
- [ ] (可選)原版觀感:長寬比校正 / CRT shader,分析見 [`docs/dos-vs-remake-ui.md`](docs/dos-vs-remake-ui.md)。

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

> **版本注意**:現有 `*.tsv` 由 **vanilla v1.31** 資料萃取,作為 baseline。主目標為
> **Community Patch v1.60**(對齊引擎,見 [ADR 0002](docs/adr/0002-target-game-version.md));
> 表以英文原文為 key,可同時相容兩版,取得 1.60 資料後補 delta。

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
