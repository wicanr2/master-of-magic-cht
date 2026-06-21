# cjk-hello — Phase 1 CJK 渲染 prototype

驗證「碼點 ≥ 0x2E80 改用 TTF rasterize 的 glyph」這條注入支線 (對應引擎 `lib/font/font.go`)。

## 跑法 (headless 截圖,docker)

```bash
# 在 repo 根目錄
docker run --rm \
  -v "$PWD/prototype/cjk-hello:/app" \
  -v /usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc:/font/noto.ttc:ro \
  -v "$PWD/proto-out:/out" \
  golang:1.25-bookworm bash -c '
    apt-get update -qq && apt-get install -y -qq xvfb gcc libgl1-mesa-dev \
      libx11-dev libxrandr-dev libxcursor-dev libxinerama-dev libxi-dev libxxf86vm-dev >/dev/null
    cd /app && go mod tidy
    xvfb-run -a go run . -font /font/noto.ttc -out /out/cjk-hello.png'
```

輸出 `proto-out/cjk-hello.png` (見 `docs/img/phase1-cjk-hello.png`)。

## 注意

- Ebiten Linux 桌面需 CGo + X11 dev headers。
- 字型用 `.ttc` 時走 `opentype.ParseCollection`;AR PL UMing 的 ttc 是 CFF/舊式,sfnt 解析失敗,改 Noto。
- 這是 prototype,正式注入見 `patches/0001-cjk-font-injection.patch`。
