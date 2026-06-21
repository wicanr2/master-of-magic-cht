# CONTEXT — 工作魔法大帝 (Master of Magic) 繁中化 Ubiquitous Language

本檔為專案 domain glossary。命名變數、寫文件、討論時優先使用以下術語;遇新概念先登錄再用。
格式:`Term — definition. _Avoid_: forbidden synonyms`。

## 專案

- **MoM / 工作魔法大帝** — Master of Magic (MicroProse, 1994),本專案中文化標的。奇幻 4X,巫師爭奪 Arcanus / Myrror 雙位面霸權。
- **kazzmir 引擎** — [kazzmir/master-of-magic](https://github.com/kazzmir/master-of-magic),Go + Ebiten 重製引擎,直接讀原版 `.lbx` 執行。 _Avoid_: 「模擬器」(它是重製,非 emulator)。
- **patch-only repo** — `master-of-magic-cht` 只放譯文表 + 字型 + patch + 文件 + 腳本,**不 vendor 引擎本體、不散布版權遊戲檔**。
- **CHT / 繁中化** — 繁體中文化。範圍見 PLAN.md。

## 資料格式

- **LBX** — MoM 的資源封裝格式 (`.lbx`),內含文字、點陣圖、字型、音樂。中文化主戰場之一。
- **itemdata.lbx** — 250 筆**預設神器名** (Black Asp、Sword of Mallana…)。
- **itempow.lbx** — 64 筆**物品能力 / 附魔** (+N Attack、Flaming、Vampiric…),神器與自製物品的詞綴。
- **itemisc.lbx / items.lbx** — 物品 sprite 圖,**不翻**。
- **fonts.lbx** — 8 種尺寸點陣字型,每尺寸 96 個 ASCII glyph (32–127)。CJK 渲染改此路徑或新增 atlas。
- **字串來源 (string source)** — 分兩處:(a) LBX 內字串、(b) hardcode 在 Go source (尤其 `units/unit.go` 單位名)。翻譯需同時覆蓋。 _Avoid_: 假設「文字全在 LBX」。

## 中文化技術

- **CJK glyph 分流** — 在 `lib/font/font.go` glyph 查找處,碼點 < 0x80 走原 ASCII glyph,≥ 0x80 改用 CJK 字形來源。引擎字串迴圈本就走 rune,是低風險注入點。
- **CJK 點陣字 (路線 A)** — 24×24 點陣中文,`build_cjk_font.py` 從自由授權 TTF 子集烘出。維持 pixel-art 風格一致。
- **TTF 即時渲染 (路線 B)** — 利用引擎已有 `golang.org/x/image/font` 直接畫 TTF。快但風格略不一致。A/B 決策見 ADR 0001。
- **覆蓋層 (override layer)** — 引擎於 LBX 載入後用本專案譯文表即時替換字串的機制,避免散布改過的版權 LBX。
- **hi-res canvas / pixel scaling** — 不縮小中文塞低解析;拉高有效解析度、底圖 nearest-neighbor 放大,CJK 字畫在放大層。見全域規則 `81-retro-cjk-hires-canvas`。
- **破版 (layout overflow)** — UI widget 座標需 `mapX()/mapY()` 比例映射,否則溢出/錯位。
- **譯文表 (string table)** — 原文 → 譯文對照,放 `docs/strings/*.tsv`,翻譯與注入的單一真實來源。

## 引擎子系統 (engine 內路徑)

- **lib/font** — 字型載入與繪製 (`font.go` / `read.go`),CJK 渲染入口。
- **lib/lbx** — LBX 讀取 (`lbx.go`),`readStringsSection()` 解析字串區。
- **game/magic** — 主程式與遊戲邏輯;`data/data.go` 含 320×200 畫布常數,`scale/scale.go` 含 `ScaleAmount`。
- **util/** — `fontviewer`、`fonteditor`、`lbxdump`、`lbxviewer`、`make-lbx`,觀察/編輯字型與 LBX 的現成工具。

## 物品詞彙 (item domain)

- **神器 (artifact)** — 玩家用 create artifact 法術製作、或地城掉落的具名強力物品。250 筆預設名在 itemdata.lbx。 _Avoid_: 「文物」。
- **物品能力 (item power)** — 附在物品上的數值加成或特殊效果 (+N 攻擊、烈焰、吸血…),64 筆在 itempow.lbx。
- **附魔 (enchantment)** — 此處作 item power 的同義語境用;城市/單位上的持續法術另稱「結界」。

## Flagged ambiguities (待釐清)

- 文字渲染路線 A (點陣) vs B (TTF) — ADR 0001 prototype 後定案。
- CJK logical 字高與原版 UI 框整合方式 (拉高內部畫布 vs 獨立高解析文字層) — Phase 1 實測後定。
- override layer 注入點:在 `lib/lbx` 載入處 vs 在各 `game/` 取字串處 — Phase 2 定。
