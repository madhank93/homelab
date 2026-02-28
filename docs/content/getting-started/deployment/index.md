+++
title = "Deployment Guide"
description = "Complete step-by-step guide for deploying the homelab from scratch — from local prerequisites to a fully running cluster with all services."
weight = 20
+++

This guide walks through the complete deployment sequence. Each phase depends on the previous one completing successfully. Most steps are fully automated — the majority of your time is spent waiting for services to come up.

## At a Glance

```
Phase 0  →  Local prerequisites + SOPS secrets setup        (one-time)
Phase 1  →  Bootstrap Kubernetes cluster                     (~15 min)
Phase 2  →  Infisical first-time setup                       (~10 min)
Phase 3  →  Authentik apps (Pulumi)                          (~2 min)
Phase 4  →  Bifrost VPS deploy — fully automated             (~8 min)
Phase 5  →  NetBird setup keys + proxy token                 (~5 min)
Phase 6  →  Publish CDK8s manifests                          (~2 min)
Phase 7  →  End-to-end verification
```

---

## Phase 0 — Local Prerequisites

### Tools

All tools are available inside the [devcontainer](/development/devcontainer), or install on macOS:

```bash
brew install pulumi talosctl kubectl just sops age go
npm install -g cdk8s-cli
```

### age key (one-time — back this up!)

```bash
mkdir -p ~/.config/sops/age
age-keygen -o ~/.config/sops/age/keys.txt
# Output: Public key: age1abc123...

# Add to shell profile — REQUIRED (sops 3.12+ doesn't auto-discover the key file)
echo 'export SOPS_AGE_KEY_FILE="$HOME/.config/sops/age/keys.txt"' >> ~/.zshrc
source ~/.zshrc
```

Register your public key in `.sops.yaml` at the repo root, under `creation_rules[*].age`.

### Populate bootstrap secrets

```bash
# Open the encrypted file in your editor — sops decrypts, re-encrypts on save
sops secrets/bootstrap.sops.yaml
```

Add all of these keys. Generate values as indicated:

```yaml
# ── Proxmox & Hetzner ─────────────────────────────────────────────────────────
PROXMOX_PASSWORD: <proxmox root password>
HCLOUD_TOKEN: <hetzner cloud API token>

# ── Cloudflare ────────────────────────────────────────────────────────────────
CLOUDFLARE_API_TOKEN: <zone:dns edit token>     # used by cert-manager + Traefik ACME

# ── Authentik (Bifrost VPS) ───────────────────────────────────────────────────
AUTHENTIK_TOKEN: <openssl rand -hex 30>         # bootstrap token → also becomes NB_IDP_MGMT_TOKEN
AUTHENTIK_SECRET_KEY: <openssl rand -base64 48> # Django secret key
AUTHENTIK_POSTGRESQL_PASSWORD: <openssl rand -hex 16>
AUTHENTIK_GITHUB_SECRET: <github oauth app secret>

# ── NetBird (Bifrost VPS) ─────────────────────────────────────────────────────
NB_DATA_STORE_KEY: <openssl rand -base64 32>    # SQLite encryption key — NEVER change after first deploy
NB_RELAY_SECRET: <openssl rand -base64 32>      # relay auth secret

# NB_PROXY_TOKEN and NB_BIFROST_SETUP_KEY are added later (Phase 5)

# ── Infisical (K8s secrets platform) ─────────────────────────────────────────
INFISICAL_DB_PASSWORD: <strong random password>
INFISICAL_ENCRYPTION_KEY: <openssl rand -hex 16>
INFISICAL_AUTH_SECRET: <openssl rand -hex 24>
REDIS_PASSWORD: <strong random password>
```

> **`NB_DATA_STORE_KEY` is permanent.** The NetBird SQLite database is encrypted with this value. Changing it means losing all peer registrations.

> **`AUTHENTIK_TOKEN`** is a strong random string you generate once. Authentik uses it as the `akadmin` bootstrap API token on first boot. The bootstrap script also uses it as the NetBird IDP management token (`NB_IDP_MGMT_TOKEN`) — no separate manual step needed.

---

## Phase 1 — Bootstrap Kubernetes Cluster

```bash
# 1. Create bootstrap k8s Secrets (Infisical DB credentials + Cloudflare token)
just create-secrets

# 2. Provision Proxmox VMs → bootstrap Talos → install Cilium + ArgoCD (~15 min)
just core talos up

# 3. Apply Cilium Gateway API, IP pool, HTTPRoutes
just core platform up
```

**Verify the cluster:**

```bash
talosctl health --nodes 192.168.1.211
kubectl get nodes
kubectl get applications -n argocd
```

ArgoCD starts syncing apps from the `v0.1.5-manifests` branch. Most apps will show `Degraded` — that's expected. They need Infisical secrets to become healthy.

---

## Phase 2 — Infisical First-Time Setup

Wait for Infisical to be running (~5 min after Phase 1):

```bash
kubectl get pods -n infisical   # wait for Running
```

