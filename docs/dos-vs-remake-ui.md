# 原版 DOS vs 重製引擎(kazzmir Go/Ebiten)的介面差異

有玩家反映「重製版玩起來沒有當年 DOS 的感覺」。這份報告整理兩者在**操作介面與視覺呈現**上的差距,
並指出造成「沒 DOS 感」的主因。結論放最前面。

> 方法:web 研究(原版規格、320×200 長寬比文獻、kazzmir 引擎說明)+ **直接查本專案引擎原始碼驗證**
> (`/tmp/mom-engine` 的 `scale`/`main.go`/`fonts`)。標「已驗證」的為原始碼直接佐證。

## 結論:最可能造成「沒 DOS 感」的三大主因

1. **長寬比沒做 CRT 校正(影響最大)** — 原版 320×200 在 4:3 CRT 上是被**垂直拉高約 20%**顯示的;
   重製版用方形像素整數倍放大(960×600,16:10),沒有把 Y 拉 1.2 倍 → 畫面整體**偏扁偏寬**,
   巫師臉、城市圖、UI 框都比記憶中「矮胖」。這是幾何上的客觀差異,輪廓就是對不上。
2. **缺 CRT 質感** — 原版在 CRT 上是柔化+掃描線+螢光暈;重製版最近鄰銳利放大,在 LCD 上是**硬邊方塊**,
   乾淨但冷、像素感過強。很多人記憶中的「DOS 感」其實是 CRT 的柔化質地。
3. **音樂音色不同** — 原版走 OPL FM(AdLib/Sound Blaster)或 MT-32;重製版用內建 MIDI 合成器 + SF2 soundfont
   (偏 General MIDI 擬真樂器)。同一段曲子,FM 的金屬電子味 vs GM 的擬真味,氛圍記憶落差大。

字型其實是忠實的(常被誤會,見下)。

---

## 規格基準

| | 原版 DOS(1994)| 重製版(kazzmir Go+Ebiten)|
|---|---|---|
| 引擎 | 原生 DOS 執行檔 | Go + Ebiten,**直接讀原版 `.lbx` 資產**(非 DOSBox 模擬)|
| 邏輯解析度 | 320×200(256 色 VGA)| 320×200(`data.ScreenWidth/Height` 已驗證)|
| 實際顯示 | CRT 拉成 4:3,**像素長寬比 PAR ≈ 1:1.2**(垂直伸長 20%)| 方形像素 ×3 → **960×600(16:10)**,無長寬比校正(已驗證 `Layout = Scale2(320,200)`)|
| 縮放濾鏡 | CRT 類比柔化 | 預設**最近鄰**(`ScaleAmount=3.0`、`ScreenScaleAlgorithm=Normal` 已驗證);另有 Scale2x/XBR 可選 |
| 字型 | LBX 內金色哥德點陣字 | **沿用同一份 LBX 點陣字**(英文走 `LbxFont`,已驗證);繁中化只對 CJK 注入 TTF |
| 音樂 | OPL FM / MT-32 | 內建 MIDI 合成器 + SF2 soundfont |
| 視窗 | 全螢幕 CRT | 預設視窗化 |

---

## 差異清單(逐項)

### 1. 長寬比未校正 ★最關鍵
- **原版**:320×200 在 4:3 CRT 上垂直拉伸至約 320×240,PAR ≈ 1:1.2;美術都是依此比例設計、看起來才「對」。
- **重製版**:方形像素整數倍,輸出 960×600(16:10),Y 沒有多乘 1.2。
- **為何沒 DOS 感**:所有圖形**偏扁約 20%**,巫師肖像、地圖、面板輪廓和記憶對不上。最廣、最客觀的失真。
- **可信度:高(已驗證)**。`Layout` 回 `scale.Scale2(320,200)` = 960×600 方形像素,無 1.2 Y 校正。

### 2. 缺 CRT 質感(掃描線 / 柔邊 / 暈光)
- **原版**:CRT 把大像素模糊柔化、帶掃描線與螢光暈,色塊間是漸層,觀感溫潤。
- **重製版**:最近鄰銳利放大,LCD 上是硬邊方塊,無掃描線、無柔邊。
- **為何沒 DOS 感**:記憶中的「DOS 感」很大一部分是 CRT 質地;銳利方塊反而「更不像當年」。
- **可信度:高(已驗證預設最近鄰)**。

### 3. 音樂音色(SF2 GM vs OPL/MT-32 FM)
- **原版**:OPL2/OPL3 FM 合成或 MT-32 的標誌音色。
- **重製版**:MIDI 合成器 + SF2(GM 取向)。
- **為何沒 DOS 感**:聽覺記憶落差直接影響「氛圍對不對」。
- **可信度:高**(引擎說明明載 MIDI+SF2)。

