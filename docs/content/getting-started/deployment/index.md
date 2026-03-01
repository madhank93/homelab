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
Phase 3  →  Bifrost VPS deploy — fully automated             (~8 min)
Phase 4  →  Authentik apps (Pulumi)                          (~2 min)
Phase 5  →  NetBird first-login + Authentik connector        (~5 min)
Phase 6  →  NetBird setup keys + proxy token                 (~5 min)
Phase 7  →  Publish CDK8s manifests                          (~2 min)
Phase 8  →  End-to-end verification
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
AUTHENTIK_TOKEN: <openssl rand -hex 30>         # bootstrap token — becomes akadmin API token on first boot
AUTHENTIK_SECRET_KEY: <openssl rand -base64 48> # Django secret key
AUTHENTIK_POSTGRESQL_PASSWORD: <openssl rand -hex 16>
AUTHENTIK_GITHUB_SECRET: <github oauth app secret>

# ── NetBird (Bifrost VPS) ─────────────────────────────────────────────────────
NB_DATA_STORE_KEY: <openssl rand -base64 32>    # SQLite encryption key — NEVER change after first deploy
NB_RELAY_SECRET: <openssl rand -base64 32>      # relay auth secret
NB_OWNER_PASSWORD: <strong password>            # initial NetBird local admin — used for first login only
NETBIRD_CLIENT_SECRET: <openssl rand -hex 32>   # Dex→Authentik OIDC connector secret

# NB_PROXY_TOKEN and NB_BIFROST_SETUP_KEY are added later (Phase 6)

# ── Infisical (K8s secrets platform) ─────────────────────────────────────────
INFISICAL_DB_PASSWORD: <strong random password>
INFISICAL_ENCRYPTION_KEY: <openssl rand -hex 16>
INFISICAL_AUTH_SECRET: <openssl rand -hex 24>
REDIS_PASSWORD: <strong random password>
```

> **`NB_DATA_STORE_KEY` is permanent.** The NetBird SQLite database is encrypted with this value. Changing it means losing all peer registrations.

> **`NB_OWNER_PASSWORD`** is the password for the initial local admin account (`admin@madhan.app`) inside NetBird's embedded Dex OIDC. Used only for the first login before an external identity provider is configured. Can be changed after Authentik SSO is working.

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

*(Skip `/netbird` for now — add it after Phase 6.)*

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

## Phase 3 — Bifrost VPS Deploy (Fully Automated)

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

   Preflight      — validates 5 required secrets, waits for cloud-init
   Step 1/5  traefik            → wait healthy (60s)
   Step 2/5  authentik-postgres → wait healthy (120s)
   Step 3/5  authentik-server   → wait healthy (300s)

   process_netbird_config():
     sed replaces ${NB_RELAY_SECRET}, ${NB_DATA_STORE_KEY}
     python3 bcrypt-hashes NB_OWNER_PASSWORD → ${NB_OWNER_HASH}
     python3 substitutes ${NB_OWNER_HASH} in netbird/config.yaml

   Step 4/5  netbird-server + netbird-dashboard → wait healthy
   Step 5/5  netbird-proxy → skipped (NB_PROXY_TOKEN not set yet)
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

---

## Phase 4 — Authentik Apps (Pulumi)

Now that Authentik is running, configure it via Pulumi:

```bash
just core authentik up
```

This creates in Authentik:

- GitHub OAuth source (for user login)
- GitHub source bound to the default identification stage → **"Login with GitHub" button appears**
- NetBird OIDC application (`aumenijDycfG1cQURqH9BNJpV3KVUCoMHGPUVUlT`) + confidential client secret
- Homelab ForwardAuth proxy provider (covers `*.madhan.app`)
- Embedded outpost for ForwardAuth

---

## Phase 5 — NetBird First Login + Authentik Connector

NetBird v0.66 runs an embedded Dex OIDC provider. On first deploy, no external identity provider is connected — you must log in with the local admin account to wire up Authentik.

### 5a. Log in with local admin

Open **`https://netbird.madhan.app`**

Sign in with:
- Email: `admin@madhan.app`
- Password: the value of `NB_OWNER_PASSWORD` from your SOPS file

### 5b. Connect Authentik as the identity provider

