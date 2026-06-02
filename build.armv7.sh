#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT_DIR"

if [[ $# -lt 1 ]]; then
  echo "usage: $0 <version>" >&2
  exit 1
fi

GIT_COMMIT="$(git log -n 1 --format=%h 2>/dev/null || echo unknown)"
VERSION="$1"
IMAGE="${IMAGE:-registry.cn-shanghai.aliyuncs.com/ou88zz/filebrowser:${VERSION}-armv7}"
PLATFORM="${PLATFORM:-linux/arm/v7}"
ALPINE_IMAGE="${ALPINE_IMAGE:-docker.m.daocloud.io/library/alpine:3.23}"
BUSYBOX_IMAGE="${BUSYBOX_IMAGE:-docker.m.daocloud.io/library/busybox:1.37.0-musl}"
export GOCACHE="${GOCACHE:-${ROOT_DIR}/.cache/go-build}"

build_frontend() {
  if command -v pnpm >/dev/null 2>&1; then
    pnpm -C frontend install --frozen-lockfile
    pnpm -C frontend run build
    return 0
  fi

  if command -v npm >/dev/null 2>&1; then
    npm --prefix frontend run build
    return 0
  fi

  echo "missing package manager: install pnpm (recommended) or npm" >&2
  return 1
}

build_backend_armv7() {
  CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build \
    -ldflags="-s -w -X \"github.com/filebrowser/filebrowser/v2/version.Version=${VERSION}\" -X \"github.com/filebrowser/filebrowser/v2/version.CommitSHA=${GIT_COMMIT}\"" \
    -o filebrowser \
    .
}

build_frontend
build_backend_armv7

echo "compiled binary:"
file filebrowser

echo "building ${IMAGE} for ${PLATFORM}"
docker buildx build \
  --platform "${PLATFORM}" \
  --progress=plain \
  -f Dockerfile.armv7 \
  --build-arg "ALPINE_IMAGE=${ALPINE_IMAGE}" \
  --build-arg "BUSYBOX_IMAGE=${BUSYBOX_IMAGE}" \
  -t "${IMAGE}" \
  --load \
  .
