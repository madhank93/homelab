+++
title = "GitOps Flow"
description = "How infrastructure and workload changes flow from code to cluster — two separate paths, one repository."
weight = 30
+++

## Two Paths, One Repo

The homelab uses two separate IaC paths that live in the same repository but are triggered and operated differently:

| Path | Tool | Trigger | Who runs it | Changes |
|------|------|---------|------------|---------|
| `core/` | **Pulumi** | Manual (`just core <stack> up`) | Engineer on laptop | VMs, cluster, VPS, DNS, TLS |
| `workloads/` | **CDK8s** | Push to `main` → GitHub Actions CI | Automated | Kubernetes app manifests |

Pulumi is intentionally manual — infrastructure changes are high-risk and require human judgment. CDK8s synthesis is safe to automate since it only generates Kubernetes manifests.

---

## Architecture Diagram

{% mermaid() %}
flowchart TB
    subgraph LOCAL["Developer Laptop"]
        DEV["Engineer"]
        SOPS["SOPS + age<br/>secrets/bootstrap.sops.yaml"]
    end

    subgraph PULUMI["Pulumi — manual, laptop only"]
        direction LR
        PUL_T["just core talos up<br/>Proxmox VMs + Talos bootstrap<br/>Cilium + ArgoCD"]
        PUL_P["just core platform up<br/>Gateway API · HTTPRoutes · cert-manager"]
        PUL_H["just core hetzner up<br/>Hetzner VPS + bootstrap.sh<br/>NetBird + Traefik + Authentik"]
        PUL_A["just core authentik up<br/>OIDC apps · GitHub OAuth<br/>ForwardAuth outpost"]
        PUL_C["just core cloudflare up<br/>DNS records for all services"]
    end

    subgraph GITHUB["GitHub"]
        REPO["main branch<br/>code changes"]
        CI["GitHub Actions<br/>CDK8s publish workflow"]
        MBRANCH["v0.1.5-manifests branch<br/>app/*/  (synthesized manifests)"]
    end

    subgraph CLUSTER["Kubernetes Cluster"]
        ARGO["ArgoCD ApplicationSet<br/>watches manifests branch"]
        APPS["Application pods<br/>Grafana · Harbor · n8n · etc."]
    end

    subgraph VPS["Bifrost VPS (Hetzner)"]
        BS["bootstrap.sh<br/>automated startup sequence"]
    end

    DEV -->|sops exec-env| SOPS
    SOPS -->|env vars injected| PUL_T & PUL_P & PUL_H & PUL_A & PUL_C
    PUL_T & PUL_P -->|provisions| CLUSTER
    PUL_H -->|remote.Command| BS
    PUL_A & PUL_C -->|API calls| VPS & CLUSTER

    DEV -->|git push| REPO
    REPO --> CI
    CI -->|go run . → app/| MBRANCH
    MBRANCH -->|ApplicationSet detects dirs| ARGO
    ARGO -->|kubectl apply SSA| APPS
{% end %}

---

## Infra Path (Pulumi)

Pulumi is run from the developer's laptop. Secrets are never stored in environment variables permanently — they're injected per-command via `sops exec-env`:

```bash
# The justfile recipe wraps every pulumi command:
SOPS_AGE_KEY_FILE="$HOME/.config/sops/age/keys.txt" \
  sops exec-env secrets/bootstrap.sops.yaml \
  'pulumi stack select <stack> && pulumi up --yes'
```

### Stacks

| Stack | Command | Manages |
|-------|---------|---------|
| `talos` | `just core talos up` | Proxmox VMs, Talos bootstrap, Cilium CNI, ArgoCD |
| `platform` | `just core platform up` | Gateway API, IP pools, HTTPRoutes, cert-manager |
| `hetzner` | `just core hetzner up` | Hetzner VPS, Bifrost config, `bootstrap.sh` execution |
| `authentik` | `just core authentik up` | Authentik OIDC apps, GitHub OAuth, ForwardAuth outpost |
| `cloudflare` | `just core cloudflare up` | DNS A records for all public hostnames |

