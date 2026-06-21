# 工作魔法大帝 繁體中文化 — Master of Magic (CHT)

> 還記得嗎?1994 年的某個午後,你在 14 吋 CRT 前,看著兩個位面在 320×200 的畫面裡緩緩展開——
> Arcanus 與 Myrror,十四本魔法書,上百種單位與法術。那是《Master of Magic》(MicroProse)。
> 它沒有官方中文版,三十年也沒人完整漢化過。這個 repo,就是想把這件事補上。

跑在開源 Go 重製引擎 [kazzmir/master-of-magic](https://github.com/kazzmir/master-of-magic) 上,
用**自製 CJK 渲染管線 + 載入後字串覆蓋**把遊戲文字一次解掉,且**全程不修改、不散布任何版權檔案**。
本專案承接先前《銀河霸主》(Master of Orion) 中文化的方法論。

> ⚠️ **進度誠實說明:專案在 Phase 2(進行中),現在還不能完整遊玩。**
> 已完成的是「中文能正確畫上真實引擎畫面」這一關(Phase 1),以及版本決策與第一批譯文。
> 接下來才是讓真實遊戲畫面的英文逐步換成中文。進度以 [`PLAN.md`](PLAN.md) 為單一真實來源。

當年華文電玩圈怎麼看這款遊戲、它的譯名到底叫什麼、為什麼這麼多年沒有中文版——
這段考古整理在 [`docs/history-chinese-reception.md`](docs/history-chinese-reception.md)。

---

## 為什麼是 Go / Ebiten 引擎

這次不像《銀河霸主》那樣改 C 引擎(1oom)。我們站在 kazzmir 用 **Go + Ebiten** 重製的引擎上——
它直接讀原版 `.lbx` 執行,不是模擬器,是把遊戲邏輯乾淨重寫的重製版。選它有一個很實際的理由:**跨平台**。

Ebiten 原生支援 Windows / macOS / Linux / Android / iOS / Web,而且上游**已經把 Web (WASM) 版部署到 itch.io**。
換句話說,語言不是跨平台瓶頸;真正的成本在「打包簽章」與「觸控/檔案存取 UX」,那是換成 C++ 一樣得做、甚至更麻煩的事。
完整評估見 [`docs/porting-difficulty.md`](docs/porting-difficulty.md)。

對中文化來說還有一個甜頭:引擎的字串繪製迴圈本來就走 `for _, c := range text`(rune / UTF-8 迭代),
只是非 ASCII 的字被靜默丟棄。所以我們不必重寫整條點陣管線,只要在 glyph 查找處加一條
「碼點 ≥ 0x80 → 改用 CJK 字形來源」的支線即可。

---

## 現在到哪了

### 已完成

**Phase 0 — 盤點與骨架**
盤點原版 142 個 LBX 資產、clone 引擎、摸清字型 / LBX / 渲染架構,確立 patch-only 散布原則,
建好 `PLAN.md` / `CONTEXT.md` / ADR 骨架。

**Phase 1 — CJK 渲染注入(已驗證)**
不只在獨立 prototype 把一行中文畫上畫面,更進一步在**真實引擎字型管線**(`lib/font` + 真實 `fonts.lbx`)上,
讓中文走完引擎三條繪字路徑——`doPrint` / `PrintOutline` / `MeasureTextWidth`——全部正確渲染。
英文續用引擎原本的金色花體點陣字,中文用 Noto Sans CJK TC。注入點集中在新增的 `lib/font/cjk.go`
與一份 patch(`patches/0001-cjk-font-injection.patch`),不碰版權檔。技術路線決策見 [ADR 0001](docs/adr/0001-cjk-rendering.md)。

![真實引擎三條繪字路徑都吃中文:標題、神器名、物品能力中英混排,drop shadow 與 shader outline 都正確套到中文 glyph](docs/img/phase1-engine-cjk.png)

*上圖是真實引擎(非 prototype)跑出來的畫面。看「神聖復仇者」「烈焰之劍」這些中文如何和原版金色英文標題並排——這證明三條繪字路徑都吃中文了。也正是這張圖暴露了 `PrintOutline` 一開始漏 patch、以及中文字級偏大導致行距重疊兩個問題。*

跑真畫面的價值就在這裡:CI 編譯全綠不代表畫面對。這一畫面當場逼出兩個必修項——`PrintOutline`
漏 patch(已修)、CJK 字級未對齊引擎字高造成行距重疊。

**Phase 2 第一刀:字級對齊(已完成)。** 把 `cjkGlyphImage` 改成 size-aware——依呼叫端字型的
`GlyphHeight` 決定 CJK 渲染尺寸、以 (rune, height) 為快取 key、各字高一份 face,三條繪字路徑都傳入字高。
結果中文不再溢出行高、與英文基線對齊、各行不再重疊:

![字級對齊後:中文與英文基線一致、行距不再重疊(對比上一張 phase1 的破版)](docs/img/phase2-cjk-aligned.png)

*同一段中英混排,經字級對齊後的樣子。和上一張比,「神聖復仇者」「烈焰之劍」不再壓到下一行,中英文坐在同一條基線上。trade-off:最小字型下中文偏小,密集面板日後可走 hi-res canvas 或固定尺寸點陣補強。*

下面這張是更早的獨立 prototype 截圖,驗證「逐 rune 預渲染 + 快取」這條注入支線本身可行:

![獨立 prototype:標題「工作魔法大帝」加第一批 item 譯文,全部經 CJK glyph 支線渲染](docs/img/phase1-cjk-hello.png)

**版本決策(ADR 0002,已實測)**
目標版本定為 **Community Patch 1.60**(對齊引擎邏輯與測試基準),同時保留 **vanilla 1.31** 相容——
做法是譯文表以「英文原文字串」為 key。我們實際用 MOMDIFFP 把正版 1.31 升級到 1.60、對前後字串做 diff,
得到一個關鍵結論:CP 1.60 維持所有**名稱**字串不變(改的是數值平衡與重寫散文),
所以神器名 / 物品能力 / 法術名的譯文表**一份通吃兩版**;只有 help / 建築描述 / 訊息這類散文須以 1.60 為準重抽。
實測 diff 細節見 [ADR 0002](docs/adr/0002-target-game-version.md)。

**第一批譯文**
物品能力 / 附魔(`itempow.lbx`)第一批已譯,放在 [`docs/strings/item-powers.tsv`](docs/strings/item-powers.tsv)。

### 進行中 / 待辦(Phase 2 起)

- **把 CJK 字型納入打包**:用自由授權字型或自製點陣,不依賴系統字型路徑。
- **字串覆蓋層**:LBX 載入後即時把英文換成中文(不改版權 LBX)。這一步做完,真實遊戲畫面才真正中文化。
- **其餘譯文**:神器名 → 法術 → 建築 → help → 單位名(單位名 hardcode 在 `units/unit.go`,須改查表)。
- **跨平台打包**:Linux / Windows → Web WASM → macOS → Android,順序與理由見 `docs/porting-difficulty.md`。

---

## 這份 repo 放什麼 / 不放什麼

patch-only。本 repo **不 vendor 引擎本體、不散布任何版權遊戲檔**,玩家自備正版資料即可套用。

| 放 | 不放(版權,列入 `.gitignore`) |
|---|---|
| 計畫 `PLAN.md`、術語 `CONTEXT.md`、決策 `docs/adr/` | 原版遊戲檔(`.lbx` / `.exe` / 手冊) |
| 英文 → 繁中譯文表 `docs/strings/*.tsv` | 引擎本體(由 `scripts/fetch-engine.sh` 取得) |
| CJK 渲染 patch、字型烘製腳本、prototype | 解壓出的任何版權資產 |

譯文表是本專案的衍生資產,版權 LBX 分毫未動。

---

## 文件導覽

| 文件 | 內容 |
|---|---|
| [`PLAN.md`](PLAN.md) | 單一真實計畫來源,階段規劃與進度回填 |
| [`CONTEXT.md`](CONTEXT.md) | 專案術語表(ubiquitous language) |
| [`docs/phase1-cjk-prototype.md`](docs/phase1-cjk-prototype.md) | Phase 1 渲染驗證全紀錄(含真實引擎截圖與必修項) |
| [`docs/porting-difficulty.md`](docs/porting-difficulty.md) | 跨平台移植難度評估(為何維持 Go,不重寫 C++) |
| [`docs/adr/0001-cjk-rendering.md`](docs/adr/0001-cjk-rendering.md) | CJK 渲染路線決策(TTF 為主線,點陣為美術升級選項) |
| [`docs/adr/0002-target-game-version.md`](docs/adr/0002-target-game-version.md) | 目標版本 = CP 1.60,1.31 相容,附實測字串 diff |
| [`docs/history-chinese-reception.md`](docs/history-chinese-reception.md) | 魔法大帝當年在華文電玩圈的譯名考據與接受史 |

---

## 參考

- 引擎:<https://github.com/kazzmir/master-of-magic>
- 銀河霸主(Master of Orion 1, 1oom)繁中化前例:patch-only,24×24 CJK 管線
