#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ABI="${1:-all}"
BUILD_TYPE="${BUILD_TYPE:-release}"

VERSION_OFFSET="${VERSION_OFFSET:-5000}"
if [[ -z "${BUILD_VERSION_LABEL:-}" ]]; then
  if [[ -n "${VERSION:-}" ]]; then
    version_base="${VERSION}"
  else
    version_base="$(git -C "${ROOT_DIR}" describe --tags --abbrev=0 2>/dev/null || echo v0.0.0)"
  fi
  short_sha="$(git -C "${ROOT_DIR}" rev-parse --short=6 HEAD)"
  export BUILD_VERSION_LABEL="${version_base}+${short_sha}"
fi
if [[ -z "${BUILD_VERSION_CODE:-}" ]]; then
  if git -C "${ROOT_DIR}" rev-parse --verify origin/main >/dev/null 2>&1; then
    commit_count="$(git -C "${ROOT_DIR}" rev-list --count origin/main)"
  else
    commit_count="$(git -C "${ROOT_DIR}" rev-list --count main 2>/dev/null || git -C "${ROOT_DIR}" rev-list --count HEAD)"
  fi
  export BUILD_VERSION_CODE="$((VERSION_OFFSET + commit_count))"
fi

case "${ABI}" in
  arm32) ABI_FILTERS=("armeabi-v7a") ;;
  arm64) ABI_FILTERS=("arm64-v8a") ;;
  all) ABI_FILTERS=("armeabi-v7a" "arm64-v8a") ;;
  *) echo "Usage: $0 [arm32|arm64|all]" >&2; exit 2 ;;
esac

mkdir -p "${ROOT_DIR}/dist/android"

for abi_filter in "${ABI_FILTERS[@]}"; do
  if [[ "${abi_filter}" == "armeabi-v7a" ]]; then
    bridge_abi="arm32"
    artifact_abi="arm32"
  else
    bridge_abi="arm64"
    artifact_abi="arm64"
  fi

  "${ROOT_DIR}/scripts/build-android-server.sh" "${bridge_abi}"

  echo "Building Android ${BUILD_TYPE} APK for ${artifact_abi} (${abi_filter})"
  (
    cd "${ROOT_DIR}/mobile/android"
    ./gradlew ":app:assemble${BUILD_TYPE^}" -PtargetAbi="${abi_filter}"
  )

  src="${ROOT_DIR}/mobile/android/app/build/outputs/apk/${BUILD_TYPE}/app-${BUILD_TYPE}.apk"
  dest="${ROOT_DIR}/dist/android/passworder-android-${artifact_abi}-${BUILD_TYPE}.apk"
  cp "${src}" "${dest}"
  echo "Done: ${dest}"
done
