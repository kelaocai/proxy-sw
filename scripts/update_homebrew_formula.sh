#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 3 ]]; then
  echo "Usage: $0 <version> <owner> <repo>"
  exit 1
fi

VERSION="$1"
OWNER="$2"
REPO="$3"

if [[ "$VERSION" != v* ]]; then
  echo "Version must start with 'v' (example: v0.1.0)"
  exit 1
fi

SHORT_VERSION="${VERSION#v}"
BASE_URL="https://github.com/${OWNER}/${REPO}/releases/download/${VERSION}"
ARM_ASSET="proxy-sw_${SHORT_VERSION}_macos_arm64.tar.gz"
AMD_ASSET="proxy-sw_${SHORT_VERSION}_macos_x86_64.tar.gz"
ARM_URL="${BASE_URL}/${ARM_ASSET}"
AMD_URL="${BASE_URL}/${AMD_ASSET}"

wait_for_asset() {
  local url="$1"
  local max_wait_seconds=180
  local elapsed=0
  local code=""
  while true; do
    code="$(curl -s -o /dev/null -w "%{http_code}" "$url" || true)"
    if [[ "$code" == "200" || "$code" == "301" || "$code" == "302" ]]; then
      return 0
    fi
    if (( elapsed >= max_wait_seconds )); then
      echo "Timed out waiting for release asset: $url"
      return 1
    fi
    sleep 5
    elapsed=$((elapsed + 5))
  done
}

sha_from_url() {
  local url="$1"
  curl -L --max-time 120 -s "$url" | shasum -a 256 | awk '{print $1}'
}

wait_for_asset "$ARM_URL"
wait_for_asset "$AMD_URL"
ARM_SHA256="$(sha_from_url "$ARM_URL")"
AMD_SHA256="$(sha_from_url "$AMD_URL")"

cat > packaging/homebrew/proxy-sw.rb <<FORMULA
class ProxySw < Formula
  desc "macOS system proxy manager for local proxy tools like Clash"
  homepage "https://github.com/${OWNER}/${REPO}"
  version "${SHORT_VERSION}"
  license "MIT"

  on_macos do
    on_arm do
      url "${ARM_URL}"
      sha256 "${ARM_SHA256}"
    end

    on_intel do
      url "${AMD_URL}"
      sha256 "${AMD_SHA256}"
    end
  end

  def install
    bin.install "proxy-sw"
  end

  test do
    assert_match "proxy-sw commands", shell_output("#{bin}/proxy-sw --help")
  end
end
FORMULA

echo "Updated packaging/homebrew/proxy-sw.rb for ${VERSION}"
