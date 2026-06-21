#!/usr/bin/env bash
# 取得 kazzmir/master-of-magic 引擎本體 (不 vendor 進本 repo)。
# 用法:./scripts/fetch-engine.sh [目標目錄,預設 ./engine]
set -euo pipefail

ENGINE_REPO="https://github.com/kazzmir/master-of-magic.git"
# TODO(Phase 0): 釘選已驗證的 commit,避免上游變動破壞 patch。
ENGINE_REF="${ENGINE_REF:-main}"
DEST="${1:-engine}"

if [ -d "$DEST/.git" ]; then
  echo "[fetch-engine] $DEST 已存在,拉取 $ENGINE_REF ..."
  git -C "$DEST" fetch --depth 1 origin "$ENGINE_REF"
  git -C "$DEST" checkout "$ENGINE_REF"
else
  echo "[fetch-engine] clone $ENGINE_REPO ($ENGINE_REF) → $DEST ..."
  git clone --depth 1 --branch "$ENGINE_REF" "$ENGINE_REPO" "$DEST" 2>/dev/null \
    || git clone --depth 1 "$ENGINE_REPO" "$DEST"
fi

echo "[fetch-engine] 完成。引擎位於 $DEST (已列入 .gitignore,不會入 git)。"
echo "[fetch-engine] 後續:套用 patches/ 下的中文化 patch,再 'cd $DEST && go build ./game/magic'。"
