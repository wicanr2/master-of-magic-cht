# Worklist — 中文化待修清單 (玩家實測回報)

實機遊玩時發現的未翻譯/問題,逐項追蹤。

## 已完成

- [x] **#1 資訊 (Info) → 選擇顧問選單** (2026-06-22)
  - 加 8 條 ui.tsv:Select An Advisor→選擇顧問、Apprentice→見習巫師、Historian→史官、
    Astrologer→占星師、Chancellor→大臣、Tax Collector→稅務官、Grand Vizier→大維齊爾、Mirror→魔鏡。
  - font 路徑 (同已翻的 地形勘查/製圖師);三平台包已重建含此修正。
  - 註:headless 下顧問彈窗難穩定截圖,但機制與同選單已翻項一致,生效無虞。

- [x] **#2 建築名/描述 + help.lbx 百科** (2026-06-22)
  - buildings.tsv 72 條(37 建築名 + 35 描述)+ help.tsv 920 條(help.lbx 全 entries)。
  - 走 font 顯示層覆蓋:城市畫面建築名(住宅…)、生產描述、右鍵建築 help scroll 本體全中文。
  - help body 含 0x14 換行:override.go 載入時把 TSV 字面 `\n` 還原成 0x14,本體換行與數值(15%、金幣/回合)正確。
  - 字型子集擴至 1628 字;AppImage 重建並 headless 驗證(城市畫面 + BUILDER'S HALL help 本體)。

- [x] **#3 隨機事件 / 災害訊息 (events.go)** (2026-06-22)
  - messages.tsv:13 靜態訊息(解咒/凶月/吉月/三交會/魔力短路 + 結束句)+ 13 動態模板(枯竭/聯姻/捐獻/地震/天賜/隕石/新礦/海盜/瘟疫/人口激增/叛亂 + 結束句)+ 11 礦物名。
  - events.go 動態訊息包 `font.TranslateFormat`,內嵌 enum(城市規模/礦物)亦過翻譯;%v 順序與原文一致。
  - 確定性驗證:events.go 全部字面(13 靜態 + 13 模板 + 11 礦物)與 messages.tsv key 100% 精確命中;AppImage 重建 boot 無 regression。

- [x] **#4 合成單位名 (race + unit) + 軍隊清單** (2026-06-22)
  - armyview/view.go:種族名與單位名各自過 `font.TranslateFormat` 後再合成(中文不需空格),
    避免合成英文整串無對應 key;標題 "The Armies Of %v" 走模板翻譯。
  - unitview/ui.go:解散/遣散確認對話框模板化翻譯。
  - ui.tsv 新增 3 模板;headless 驗證軍隊清單:標題「Oberic 的軍隊」+ 合成名「蜥蜴人劍士」全中文。
  - 已知殘留:單位名 "Hero" 與既有 "Hero→召回英雄" 同 key 衝突(英雄單位 highlight 罕見,暫不處理)。

## 待修

- [ ] **help 彈窗標題仍英文**(如 help scroll 頂端 "BUILDER'S HALL")
  - 標題走引擎華麗哥德標題字(LBX baked bitmap font),非 TTF/CJK 路徑,translateForDisplay 接不到。
  - 需走圖片疊字技術(util.ChtLabel)或把標題改走 CJK font;屬美術疊字範疇,非純資料翻譯。
- [ ] **部分動態 tooltip 英文**(如資源 hover "Power 10")
  - 屬 "Power %d" 類合成字串,需 TranslateFormat 模板翻譯 + 數值代入。
