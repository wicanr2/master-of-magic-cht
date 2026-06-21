# 工作魔法大帝 繁體中文化 — Master of Magic (CHT)

> 1994 年的奇幻 4X 經典《Master of Magic》(MicroProse),正在被做成**全程繁體中文**可玩。
> 跑在開源 Go 重製引擎 [kazzmir/master-of-magic](https://github.com/kazzmir/master-of-magic) 上。

兩個位面 (Arcanus / Myrror)、十四個魔法書學派、上百種單位與法術、數百件神器——當年沒有官方中文版,
三十年也沒人完整漢化。本專案承接先前《銀河霸主》(Master of Orion) 中文化的方法論,用**自製 CJK
渲染管線 + 載入後字串覆蓋**把遊戲文字一次解掉,且**全程不修改、不散布任何版權檔案**。

> ⚠️ 專案處於 **Phase 0 (起步)**。目前內容為計畫、術語表、引擎架構盤點與第一批物品譯文。
> 進度以 [`PLAN.md`](PLAN.md) 為準。

---

## 這份 repo 放什麼 / 不放什麼

| 放 | 不放 (版權,列入 `.gitignore`) |
|---|---|
| 計畫 `PLAN.md`、術語 `CONTEXT.md`、決策 `docs/adr/` | 原版遊戲檔 (`.lbx` / `.exe` / 手冊) |
| 英文 → 繁中譯文表 `docs/strings/*.tsv` | 引擎本體 (由 `scripts/fetch-engine.sh` 取得) |
| CJK 字型烘製腳本、patch、字型 | 解壓出的任何版權資產 |

玩家自備正版 Master of Magic 資料檔,套用本專案即可。譯文表是本專案的衍生資產。

---

## 技術路線一覽

- **引擎**:Go 1.25 + Ebiten。字串渲染本就走 rune 迴圈,CJK 注入點在 `lib/font/font.go` 的 glyph 查找。
- **渲染**:不縮小中文硬塞 320×200;走 24×24 點陣 atlas (或 TTF 即時渲染),維持 pixel-art 銳利。決策見 [ADR 0001](docs/adr/0001-cjk-rendering.md)。
- **翻譯**:LBX 載入後**覆蓋層**即時替換,版權檔不動;單位名因 hardcode 在 Go source,另以查表處理。

完整盤點與階段規劃見 [`PLAN.md`](PLAN.md)。

---

## 參考

- 引擎:<https://github.com/kazzmir/master-of-magic>
- 前例:銀河霸主 (Master of Orion 1, 1oom) 繁中化
