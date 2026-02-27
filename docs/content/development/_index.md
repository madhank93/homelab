+++
title = "Development"
description = "Development environment setup and tooling."
weight = 30
sort_by = "weight"
+++

The homelab repo ships a devcontainer for a consistent, reproducible development environment.

## Quick Start

1. Open the repo in VS Code
2. Run **Dev Containers: Reopen in Container**
3. All tools are available immediately

## Tools Available in Devcontainer

| Tool | Purpose |
|------|---------|
| Go | CDK8s app language; Pulumi providers written in Go |
| Pulumi | Infrastructure provisioning (Proxmox, Hetzner, Kubernetes) |
| cdk8s CLI | Kubernetes manifest synthesis |
| talosctl | Talos cluster management |
| kubectl | Kubernetes resource management |
| just | Task runner (see `justfile` in repo root) |
| sops | Secrets encryption/decryption |
| age | Encryption key management |
| golangci-lint | Go linting |
| gofumpt | Go formatting |
| Docker (outside-of-docker) | Container builds |
