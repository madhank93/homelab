+++
title = "Cloud"
description = "core/cloud/ Pulumi stacks: Hetzner Bifrost VPS, Authentik SSO, NetBird VPN, Cloudflare DNS/TLS."
weight = 40
sort_by = "weight"
+++

Cloud infrastructure is managed entirely by **Pulumi** (Go) from `core/cloud/`, running from a developer laptop. It is never run in CI — infrastructure changes are intentional, human-reviewed operations.

All secrets are injected at runtime via SOPS. No plaintext secrets exist on disk or in CI.

---

## Pulumi Stacks (`core/cloud/`)

| Stack | Command | Manages |
|-------|---------|---------|
| `hetzner` | `just core hetzner up` | Hetzner VPS + full Bifrost bootstrap |
| `authentik` | `just core authentik up` | OIDC apps, GitHub OAuth, ForwardAuth outpost |
| `cloudflare` | `just core cloudflare up` | DNS A records, public service exposure |

Kubernetes cluster provisioning lives in [Platform](/platform/) (`core/platform/`).

---

## Cluster Architecture

The cluster topology this sits in front of is drawn in
[Kubernetes Architecture](@/architecture/kubernetes-architecture/index.md).

---

## How Pulumi Runs

Every stack command uses `sops exec-env` to inject secrets as environment variables for the duration of the `pulumi up` call:

```bash
# Under the hood of every `just core <stack> up`:
SOPS_AGE_KEY_FILE="$HOME/.config/sops/age/keys.txt" \
  sops exec-env secrets/bootstrap.sops.yaml \
  'pulumi stack select <stack> && pulumi up --yes'
```

Secrets are never written to disk as plaintext. They exist only in memory during the Pulumi run.

---

## Source Layout

```
core/
├── main.go              # routes ctx.Stack() to the right Deploy function
├── config.go            # koanf-based config loader
├── config.yml           # per-stack settings (IPs, server names, locations)
├── internal/cfg/        # shared config helpers
├── cloud/
│   ├── hetzner.go       # Hetzner VPS + file generation + remote.Command
│   ├── cloudflare.go    # DNS records + publicServices toggle slice
│   ├── authentik.go     # OIDC apps + GitHub OAuth + ForwardAuth outpost
│   └── bifrost/         # All files uploaded to /etc/bifrost/ on the VPS
│       ├── bootstrap.sh           # Automated startup + secret provisioning
│       ├── docker-compose.yml     # All Bifrost services
│       ├── traefik/               # traefik.yml + dynamic/ routes
│       └── netbird/               # config.yaml (template) + dashboard.env + proxy.env
└── platform/
    ├── talos.go          # Proxmox VMs + Talos machine configs + bootstrap
    ├── proxmox.go        # Proxmox provider setup
    ├── argocd.go         # ArgoCD Helm chart + ApplicationSet
    ├── cilium.go         # Cilium CNI + Gateway API + L2 announcements
    └── cert_manager.go   # cert-manager Helm + ClusterIssuer
```
