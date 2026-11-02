---
title: Devcontainer
description: Tools, mounts and environment variables of the VS Code devcontainer.
sidebar:
  order: 1
---

Defined in [`.devcontainer/devcontainer.json`](src:.devcontainer/devcontainer.json), built from [`.devcontainer/Dockerfile`](src:.devcontainer/Dockerfile). The repo mounts at `/workspace`; every recipe assumes that path.

## Tools

Versions are `ARG`s at the top of the Dockerfile.

| Tool | Used for |
|---|---|
| Go, Delve, gofumpt, golangci-lint | CDK8s and Pulumi code |
| Pulumi | `just core …` |
| cdk8s CLI | `cdk8s import` |
| talosctl, kubectl, helm | Cluster operations |
| just, yq | Recipes and reading `config/` |
| Node.js (nvm), Bun | Docs site and tooling |
| kubeseal | Sealed Secrets CLI |
| Docker CLI | Docker-outside-of-Docker feature, for `just build-push` and secret scans |
| Claude Code | Coding assistant; `postCreate.sh` installs its plugins |

`sops` and `age` are not in the image: recipes that decrypt (`just core`, `just create-secrets`) need them on the host or installed separately, with the age key at `~/.config/sops/age/keys.txt`.

## Runtime

`runArgs: --network=host`, so Pulumi and `talosctl` reach Proxmox and the cluster directly.

| Host path | Container path | Notes |
|---|---|---|
| `~/.ssh/id_ed25519{,.pub}` | `/home/vscode/.ssh/` | Read-only; SSH to the VPS |
| `~/.kube` | `/home/vscode/.kube` | kubeconfig |
| `~/.claude.json`, `~/.claude/settings.json`, `~/.claude/projects` | `/home/vscode/.claude…` | Claude Code config and history |
| volume `vscode-claude-plugins` | `/home/vscode/.claude/plugins` | Persists across rebuilds |
| volume `vscode-go-modules` | `/home/vscode/go/pkg/mod` | Go module cache |

| Variable | Value |
|---|---|
| `TALOSCONFIG` | `/workspace/core/talosconfig` |
| `KUBECONFIG` | `/workspace/core/kubeconfig` |
| `GOPATH` | `/home/vscode/go` |
| `PULUMI_SKIP_UPDATE_CHECK` | `true` |
| `LOCAL_WORKSPACE_FOLDER`, `HOST_HOME` | Host paths, for scripts |

[`postCreate.sh`](src:.devcontainer/postCreate.sh) adds `vscode` to the docker group, fixes `.claude` ownership, prints the cdk8s, Pulumi and talosctl versions, and installs the Claude plugins. Host needs 2+ CPUs and 4 GB+ RAM.
