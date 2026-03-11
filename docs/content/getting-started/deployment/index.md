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
Phase 2  →  OpenBao init + K8s auth + write app secrets      (~10 min)
Phase 3  →  Bifrost VPS deploy — fully automated             (~8 min)
Phase 4  →  Authentik apps (Pulumi)                          (~2 min)
Phase 5  →  NetBird first-login + Authentik connector        (~5 min)
Phase 6  →  NetBird setup keys + proxy token                 (~5 min)
Phase 7  →  Grafana Authentik SSO setup                      (~5 min)
Phase 8  →  Publish CDK8s manifests                          (~2 min)
Phase 9  →  End-to-end verification
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

# ── OpenBao (K8s secrets platform) ───────────────────────────────────────────
OPENBAO_UNSEAL_KEY: placeholder                 # replaced in Phase 2 after first init

# NB_PROXY_TOKEN and NB_BIFROST_SETUP_KEY are added later (Phase 6)
```

> **`NB_DATA_STORE_KEY` is permanent.** The NetBird SQLite database is encrypted with this value. Changing it means losing all peer registrations.

> **`OPENBAO_UNSEAL_KEY`** starts as a placeholder. You'll replace it with the real value after `just openbao-init` in Phase 2.

---

## Phase 1 — Bootstrap Kubernetes Cluster

```bash
# 1. Create bootstrap k8s Secrets (OpenBao unseal key + Cloudflare token)
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

ArgoCD starts syncing apps from the manifests branch. Most apps will show `Degraded` — that's expected. OpenBao needs to be initialised and unsealed before apps can fetch their secrets.

---

## Phase 2 — OpenBao Init + K8s Auth + Write App Secrets

Wait for OpenBao to be running (but sealed) after Phase 1:

```bash
kubectl get pods -n openbao   # wait for Running
```

### 2a. Initialise OpenBao (one-time)

```bash
just openbao-init
```

This generates the root token and unseal key, writes them to `/tmp/openbao-init.json`, then unseals OpenBao.

### 2b. Store the unseal key in SOPS

```bash
UNSEAL_KEY=$(python3 -c "import json; print(json.load(open('/tmp/openbao-init.json'))['keys'][0])")
echo "OPENBAO_UNSEAL_KEY: $UNSEAL_KEY"

# Add to SOPS (replaces the placeholder from Phase 0)
sops secrets/bootstrap.sops.yaml
# → update OPENBAO_UNSEAL_KEY with the value above
```

### 2c. Update the bootstrap secret

```bash
just create-secrets
kubectl rollout restart statefulset/openbao -n openbao
# The unseal sidecar reads the updated secret and unseals on startup
```

### 2d. Configure K8s auth + policies + roles

```bash
just openbao-setup
```

This enables Kubernetes auth, creates per-app policies and roles, and writes placeholder secrets at each path.

### 2e. Write real app secrets

```bash
# Grafana — admin password + Authentik OAuth client secret (set OAUTH_CLIENT_SECRET after Phase 7)
kubectl exec -n openbao openbao-0 -- bao kv put secret/grafana \
  ADMIN_PASSWORD="<strong password>" \
  OAUTH_CLIENT_SECRET="placeholder"

# Harbor
kubectl exec -n openbao openbao-0 -- bao kv put secret/harbor \
  HARBOR_ADMIN_PASSWORD="<strong password>"

# N8n (DB password managed by CNPG — only encryption key needed)
kubectl exec -n openbao openbao-0 -- bao kv put secret/n8n \
  N8N_ENCRYPTION_KEY="<32-char random — record this, required on every rebuild>"

# Rancher
kubectl exec -n openbao openbao-0 -- bao kv put secret/rancher \
  BOOTSTRAP_PASSWORD="<strong password>"

# NetBird — add the setup key after Phase 6
kubectl exec -n openbao openbao-0 -- bao kv put secret/netbird \
  NETBIRD_SETUP_KEY="placeholder"
```

