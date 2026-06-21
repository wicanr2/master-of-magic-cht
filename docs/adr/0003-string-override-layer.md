# ADR 0003 — 字串覆蓋層:在顯示層翻譯,不碰邏輯字串

- 狀態:**Accepted**(第一個 slice 已實作並在真實引擎驗證)
- 日期:2026-06-21
- 相關:[ADR 0002](0002-target-game-version.md)、`docs/strings/*.tsv`、`patches/0001-cjk-font-injection.patch`

## 背景

要讓真實遊戲畫面的英文變中文,且**不修改版權 LBX**。需決定「在哪裡把英文換成中文」。
兩個候選:**資料讀取層**(LBX 讀出字串時就換)vs **顯示層**(font 繪製時才換)。

## 決策:在顯示層 (lib/font) 翻譯

對照引擎原始碼後排除資料讀取層,因為**引擎把英文名稱當邏輯 key 比對**:

- `game/magic/artifact/create-artifact.go:198` → `if string(name) == "+6 Defense" { amount = 6 }`
- 引擎讀 LBX 名稱時 `bytes.Trim(name, " \x00")` (同檔 :149) → 內部字串是 trim 後的英文。

若在資料層把 `name` 換成中文,這類比較與各種 `map[string]...` 查找全會失效 → 破壞遊戲邏輯。
因此**翻譯只在最末端的顯示路徑做**,引擎內部一律維持英文。

實作 (`lib/font/override.go`):在 `lib/font` 的三個繪字進入點
(`doPrint` / `PrintOutline` / `MeasureTextWidth`) 開頭呼叫 `translateForDisplay(text)`,
把要畫的字串換成中文;量測寬度同步翻譯,讓置中/置右正確。

## 翻譯表

- 來源:環境變數 `MOM_CHT_STRINGS` 指向目錄,載入其中所有 `*.tsv`(`docs/strings/` 那批)。
- 比對:**TrimSpace 後精確比對**(對齊引擎 `bytes.Trim` 的正規化形式;故 TSV source 欄即使帶前導空白也能命中)。
- 空譯文 (zh 欄空,如 `artifacts.tsv` 的 TODO) 自動略過 → 該字串保持英文。**覆蓋是選擇性的**。
- 未設定 `MOM_CHT_STRINGS` 或查無對應 → 回傳原字串,不影響英文版。

## 驗證

真實引擎 `test/cjk-render` 餵入引擎實際使用的英文 power 名,覆蓋層自動轉中文
(截圖 `docs/img/phase2-override.png`):`Holy Avenger→神聖復仇者`、`Flaming→烈焰`、
`Vampiric→吸血`、`+3 Attack→+3 攻擊`、`Magic Immunity→魔法免疫`、`Invisibility→隱形`、
`Flight→飛行`;未翻的 `Sword of Mallana` 保持英文。引擎邏輯字串未動。

## 散文換行 (已完成)

`PrintWrap` → `CreateWrappedText` → `splitText`。兩處改動讓散文也能正確中文化換行:

1. **整段翻譯**:`CreateWrappedText` 進入點呼叫 `translateForDisplay(text)`,整段英文先翻成中文
   再斷行 (逐段翻譯無法命中整句譯文)。
2. **CJK 逐字斷行**:重寫 `splitText` — 英文仍在空白處斷行;中文 (無空白) 可在任意字之間逐字斷,
   並保證至少切一個 rune (原版對無空白長串回傳空字串,會讓整段中文被 `CreateWrappedText` 的
   `if show==""{break}` 丟棄)。

驗證 (基於 **Community Patch 1.60** 資料,截圖 `docs/img/phase2-prose-wrap-1.60.png`):
長中文描述「烈焰之劍:每次攻擊都附帶火焰傷害…無人能擋其鋒芒。」自動逐字換成多行;
同畫面 power 名英→中覆蓋、未翻的 `Sword of Mallana` 保持英文。

## 已知限制 / 待辦

- **help.lbx 的格式碼**:help 條目含 0x14 等控制碼,整段 key 比對需連控制碼一起存或先剝離 — 待接 help.tsv 時定。
- **組合字串**(英文片段 + 數值串接後才 Print)無法整串命中 → 視情況在組裝前翻譯或拆詞。
- 翻譯表目前以英文原文為唯一 key;1.31/1.60 同義變體並存策略見 ADR 0002。
