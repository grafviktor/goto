#!/usr/bin/env bash
# Embeds icon.ico into the Windows binary via a .syso file in cmd/.
# Requires: go install github.com/akavel/rsrc@latest
set -e

arch="${1:?usage: embed-icon.sh <amd64|arm64>}"

cd "$(dirname "$0")"

if [[ ! -f icon.ico ]]; then
  echo "build/exe/icon.ico not found (convert build/exe/icon_source.svg first)" >&2
  exit 1
fi

rsrc -arch "$arch" -ico icon.ico -o "icon_windows_${arch}.syso"
mv -f "icon_windows_${arch}.syso" ../../cmd/goto/
