class Multiplan < Formula
  desc "4-model parallel planning workflow — Claude, Gemini, Codex, GLM-5"
  homepage "https://github.com/cyperx84/multiplan"
  url "https://github.com/cyperx84/multiplan/archive/refs/tags/v0.1.0.tar.gz"
  # sha256 — update after tagging: `brew fetch --build-from-source multiplan` to get hash
  sha256 "PLACEHOLDER_UPDATE_AFTER_RELEASE"
  license "MIT"

  depends_on "node"

  def install
    system "npm", "install", *Language::Node.std_npm_install_args(libexec)
    bin.install_symlink Dir["#{libexec}/bin/*"]
  end

  test do
    assert_match "multiplan", shell_output("#{bin}/multiplan --version")
  end
end