### File structure

```
core/
├── main.go          # dispatch by ctx.Stack()
├── config.go        # koanf config loader
├── config.yml       # stack-specific settings (IPs, names, etc.)
├── cloud/
│   ├── hetzner.go        # Hetzner VPS + bifrost bootstrap
│   ├── cloudflare.go     # DNS records + publicServices slice
│   └── authentik.go      # OIDC apps + ForwardAuth outpost
└── platform/
    ├── talos.go           # VMs + cluster bootstrap
    ├── proxmox.go         # Proxmox provider
    ├── argocd.go          # ArgoCD Helm + ApplicationSet
    ├── cilium.go          # CNI + Gateway API
    └── cert_manager.go    # TLS cert automation
```

---

## Workload Path (CDK8s + GitHub Actions)

CDK8s is a Go application that synthesizes Kubernetes YAML from typed Go code:

```bash
# Local synthesis (run from repo root)
just synth
# → cd workloads && go run . → writes to ../app/
```

In CI, a GitHub Actions workflow runs `go run .` and pushes the output to the `v0.1.5-manifests` branch:

```yaml
# .github/workflows/publish.yml (simplified)
- run: go run .
  working-directory: workloads
- run: |
    git checkout v0.1.5-manifests
    cp -r app/* .
    git add . && git commit -m "chore: Synthesize manifests" && git push
```

### Adding a new workload app

1. Create `workloads/<category>/<name>.go` with a `Deploy<Name>(app cdk8s.App)` function
2. Register it in `workloads/main.go`
3. Push to `main` → CI synthesizes manifests → ArgoCD syncs automatically

### File structure

```
workloads/
├── main.go              # registers all apps, calls cdk8s.App.Synth()
├── go.mod               # module: github.com/madhank93/homelab/workloads
├── imports/             # generated CDK8s type bindings
├── ai/                  ollama.go  comfyui.go
├── automation/          n8n.go
├── hardware/            nvidia_gpu_operator.go
├── management/          headlamp.go  fleet_device_manager.go  rancher.go
├── monitoring/          grafana.go
├── networking/          netbird_peer.go
├── observability/       victoria_metrics.go  victoria_logs.go  otel_collector.go  alert_manager.go
├── registry/            harbor.go
├── secrets/             infisical.go
├── security/            falco.go  keyverno.go  trivy.go
├── storage/             longhorn.go
└── support/             reloader.go
```

---

## ArgoCD ApplicationSet

One `ApplicationSet` watches the manifests branch. Every top-level directory under `app/` becomes one ArgoCD Application automatically:

```yaml
generators:
  - git:
      repoURL: https://github.com/madhank93/homelab.git
      revision: v0.1.5-manifests
      directories:
        - path: "*"
template:
  spec:
    syncPolicy:
      syncOptions:
        - ServerSideApply=true   # required for CRDs >262KB (kube-prometheus-stack)
```

> **`Prune=false` on bootstrap Secrets**: `infisical-secrets` and `cloudflare-api-token` are created by `just create-secrets`, not by CDK8s. Both carry `argocd.argoproj.io/sync-options: Prune=false` so ArgoCD never tries to delete them.

---

## What Changes What

| You want to change | Edit | Run |
|-------------------|------|-----|
| VPS firewall rules | `core/cloud/hetzner.go` | `just core hetzner up` |
| Public internet-exposed services | `core/cloud/cloudflare.go` (`publicServices` slice) | `just core cloudflare up && just core hetzner up` |
| Traefik routes (static) | `core/cloud/bifrost/traefik/dynamic/services.yml` | `just core hetzner up` |
| Authentik OIDC apps | `core/cloud/authentik.go` | `just core authentik up` |
| Kubernetes app version/config | `workloads/<category>/<app>.go` | `git push` → CI |
| Cluster node count or specs | `core/platform/talos.go` | `just core talos up` |
| Gateway API routes | `core/platform/cilium.go` | `just core platform up` |
