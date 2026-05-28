#!/usr/bin/env bash
set -euo pipefail

# Add vscode to docker group
sudo usermod -aG docker vscode

# Fix .claude dir ownership so vscode can create subdirs (e.g. session-env)
sudo chown vscode:vscode /home/vscode/.claude
# Fix named volume ownership (vscode-claude-plugins mounts as root)
sudo chown -R vscode:vscode /home/vscode/.claude/plugins

# Claude: installMethod=native in .claude.json expects binary at ~/.local/bin/claude
# Container installs via npm — create symlink to satisfy native path
mkdir -p /home/vscode/.local/bin
CLAUDE_BIN=$(command -v claude)
ln -sf "${CLAUDE_BIN}" /home/vscode/.local/bin/claude

# Verify tool versions
cdk8s --version
pulumi version
talosctl version --client

# Install Claude plugins into named volume (idempotent — volume persists across rebuilds)
# || true: "already installed/added" exits non-zero; that is a success state here
claude plugin marketplace add anthropics/claude-plugins-official || true
claude plugin install gopls-lsp@claude-plugins-official || true
claude plugin marketplace add JuliusBrussee/caveman || true
claude plugin install caveman@caveman || true
claude plugin marketplace add thedotmack/claude-mem || true
claude plugin install claude-mem@thedotmack || true