**Settings → Identity Providers → Add → Authentik**

| Field | Value |
|-------|-------|
| Client ID | `aumenijDycfG1cQURqH9BNJpV3KVUCoMHGPUVUlT` |
| Client Secret | value of `NETBIRD_CLIENT_SECRET` from SOPS |
| Issuer | `https://auth.madhan.app/application/o/netbird/` |

Save. The redirect URI shown should be `https://netbird.madhan.app/oauth2/callback`.

### 5c. Verify GitHub login works

Open a private browser window → go to `https://netbird.madhan.app` → you should now see "Login with Authentik" → which shows the GitHub button via Authentik.

---

## Phase 6 — NetBird Setup Keys + Proxy Token

### 6a. Create setup keys and personal access token

Still in **`https://netbird.madhan.app`** (logged in via GitHub now):

**Setup Keys → Add Key:**

| Key name | Type | Stored as |
|----------|------|-----------|
| `bifrost-agent` | Reusable | `NB_BIFROST_SETUP_KEY` in SOPS |
| `k8s-routing-peer` | Reusable | `NETBIRD_SETUP_KEY` in Infisical |

**Settings → Access Tokens → Create Personal Access Token** → copy it → stored as `NB_PROXY_TOKEN` in SOPS.

### 6b. Add tokens to SOPS, re-run Pulumi

```bash
sops secrets/bootstrap.sops.yaml
```

Add:
```yaml
NB_PROXY_TOKEN: <personal access token from 6a>
NB_BIFROST_SETUP_KEY: <bifrost-agent setup key from 6a>
```

Then re-deploy Bifrost:

```bash
just core hetzner up
```

bootstrap.sh re-runs. This time:
- `NB_PROXY_TOKEN` is now set → `netbird-proxy` starts
- `NB_BIFROST_SETUP_KEY` is now set → `netbird-agent` connects to the mesh

### 6c. Add K8s routing peer key to Infisical

Open **`http://infisical.madhan.app`** → Project `homelab-prod` → Env `prod` → Path `/netbird`:

```
NETBIRD_SETUP_KEY = <k8s-routing-peer key from step 6a>
```

Within 60 seconds the `InfisicalSecret` in the `netbird` namespace syncs. The `netbird-peer` pod starts and connects:

```bash
kubectl get pods -n netbird   # → Running
```

### 6d. Configure the LAN route

In the NetBird dashboard, **Network Routes → Add Route**:

- **Network:** `192.168.1.0/24`
- **Routing Peer:** `k8s-routing-peer` (appears once the pod connects)
- **Enabled:** Yes

Once active, the Bifrost `netbird-agent` can reach `192.168.1.220` through the mesh — enabling Traefik to proxy public services to the cluster gateway.

---

## Phase 7 — Publish CDK8s Manifests

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

## Phase 8 — End-to-End Verification

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
| `AUTHENTIK_TOKEN` | SOPS → `.secrets.env` as `AUTHENTIK_BOOTSTRAP_TOKEN` | Phase 0 | Authentik akadmin API token |
| `AUTHENTIK_SECRET_KEY` | SOPS → `.env` | Phase 0 | Django secret key |
| `AUTHENTIK_POSTGRESQL_PASSWORD` | SOPS → `.env` | Phase 0 | Authentik Postgres |
| `AUTHENTIK_GITHUB_SECRET` | SOPS | Phase 0 | GitHub OAuth app secret |
| `NB_DATA_STORE_KEY` | SOPS → `.secrets.env` | Phase 0 | **Never rotate** |
| `NB_RELAY_SECRET` | SOPS → `.secrets.env` | Phase 0 | |
| `NB_OWNER_PASSWORD` | SOPS → `.secrets.env` | Phase 0 | Local admin for first NetBird login |
| `NETBIRD_CLIENT_SECRET` | SOPS (used by Pulumi) | Phase 0 | Dex→Authentik OIDC connector |
| `NB_PROXY_TOKEN` | SOPS → `.secrets.env` | Phase 6 | |
| `NB_BIFROST_SETUP_KEY` | SOPS → `.secrets.env` | Phase 6 | |
| `NETBIRD_SETUP_KEY` (k8s) | Infisical `/netbird` | Phase 6 | k8s-routing-peer setup key |
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
