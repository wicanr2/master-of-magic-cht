# ADR 0002 — 目標遊戲版本與 Community Patch

- 狀態:**Accepted**(主目標 CP 1.60,1.31 相容;已實測版本差異)
- 日期:2026-06-21(2026-06-21 更新:完成 1.31→1.60 升級與字串 diff)
- 相關:[ADR 0001](0001-cjk-rendering.md)、`docs/strings/*.tsv`

## 背景

Master of Magic 有多個版本:官方最終 DOS 版 **v1.31**(Simtex/MicroProse, 1995),社群維護的
非官方/Community Patch **v1.40n → v1.60**(目前社群現行版,由 drake178 等維護)。

兩項已查證事實決定本決策:

1. **kazzmir 引擎朝 Community Patch 開發。** 遊戲邏輯實作 1.50/1.60 行為:
   - `game/magic/player/relations.go:11` → `// 1.50 patch formula`
   - `game/magic/city/city_test.go:705,789` → `// Test against values from a city screen of original MoM v1.60`
   引擎 README 不硬性指定版本,只要求「放入 MoM 的 lbx」,即**讀玩家現有資料**。
2. **本專案目前手上的資料是 vanilla v1.31。** `extracted/magic.exe` 內含
   `Copyright Simtex Software, 1995 V1.31`,且 `help.lbx` 無 CP 標誌的第 808 筆文件 entry。

差距:引擎邏輯=1.60,我們萃取字串的資料=1.31。字串覆蓋層需**精確比對原文**,版本不一致 → 漏譯/不命中。

## 升級路徑(已知工具)

`original_game/community_patch.txt` = **MOMDIFFP** 差分補丁器的說明。它能把正版 MoM 在
1.20/1.30/1.31/1.40n/1.50/1.51/1.52.03/**1.60** 間互轉,不需 Slitherine launcher;
也能降級到 1.00/1.01/1.10。符合本專案「玩家自備正版資料」模型,可納入玩家安裝說明。
注意:該套件**不含** changelog / 字串 / FILESET patch(明示請查 masterofmagic.fandom.com wiki);
且本 repo 目前只有說明 `.txt`,未有 `MOMDIFFP.EXE` 本體——要產 1.60 資料需另取 patcher 並於 DOSBox/Windows 執行。

## 決策

1. **主目標 = Community Patch v1.60**(對齊引擎邏輯與測試基準),於玩家文件建議升級到 1.60。
2. **保留 v1.31 相容**,做法:**譯文表以「英文原文字串」為 key**。
   - 1.31 與 1.60 相同的字串:一條譯文同時覆蓋兩版。
   - 1.60 改過/新增的字串:加一條變體,對應同一中文。
   - → 單一份 `docs/strings/*.tsv` 支援兩個版本,且現有 1.31 萃取不浪費。
3. **現有 1.31 萃取為 baseline**(`item-powers.tsv` 64、`artifacts.tsv` 250);取得 1.60 資料後
   做 string diff,補上 delta(改名/新增的 item power、法術、help 等)。

## 實測版本差異 (1.31 → 1.60)

用 MOMDIFFP (DOSBox in docker) 將正版 1.31 升級到 v1.60 (`MAGIC.EXE` 內 `Seravy, Drake178, 2021 v1.60`),
對升級前後字串做 diff。**升級結果另存本地 `original_game/msdos_mom_cp160.zip` 供比對 (版權資料,不入庫)。**

CP 1.60 改寫了這些字串相關 LBX (DOSBox 重寫成大寫 8.3 檔名):
`BUILDESC HELP HLPENTRY SPELLDAT DIPLOMSG MESSAGE NEWGAME UNITVIEW MAGIC ARMYLIST ITEMPOW FONTS …`

關鍵 diff 結果:

| 類別 | 檔案 | 名稱字串 1.31→1.60 | 結論 |
|---|---|---|---|
| 物品能力名 | `itempow.lbx` | 66→66,**0 差異** (md5 變=只改數值) | `item-powers.tsv` 兩版通用 ✅ |
| 神器名 | `itemdata.lbx` | 250→250,**0 差異** (檔未被 patch) | `artifacts.tsv` 兩版通用 ✅ |
| 法術名 | `spelldat.lbx` | 214→214,**0 差異** (data 改、名稱不變) | 法術名表兩版通用 ✅ |
| Help 文字 | `HELP/HLPENTRY.LBX` | **大幅改寫** (散文) | 須從 1.60 萃取 |
| 建築描述 | `BUILDESC.LBX` | 改寫 (散文) | 須從 1.60 萃取 |
| 訊息/外交 | `MESSAGE/DIPLOMSG.LBX` | 改寫 (散文,含動態 token) | 須從 1.60 萃取 |
| 字型 | `FONTS.LBX` | CP 改了 ASCII 字型 | CJK 注入在 ASCII 之上獨立,不受影響 |

**核心結論**:CP 1.60 維持所有**名稱**字串不變 (改的是數值/平衡 + 重寫散文)。
→ 名稱類譯文表 (神器/物品能力/法術) **版本無關,一份通殼兩版,現有萃取不必重做**;
須以 1.60 為準萃取的只有**散文類** (help / 建築描述 / 訊息)。

## 待辦 / 風險

- [x] 取得實際 1.60 LBX、升級、對 1.31 做字串 diff (見上)。
- [x] 量化差異:名稱不變,散文 (help/desc/message) 改寫。
- [ ] 散文類譯文表 (help / buildings / messages) 以 1.60 為來源萃取。

> **2026-06-28 事故根因(玩家回報法術/建築畫面英文)**:上面這條「散文類萃取」TODO **從未完成** ——
> 這就是法術 help 卷軸、建造畫面建築描述大量顯示英文的主因。三個具體缺口:
> 1. **`desc.lbx`(法術書施法卷軸的法術描述,215 條)整類從未 dump/翻譯** —— `cht-dump` 與方法論 §4
>    的 dump 源清單當初**漏列**這個源(已補,見 `localization-methodology.md` §4 [HARD] 清單)。
> 2. **`help.lbx` 868 條只有 393 條 key 命中當前 `extracted/` dump(475 條過時)** —— `help.tsv` 的 source 欄
>    不是從實機 reader dump 的字面(混了 wiki/舊版措辭),與「散文須從實際資料萃取」決策不符。
> 3. **`buildesc.lbx` ~13 條建築描述缺**。
> **修法**:一律改用「從當前 `extracted/` 實機 reader dump 的精確字面」當 key(方法論 §2 [HARD] key 對齊),
> 重新對齊 + 補譯 desc/buildesc;並把缺漏的 dump 源補進 `cht-dump`。
- [ ] **檔名大小寫**:DOSBox 把 patch 過的檔寫成大寫 (`HELP.LBX`),未變的維持小寫;
      Linux 檔案系統大小寫敏感,實際拿 1.60 資料餵引擎時需統一檔名大小寫 (engine 載入測試時處理)。
- 引擎 commit 已釘選 (`0c7669b`);CP 釘選 1.60。風險:雙邊持續演進,日後需重驗。
