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

### T1 明確 bug(純 bug → 兩模式都修;非「規則差異」不值得留著對照)
- [x] **#5 創角名字輸入打一半被踢回** (2026-06-25)
  - 根因:`inputmanager.IsQuitKey` 誤含 `CapsLock`(`Escape || CapsLock`)→ 打名字按大小寫切換
    被當取消,`HandleKeys` 呼 `previousUI()` 退回選擇畫面。影響全部 5 個建角子畫面(都用此 handler)。
  - 柵欄檢查(Chesterton):web 版 `input_js.go` 刻意 return false「避免誤觸退出」=作者已知誤觸是問題;
    CapsLock 全碼庫無其他用途;柵欄正當目的是「Escape 取消」,CapsLock 是搭便車意外。
  - 修:`input_default.go` 的 `IsQuitKey`/`IsQuitPressed` 移除 CapsLock(保留 Escape)。根因一處修好 5 畫面。
  - 驗:`IsQuitKey(CapsLock)=false, Escape=true`(xvfb 確定性)。
- [x] **#1 走進怪物點直接判敗** (2026-06-25)
  - **根因(testing 證明,不靠遊玩)**:`doCombat.createArmy`(game.go:4626)規則「水域戰場濾掉所有
    `IsLandWalker` 陸行單位」。玩家攻擊**水上據點**時,combat 地形取防守方(據點)的水格 →
    整支陸軍被濾空 → 戰鬥開始即無活單位 → 判敗、完全不開打。對上「有些(水邊)怪物點走進去秒敗」。
  - `cht_repro_test.go` 用 BarbarianSpearmen 數值直接驗:grass→2 單位、water→**0 單位**(空軍隊)。
    確認 `DoStrategicCombat`(純比 power 的自動結算)是**死碼無呼叫者**,排除該假說。
  - **Oracle**:原版 MoM 載於運輸船的陸軍會在海戰作戰,不會憑空消失。
  - **修(Classic 守護)**:過濾若導致空軍隊但原本有單位,保留被濾單位讓其開打。測試驗 Classic→2 單位。
    戰鬥語意改動 → flag 守護,Remake 保留原行為供對照。
- [x] **#9 召喚第七英雄靜默失效** (2026-06-25)
  - 根因:`doHireHero`(召喚 cost 0 與雇用共用此路徑,經 `GameEventHireHero`)的 `if added {...}`
    沒有 else——英雄滿 6(`AddHero` 找不到空 slot 回 false)時靜默吞掉。上限本身有強制(維持 6),只缺警告。
  - 原版會警告「無空間,需先解僱」(對上玩家記憶 = oracle)。
  - 修:`added==false && IsHuman()` 時 `GameEventNotice` 警告「你沒有空間容納更多英雄,請先解僱一位。」
    一處修好召喚 + 雇用兩條路徑。純 UX bug → 兩模式都修。misc-ui.tsv +1。

### T2 經典規則
- [x] **#6/#8 書系限制 — 經查證為正常 MoM 行為,非 bug**(2026-06-25)
  - 測試 `cht_books_test.go` 證明:`InitializeResearchableSpells` 只迭代 `player.Wizard.Books`,
    只有 Nature 書 → 研究池 Nature=8、**Chaos=0、Death=0**。研究 gating 嚴格依書系,正確。
  - 玩家 KnownSpells 唯一增長路徑 `LearnSpell` 只在研究完成時呼叫(game.go:7041),與英雄無關 →
    玩家法術書不可能沒研究就有他系法術。
  - 他系法術來自**英雄/物品 spell charges**(`hero.GetSpellChargeSpells`,合法 MoM,不受書系限制);
    MoM 中**召喚英雄不依書系 gate**(可同時有 Roland+Mortu)。Life/Death 互斥已在建角強制(new-wizard.go:1213)。
  - **結論**:正常 MoM 行為被誤記。柵欄原則 + 別猜——不對正確機制加錯誤限制。`cht_books_test.go` 留作證明。
- [x] **#2 製造神器成本(存讀檔丟失 OverrideCost)** (2026-06-25)
  - **手冊 oracle**(manual.pdf p.88):製作神器時間取決於「神器法力成本、施法技能、每回合法力」;
    強大物品極貴、需很長時間。故 4000 神器 2 回合完成確是 bug。
  - **live 流向其實正確**(我先前假說錯):game.go:5426 `spell.OverrideCost=created.Cost` → 5450
    `player.CastingSpell=spell` → 7009 多回合用 `ComputeEffectiveSpellCost`(尊重 OverrideCost)。
    測試 `cht_cost_test` 證 `calculateCost` 正確加總(100+1600+1600=3300)。
  - **真根因(存讀檔特定)**:serialize.go:250 只存 `CastingSpell.Name`,**OverrideCost 丟失**;
    讀檔 426 `FindByName` 回 base spell(OverrideCost=0)→ 進行中貴神器成本退回 base → 幾回合做好+法力不足照做。
    **對上玩家「會閃退、注意存檔」=常存讀檔。**
  - **修(兩模式)**:讀檔後 CastingSpell=Create Artifact/Enchant Item 時,從另存的 `CreateArtifact.Cost`
    還原 OverrideCost。`cht_castcost_test` 證序列化丟失 + CreateArtifact.Cost 可還原。

