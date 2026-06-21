#!/usr/bin/env bash
# 取得 kazzmir/master-of-magic 引擎本體 (不 vendor 進本 repo)。
# 用法:./scripts/fetch-engine.sh [目標目錄,預設 ./engine]
set -euo pipefail

ENGINE_REPO="https://github.com/kazzmir/master-of-magic.git"
# 釘選已驗證的引擎 commit (2026-06-03),避免上游變動破壞 patch。
# 需追上游時覆寫 ENGINE_REF=main。
ENGINE_REF="${ENGINE_REF:-0c7669b85df024e77c9a492e7ce7cd1a5b33ff31}"
DEST="${1:-engine}"

if [ ! -d "$DEST/.git" ]; then
  echo "[fetch-engine] clone $ENGINE_REPO → $DEST ..."
  git clone "$ENGINE_REPO" "$DEST"
fi
echo "[fetch-engine] checkout $ENGINE_REF ..."
git -C "$DEST" fetch origin
git -C "$DEST" checkout "$ENGINE_REF"

echo "[fetch-engine] 完成。引擎位於 $DEST (已列入 .gitignore,不會入 git)。"
echo "[fetch-engine] 後續:套用 patches/ 下的中文化 patch,再 'cd $DEST && go build ./game/magic'。"
