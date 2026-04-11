class ProxySw < Formula
  desc "macOS system proxy manager for local proxy tools like Clash"
  homepage "https://github.com/kelaocai/proxy-sw"
  version "0.1.4"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/kelaocai/proxy-sw/releases/download/v0.1.4/proxy-sw_0.1.4_macos_arm64.tar.gz"
      sha256 "51d35b5af9fb47b9d4103177db7dd6511838694f67589f85bf1d5743d466a6e7"
    end

    on_intel do
      url "https://github.com/kelaocai/proxy-sw/releases/download/v0.1.4/proxy-sw_0.1.4_macos_x86_64.tar.gz"
      sha256 "e3b2c7602ba77453c1aee7e8eb7860785269d94a3d3e4af41bf17b88c3b8ecd1"
    end
  end

  def install
    bin.install "proxy-sw"
  end

  test do
    assert_match "proxy-sw commands", shell_output("#{bin}/proxy-sw --help")
  end
end
