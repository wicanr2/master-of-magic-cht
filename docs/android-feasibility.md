# Android 版可行性評估

結論:**技術可行,但工作量明顯高於桌面三平台**,主要卡在「觸控 UX 重設計」與「版權遊戲檔在沙箱的匯入」,
非編譯問題。建議列為桌面版穩定後的獨立階段。

## 可行的部分

- **Ebiten 原生支援 Android**:go.mod 已有 `github.com/ebitengine/gomobile` 依賴,標準路徑為
  `ebitenmobile bind` 產出 `.aar`,再套進 Android Studio 專案(`EbitenView`)。
- **中文化資產零負擔**:CJK 字型子集 + 譯文表已 `go:embed` 進 Go 程式,Android build 一起帶,
  跨平台行為一致,不需 per-platform 處理。
- **CPU/記憶體**:MoM 是 1994 年遊戲、邏輯輕量,中階手機綽綽有餘。

## 主要成本與風險(由高到低)

1. **觸控 UX 重設計(最大)**:MoM 是密集的滑鼠 + 鍵盤 UI——右鍵、hover 提示、小按鈕、
   多欄位面板。直接搬到觸控會難用。需要:點擊區放大、右鍵改長按、hover 改點選、必要時加縮放/平移手勢。
   這是設計 + 實作的真功夫,非自動化。
2. **版權遊戲檔匯入**:Android 沙箱不能像桌面直接讀任意路徑的 LBX。需做「匯入遊戲檔」流程
   (SAF 檔案選擇器 / 首次啟動引導),讓玩家提供自己的正版 1.60 資料。桌面版是 `-data` 一行,
   Android 要 UI + 儲存權限處理。
3. **打包鏈**:Android SDK/NDK + Gradle + 簽章(自簽即可側載;上架 Play 另需開發者帳號)。
4. **圖片 UI 中文化**:與桌面共用同一未解問題(烘進 LBX 的圖片文字),Android 不會更省。

## 與桌面的差異點

| 面向 | 桌面 | Android |
|---|---|---|
| 編譯 | Linux/Win 免 CGo 交叉編譯;mac 走 CI | `ebitenmobile bind` + Android Studio |
| 輸入 | 滑鼠鍵盤,直接可用 | **需重設計觸控** |
| 遊戲檔 | `-data` 指目錄 / 內含 | **需 SAF 匯入流程** |
| 中文化 | 已 go:embed | 同(共用) |

## 建議

- **先決條件**:桌面三平台穩定、圖片 UI 中文化有結論後再啟動。
- **首版範圍**:`ebitenmobile` 產 .aar + 最小 Android wrapper + 遊戲檔匯入流程,觸控先做「點擊放大 + 長按=右鍵」基本層,
  進階手勢與面板重排列為後續。
- **驗證**:Android emulator(headless 不易,需實機或 GUI emulator)。

預估:桌面已備的前提下,可玩的 Android 首版約是「觸控層 + 匯入流程 + 打包」三塊的中等工程,
遠大於 Windows/macOS 那種「重編譯 + 換殼」。
