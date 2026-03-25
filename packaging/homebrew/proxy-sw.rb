class ProxySw < Formula
  desc "macOS system proxy manager for local proxy tools like Clash"
  homepage "https://github.com/kelaocai/proxy-sw"
  version "0.1.3"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/kelaocai/proxy-sw/releases/download/v0.1.3/proxy-sw_0.1.3_macos_arm64.tar.gz"
      sha256 "f53f340f025d104d425eba074dcc160c70dfc6303a6fd63468bb3b112a98cceb"
    end

    on_intel do
      url "https://github.com/kelaocai/proxy-sw/releases/download/v0.1.3/proxy-sw_0.1.3_macos_x86_64.tar.gz"
      sha256 "891b5e2d383699aaf5f8a0e12206b7f188964d5994c2c17add87bc9eb0ee0ec3"
    end
  end

  def install
    bin.install "proxy-sw"
  end

  test do
    assert_match "proxy-sw commands", shell_output("#{bin}/proxy-sw --help")
  end
end
