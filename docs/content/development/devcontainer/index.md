+++
title = "Devcontainer"
description = "VS Code devcontainer walkthrough: tools, mounts, and environment variables."
weight = 10
+++

The devcontainer is defined in `.devcontainer/devcontainer.json`. It provides a fully configured Go development environment with all homelab tooling pre-installed.

## Base Image

Built from `.devcontainer/Dockerfile` with `--network=host` run args (required for Pulumi to reach the Proxmox API and the Talos cluster directly).

## VS Code Extensions

The following extensions are auto-installed:

| Extension | Purpose |
|-----------|---------|
| `golang.go` | Go language support (gopls, test runner, debugger) |
| `ms-azuretools.vscode-docker` | Docker and compose file support |
| `ms-kubernetes-tools.vscode-kubernetes-tools` | kubectl, YAML validation, cluster explorer |
| `redhat.vscode-yaml` | YAML schema validation |
| `nefrob.vscode-just-syntax` | `justfile` syntax highlighting |

## Editor Settings

```json
{
  "go.useLanguageServer": true,
  "go.toolsManagement.autoUpdate": false,
  "editor.formatOnSave": true,
  "go.lintTool": "golangci-lint",
  "go.formatTool": "gofumpt"
}
```

## File Mounts

The devcontainer mounts several host paths into the container:

| Host Path | Container Path | Notes |
|-----------|----------------|-------|
| `~/.ssh/id_ed25519` | `/home/vscode/.ssh/id_ed25519` | SSH key for Hetzner VPS access (read-only) |
| `~/.ssh/id_ed25519.pub` | `/home/vscode/.ssh/id_ed25519.pub` | (read-only) |
| `~/.kube` | `/home/vscode/.kube` | kubeconfig for kubectl access |
| `vscode-go-modules` (Docker volume) | `/home/vscode/go/pkg/mod` | Go module cache (persists across rebuilds) |

## Environment Variables

| Variable | Value | Purpose |
|----------|-------|---------|
| `GOPATH` | `/home/vscode/go` | Go workspace |
| `PULUMI_SKIP_UPDATE_CHECK` | `true` | Suppress Pulumi update nags |
| `TALOSCONFIG` | `/workspace/infra/pulumi/talosconfig` | Points `talosctl` to the cluster config |
| `KUBECONFIG` | `/workspace/infra/pulumi/kubeconfig` | Points `kubectl` to the cluster |
| `LOCAL_WORKSPACE_FOLDER` | `${localWorkspaceFolder}` | Host workspace path (for scripts) |
| `HOST_HOME` | `${env:HOME}` | Host home directory (for SOPS age key path) |

## Post-Create Command

After container creation, the following runs to verify tooling:

```bash
sudo usermod -aG docker vscode && cdk8s --version && pulumi version && talosctl version --client
```

## Host Requirements

- **CPUs:** 2+
- **Memory:** 4 GB+

## Workspace

The repo is mounted at `/workspace` inside the container. All `just` recipes and Pulumi commands assume this path.
