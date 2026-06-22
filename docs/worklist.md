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

- [x] **#5 help 彈窗標題 + 資源 tooltip + Hero 衝突** (2026-06-22)
  - **help 標題**:help.lbx 的 headline 全大寫(HOUSING/BUILDER'S HALL),原 TSV key 是混合大小寫故配不上。
    正規化(小寫去標點)自動回填 346 條,4 個 agent 新譯 410 條 → headlines.tsv 753 條。
    HelpTitleFont 走 translateForDisplay 且能渲染 CJK,headless 確認 help scroll 標題「建築選項」中文。
  - **資源 tooltip**:city-screen.go 6 個 `Sprintf("X %v")` 包 TranslateFormat;headless 確認「金幣盈餘 2/食物盈餘 4」。
    資源對話框標題在繪製點翻譯;`Power` 與既有「力量戒指」衝突 → 該 call site 改獨立 key `Magic Power`→魔力。
  - **Hero 衝突**:armyview 對英雄略過種族前綴(避開 `Hero`→「召回英雄」同 key 衝突)。
  - 字型子集 1657 字,AppImage 重建驗證。

- [x] **#6 動態數值模板 / 組合句(模板大隊)** (2026-06-22)
  - 7 個 agent 分檔掃出玩家面 `fmt.Sprintf` 模板,包 `font.TranslateFormat`,共 **162 個 call site**
    (game.go 25、magicview/artifact/abilities/new-wizard 48、surveyor/treasure/vault/hero 24、cast 18、
    cityview 16、combat 16、misc[hero/unitview/spellbook/settings/summon] 15)。
  - templates.tsv **139 條**:建造完成/拾獲/解散/城市放棄/前哨/勘查/寶物/法術回饋(無法對此城市施放/施法失敗)/
    戰鬥結算/人口/農民工人叛民 tooltip/城市清單欄位等。佔位符數量英中 100% 一致。
  - treasure/model/item/abilities/hero/cast-select-wizard 補 font import;`font.Translate`(不存在)修為 `TranslateFormat`。
  - `%v of %s` 城市標題的 size 引數補翻;headless 驗證:「村落 的 Xanten」「Lo Pan 的城市」「獸人 農民 4」。
  - 全引擎 build 通過(45 檔變更),字型子集 1659 字,AppImage 重建驗證。

- [x] **#7 專有名詞 中文(英文) + A 顯示缺口 + B 位圖按鈕** (2026-06-22)
  - **專有名詞**:names.tsv 49 條(巫師 14 + 英雄 35)走「中文(英文)」格式,原文與中文並存;
    直接列印處(mirror/overworld)自動生效,Sprintf 引數處(軍隊/城市標題、diplomacy、hero 全名)補 wrap。
    headless 驗證「羅潘(Lo Pan) 的軍隊」。英雄名 names.tsv 因字母序先載入勝過 units.tsv 舊純中文。
  - **A 顯示缺口**:misc-ui.tsv 28 條(節點/結界 enum、神器鍛造/巫師建立/戰鬥/復活對話、施法目標提示);
    undead/road 2 模板補 wrap。56 個 enum(Bless/Haste…)早已被 spells.tsv 覆蓋。
  - **B 位圖按鈕**:armyview Items/Ok → 疊「物品」「確定」(util.ChtLabel);headless 驗證。
  - 字型子集 1670 字,全引擎 build 通過(47 檔),AppImage 重建驗證。

- [x] **#8 城市名 中文(英文)** (2026-06-22)
  - cityname.lbx 280 名(德/英/古典/奇幻)→ citynames.tsv 279 條(去 Bremen 重複),真實地名通用譯、奇幻名意譯/音譯。
  - cities list 直接列印自動生效;城市標題 `%v of %s`、Outpost、事件訊息的 city.Name 引數補 wrap。
  - headless 驗證:城市清單「烏爾納(Ulna)」、標題「荷魯斯(Horus) 的城市」。
  - **[雷] 新增 TSV 檔需清 GOCACHE**:go:embed 增量 build 不重嵌新檔名 → 清 cache 才生效。

## 待修

- [ ] **英文小字** — 目前「中文(英文)」同字級;英文縮小需逐顯示點改雙重渲染,屬美術強化。
- [ ] **零星 enum 引數/純字串** — 次要畫面若實測見英文再個案補(主要畫面已涵蓋)。
- [ ] **Windows/macOS 包同步** — dist 的 Win/Mac 仍停在數輪前。
