# Go / Ebiten 老遊戲繁中化方法論

把一款跑在 Go + Ebiten 重製引擎上的老遊戲完整中文化的可重用流程與踩雷紀錄。
本專案(工作魔法大帝 / Master of Magic)從 ~937 條起步、收斂到 3,200+ 條、三平台可玩,過程萃取成這份 playbook。
與前例《銀河霸主》(Master of Orion / 1oom,**C 引擎**,走離線點陣烘字)分岔:Go 引擎讓「**載入後字串覆蓋 + go:embed 內嵌**」這條路徑成立,不必改版權資料、不必重寫點陣管線。

---

## 1. 何時適用

- 目標遊戲有**開源 Go / Ebiten 重製引擎**(直接讀原版資料,非模擬器)。
- 字串來源混合:資料檔(LBX 等)+ 少量 hardcode 在原始碼。
- 要保持 **patch-only**:不散布版權資料,只散布譯文表 + 引擎 patch + 字型子集。

不適用:純 C/ASM 引擎(走 1oom 那條離線烘字路線)、無重製引擎只能改 binary 的情況。

---

## 2. 核心架構決策:只在「顯示層」翻譯

**[HARD] 不要在資料讀取層把英文改成中文。** 引擎常把英文字串當**邏輯 key** 比對
(例:`if string(name) == "+6 Defense"`)。資料層改中文會破壞遊戲邏輯。

正解:在 `lib/font` 的繪字進入點(`doPrint` / `PrintOptions` / `MeasureTextWidth` / `CreateWrappedText`)
掛一個 `translateForDisplay(text)`,**比對採 TrimSpace 後精確比對**,查表命中才換中文,引擎內部仍用英文。
- 表來源:`go:embed cht_strings/*.tsv`(出包零依賴)+ 可選 `MOM_CHT_STRINGS` env 覆寫(dev 用)。
- 載入採**先到先得**(同 key 第一個贏),用檔名字母序控制優先權。
- 見 `ADR 0003`、`lib/font/override.go`。

---

## 3. CJK 渲染注入

引擎的繪字迴圈本來就走 `for _, c := range text`(rune 迭代),只是非 ASCII 被靜默丟棄。
在 glyph 查找處加分流即可,不必重寫管線:

- `glyphIndex := int(c) - 32`(ASCII 32–127);超範圍且 `isCJK(c)` → 走 `cjkGlyphImage(rune, height)`
  逐 rune TTF rasterize + 快取(key = `(rune, height)`)。字型來自 env / 系統路徑 / 內嵌子集。
- **[HARD] 字級要對齊呼叫端字高**:依 `font.GlyphHeight` 決定 CJK 渲染尺寸,否則行距重疊破版。
- **[HARD] CJK 銳利化 = supersample**:`cjkSupersample=4`,以 logical 字高×4 高解析 rasterize、
  doPrint 以 `scale/4` 畫。對應 logical 320×200 → 視窗 4× 放大。**不可回退成「小字放大」(會糊)**。
  (這跟 rule 81「拉高內部畫布」同源:不縮字,改提高內部解析度。)
- `.ttc` 用 `opentype.ParseCollection`+`Font(0)`;**AR PL UMing(CFF/舊式)x/image/sfnt 解析失敗**,
  即時 rasterize 改用 Noto Sans CJK TC。

---

## 4. 譯文工作流:dump → 翻譯 → embed → 重生字型

1. **dump 精確英文字串**:寫一支 `test/cht-dump`,用**引擎自己的 reader**讀出**所有**顯示文字源的
   精確字串(LBX 二進位別自己 parse,易錯)。需在 **xvfb** 下跑(Ebiten init 要 display)。
   **[HARD] 必 dump 的文字源清單(漏一個 = 整類畫面英文)**:
   - `help.lbx`(右鍵 help 卷軸、法術 help)— `helplib.ReadHelp`
   - `desc.lbx`(法術書施法卷軸的法術描述)— `spellbook.ReadSpellDescriptionsFromCache` ← **2026-06 漏掉這個整類沒翻**
   - `buildesc.lbx`(城市建造畫面的建築短描述)— `building.MakeBuildDescriptions`(與 help.lbx 的長版**不同字串**)
   - `builddat.lbx` 建築名、`cityname.lbx` 城市名、item-powers / artifacts / spells / units / messages / templates…
   > 教訓:`desc.lbx` / `buildesc.lbx` 是**獨立於 help.lbx** 的文字源;只 dump help.lbx 會漏掉施法卷軸與建造畫面描述。
