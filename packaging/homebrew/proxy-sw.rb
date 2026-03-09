class ProxySw < Formula
  desc "macOS system proxy manager for local proxy tools like Clash"
  homepage "https://github.com/kelaocai/proxy-sw"
  version "0.0.0"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/kelaocai/proxy-sw/releases/download/v0.0.0/proxy-sw_0.0.0_macos_arm64.tar.gz"
      sha256 "REPLACE_ARM64_SHA256"
    end

    on_intel do
      url "https://github.com/kelaocai/proxy-sw/releases/download/v0.0.0/proxy-sw_0.0.0_macos_x86_64.tar.gz"
      sha256 "REPLACE_AMD64_SHA256"
    end
  end

  def install
    bin.install "proxy-sw"
  end

  test do
    assert_match "proxy-sw commands", shell_output("#{bin}/proxy-sw --help")
  end
end
