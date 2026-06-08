#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT="${ROOT_DIR}/dist/passworder-darwin-amd64"
if [[ -n "${BUILD_VERSION_LABEL:-}" ]]; then
  VERSION="${BUILD_VERSION_LABEL}"
elif [[ -n "${VERSION:-}" ]]; then
  version_base="${VERSION}"
  short_sha="$(git -C "${ROOT_DIR}" rev-parse --short=6 HEAD)"
  VERSION="${version_base}+${short_sha}"
else
  version_base="$(git -C "${ROOT_DIR}" describe --tags --abbrev=0 2>/dev/null || echo v0.0.0)"
  short_sha="$(git -C "${ROOT_DIR}" rev-parse --short=6 HEAD)"
  VERSION="${version_base}+${short_sha}"
fi

mkdir -p "$(dirname "${OUT}")"

echo "Building macOS amd64 -> ${OUT}"
(
  cd "${ROOT_DIR}"
  GOOS=darwin GOARCH=amd64 CGO_ENABLED="${CGO_ENABLED:-1}" \
    go build -trimpath -ldflags="-s -w -X passworder/internal/embedded.Version=${VERSION}" -o "${OUT}" ./cmd/server
)

echo "Done: ${OUT}"
