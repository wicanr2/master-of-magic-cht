# Worklist — 中文化待修清單 (玩家實測回報)

實機遊玩時發現的未翻譯/問題,逐項追蹤。

## 已加譯 (待下次 AppImage 重建生效)

- [x] #1 資訊顧問選單:選擇顧問/見習巫師/史官/占星師/大臣/稅務官/大維齊爾/魔鏡 已加入 ui.tsv (2026-06-22)

## 待修

### 1. 資訊 (Info) → 選擇顧問選單 未翻 [font, 可加 ui.tsv]
畫面:進入遊戲 → 頂部「資訊」→ 顧問選單。
- 標題「Select An Advisor」未翻。
- F3-F9 顧問名未翻:Apprentice / Historian / Astrologer / Chancellor / Tax Collector / Grand Vizier / Mirror。
- (F1 地形勘查 / F2 製圖師 已翻)
- 來源:`game/magic/game/game.go` 顧問清單 `Name:` + `MakeSelectionUI(..., "Select An Advisor", ...)`。font 路徑,加 ui.tsv 即可。
- 回報:2026-06-22 螢幕快照 06-52-12。
