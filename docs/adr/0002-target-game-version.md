# ADR 0002 — 目標遊戲版本與 Community Patch

- 狀態:**Proposed**(建議主目標 CP 1.60,保留 1.31 相容)
- 日期:2026-06-21
- 相關:[ADR 0001](0001-cjk-rendering.md)、`docs/strings/*.tsv`

## 背景

Master of Magic 有多個版本:官方最終 DOS 版 **v1.31**(Simtex/MicroProse, 1995),社群維護的
非官方/Community Patch **v1.40n → v1.60**(目前社群現行版,由 drake178 等維護)。

兩項已查證事實決定本決策:

1. **kazzmir 引擎朝 Community Patch 開發。** 遊戲邏輯實作 1.50/1.60 行為:
   - `game/magic/player/relations.go:11` → `// 1.50 patch formula`
   - `game/magic/city/city_test.go:705,789` → `// Test against values from a city screen of original MoM v1.60`
   引擎 README 不硬性指定版本,只要求「放入 MoM 的 lbx」,即**讀玩家現有資料**。
2. **本專案目前手上的資料是 vanilla v1.31。** `extracted/magic.exe` 內含
   `Copyright Simtex Software, 1995 V1.31`,且 `help.lbx` 無 CP 標誌的第 808 筆文件 entry。

差距:引擎邏輯=1.60,我們萃取字串的資料=1.31。字串覆蓋層需**精確比對原文**,版本不一致 → 漏譯/不命中。

## 升級路徑(已知工具)

`original_game/community_patch.txt` = **MOMDIFFP** 差分補丁器的說明。它能把正版 MoM 在
1.20/1.30/1.31/1.40n/1.50/1.51/1.52.03/**1.60** 間互轉,不需 Slitherine launcher;
也能降級到 1.00/1.01/1.10。符合本專案「玩家自備正版資料」模型,可納入玩家安裝說明。
注意:該套件**不含** changelog / 字串 / FILESET patch(明示請查 masterofmagic.fandom.com wiki);
且本 repo 目前只有說明 `.txt`,未有 `MOMDIFFP.EXE` 本體——要產 1.60 資料需另取 patcher 並於 DOSBox/Windows 執行。

## 決策

1. **主目標 = Community Patch v1.60**(對齊引擎邏輯與測試基準),於玩家文件建議升級到 1.60。
2. **保留 v1.31 相容**,做法:**譯文表以「英文原文字串」為 key**。
   - 1.31 與 1.60 相同的字串:一條譯文同時覆蓋兩版。
   - 1.60 改過/新增的字串:加一條變體,對應同一中文。
   - → 單一份 `docs/strings/*.tsv` 支援兩個版本,且現有 1.31 萃取不浪費。
3. **現有 1.31 萃取為 baseline**(`item-powers.tsv` 64、`artifacts.tsv` 250);取得 1.60 資料後
   做 string diff,補上 delta(改名/新增的 item power、法術、help 等)。

## 待辦 / 風險

- [ ] 取得實際 1.60 LBX(跑 MOMDIFFP 或他法),重新萃取 item/spell/help 字串,對 1.31 做 diff。
- [ ] 量化 1.31↔1.60 字串差異規模(目前未知,需實測;勿假設「差異很小」)。
- [ ] 引擎 commit 已釘選(`0c7669b`);CP 版本同樣釘選 1.60,避免雙邊漂移。
- 風險:CP 與引擎皆持續演進;若引擎日後改變對某 LBX 結構的假設,需重驗。
