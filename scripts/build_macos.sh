#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

./scripts/preflight.sh

echo
echo "[1/5] install frontend deps"
cd frontend
npm install
npm run build
cd ..

echo
echo "[2/5] tidy go modules"
go mod tidy

echo
echo "[3/5] run wails doctor"
wails doctor

echo
echo "[4/5] build macOS universal app"
wails build -platform darwin/universal

echo
echo "[5/5] package release"
./scripts/package_release.sh

echo
echo "done. artifacts are in build/bin and release"
