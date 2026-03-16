class Multiplan < Formula
  desc "4-model parallel planning workflow with lens-based prompts and eval framework"
  homepage "https://github.com/cyperx84/multiplan"
  url "https://github.com/cyperx84/multiplan/archive/refs/tags/v0.3.0.tar.gz"
  sha256 "PLACEHOLDER_UPDATE_AFTER_RELEASE"
  license "MIT"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w"), "-o", bin/"multiplan", "."
  end

  test do
    assert_match "0.3.0", shell_output("#{bin}/multiplan --version 2>&1")
  end
end
