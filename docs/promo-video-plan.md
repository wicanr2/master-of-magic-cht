# 工作魔法大帝 繁中化 — 推廣影片錄製腳本(規劃)

> 狀態:**規劃,尚未錄製**。本文是「拍什麼、怎麼錄、怎麼合成」的可執行劇本,確認後再跑。
> Pipeline 參考:u1-cht《創世紀一代》的 [`docs/llm-promo-video-pipeline.md`](https://github.com/) 三段式
> (擷取 → 素材 → ffmpeg 合成,設計 token 換皮,Ken Burns + 字幕 + 配樂)。本片沿用同一套合成引擎,
> 只換「MoM 的皮(色票/字體/母題)+ 內容(分鏡/字幕)」。
> 目標:**60–75 秒** YouTube / README 短片。主打三件事:**全程繁中**、**重現經典**、**原版 DOS 觀感**。

---

## 0. 與 u1 pipeline 的三點關鍵差異(先講,影響錄製腳本)

| 維度 | u1(創世紀)| 本片(魔法大帝)| 影響 |
|---|---|---|---|
| **操作方式** | 鍵盤驅動(`xdotool key`)| **滑鼠驅動**(選單/法術書/城市畫面都靠點)| 擷取要用 `xdotool mousemove + click`,座標脆弱 → **以靜態截圖為主**(見 §3)|
| **引擎/音訊** | SDL → `SDL_AUDIODRIVER=disk` 直接錄樂 | **Ebiten(Go),非 SDL** | 不能用 SDL disk;改**離線抽 `music.lbx` XMI + SoundFont 算樂**(見 §5,更乾淨)|
| **視覺 DNA** | 古卷 + EGA + 安卡/月之門 | **巫師法典 + 五色魔法 + 哥德血字**(見 §2)| 換一組 token |

---

## 1. 一句話風格定義

**「一本攤開的巫師法典——自然綠、秘術藍、混沌紅、死亡紫、生命白,五色魔法之力在書頁與星空間流轉;
哥德血字標題如咒文浮現。你是爭奪『魔法大帝』王座的十一位巫師之一,而這一切,如今字字繁體中文。」**

張力公式:**哥德奇幻法典(70%)+ 4X 策略征服的恢宏(20%)+ 「全中文」的字幕鉤子(10%)**。
與創世紀的「古卷史詩」區隔:MoM 不是羊皮紙文物感,而是**法師的咒文書 + 五色魔法 + 雙位面征服**。

---

## 2. 設計 token(換皮,直接寫進 ffmpeg/ImageMagick)

```bash
# ===== MoM 設計 token(換皮只改這段)=====
BG='#1a1230'        # 深奧紫夜(主選單星空底)
BG_DEEP='#0c0818'   # 近黑紫,暗角/結尾漸暗
GOLD='#c9a227'      # 哥德鎏金(標題主金,古銅不螢光)
GOLD_HI='#f0d27a'   # 高光金(描邊內側/咒文浮現發光)
GOLD_SH='#7a5c14'   # 暗金陰影(浮雕厚度)
BLOOD='#8c1c13'     # 哥德血紅(標題副色/危機強調,呼應原版 logo 紅)
PARCH='#d8c79a'     # 法典書頁米黃(字幕底卡/卷軸面)
CREAM='#f2ead2'     # 字幕米白(不用純白,避免數位廉價感)
# 五色魔法(MoM 簽名,做色帶/點綴/轉場):
R_NATURE='#3f8f3a'; R_SORCERY='#2f6fc4'; R_CHAOS='#c0392b'; R_DEATH='#6a2a86'; R_LIFE='#e6d8a8'
```

**字體**:
- 中文標題/字幕:**Noto Serif CJK TC**(Black 標題 / Medium 字幕)——西方高奇幻=襯線,撐史詩感(勿用黑體=手遊味)。
- 中文點綴(法術/巫師專名一閃):**TW-Kai 楷書**,增咒文手寫感。
- 英文標題 "Master of Magic":**哥德花體(blackletter,如 UnifrakturMaguntia / Old English Text)**——貼原版 logo;副標可用 Cinzel 碑體。

**母題**(認得出是 MoM 的鉤子):① 翻開的法術書卷軸 ② 五色魔法球/色帶 ③ 巫師法杖與肖像 ④ 雙位面(Arcanus/Myrror)大地圖 ⑤ 哥德花體標題 ⑥ 戰術戰鬥棋盤。

---

## 3. ★錄製腳本(capture)— 本文重點

### 3.1 策略:以「靜態截圖 + Ken Burns」為主,live 片段為輔

MoM 是滑鼠驅動,用 `xdotool` 精準點 UI 座標很脆(retro-game-playtest 的教訓:合成事件常不穩)。
而 u1 的合成段本來就是 **slide(截圖)+ Ken Burns** 為主、live `cap.mp4` 為輔。所以本片:

- **主**:對每個展示畫面**截一張乾淨高解析 PNG**,合成時做緩慢 Ken Burns。穩、可重跑、可逐張挑。
- **輔(可選)**:1–2 段短 live 片段(大地圖捲動、戰鬥動畫),用 `xdotool` 驅動 + `ffmpeg x11grab` 錄;
  若 xdotool 不穩就放棄 live,全用截圖(不影響成片)。

### 3.2 截圖取得方式(兩條路,皆已有基礎)

1. **跑實機 + 定格截圖**:docker `mom-art-build` + Xvfb,跑含資料 AppImage,用既有 `-dosaspect`/`-crt` flag
   與 `SHOT`/`SHOTFRAME` env(`test/cht-settings`、`test/combat` 已有此機制)截 PNG。
2. **沿用既有截圖**:`repo/docs/img/` 已有 `playtest-mainmenu`、`final-overworld`、`armylist-cht-names`、
   `citylist-cht-names`、`help-builder-cht`、`ui-aspect-compare`、`crt-compare` 等——部分鏡頭直接用。

> 新截圖務必用**新版含資料 build**(才有本次補譯的法術書/建築中文)。截圖**保留遊戲原色不調色**,只在合成時加金框。

### 3.3 Shot list(鏡頭清單,★=主打中文化,務必新截)

| # | 畫面 | 怎麼到 | 重點 | 來源 |
|---|---|---|---|---|
| S1 | 主選單(全中文)| 開啟即是 | 快速開始/繼續/讀取/新遊戲 + 哥德標題 | 已有 `playtest-mainmenu` |
| S2 | 選巫師/種族 | 新遊戲流程 | 巫師肖像 + 中文選項 | 新截 |
| S3 | 雙位面大地圖 | 進遊戲 | 中文 HUD(金/食/法力、城市名)| 已有 `final-overworld` |
| S4 | ★**法術書施法卷軸** | 點「法術」鈕 → 選法術 | **法術名(魔法精靈/娜迦/浮空島)+ 翻開的中文說明卷軸**(本次修好)| **新截(關鍵)** |
| S5 | ★**城市建造畫面** | 點城市 → 建造 | **建築清單 + 中文建築描述(城牆/市集/神龕)**(本次修好)| **新截(關鍵)** |
| S6 | 魔法總覽 / 五色法術書 | 點「魔法」鈕 | 五色魔法分頁,呼應 token | 新截 |
| S7 | 戰術戰鬥 | 觸發戰鬥 | 中文單位名 + 完成/巡邏/等待/建造 | 新截(可配 live)|
| S8 | 軍隊/英雄清單 | 點「軍隊」鈕 | 中英並存名(羅潘(Lo Pan)格式)| 已有 `armylist-cht-names` |
| S9 | 設定:原版觀感 | 遊戲→設定 | **DOS 原版長寬比 / CRT 質感** 切換 | 已有 `crt-compare`/`ui-aspect-compare` |

### 3.4 capture 腳本骨架(規劃,尚未跑)

```bash
#!/usr/bin/env bash
# capture_promo.sh — 規劃版:逐畫面截 PNG(實機 + Xvfb),live 片段可選。
set -u
APP=dist/MasterOfMagic-CHT-x86_64.AppImage   # 含資料、含本次補譯
OUT=promo/shots; mkdir -p "$OUT"
DISP=:97; WH=1920x1440                        # 960x720 的 2x,夠清晰做 Ken Burns

run_shot(){  # $1=旗標  $2=SHOTFRAME  $3=輸出名 —— 需引擎支援對應 SHOT 點(部分待加 hook)
  Xvfb $DISP -screen 0 ${WH}x24 >/dev/null 2>&1 & XP=$!; sleep 1
  DISPLAY=$DISP APPIMAGE_EXTRACT_AND_RUN=1 SHOT="$OUT/$3.png" SHOTFRAME="$2" \
    "$APP" $1 -music=false >/dev/null 2>&1 &
  sleep 8; kill %1 $XP 2>/dev/null
}
# S4/S5/S6/S7 需要「開到特定畫面」:
#   方案 A(穩):為 promo 加極簡 demo hook(沿用 cht-settings/combat 的 SHOT 機制),
#               直接渲染法術書/城市/戰鬥畫面到 PNG —— 工程量小、最可靠。
#   方案 B(快):xdotool 滑鼠點到對應畫面再截(座標脆,需逐張看圖校正)。
# live(可選):ffmpeg -f x11grab -video_size 960x720 -framerate 25 -i $DISP -t 5 ...(大地圖捲動/戰鬥)
```

> **建議**:S4/S5/S6/S7 用**方案 A(加 promo 截圖 hook)**——我們已有 `test/cht-settings`、`test/combat`
> 這類「建 UI → 第 N 幀存 PNG」的 harness,複製成 `test/promo-spellbook` / `test/promo-cityview` 即可穩定產圖,
> 不必跟 xdotool 滑鼠座標纏鬥。這是本片**唯一需要的少量引擎側工作**。

---

## 4. 分鏡 + 中文字幕(70 秒)

節奏:**慢起(法典)→ 漸快(亮點蒙太奇)→ 收於莊嚴**。字幕走史詩奇幻 + 「全中文」鉤子。

### 段 1｜標題卡(0–8s)
- 視覺:`#1a1230` 星空底 + 中央哥德花體 "Master of Magic" 鎏金浮雕,下方中文「工作魔法大帝」,
  兩側五色魔法球緩亮,底部一行半透明咒文。
- 動態:標題淡入 + 極輕放大,金高光掃一次。
- 字幕:標題本身 +(下緣小字)「重現經典 · 全程繁體中文」。

### 段 2｜世界觀引子(8–20s,大地圖 Ken Burns）
- 字幕逐句(襯線米白):
  - 「十一位巫師,爭奪掌控一切魔法的王座——『魔法大帝』。」
  - 「召喚生物、研發法術、橫跨 Arcanus 與 Myrror 雙重位面。」
  - 「而這一次,字字繁體中文。」

### 段 3｜亮點蒙太奇(20–58s,核心)
每鏡 4–6s,Ken Burns + 金框 + 字幕。順序由「介面」到「魔法」到「征服」到「年代感」:

| 鏡 | Shot | 字幕(史詩 + 中文化鉤子)|
|---|---|---|
| 1 | S3 大地圖 | 「踏遍雙重位面的廣袤疆土」 |
| 2 | S4 法術書 ★ | 「翻開法典——**連施法卷軸的每一句說明,都是繁中**」 |
| 3 | S6 五色魔法 | 「自然、秘術、混沌、死亡、生命——駕馭五色之力」 |
| 4 | S5 城市建造 ★ | 「經營城市,**建築說明字字看得懂**」 |
| 5 | S7 戰術戰鬥 | 「於棋盤上指揮軍隊,以法術扭轉戰局」 |
| 6 | S9 原版觀感 | 「想要年代感?一鍵切回 DOS 原版長寬比與 CRT 掃描線」 |

> 鏡 2、4 是本次修補的**法術書/建築中文化**,給足停留——這是相對舊版最有感的進步。
> 鏡 6 帶出 DOS 長寬比 / CRT 賣點(可直接用 `crt-compare.png` 的對比動態)。

### 段 4｜結尾卡(58–70s)
- 視覺:漸暗回 `#0c0818`,中央哥德花體標題再現 + 五色魔法球收束成一點金光。
- 副標:「繁體中文版 · 免費開源 · 玩家自備遊戲檔」。
- 底部小字:GitHub `wicanr2/master-of-magic-cht`。
- 音樂收束長音,定格 2s 後淡出。

---

## 5. 配樂(用 remake 自己的合成音樂)

> **決策**:採用 remake 的合成配樂(使用者 2026-06-28 確認可用)。
> 註:依鐵則 `rulebook/93-promo-video-original-assets.md`,推廣片配樂預設應用**原版實機音訊**
> (MoM 原版 DOS 是 AdLib/OPL2 或 MT-32 FM);remake 用 `music.lbx` 的 **XMI(MIDI)+ SoundFont 即時合成**
> (`game/magic/music/music.go`),屬「重編渲染」而非原版晶片輸出。**此處經使用者明確同意採用 remake 版**,故不另錄原版。

**取得方式(乾淨、可重現,優於 live 錄音)**:
1. 用引擎的 `xmi.ReadMidiFromCache(cache, "music.lbx", <index>)` 把標題/大地圖曲的 **XMI 抽成標準 MIDI**。
2. 用 remake 實際載入的同一顆 **SoundFont(.sf2/.sf3)** + `fluidsynth` **離線算成 wav**:
   `fluidsynth -ni -F title.wav <soundfont>.sf2 title.mid`。
   → 得到與遊戲內**一模一樣**的音色,且無 Xvfb 無音效卡的雜訊問題。
3. 結構對齊分鏡:前奏對標題(段1)、主題展開鋪亮點段(段3)、收束長音對結尾(段4);
   `ffmpeg afade` 淡入(2s)淡出(3s)。

> 若日後想改用原版 FM 音色,可在 DOSBox 跑 `original_game/msdos_mom.zip` 設 AdLib/MT-32 錄製替換 —— 屬鐵則 93 的標準路徑,不影響本 pipeline 其餘部分。

---

## 6. 合成(compose)— 直接沿用 u1 的 ffmpeg 骨架

`make_promo.sh` = u1 的 `make_gameplay_video.sh` 換上 §2 的 MoM token + §4 分鏡:
- `card()` / `slide()` / `kenburns()` 函式不動;把 `BG/GOLD/FONT...` 換成 MoM 五行 token。
- 標題卡用哥德花體 + 五色魔法球 overlay;轉場用「五色魔法球放大吞畫面」取代 u1 的月之門光環。
- 截圖加細金框(`#c9a227`),保原色;字幕走 Noto Serif CJK TC。
- concat → 鋪 §5 的 wav(afade)→ `-movflags +faststart` 出 `dist/video/momcht-promo.mp4`。

---

## 7. 注意 / 驗證 / 待辦

- **唯一引擎側工作**:為 S4/S5/S6/S7 加 `test/promo-*` 截圖 harness(沿用 cht-settings/combat 的 SHOT 機制),
  比 xdotool 滑鼠穩。其餘全是腳本 + 素材。
- **字幕別被裁**:中文全形寬,合成後**算 3–4 幀丟回來看圖**,確認標題不糊、字幕不裁、五色不偏。
- **截圖用新版含資料 build**:才有本次補譯的法術書/建築中文(舊截圖會露英文)。
- **有界執行**:擷取固定秒數 SIGKILL 收尾(無頭/CI 友善)。
- **節奏**:前段 2–4s/鏡莊嚴,亮點段 1.5–2.5s 卡樂句,Ken Burns 幅度小勻速,忌甩鏡。
- **音色查證**(鐵則 93-2):算完樂先 `ffprobe` 看時長/bitrate/聲道正常、非空白檔再用。

> 本文為**規劃**;確認分鏡與字幕後,再依序產:① promo 截圖 harness → ② capture 截圖 → ③ 抽樂算樂 →
> ④ make_promo.sh 合成 → ⑤ 看圖迭代 → ⑥ 出片放 README/YouTube。