### T3 需先對照 1.31 oracle 再決策(可能 bug 也可能 CP1.60 故意改)
- [x] **#3 英雄戰死獲勝後復活 — 資料層證明引擎已正確處理,無復活機制**(2026-06-25)
  - **手冊 oracle**(p.71):「英雄是獨特存在,死了除非被復活否則就是死了」→ 戰死英雄不該因獲勝復活。
  - **死亡判定鏈(靜態追溯,皆正確)**:combat 的 `ArmyUnit.TakeDamage`(model.go:1782)對 `unit.Unit`
    呼叫 `AdjustHealth(-damage)`;`unit.Unit` 是真英雄物件(`Army.AddUnit` by reference,model.go:2119)→
    dispatch 到 `Hero.AdjustHealth`(hero.go:548)→ health<=0 時 `SetStatus(StatusDead)`。
  - **戰後 reconciliation(皆正確)**:存活 re-add(game.go:4778)只加回 `GetHealth()>0`(死亡英雄略過);
    `killUnits`(game.go:4922)對 `GetHealth()<=0` 呼 `player.RemoveUnit`。關鍵:`RemoveUnit`(player.go:1480)
    **只在 `Status==StatusEmployed` 時**才改 `StatusAvailable`;死亡英雄此時已是 `StatusDead` → 不被改回 →
    僅 `Heroes[i]=nil`,英雄保持死亡並移出名冊。`FinishCombat`(model.go:4393)勝利時只治療 Regeneration
    單位(英雄無此特例);亡靈升起(4423)排除 RaceHero。`NaturalHeal`(game.go:7199)只作用於 stack 內單位,
    死亡英雄已被移出 stack,不會被治療復活。
  - **測試證明**:`cht_herodeath_test.go`(player 套件)精確重現 doCombat(AttackerWin)對英雄的兩步
    reconciliation,致死走真實 `AdjustHealth` 路徑。斷言戰死英雄結束時 `StatusDead`、不在 `Heroes` 名冊、
    `AliveHeroes()==0`、未重回軍隊 → **PASS**,資料層無 #3 復活問題。
  - **為何不寫完整 combat integration test**:`MakeGame` 需版權 LBX(lbxCache),CI/單元測試拿不到 →
    全 doCombat 路徑無法 headless 跑(這也是引擎本身無此類測試的原因)。故改測「資料層 reconciliation seam」。
  - **結論(柵欄原則)**:此引擎版本英雄死亡處理正確,**不下推測性補丁**(同 #6/#8)。玩家回報的「復活」最可能是
    「英雄戰鬥中掉到低血、未真正死亡(末圖動畫像死),戰後 NaturalHeal 回血」=正常 MoM,非真復活。
    要 100% 蓋掉 UI/多 stack 路徑需實機 playtest(`retro-game-playtest` 紀律,需 LBX 實玩)。測試留作回歸保護。
- [x] **#10 打怪英雄當場穿裝備免 20 傳送法力 — 確認 bug 並修正**(2026-06-25)
  - **手冊 oracle**(pp.28-29):與物品**當前位置**同格的英雄裝備顯示「Same Location」→ 移動免費;
    異地英雄顯示「Item Teleport」→ 付 20 mana。免費條件綁「物品位置」。
  - **根因**:`vault.go:105` 把**新發現戰利品**的 `selectedItem.Location` 寫死為 `fortressLocation`。
    但打怪戰利品的「位置」應是**發現地點**(戰鬥 tile)。已驗證 stack 移動鏈(game.go:3454→3459):打怪後
    英雄 stack 正位於 encounter tile = `treasure.Point`。寫死 fortress → 打怪英雄被誤判異地 → 一律 20 mana。
  - **柵欄檢查**:vault 畫面共用於「軍隊畫面瀏覽 vault」(物品確在 fortress)與「新發現戰利品」兩情境;
    作者用 fortress 當預設對前者正確、對後者錯。修正只對「新戰利品 + 有發現地點」改用發現地點,前者維持 fortress。
  - **修(兩模式)**:`Treasure` 已有 `Point`(發現地點)。把它經 `ApplyTreasure`→`doVault(...,foundLocation)`
    →`showVaultScreen(...,foundLocation)` 帶入;新戰利品的初始 `Location` 用發現地點,其餘 caller(軍隊畫面、
    商人購入、死亡英雄裝備回庫)傳 nil → 維持 fortress。打怪英雄當場穿戴 = Same Location = 免費,符合手冊。
  - **驗證**:全引擎 `go build ./...` 通過(4 個 doVault caller + showVaultScreen 簽名一致);UI seam 無法
    unit test(需 LBX),以手冊 oracle + 移動程式碼路徑(stack==treasure.Point)+ 編譯為據。
- [ ] #4 船斜向/長程海上尋路錯誤聲(pathfinding)
- [ ] #7 法術書顏色排序選項 + 設定畫面補項

> 註:閃退(穩定性)獨立追,不在規則範疇。

## 進度

(逐項回填)