Open **`http://infisical.madhan.app`** (LAN only — no SSO yet).

### 2a. Create organization and project

1. Create an account
2. **Organization** → **New Project** → name: `homelab`, slug: `homelab-prod`
3. **Environments** → Add → name: `prod`, slug: `prod`

### 2b. Add secrets

Navigate to **Secrets** → `prod` environment and add each path:

**Path `/grafana`**
```
ADMIN_PASSWORD = <strong password>
```

**Path `/harbor`**
```
HARBOR_ADMIN_PASSWORD = <strong password>
```

**Path `/n8n`**
```
DB_PASSWORD         = <strong password>
N8N_ENCRYPTION_KEY  = <32-char random — record this, needed on every rebuild>
```

**Path `/rancher`**
```
BOOTSTRAP_PASSWORD = <strong password>
```

*(Skip `/netbird` for now — add it after Phase 5.)*

### 2c. Create service token

**Settings → Service Tokens → Add Token**:
- Name: `k8s-homelab`
- Environment: `prod`
- Path: `/`
- No expiry
- Read only

Copy the token (shown once only).

### 2d. Store token in K8s

```bash
kubectl create secret generic infisical-service-token \
  --from-literal=infisicalToken=<paste-token> \
  -n infisical
```

Within 60 seconds, `InfisicalSecret` CRs sync and apps start becoming healthy:

```bash
kubectl get applications -n argocd
# grafana, harbor, n8n, rancher → Healthy
```

---

## Phase 3 — Authentik (Pulumi)

```bash
just core authentik up
```

This creates in Authentik (via the Authentik API using your `AUTHENTIK_TOKEN`):

- GitHub OAuth source (for user login)
- NetBird OIDC application + client ID/secret
- Homelab ForwardAuth proxy provider (covers `*.madhan.app`)
- Embedded outpost for ForwardAuth

> The Authentik stack runs **after** Bifrost is up. If you're deploying for the first time and Authentik isn't running yet, skip this phase and come back after Phase 4.

---

## Phase 4 — Bifrost VPS Deploy (Fully Automated)

This is the phase that used to require manual SSH steps. **It is now entirely automated** by a single command.

```bash
# Create Cloudflare DNS records for all public hostnames
just core cloudflare up

# Deploy/update Hetzner VPS + run bootstrap sequence
just core hetzner up
```

### What `just core hetzner up` does

```
1. Reads SOPS secrets → writes .secrets.env and .env on laptop
2. Uploads entire ./cloud/bifrost/ directory to /etc/bifrost/ on VPS
3. Runs bootstrap.sh on the VPS, which:

   Preflight      — validates required secrets, waits for cloud-init
   Step 1/6  traefik            → wait healthy (60s)
   Step 2/6  authentik-postgres → wait healthy (120s)
   Step 3/6  authentik-server   → wait healthy (300s)

   Auto-provision NB_IDP_MGMT_TOKEN:
     docker exec authentik-server ak shell
     → Token.objects.get_or_create(identifier='netbird-mgmt-token')
     → NB_IDP_MGMT_TOKEN appended to .secrets.env

   Substitute config.yaml template:
     sed replaces ${NB_RELAY_SECRET}, ${NB_DATA_STORE_KEY}, ${NB_IDP_MGMT_TOKEN}

   Step 4/6  netbird-server     → wait healthy (120s)
   Step 5/6  netbird-dashboard  → wait healthy (60s)
   Step 6/6  netbird-proxy      → skipped (NB_PROXY_TOKEN not set yet)
```

**Expected output at the end:**

```
NAMES                STATUS
netbird-dashboard    Up X seconds
netbird-server       Up X seconds
authentik-server     Up X minutes (healthy)
authentik-worker     Up X minutes (healthy)
authentik-postgres   Up X minutes (healthy)
traefik              Up X minutes
Bootstrap complete.
```

### Verify Bifrost is reachable

```bash
curl -I https://auth.madhan.app          # → 200 Authentik login page
curl -I https://netbird.madhan.app       # → 200 NetBird dashboard
curl -sI https://grafana.madhan.app | grep -i location
# → location: https://auth.madhan.app/...  (ForwardAuth working)
```

### Go back and run Authentik (if skipped)

```bash
just core authentik up    # now Authentik is running, this can complete
```

---

## Phase 5 — NetBird Setup Keys + Proxy Token

### 5a. Create setup keys

Open **`https://netbird.madhan.app`** → Log in with GitHub.

**Setup Keys → Add Key:**

| Key name | Type | Used for |
|----------|------|---------|
| `bifrost-agent` | Reusable | `netbird-agent` container on the VPS |
| `k8s-routing-peer` | Reusable | `netbird-peer` pod in Kubernetes |

**Settings → Access Tokens → Create Personal Access Token** — copy it immediately.

### 5b. Add tokens to SOPS, re-run Pulumi

```bash
# Add to bootstrap.sops.yaml
sops secrets/bootstrap.sops.yaml
```