> After writing secrets, ArgoCD syncs and app pods start. Apps will become `Healthy` progressively as their CSI volumes mount.

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

**Verify Bifrost is reachable:**

```bash
curl -I https://auth.madhan.app        # → 200 Authentik login page
curl -I https://netbird.madhan.app     # → 200 NetBird dashboard
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
- NetBird OIDC application + confidential client secret
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

Open a private browser window → `https://netbird.madhan.app` → "Login with Authentik" → GitHub button should appear.

---

## Phase 6 — NetBird Setup Keys + Proxy Token

### 6a. Create setup keys and personal access token

In **`https://netbird.madhan.app`** (logged in via GitHub):

**Setup Keys → Add Key:**

| Key name | Type | Stored as |
|----------|------|-----------|
| `bifrost-agent` | Reusable | `NB_BIFROST_SETUP_KEY` in SOPS |
| `k8s-routing-peer` | Reusable | written to OpenBao `secret/netbird` |

**Settings → Access Tokens → Create Personal Access Token** → copy it → `NB_PROXY_TOKEN` in SOPS.

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
- `NB_PROXY_TOKEN` is set → `netbird-proxy` starts
- `NB_BIFROST_SETUP_KEY` is set → `netbird-agent` connects to the mesh

### 6c. Write K8s routing peer setup key to OpenBao

```bash
kubectl exec -n openbao openbao-0 -- bao kv patch secret/netbird \
  NETBIRD_SETUP_KEY="<k8s-routing-peer key from step 6a>"
```

The `netbird-peer` StatefulSet CSI volume mounts the secret and the pod starts.

```bash
kubectl get pods -n netbird   # → netbird-peer-0 Running
```

### 6d. Configure the LAN route

In the NetBird dashboard, **Network → Routes → Add Route**:

- **Network:** `192.168.1.0/24`
- **Routing Peer:** `k8s-routing-peer` (appears once pod connects)
- **Enabled:** Yes

Once active, the Bifrost `netbird-agent` can reach `192.168.1.220` through the mesh — enabling Traefik to proxy public services to the cluster gateway.

---

## Phase 7 — Grafana Authentik SSO Setup

Grafana uses Authentik as an OIDC provider. GitHub login flows through Authentik → Grafana.

### 7a. Create OAuth2/OIDC provider in Authentik

Authentik UI → **Applications → Providers → Create → OAuth2/OpenID Provider**

| Field | Value |
|-------|-------|
| Name | `Grafana` |
| Client type | `Confidential` |
| Redirect URI | `https://grafana.madhan.app/login/generic_oauth` |
| Scopes | `openid`, `email`, `profile` |

Copy the **Client ID** and **Client Secret**.

### 7b. Create Application in Authentik

Authentik UI → **Applications → Applications → Create**
- Name: `Grafana`, Slug: `grafana`
- Provider: bind to the provider from 7a

### 7c. Create grafana-admins group

Authentik UI → **Directory → Groups → Create** → name: `grafana-admins`

Add yourself to this group for Admin role in Grafana.

### 7d. Update client_id in code

In `workloads/monitoring/grafana.go`, replace:
```
"client_id": "REPLACE_WITH_AUTHENTIK_CLIENT_ID",
```
with the Client ID from step 7a. Then:
```bash
just synth && git add -A && git commit -m "feat: set Grafana Authentik client_id" && git push
```

### 7e. Write client secret to OpenBao

```bash
kubectl exec -n openbao openbao-0 -- bao kv patch secret/grafana \
  OAUTH_CLIENT_SECRET="<client-secret-from-7a>"
```

Grafana pod will start and SSO will work.

---

## Phase 8 — Publish CDK8s Manifests

If you haven't already, synthesize and push the manifests:

```bash
just synth
git add app/
git commit -m "chore: synth manifests"
git push
```

ArgoCD auto-syncs within 3 minutes:

```bash
kubectl get applications -n argocd
```

