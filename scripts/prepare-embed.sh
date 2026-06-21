#!/usr/bin/env bash
# 把中文化內嵌資產 (CJK 字型子集 + 譯文表) 放進引擎 lib/font/,供 go:embed 編譯。
# 在 fetch-engine.sh 取得引擎、套用 patches/0001 之後、go build 之前執行。
# 用法:./scripts/prepare-embed.sh <engine 目錄,預設 ./engine>
set -euo pipefail

REPO="$(cd "$(dirname "$0")/.." && pwd)"
ENGINE="${1:-engine}"
DST="$ENGINE/lib/font"

if [ ! -d "$DST" ]; then
  echo "[prepare-embed] 找不到 $DST,請先 fetch-engine 並確認路徑" >&2
  exit 1
fi

# 1) CJK 字型子集 (go:embed cht_font.otf)
cp "$REPO/assets/cht-subset.otf" "$DST/cht_font.otf"

# 2) 譯文表 (go:embed cht_strings/*.tsv)
mkdir -p "$DST/cht_strings"
cp "$REPO"/docs/strings/*.tsv "$DST/cht_strings/"

echo "[prepare-embed] 已放置:"
echo "  $DST/cht_font.otf ($(wc -c < "$DST/cht_font.otf") bytes)"
echo "  $DST/cht_strings/ ($(ls "$DST/cht_strings" | wc -l) 個 tsv)"
echo "[prepare-embed] 接著:cd $ENGINE && go build -buildvcs=false ./game/magic"