Add these keys:
```yaml
NB_PROXY_TOKEN: <personal access token from step 5a>
NB_BIFROST_SETUP_KEY: <bifrost-agent setup key from step 5a>
```

Then re-deploy Bifrost:

```bash
just core hetzner up
```

bootstrap.sh re-runs. This time:
- `NB_IDP_MGMT_TOKEN` is already in `.secrets.env` → provisioning skipped
- `NB_PROXY_TOKEN` is now set → `netbird-proxy` starts
- `NB_BIFROST_SETUP_KEY` is now set → `netbird-agent` connects to the mesh

### 5c. Add K8s routing peer key to Infisical

Open **`http://infisical.madhan.app`** → Project `homelab-prod` → Env `prod` → Path `/netbird`:

```
NETBIRD_SETUP_KEY = <k8s-routing-peer key from step 5a>
```

Within 60 seconds the `InfisicalSecret` in the `netbird` namespace syncs. The `netbird-peer` pod starts and connects:

```bash
kubectl get pods -n netbird   # → Running
```

### 5d. Configure the LAN route

In the NetBird dashboard, **Network Routes → Add Route**:

- **Network:** `192.168.1.0/24`
- **Routing Peer:** `k8s-routing-peer` (appears once the pod connects)
- **Enabled:** Yes

Once active, the Bifrost `netbird-agent` can reach `192.168.1.220` through the mesh — enabling Traefik to proxy public services to the cluster gateway.

---

## Phase 6 — Publish CDK8s Manifests

The NetBird peer chart needs to be synthesized and pushed to the manifests branch:

```bash
# From repo root — generates app/netbird/ and other workload dirs
just synth

# Commit and push to the manifests branch
git add app/
git commit -m "feat: Add NetBird routing peer"
git push origin v0.1.5-manifests
```

ArgoCD auto-syncs within 3 minutes:

```bash
kubectl get application netbird -n argocd
```

---

## Phase 7 — End-to-End Verification

```bash
# DNS split working correctly
dig grafana.madhan.app    # → 178.156.199.250  (public, via Hetzner)
dig headlamp.madhan.app   # → 192.168.1.220    (LAN wildcard, private)

# ForwardAuth redirecting unauthenticated requests
curl -sI https://grafana.madhan.app | grep -i location
# → location: https://auth.madhan.app/outpost.goauthentik.io/start?rd=...

# NetBird mesh
# NetBird UI → Peers: bifrost-agent Connected, k8s-routing-peer Connected
# NetBird UI → Routes: 192.168.1.0/24 Active

# Full browser flow
# 1. Open https://grafana.madhan.app
# 2. Redirected to auth.madhan.app → log in with GitHub
# 3. Land on Grafana dashboard

# All apps healthy
kubectl get applications -n argocd
```

---

## Quick Reference: Secrets and Where They Live

| Secret | Stored in | Set during | Notes |
|--------|-----------|-----------|-------|
| `HCLOUD_TOKEN` | SOPS | Phase 0 | Hetzner API access |
| `PROXMOX_PASSWORD` | SOPS | Phase 0 | Proxmox API access |
| `CLOUDFLARE_API_TOKEN` | SOPS → `.secrets.env` + k8s Secret | Phase 0 | cert-manager + Traefik ACME |
| `AUTHENTIK_TOKEN` | SOPS → `.secrets.env` as `AUTHENTIK_BOOTSTRAP_TOKEN` | Phase 0 | Also becomes `NB_IDP_MGMT_TOKEN` |
| `AUTHENTIK_SECRET_KEY` | SOPS → `.env` | Phase 0 | Django secret key |
| `AUTHENTIK_POSTGRESQL_PASSWORD` | SOPS → `.env` | Phase 0 | Authentik Postgres |
| `NB_DATA_STORE_KEY` | SOPS → `.secrets.env` | Phase 0 | **Never rotate** |
| `NB_RELAY_SECRET` | SOPS → `.secrets.env` | Phase 0 | |
| `NB_IDP_MGMT_TOKEN` | Auto-written to `.secrets.env` by `bootstrap.sh` | Phase 4 (auto) | Via `ak shell` |
| `NB_PROXY_TOKEN` | SOPS → `.secrets.env` | Phase 5 | |
| `NB_BIFROST_SETUP_KEY` | SOPS → `.secrets.env` | Phase 5 | |
| `NETBIRD_SETUP_KEY` (k8s) | Infisical `/netbird` | Phase 5 | |
| App passwords | Infisical paths | Phase 2 | |
| Infisical service token | k8s Secret `infisical-service-token` | Phase 2 | |

---

## LAN Access Without the Hetzner Hop

Public services (`grafana.madhan.app`, `harbor.madhan.app`) DNS-resolve to the Hetzner VPS for all clients — including LAN users. To bypass this on your machine:

```bash
# /etc/hosts — direct LAN access, skips SSO
192.168.1.220  grafana.madhan.app harbor.madhan.app
```

For LAN-wide bypass, add overrides in your router's DNS (Pi-hole, Unbound, etc.).