---

## Phase 9 — End-to-End Verification

```bash
# DNS split working correctly
dig grafana.madhan.app    # → 178.156.199.250  (public, via Hetzner)
dig headlamp.madhan.app   # → 192.168.1.220    (LAN wildcard, private)

# OpenBao unsealed
kubectl exec -n openbao openbao-0 -- bao status | grep Sealed
# → Sealed  false

# NetBird mesh
# NetBird UI → Peers: bifrost-agent Connected, k8s-routing-peer Connected
# NetBird UI → Routes: 192.168.1.0/24 Active

# Full browser flow
# 1. Open https://grafana.madhan.app
# 2. Redirected to Authentik → log in with GitHub
# 3. Land on Grafana as Viewer (or Admin if in grafana-admins group)

# All apps healthy
kubectl get applications -n argocd
```

---

## Quick Reference: Secrets and Where They Live

| Secret | Stored in | Set during | Notes |
|--------|-----------|-----------|-------|
| `HCLOUD_TOKEN` | SOPS | Phase 0 | Hetzner API access |
| `PROXMOX_PASSWORD` | SOPS | Phase 0 | Proxmox API access |
| `CLOUDFLARE_API_TOKEN` | SOPS → k8s Secret `cert-manager/cloudflare-api-token` | Phase 0 | cert-manager + Traefik ACME |
| `AUTHENTIK_TOKEN` | SOPS → `.secrets.env` | Phase 0 | Authentik akadmin API token |
| `AUTHENTIK_SECRET_KEY` | SOPS → `.env` | Phase 0 | Django secret key |
| `AUTHENTIK_POSTGRESQL_PASSWORD` | SOPS → `.env` | Phase 0 | Authentik Postgres |
| `AUTHENTIK_GITHUB_SECRET` | SOPS | Phase 0 | GitHub OAuth app secret |
| `NB_DATA_STORE_KEY` | SOPS → `.secrets.env` | Phase 0 | **Never rotate** |
| `NB_RELAY_SECRET` | SOPS → `.secrets.env` | Phase 0 | |
| `NB_OWNER_PASSWORD` | SOPS → `.secrets.env` | Phase 0 | Local admin for first NetBird login |
| `NETBIRD_CLIENT_SECRET` | SOPS (used by Pulumi) | Phase 0 | Dex→Authentik OIDC connector |
| `OPENBAO_UNSEAL_KEY` | SOPS → k8s Secret `openbao/openbao-unseal-key` | Phase 2 | Generated by `just openbao-init` |
| `NB_PROXY_TOKEN` | SOPS → `.secrets.env` | Phase 6 | |
| `NB_BIFROST_SETUP_KEY` | SOPS → `.secrets.env` | Phase 6 | |
| `ADMIN_PASSWORD` (Grafana) | OpenBao `secret/grafana` | Phase 2 | |
| `OAUTH_CLIENT_SECRET` (Grafana) | OpenBao `secret/grafana` | Phase 7 | Authentik OIDC client secret |
| `HARBOR_ADMIN_PASSWORD` | OpenBao `secret/harbor` | Phase 2 | |
| `N8N_ENCRYPTION_KEY` | OpenBao `secret/n8n` | Phase 2 | **Never rotate** — re-entering workflows |
| `BOOTSTRAP_PASSWORD` (Rancher) | OpenBao `secret/rancher` | Phase 2 | |
| `NETBIRD_SETUP_KEY` | OpenBao `secret/netbird` | Phase 6 | k8s-routing-peer setup key |

---

## LAN Access Without the Hetzner Hop

Public services (`grafana.madhan.app`) DNS-resolve to the Hetzner VPS for all clients — including LAN users. To bypass this on your machine:

```bash
# /etc/hosts — direct LAN access, skips SSO
192.168.1.220  grafana.madhan.app
```

For LAN-wide bypass, add overrides in your router's DNS (Pi-hole, Unbound, etc.).
