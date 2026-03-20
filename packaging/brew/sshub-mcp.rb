# Homebrew formula (отдельный tap-репозиторий).
# После релиза на GitHub: подставь OWNER/REPO, version, sha256 для каждого архива.
#
#   shasum -a 256 dist/sshub-mcp_darwin_arm64.tar.gz
#
# URL: https://github.com/OWNER/REPO/releases/download/v#{version}/sshub-mcp_....tar.gz
#
class SshubMcp < Formula
  desc "Local MCP HTTP server for SSH projects and remote commands"
  homepage "https://github.com/OWNER/REPO"
  version "0.0.0"

  if OS.mac? && Hardware::CPU.arm?
    url "https://github.com/OWNER/REPO/releases/download/v#{version}/sshub-mcp_darwin_arm64.tar.gz"
    sha256 "REPLACE_SHA256_DARWIN_ARM64"
  elsif OS.mac?
    url "https://github.com/OWNER/REPO/releases/download/v#{version}/sshub-mcp_darwin_amd64.tar.gz"
    sha256 "REPLACE_SHA256_DARWIN_AMD64"
  elsif OS.linux? && Hardware::CPU.arm?
    url "https://github.com/OWNER/REPO/releases/download/v#{version}/sshub-mcp_linux_arm64.tar.gz"
    sha256 "REPLACE_SHA256_LINUX_ARM64"
  else
    url "https://github.com/OWNER/REPO/releases/download/v#{version}/sshub-mcp_linux_amd64.tar.gz"
    sha256 "REPLACE_SHA256_LINUX_AMD64"
  end

  def install
    bin.install "sshub-mcp"
  end

  service do
    run [opt_bin/"sshub-mcp"]
    environment_variables SSHUB_MCP_HTTP_ADDR: "127.0.0.1:8787"
    keep_alive true
    log_path var/"log/sshub-mcp.log"
    error_log_path var/"log/sshub-mcp.err.log"
  end
end
