class ProxySw < Formula
  desc "macOS system proxy manager for local proxy tools like Clash"
  homepage "https://github.com/kelaocai/proxy-sw"
  version "0.1.1"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/kelaocai/proxy-sw/releases/download/v0.1.1/proxy-sw_0.1.1_macos_arm64.tar.gz"
      sha256 "395592cb89ef5d5147dd65ff6e2d6a58bd930460705c056f49780a3276567543"
    end

    on_intel do
      url "https://github.com/kelaocai/proxy-sw/releases/download/v0.1.1/proxy-sw_0.1.1_macos_x86_64.tar.gz"
      sha256 "7518479c0d70ea11152b8b80ce99b5da977a108cff3a3c12401360a70300b9ef"
    end
  end

  def install
    bin.install "proxy-sw"
  end

  test do
    assert_match "proxy-sw commands", shell_output("#{bin}/proxy-sw --help")
  end
end
