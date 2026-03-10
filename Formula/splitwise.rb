# This is a template. GoReleaser generates the real formula automatically.
# It's published to github.com/barronlroth/homebrew-tap/Formula/splitwise.rb

class Splitwise < Formula
  desc "A command-line interface for Splitwise"
  homepage "https://github.com/barronlroth/splitwise-cli"
  license "MIT"

  # GoReleaser fills in the URL, sha256, and version automatically.
  # This template is for reference only.

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/barronlroth/splitwise-cli/releases/download/v#{version}/splitwise_#{version}_darwin_arm64.tar.gz"
    else
      url "https://github.com/barronlroth/splitwise-cli/releases/download/v#{version}/splitwise_#{version}_darwin_amd64.tar.gz"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/barronlroth/splitwise-cli/releases/download/v#{version}/splitwise_#{version}_linux_arm64.tar.gz"
    else
      url "https://github.com/barronlroth/splitwise-cli/releases/download/v#{version}/splitwise_#{version}_linux_amd64.tar.gz"
    end
  end

  def install
    bin.install "splitwise"
  end

  test do
    system "#{bin}/splitwise", "--version"
  end
end