2. **[HARD] key 必須對齊「實際 shipping 資料版本」的 dump**:`source` 欄不可用 wiki/手打/憑記憶的英文,
   一律用 §1 從**當前 `extracted/` 實機 reader dump 出來的字面**。動翻譯前先跑 `comm -12`(dump 的 source ∩ TSV 的 key),
   命中率 <95% 代表 key 過時/版本漂移,先重新對齊再翻(見 §「版本漂移」與 ADR 0002)。
3. **切塊翻譯**(見 §8 並行 agent)。
4. **回填 TSV** → `docs/strings/*.tsv`(欄位 `source<TAB>zh<TAB>note`,英文原文即 key)。
5. **複製進 embed** + **全面重生字型子集**(否則新字顯示空白方塊):
   收集「引擎所有 .go 的非 ASCII 字 + 全 TSV zh 欄」→ `sort -u` → `pyftsubset` Noto **Regular**
   (非 DemiLight,字重一致)`--font-number=0`。
6. **重建 + 驗證**(見 §9、§10)。

格式碼處理:help.lbx 本體用 `0x14` 當換行,TSV 以字面 `\n` 表示,override 載入時 `\n`→`0x14`
還原,讓 source 精確比對、zh 也保有換行。

---

## 5. 動態字串(Sprintf 模板)

組合字串在 `fmt.Sprintf` 時 `%v` 已填值,顯示層整串比對不會命中。解法:**把模板字面先翻**:
```go
fmt.Sprintf(font.TranslateFormat("Cost %v"), cost)   // 模板進 TSV: "Cost %v" → "花費 %v"
```
- **[HARD] 佔位符數量/順序英中必一致**(否則 Sprintf 行為錯亂/panic);驗證可逐行比對 `%` 數量。
- **內嵌列舉引數**(城市規模/種族/礦物 `.String()`)要個別再翻:
  `fmt.Sprintf(font.TranslateFormat("%v of %s"), font.TranslateFormat(size.String()), name)`。
- **只包顯示用的 Sprintf**:跳過 `log.*` / `Printf` / `Errorf` / `AddLogEvent` / 存檔序列化 / `==` 邏輯比對。
- 靜態(非 Sprintf)訊息字串免改碼,直接進 TSV 即可(走顯示層覆蓋)。

---

## 6. 專有名詞「中文(英文)」+ 英文小字

需求:巫師/英雄/城市等專名既保留原文又看得懂 →「羅潘(Lo Pan)」格式。

- **資料**:`names.tsv` / `citynames.tsv` 用「中文(半形英文)」。直接列印的名字自動生效;
  Sprintf 引數處的名字要 `font.TranslateFormat(name)` 包起來(軍隊/城市標題、外交、英雄全名、事件訊息)。
- **英文小字(scoped 不誤傷)**:override loader 只對 names/citynames 把尾端「(英文)」包進控制碼
  `SmallStart=\x0e`…`SmallEnd=\x0f`;`doPrint`/`MeasureTextWidth` 遇標記把字級降 `0.65`、下移
  `GlyphHeight*(scale-curScale)` 對齊基線、不繪標記。標記是 ASCII 控制碼,未處理的渲染路徑只是不縮、不出錯。
- **避免全域 collision**:名稱用獨立檔 + 字母序載入,蓋過 units.tsv 舊純中文;**注意誤判**——
  神器「Golden Staff of Sharee」的 `grep -F "Sharee\t"` 會誤匹配子字串,核對時要錨定行首。

---

## 7. 圖片疊字(烘進 LBX 的英文按鈕/標題)

有些英文是烘進美術圖的點陣,不走字串路徑。用「擦底色 + 疊銳利中文」:
- `util.ChtLabel(screen, rect, img, fnt, label)`(吃按鈕 img 取樣底色)/
  `util.ChtLabelRect(...fill...)`(吃面板紋理填色)。在 `ui.StandardDraw(screen)` **之後**疊。
- 字級由傳入的 `fnt` 決定(NormalFont 太小就改 BigFont)。
- Draw closure 內若要用後面才宣告的 LBX 圖,**重抓**(`GetImages` 有快取,便宜),避免 Go 閉包作用域問題。
- 哥德花體標題字也是 `*font.Font`,會走 translateForDisplay + CJK 注入,headline 進 TSV 即可中文。

---

## 8. 並行 agent 翻譯工作流

- **切塊**:920 條 help 切 10 塊、162 個 Sprintf 模板按**檔案**分給 7 個 agent,各寫獨立輸出檔避免衝突。
- **agent 紀律**:純文字翻譯**不碰 git**(classifier 會擋 push)、不 build、各檔不重疊;主 agent 統一整合 + commit。
- **逐塊核對行數**(in vs out)抓漏譯;模板類額外做**字面 vs TSV key 100% 命中**的確定性比對。
- 機械包裹(wrap Sprintf)交給 agent 判斷「顯示 vs 內部」,比 grep 可靠;但仍要主 agent 全 build 收尾。
- transient API 中斷會讓 agent 早夭(產出空檔)→ 檢查產出、必要時改自己翻或重派。