### 4. 部分畫面是「重排版面」而非螢幕 1:1 還原
- **原版**:固定點陣構圖。
- **重製版**:讀原始 sprite/底圖,但**版面由程式重新排版繪製**;部分畫面用原 baked 底圖、其餘逐元素重畫
  (本專案繁中化工作中已實證:有一批 baked 圖片畫面用原圖,其餘是引擎重組)。minimap/縮放等互動也與原版不同。
- **為何沒 DOS 感**:元素間距、對齊、留白有細微位移,熟悉原版的人會覺得「排版怪怪的」。
- **可信度:中**(本專案實作經驗 + 引擎承認 UI 取捨;非每畫面都有逐一截圖比對)。

### 5. 字型其實忠實(澄清常見誤判)
- **原版 vs 重製版**:重製版**直接沿用 LBX 原始點陣金色哥德字**,英文字形 1:1,**沒有換成 TTF**。
- **為何容易誤會**:疊在第 1、2 點之上——同樣的點陣字,被方形像素+無 CRT 柔化呈現,看起來比記憶中「硬」,
  但那不是字型換了,是**顯示鏈差了**。
- **可信度:高(已驗證)**。英文走 `font.LbxFont`;繁中化只對 CJK 走 TTF,英文不動。

### 6. 滑鼠游標 / 動畫時序(次要、推測)
- 重製版 Ebiten 固定 60 TPS 重跑時序;游標慣性、動畫快慢的微差可能累積成「不是當年那台」的感覺。
- **可信度:低**(合理推測,未逐項公開驗證)。

---

## 若要讓重製版更接近 DOS 觀感,技術上可調的方向

- **長寬比校正(CP 值最高)**:邏輯 320×200 算完後,輸出時 X、Y 用不同倍率,Y 多乘 1.2
  (例如 X 5× / Y 6× → 1600×1200,或用 Ebiten `GeoM` 個別縮放)。Ebiten 完全支援,改動集中在 `scale`/`Layout`。
- **CRT 後製 shader**:用 Ebiten Kage 寫一個 post-process pass,加掃描線 + 遮罩 + 輕微 bloom/blur,模擬 CRT 柔化。
- **音樂**:改走 OPL/AdLib 模擬或 MT-32(munt)還原原音色;若留在 GM,至少換一個接近 Roland 的 SF2。
- **palette**:確認沿用 LBX 的 VGA 256 色盤(引擎本就讀 LBX 色盤,屬低風險檢查項)。
- **字型**:**維持沿用原 LBX 點陣字**,英文別 TTF 化;「字看起來硬」交給長寬比 + CRT shader 解決,不是改字型。

> 對本繁中化專案的意義:這些都是**引擎層**的觀感差異,與中文化無關(中文化只動顯示層字串覆蓋 + CJK 字形)。
> 若要追求 DOS 觀感,優先做「長寬比校正」與「可選 CRT shader」,且都用**設定開關**包住、預設維持現況,
> 與本專案「修正用旗標守護、預設不變」的原則一致。

## 其實是原版本來就有、屬玩家記憶誤差的「差異」

- **「重製版像素太大太低解析」**:320×200 是原版真實規格,不是重製版降規。
- **「原版比較銳利清楚」**:相反——原版在 CRT 上是偏**模糊柔化**的;記憶中的「清楚」多半是腦補,
  重製版的銳利其實「比原版更不像原版」。
- **「字型被換掉了」**:沒換,英文沿用 LBX 原點陣字。
- **icon 面板、右鍵資訊卷軸、金色標題字**:都保留,屬忠實還原。

## 來源

- [kazzmir/master-of-magic GitHub](https://github.com/kazzmir/master-of-magic) · [itch.io demo](https://kazzmir.itch.io/magic)
- [scale 套件文件(預設 3.0× / 最近鄰 / 無 CRT 校正)](https://pkg.go.dev/github.com/kazzmir/master-of-magic/game/magic/scale)
- [Felipe Pepe: No, MS-DOS games weren't widescreen(320×200 PAR 1:1.2)](https://felipepepe.medium.com/no-ms-dos-games-weren-t-widescreen-tips-on-correcting-aspect-ratio-37f86343ad65)
- [Hacker News: 320×200 PAR 1:1.2 討論](https://news.ycombinator.com/item?id=20206016)
- [Lilura1: Master of Magic 1994 回顧](https://lilura1.blogspot.com/2021/12/Master-of-Magic-Retrospective-Review.html)
- 本機引擎原始碼(`/tmp/mom-engine`):`game/magic/scale/scale.go`(ScaleAmount=3.0、Normal)、
  `game/magic/main.go:854`(Layout=Scale2(320,200))、`game/magic/data/data.go`(320×200)、
  `game/magic/fonts/fonts.go`(英文 LbxFont)。
