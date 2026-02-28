+++
title = "Getting Started"
description = "Prerequisites, architecture overview, and fresh cluster bootstrap."
weight = 10
sort_by = "weight"
+++

This section covers what you need before running the homelab, how the four layers fit together, and the commands to bootstrap a fresh cluster.

---

## Architecture Overview

The homelab is organized in four layers. Each layer is managed by a specific tool:

![Architecture Diagram](/images/architecture.svg)

| Layer | What's in it | Managed by |
|-------|-------------|------------|
| **Apps** | ComfyUI, Ollama, Grafana, Harbor, n8n, Falco, … | ArgoCD + CDK8s |
| **Platform** | Talos K8s, Cilium CNI, Gateway API, ArgoCD, cert-manager | Pulumi |
| **Infrastructure** | Proxmox VMs (7 nodes), Hetzner VPS (Bifrost edge) | Pulumi |
| **Hardware** | Proxmox host, NVIDIA RTX 5070 Ti | Manual |

---

## Prerequisites

### Local Tools

All tools are available inside the [devcontainer](/development/devcontainer), or install manually:

| Tool | Purpose | Install |
|------|---------|---------|
| `pulumi` | Infrastructure provisioning | `brew install pulumi` |
| `talosctl` | Talos cluster management | `brew install siderolabs/tap/talosctl` |
| `kubectl` | Kubernetes access | `brew install kubectl` |
| `just` | Task runner | `brew install just` |
| `sops` | Secrets encryption | `brew install sops` |
| `age` | Encryption key management | `brew install age` |
| `cdk8s` | K8s manifest synthesis CLI | `npm install -g cdk8s-cli` |
| `go` 1.23+ | CDK8s + Pulumi app language | `brew install go` |

### Secrets Setup (one-time)

Before running any Pulumi commands, set up the SOPS age key:

```bash
# Generate age key pair
mkdir -p ~/.config/sops/age
age-keygen -o ~/.config/sops/age/keys.txt
# Prints: Public key: age1abc123...

# Add to shell profile — REQUIRED (sops 3.12+ won't auto-discover the key)
echo 'export SOPS_AGE_KEY_FILE="$HOME/.config/sops/age/keys.txt"' >> ~/.zshrc
source ~/.zshrc

# Register public key in .sops.yaml at repo root, then populate secrets:
sops secrets/bootstrap.sops.yaml
```

See [Secrets Architecture](/infrastructure/secrets) and the [Deployment Guide](/getting-started/deployment) for the full secrets list.

### Proxmox Host

- CPU/RAM for 7 VMs (4 vCPU + 6 GiB each ≈ 28 vCPU, 42 GiB)
- ~900 GiB storage (7 × 125 GiB workers + 3 × 30 GiB controllers)
- Proxmox API access (the Pulumi provider uses it)
- NVIDIA GPU for PCIe passthrough to `k8s-worker4`

---

## Bootstrap Quick Reference

```bash
# 1. Create bootstrap k8s Secrets (Infisical + Cloudflare)
just create-secrets

# 2. Provision VMs, bootstrap Talos, install Cilium + ArgoCD  (~15 min)
just core talos up

# 3. Apply Cilium Gateway, IP pool, HTTPRoutes
just core platform up

# 4. Deploy Bifrost VPS + run automated bootstrap sequence   (~8 min)
just core cloudflare up && just core hetzner up

# 5. Watch ArgoCD sync apps
kubectl get applications -n argocd
```

> After phase 4, log in to `https://netbird.madhan.app` to complete the one-time NetBird setup (setup keys + proxy token). See the [Deployment Guide](/getting-started/deployment) for the full sequence.

---

## Repository Structure

```
homelab/
├── core/                # Pulumi Go code (infra + Bifrost)
│   ├── cloud/           #   hetzner.go · cloudflare.go · authentik.go
│   │   └── bifrost/     #   docker-compose.yml · bootstrap.sh · traefik/ · netbird/
│   └── platform/        #   talos.go · argocd.go · cilium.go · cert_manager.go
├── workloads/           # CDK8s Go code → generates Kubernetes manifests
│   ├── ai/              #   ollama.go · comfyui.go
│   ├── observability/   #   victoria_metrics.go · victoria_logs.go · otel_collector.go
│   ├── monitoring/      #   grafana.go
│   ├── security/        #   falco.go · keyverno.go · trivy.go
│   └── ...              #   storage · registry · automation · management · networking
├── secrets/             # bootstrap.sops.yaml (age-encrypted — safe to commit)
├── scripts/             # create-bootstrap-secrets.sh
├── app/                 # CDK8s synthesis output (gitignored; published to *-manifests)
├── docs/                # This site (Zola + Goyo theme)
├── .devcontainer/       # VS Code devcontainer
├── .github/workflows/   # CDK8s publish + docs deploy
└── justfile             # Task runner recipes
```

---

## Justfile Recipes

| Command | What it does |
|---------|-------------|
| `just create-secrets` | Create bootstrap k8s Secrets from SOPS |
| `just core talos up` | Provision cluster (Talos + Cilium + ArgoCD) |
| `just core platform up` | Apply Gateway API + cert-manager config |
| `just core hetzner up` | Deploy Bifrost VPS + automated bootstrap |
| `just core authentik up` | Create OIDC apps + ForwardAuth in Authentik |
| `just core cloudflare up` | Create/update DNS records |
| `just synth` | Synthesize CDK8s manifests → `app/` |

The `core <stack> <action>` recipe injects SOPS secrets as environment variables for every Pulumi run:
```bash
SOPS_AGE_KEY_FILE="$HOME/.config/sops/age/keys.txt" \
  sops exec-env secrets/bootstrap.sops.yaml \
  'pulumi stack select <stack> && pulumi <action> --yes'
```