---

## 9. 打包(三平台,全 docker)

- **Linux AppImage**:`golang:1.25-bookworm` + CGo + X11 dev headers,內含全部遊戲檔。
- **Windows x64**:`CGO_ENABLED=0 GOOS=windows`(purego/DirectX)→ **magic.exe 只依賴 kernel32.dll,免外部 DLL**。
- **macOS arm64**:**需 CGo(Metal/Cocoa),不能從 Linux 交叉編譯** → GitHub Actions `macos-14` runner
  套 `0099` patch + `prepare-embed.sh` 建 binary;下載 artifact 後**本地加遊戲檔 + wrapper 帶 `-data`** 組 .app。
- 三包同源(同一引擎工作目錄 + 同一 patch + 同一字型子集)才能保證譯文一致。
- **權威 patch = `patches/0099-all-engine-cht.patch`**(`git diff HEAD` 全 .go);分散的小號 patch 有重疊勿混用。

---

## 10. 驗證紀律

- **[HARD] CI 編譯全綠 ≠ 畫面對**:每批改動用 **dist AppImage 在 headless(xvfb)逐畫面截圖**,
  `xdotool` 導航 + `import -window root` 抓圖 + Read 看。
- **確定性 key 比對**補強截圖難觸發的(事件、深層對話):引擎字面 vs TSV key 全命中。
- **go build 抓我方錯誤看「型別檢查」階段**:`./game/magic/...` 編到 link 才失敗(缺 -dev 庫)
  = .go 改動 OK;權威驗證直接跑 build-appimage.sh(含完整 apt + link)。

---

## 11. 踩雷清單(最省時間的部分)

| 雷 | 症狀 | 解 |
|---|---|---|
| **gofmt 引擎既有檔** | 4-space 縮排檔被轉 tab,整檔變 670 行空白雜訊 patch | 只 gofmt 自己新增的檔;誤 gofmt 後 `git checkout HEAD -- file` 重套 |
| **go:embed 新檔名增量不重嵌** | 加**全新** TSV 後 build 仍用舊表(獨立 `go run` 測 TranslateFormat 卻正確) | 清 GOCACHE 或 `touch lib/font/override.go` 強制 lib/font 重編 |
| **noble libasound2 改名** | ubuntu:24.04 `apt install libasound2` → "no installation candidate" exit 100 | 改 `libasound2t64`;且 install 前必先 `apt-get update` |
| **font.Translate 不存在** | agent 自創不存在函式 → 編譯失敗 | 只有 `TranslateFormat`(純字串也能用,無 %v 即整串比對) |
| **CGO_ENABLED=0 測 font** | ebiten glfw undefined / GLFW not initialized | font 拉進 ebiten,需 CGo+X11+xvfb;或測純資料時仍要 xvfb |
| **oto ALSA headless crash** | 無音效裝置 panic | `/tmp/.asoundrc` 設 null PCM;或 `-music=false` |
| **lbxdump/dumper "GLFW not initialized"** | Ebiten init 要 display | xvfb-run 包起來 |
| **單 key 翻譯 collision** | 同英文 key 用在不同語境(資源 "Power" vs 神器 "Power") | 改 call site 用獨立 key(如 "Magic Power");或專名走獨立檔 |
| **背景 sentinel 迴圈空轉** | `until [ -f x ]; sleep` 等不到檔空轉整夜 | 禁 sentinel;docker 同步前景跑、等回傳 |
| **headless dump 無 frame 上限** | dummy SDL `--dump` 無限 poll ~70% CPU | 帶 `--frames N` 或 `timeout` |

---

## 12. 對應檔案速查

| 主題 | 檔 |
|---|---|
| 顯示層覆蓋 + 小字標記 | `lib/font/override.go`(patch 內) |
| CJK 注入 + supersample | `lib/font/cjk.go` |
| 繪字三路徑 + 標記渲染 | `lib/font/font.go` |
| 圖片疊字 helper | `game/magic/util/cht_label.go` |
| dump 工具 | `test/cht-dump/main.go` |
| 譯文表 | `docs/strings/*.tsv` |
| 權威 patch | `patches/0099-all-engine-cht.patch` |
| 打包腳本 | `docker-scripts/`(本地)/ `.github/workflows/macos.yml` |
| 決策 | `docs/adr/0001`(渲染)`0002`(版本)`0003`(覆蓋層) |
