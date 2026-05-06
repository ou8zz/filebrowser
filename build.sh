#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT_DIR"

GIT_COMMIT="$(git log -n 1 --format=%h 2>/dev/null || echo unknown)"
VERSION="$1"

build_frontend() {
  if command -v pnpm >/dev/null 2>&1; then
    pnpm -C frontend install --frozen-lockfile
    pnpm -C frontend run build
    return 0
  fi

  if command -v npm >/dev/null 2>&1; then
    npm --prefix frontend install --legacy-peer-deps
    npm --prefix frontend run build
    return 0
  fi

  echo "missing package manager: install pnpm (recommended) or npm" >&2
  return 1
}

build_backend() {
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -X \"github.com/filebrowser/filebrowser/v2/version.Version=${VERSION}\" -X \"github.com/filebrowser/filebrowser/v2/version.CommitSHA=${GIT_COMMIT}\"" \
    -o filebrowser \
    .
}

build_frontend
build_backend

echo 'version:'$1
docker build -t registry.cn-shanghai.aliyuncs.com/ou88zz/filebrowser:$1 .
docker push registry.cn-shanghai.aliyuncs.com/ou88zz/filebrowser:$1

git add .
git commit -m "auto submit $1"
git push
echo $(date "+%Y-%m-%d %H:%M:%S")
