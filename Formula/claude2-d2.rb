# This file is updated automatically by GoReleaser on each release.
# Do not edit the url or sha256 manually.
class Claude2D2 < Formula
  desc "Connects Claude Code lifecycle events to a Sphero R2-D2 droid"
  homepage "https://github.com/peterfox/claude2-d2"
  url "https://github.com/peterfox/claude2-d2/archive/refs/tags/v0.1.0.tar.gz"
  sha256 "0000000000000000000000000000000000000000000000000000000000000000"
  license "MIT"
  head "https://github.com/peterfox/claude2-d2.git", branch: "main"

  depends_on "go" => :build
  depends_on :macos

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w"), "-o", bin/"r2", "."
  end

  service do
    run [opt_bin/"r2", "daemon"]
    keep_alive true
    log_path "/tmp/r2d2.log"
    error_log_path "/tmp/r2d2.log"
  end

  def caveats
    <<~EOS
      To get started, power on your R2-D2 and run:
        r2 setup

      This scans for your droid over Bluetooth and saves its address to ~/.r2d2.
      Only needed once.

      Then start the daemon:
        brew services start claude2-d2

      Finally, install the Claude Code plugin so hooks are wired up automatically:
        Open Claude Code and run: /plugin install peterfox/claude2-d2 https://github.com/peterfox/claude2-d2

      R2-D2 will react to Claude Code lifecycle events in any session.
    EOS
  end

  test do
    assert_match "r2", shell_output("#{bin}/r2 --help 2>&1")
  end
end
