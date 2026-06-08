#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ABI="${1:-all}"
ANDROID_API="${ANDROID_API:-24}"
GO_BIN_DIR="${GO_BIN_DIR:-$(go env GOPATH)/bin}"
ANDROID_HOME="${ANDROID_HOME:-${ANDROID_SDK_ROOT:-}}"

if [[ -z "${ANDROID_HOME}" ]]; then
  echo "ANDROID_HOME or ANDROID_SDK_ROOT is required" >&2
  exit 1
fi

export ANDROID_HOME
export ANDROID_SDK_ROOT="${ANDROID_SDK_ROOT:-${ANDROID_HOME}}"
export PATH="${GO_BIN_DIR}:${ANDROID_HOME}/cmdline-tools/latest/bin:${ANDROID_HOME}/platform-tools:${ANDROID_HOME}/emulator:${PATH}"

case "${ABI}" in
  arm32) TARGETS=("android/arm") ;;
  arm64) TARGETS=("android/arm64") ;;
  all) TARGETS=("android/arm" "android/arm64") ;;
  *) echo "Usage: $0 [arm32|arm64|all]" >&2; exit 2 ;;
esac

mkdir -p "${ROOT_DIR}/mobile/android/app/libs" "${ROOT_DIR}/dist/android"

echo "Installing gomobile tools"
GOBIN="${GO_BIN_DIR}" go install golang.org/x/mobile/cmd/gomobile@latest
GOBIN="${GO_BIN_DIR}" go install golang.org/x/mobile/cmd/gobind@latest
gomobile init

for target in "${TARGETS[@]}"; do
  suffix="${target#android/}"
  [[ "${suffix}" == "arm" ]] && abi_name="arm32" || abi_name="arm64"
  aar_tmp="${ROOT_DIR}/dist/android/passworder-mobilebridge-${abi_name}.aar"
  echo "Building Android mobile bridge ${abi_name} (${target}) -> ${aar_tmp}"
  (
    cd "${ROOT_DIR}"
    gomobile bind -target="${target}" -androidapi "${ANDROID_API}" -o "${aar_tmp}" passworder/mobile/bridge
  )
  cp "${aar_tmp}" "${ROOT_DIR}/mobile/android/app/libs/passworder-mobilebridge.aar"
done

echo "Done. AAR artifacts are in dist/android/."
