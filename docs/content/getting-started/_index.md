+++
title = "Getting Started"
description = "Prerequisites, architecture overview, and fresh cluster bootstrap."
weight = 10
sort_by = "weight"
+++

This guide walks through what you need to know before running the homelab, how the components fit together, and how to bootstrap a fresh cluster.

## Architecture Overview

The homelab is organized in four layers. Each layer is managed by a specific tool:

![Architecture Diagram](/images/architecture.svg)

| Layer | What's in it | Managed by |
|-------|-------------|------------|
| **Apps** | ComfyUI, Ollama, Grafana, Harbor, n8n, Falco, … | ArgoCD + CDK8s |
| **Platform** | Talos k8s, Cilium CNI, Gateway API, ArgoCD, cert-manager | Pulumi |
| **Infrastructure** | Proxmox VMs (7 nodes), Hetzner VPS (Bifrost) | Pulumi |
| **Hardware** | Proxmox host, NVIDIA RTX 5070 Ti | Manual |

## Prerequisites

### Local Tools

All tools are available inside the [devcontainer](/development/devcontainer), or install them manually:

| Tool | Purpose | Install |
|------|---------|---------|
| `pulumi` | Infrastructure provisioning | `brew install pulumi` |
| `talosctl` | Talos cluster management | `brew install siderolabs/tap/talosctl` |
| `kubectl` | Kubernetes access | `brew install kubectl` |
| `just` | Task runner | `brew install just` |
| `sops` | Secrets encryption | `brew install sops` |
| `age` | Encryption key management | `brew install age` |
| `cdk8s` | Kubernetes manifest synthesis | `npm install -g cdk8s-cli` |
| `go` 1.23+ | CDK8s app language | `brew install go` |

### Secrets Setup

Before running any Pulumi commands, you must set up the SOPS age key:

```bash
# Generate an age key pair (one-time)
mkdir -p ~/.config/sops/age
age-keygen -o ~/.config/sops/age/keys.txt

# Add to your shell profile — REQUIRED (sops 3.12+ won't auto-discover the key)
echo 'export SOPS_AGE_KEY_FILE="$HOME/.config/sops/age/keys.txt"' >> ~/.zshrc
source ~/.zshrc

# Register your public key in .sops.yaml at the repo root
# Then encrypt the bootstrap secrets:
cp infra/secrets/bootstrap.sops.yaml.example infra/secrets/bootstrap.yaml
$EDITOR infra/secrets/bootstrap.yaml   # fill in real values
sops --encrypt infra/secrets/bootstrap.yaml > infra/secrets/bootstrap.sops.yaml
rm infra/secrets/bootstrap.yaml
```

See [Secrets Architecture](/infrastructure/secrets) for full details.

### Proxmox

A running Proxmox host with:
- Enough CPU/RAM for 7 VMs (4 vCPU + 6 GiB each = 28 vCPU, 42 GiB total)
- ~900 GiB storage (7 VMs × 125 GiB workers + 3 × 30 GiB controllers)
- API access (Pulumi uses the Proxmox provider)
- NVIDIA GPU available for PCIe passthrough to `k8s-worker4`

## Fresh Cluster Bootstrap

```bash
# 1. Create bootstrap k8s Secrets (Infisical + Cloudflare)
just create-secrets

# 2. Provision VMs, bootstrap Talos, install Cilium + ArgoCD
just pulumi talos up

# 3. Apply platform layer (Gateway, IP pool, HTTPRoutes)
just pulumi platform up

# 4. ArgoCD will auto-sync and deploy all apps from the manifests branch
# Watch progress:
kubectl get applications -n argocd
```

## Repository Structure

```
homelab/
├── infra/
│   ├── pulumi/          # Pulumi Go code (Talos, Cilium, ArgoCD, Hetzner)
│   └── secrets/         # bootstrap.sops.yaml (encrypted, safe to commit)
├── platform/
│   └── cdk8s/           # CDK8s Go code → generates Kubernetes manifests
├── docs/                # This documentation site (Zola + Goyo)
├── .devcontainer/       # VS Code devcontainer definition
├── .github/workflows/   # CI: CDK8s publish + docs deploy
└── justfile             # Task runner recipes
```
