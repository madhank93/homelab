+++
title = "Infrastructure"
description = "Pulumi stacks: Proxmox/Talos cluster, Cilium networking, Hetzner Bifrost VPS, DNS/TLS, and secrets."
weight = 40
sort_by = "weight"
+++

Infrastructure is managed entirely by **Pulumi** (Go), running from a developer laptop. It is never run in CI — infrastructure changes are intentional, human-reviewed operations.

All secrets are injected at runtime via SOPS. No plaintext secrets exist on disk or in CI.

---

## Pulumi Stacks

| Stack | Command | Manages |
|-------|---------|---------|
| `talos` | `just core talos up` | Proxmox VMs, Talos bootstrap, Cilium CNI, ArgoCD |
| `platform` | `just core platform up` | Gateway API, IP pool, HTTPRoutes, cert-manager |
| `hetzner` | `just core hetzner up` | Hetzner VPS + full Bifrost bootstrap automation |
| `authentik` | `just core authentik up` | OIDC apps, GitHub OAuth, ForwardAuth outpost |
| `cloudflare` | `just core cloudflare up` | DNS A records, public service exposure |

---

## Cluster Architecture

{% mermaid() %}
flowchart TD
    DEV["Developer Laptop<br/>Pulumi + SOPS"]

    subgraph PROX["Proxmox Host"]
        direction LR
        CP["k8s-controller1/2/3<br/>192.168.1.211–213<br/>VIP: 192.168.1.210"]
        W13["k8s-worker1/2/3<br/>192.168.1.221–223"]
        W4["k8s-worker4<br/>192.168.1.224<br/>NVIDIA RTX 5070 Ti"]
    end

    subgraph PLATFORM["Platform Layer"]
        CIL["Cilium CNI<br/>Gateway API L2 LB<br/>192.168.1.220"]
        ARGO["ArgoCD<br/>ApplicationSet"]
        CERT["cert-manager<br/>DNS-01 wildcard TLS"]
    end

    subgraph HETZNER["Hetzner Cloud"]
        VPS["Bifrost VPS<br/>178.156.199.250<br/>Traefik · NetBird · Authentik"]
    end

    subgraph CF["Cloudflare"]
        DNS["DNS zones<br/>*.madhan.app → 192.168.1.220<br/>auth/netbird/grafana → 178.156.199.250"]
    end

    DEV -->|just core talos up| PROX
    DEV -->|just core hetzner up| HETZNER
    DEV -->|just core cloudflare up| CF
    CP --> CIL
    CIL --> ARGO & CERT
    ARGO -->|syncs workloads| W13 & W4
    VPS <-->|WireGuard mesh| CIL
{% end %}

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
