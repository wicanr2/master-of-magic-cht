# 重現經典:規則切換 + bug 修正計畫(0200 patch)

> 目標:讓《魔法大帝》不只中文化,而是**玩起來就是老玩家記憶中的經典**。
> 手段:在「新遊戲」設定加 **規則切換**,玩家可選「重製原版」或「經典強化」,並排比較差異。

## 架構(已驗證可行)

- 旗標:`NewGameSettings.RuleSet`(enum:`RuleSetRemake` / `RuleSetClassic`)。
- 流向:新遊戲設定 UI 設定 → `MakeGame(...settings)` 存進 `Game` → 召喚/施法直接讀 `game.RuleSet`;
  combat 由 game 建立,把旗標傳進 `CombatModel`。
- 每個改動寫成 `if classic { 新行為 } else { 原行為 }` → **重製路徑零改動、零風險、可比較**。
- 獨立 patch:`patches/0200-classic-rules.patch`(疊在 `0099` 中文化之上,可獨立維護 / rebase 上游)。

## 紀律(從 retro-game-remake + retro-game-playtest)

1. **Oracle = vanilla 1.31 + 手冊**:每條「經典規則」先查原版實際怎麼跑(DOSBox / `original_game/` / 手冊),
   不憑記憶。分不清是 bug 還是 CP1.60 故意改時,以 1.31 為「經典」基準。
2. **驗正常玩家路徑**:玩法改動**不能只看編譯綠 / 截圖**——用 game_tester 走玩家真會走的路徑,
   且**兩個模式都驗**(經典模式有修、重製模式維持原樣)。
3. **逐項 worklist**:一次一項,根因 → 旗標保護的修正 → 雙模式 game_tester → commit。

## Worklist(分三層 + 基礎建設)

### T0 基礎建設(已完成,headless 驗證)
- [x] `data.RuleSet` enum(Remake/Classic + String())+ `NewGameSettings.RuleSet` 欄位(隨存檔序列化)
- [x] **主選單**左上角「規則:重製原版 ↔ 經典強化」切換鈕(點擊循環);seed 進新局 settings → `model.Settings.RuleSet`
- [x] `-skipintro` flag(片頭 10+ 場景太長,測試 + 玩家 QoL 都用);headless 驗證切換鈕渲染 + 點擊切換
- [ ] `CombatModel` 收 `RuleSet`(combat 建立點帶入)— 待第一個戰鬥相關修正(#1)時做
- [ ] game_tester harness(逐步補,雙模式)

> 模組化 = **執行期切換**(主選單 toggle),非分離 patch 檔。classic-rules 改動全用
> `if RuleSet==Classic { 新 } else { 原 }` 包住 + 註解標記,收進權威 `0099` patch;
> RuleSetRemake 為預設,重製路徑零改動。

### T1 明確 bug(經典模式修正;重製模式維持原狀供對照)
- [ ] #5 創角名字輸入打一半被踢回(`setup/new-wizard.go` key handler;最高日常影響)
- [ ] #1 走進怪物點直接判敗、無法開打(戰鬥初始化/自動結算)
- [ ] #9 召喚第七英雄靜默失效(缺上限警告 + 靜默吞掉)

### T2 經典規則未強制(經典模式新增)
- [ ] #6/#8 Life/Death 書系互斥 + 法術依書色 gate(召喚英雄、戰鬥法術可用性)
- [ ] #2 製造神器/施法成本與法力強制(分回合扣、不足不能施)

### T3 需先對照 1.31 oracle 再決策(可能 bug 也可能 CP1.60 故意改)
- [ ] #3 英雄戰死、我方獲勝後是否該復活(查原版規則)
- [ ] #4 船斜向/長程海上尋路錯誤聲(pathfinding)
- [ ] #10 打怪英雄當場穿裝備是否免 20 傳送法力(查原版 Enchant Item move 規則)
- [ ] #7 法術書顏色排序選項 + 設定畫面補項

> 註:閃退(穩定性)獨立追,不在規則範疇。

## 進度

(逐項回填)
