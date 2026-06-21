# 跨平台移植難度評估 (Android / macOS / Windows)

日期:2026-06-21。結論先講:**維持 Go + Ebiten,不需要重寫成 C++**。

引擎 kazzmir/master-of-magic 用 Go + Ebiten,Ebiten 原生支援 Windows / macOS / Linux /
Android / iOS / Web。**語言不是跨平台瓶頸**;各平台的真正成本在「打包簽章」與「觸控/檔案存取 UX」,
這些換成 C++ 一樣要做、甚至更麻煩 (C++ 沒有 Ebiten 這種一體化跨平台 runtime)。

## 上游已具備的證據

| 證據 | 意義 |
|---|---|
| `Makefile` 有 `wasm` / `itch.io` target,`itch.io/` 內含 `magic.wasm` + `index.html` | **Web 版已可建置並部署** (上游推到 itch.io) |
| go.mod 依賴 `ebitengine/gomobile` | **Android / iOS** 綁定路徑已就緒 |
| go.mod 依賴 `ebitengine/purego` | 桌面多平台**免 CGo**,跨編譯門檻低 |
| go.mod 依賴 `ebitengine/hideconsole` | **Windows** GUI app (隱藏 console) 是既定目標 |
| go.mod 依賴 `oto/v3` | 跨平台音訊後端 |
| `input_js.go` / `fs_js.go` (build tag) | 作者已分平台抽象輸入與檔案系統 |

## 各平台難度

| 平台 | 難度 | 建置方式 | 主要成本 / 風險 |
|---|---|---|---|
| **Linux** | 🟢 低 | `go build ./game/magic` | runtime 需系統庫 (libGL / X11 或 Wayland / libasound)。開發主機即此環境。 |
| **Windows** | 🟢 低 | `GOOS=windows GOARCH=amd64 go build` (Linux 上交叉編譯,purego 免 CGo) | 打包 `.exe` + ico;選擇性 code signing。`hideconsole` 已處理 console。 |
| **Web (WASM)** | 🟢 低 | `GOOS=js GOARCH=wasm`,**上游已在做** | 版權 `.lbx` 不能內嵌 → 需玩家上傳/匯入機制;首次載入體積。 |
| **macOS** | 🟡 中 | `go build` (purego,免 CGo);在 macOS 上產 `.app` | **跨編譯到 mac 不易**,實務用 GitHub Actions `macos-14` runner 建 universal (arm64+x86_64);`.app`/`.dmg` 打包 + (選擇性) 簽章/公證。 |
| **Android** | 🟠 中高 | `ebitenmobile bind` → `.aar`,套進 Android Studio 專案 (`EbitenView`) | ① Android SDK/NDK 工具鏈;② **觸控 UX**:MoM 是密集滑鼠+鍵盤 UI,右鍵/hover 提示需改觸控;③ **版權遊戲檔匯入** (SAF/檔案選擇器),不能內建。 |
| **iOS** | 🟠 中高 | 同 Android (ebitenmobile),另加 Apple 簽章/上架 | 觸控 UX + Apple 開發者帳號/簽章;非首期目標。 |

## 中文化資產在各平台的處理

- **CJK 字型 (我們的資產)**:不論點陣 atlas 或 TTF 子集,都是**本專案自製/自由授權資產**,可隨各平台一起打包,跨平台行為一致,不受版權限制。
- **版權 `.lbx` 遊戲檔**:桌面由玩家指向安裝目錄即可;**Web / Android / iOS 沙箱**限制檔案存取,需提供「匯入遊戲檔」流程 (上傳 / SAF)。這是**引擎層**議題,與中文化、與語言選擇 (Go vs C++) 都無關。
- **字串覆蓋層 + glyph 分流**:純 Go 邏輯,所有平台共用同一份程式碼,不需 per-platform 分支。

## 為什麼不重寫 C++

1. **跨平台會變難不會變易**:Ebiten 提供一體化的 render/input/audio/mobile 綁定;C++ 要自己拼 SDL/ANGLE/各平台音訊 + 自寫 Android JNI 與 iOS 橋接。
2. **打包簽章與觸控 UX 與語言無關**:mac 公證、Android SAF、觸控重設計——換 C++ 一樣全部要做。
3. **丟掉可運作的引擎**:上游已有完整遊戲邏輯 + WASM 部署;重寫等同砍掉重練,風險與工時遠高於收益,違反「正確性/可落地 > 時程」後仍不划算。
4. **中文化注入點更小**:Go 字串迴圈本就走 rune,CJK 只需在 glyph 查找加分流;C++ 從零反而要自建整條文字管線。

**唯一會考慮另尋路徑的情境**:若上游 Ebiton/Ebiten 在某平台出現無法繞過的 runtime 缺陷 (目前無證據)。
在那之前,Go + Ebiten 是正確選擇。

## 建議首期目標順序

1. **Linux + Windows 桌面** (🟢,交叉編譯一次到位) — 先讓中文化在桌面可玩。
2. **Web (WASM)** (🟢,上游已通) — 加「匯入遊戲檔」後即可分享 demo。
3. **macOS** (🟡,GitHub Actions runner) — 補 mac 玩家。
4. **Android** (🟠,觸控 UX 為主要工作) — 列為後期,需專門的觸控介面調整。
