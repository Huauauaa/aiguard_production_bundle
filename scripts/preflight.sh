#!/usr/bin/env bash
set -euo pipefail

check_cmd() {
  if command -v "$1" >/dev/null 2>&1; then
    echo "[ OK ] $1 -> $(command -v "$1")"
  else
    echo "[FAIL] missing command: $1" >&2
    return 1
  fi
}

check_cmd go
check_cmd git
check_cmd npm
check_cmd wails

if ! xcode-select -p >/dev/null 2>&1; then
  echo "[WARN] Xcode Command Line Tools 未安装，请先执行: xcode-select --install" >&2
else
  echo "[ OK ] Xcode Command Line Tools 已安装"
fi

echo
echo "running wails doctor..."
wails doctor
