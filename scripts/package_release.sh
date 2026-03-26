#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

mkdir -p release
STAMP="$(date +%Y%m%d-%H%M%S)"
APP_PATH="build/bin/AIGuard.app"
ZIP_PATH="release/AIGuard-macos-${STAMP}.zip"

if [ ! -d "$APP_PATH" ]; then
  echo "build/bin/AIGuard.app not found, please build first" >&2
  exit 1
fi

ditto -c -k --keepParent "$APP_PATH" "$ZIP_PATH"
echo "packaged: $ZIP_PATH"
