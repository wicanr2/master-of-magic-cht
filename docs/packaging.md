# 打包 (Packaging)

各平台從 patch-only repo 組成可執行包的流程。**完整包內含版權 1.60 遊戲檔,僅供自用,不入公開 repo**;
本 repo 只放建置腳本與中文化資產。

共同前置:`scripts/fetch-engine.sh`(取引擎,釘選 commit)→ `git -C engine apply patches/0001 patches/0002`
→ `scripts/prepare-embed.sh engine`(放入字型子集 + 譯文表供 go:embed)。

## Linux — AppImage ✅

- `CGO_ENABLED=1 go build -o magic ./game/magic`(Ebiten Linux 需 CGo + X11/GL,docker 內裝 dev headers)。
- AppDir:`usr/bin/magic` + `usr/share/mom-cht/data/`(全 1.60 LBX)+ bundle libGL/X11/asound + AppRun(`-data` 指向內含資料)。
- `appimagetool AppDir`(`APPIMAGE_EXTRACT_AND_RUN=1`)→ `MasterOfMagic-CHT-x86_64.AppImage`(~40MB)。
- **已驗證**:headless 啟動、開局、CJK 渲染、可玩(見 `playtest-validation.md`)。

## Windows — zip(免 DLL)✅

- `CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-H windowsgui -s -w" -o magic.exe ./game/magic`。
- **Ebiten Windows 免 CGo**(DirectX via purego)→ **不需任何外部 DLL**(import 表僅 `kernel32.dll`,其餘系統 DLL 執行時動態載入)。
- 包:`magic.exe` + `data/`(全 1.60 LBX)+ `玩遊戲.bat`(`magic.exe -data data`)+ 說明 → zip(~33MB)。
- 從 Linux 交叉編譯;Windows 端實機執行待使用者驗證。

## macOS — 需 GitHub Actions ⚠️

- **不能從 Linux 交叉編譯**:Metal/Cocoa 圖形驅動需 objc/cgo + macOS SDK
  (`CGO_ENABLED=0` 會缺 `finishDrawableUsage`/`nextDrawable` 等;`CGO_ENABLED=1` 需 mac clang)。
- 走 **`.github/workflows/macos.yml`**(macos-14 runner):fetch 引擎 → apply patches → prepare-embed →
  `go build`(arm64,CGo)→ 組 `.app` → 上傳 artifact。下載後本地加入 1.60 遊戲檔組完整 `.app`/`.dmg`。
- universal(arm64+x86_64)為後續;首版先 arm64(Apple Silicon)。

## Android — 評估中

見 task / 後續評估文件(`ebitenmobile bind` + 觸控 UX + 沙箱資料匯入)。
