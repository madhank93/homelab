+++
title = "Deployment Guide"
description = "Step-by-step sequence for deploying the full homelab stack from scratch."
weight = 20
+++

This guide covers the complete deployment sequence. Each phase depends on the previous one completing successfully.

---

## Phase 0 — Local prerequisites (one-time)

```bash
# Tools
brew install pulumi talosctl kubectl just sops age go
npm install -g cdk8s-cli

# age key (one-time — back this up securely)
mkdir -p ~/.config/sops/age
age-keygen -o ~/.config/sops/age/keys.txt
# ↑ prints: Public key: age1abc123...

# Add to shell profile (required on every machine you run pulumi from)
echo 'export SOPS_AGE_KEY_FILE="$HOME/.config/sops/age/keys.txt"' >> ~/.zshrc
source ~/.zshrc
```

Register your public key in `.sops.yaml` at the repo root, then populate the bootstrap secrets file:

```bash
sops infra/secrets/bootstrap.sops.yaml
```

The file must contain these keys (add any missing ones):

```yaml
INFISICAL_DB_PASSWORD: <strong-random-password>
INFISICAL_ENCRYPTION_KEY: <32-char-hex>
INFISICAL_AUTH_SECRET: <32-char-hex>
REDIS_PASSWORD: <strong-random-password>
CLOUDFLARE_API_TOKEN: <cloudflare-dns-zone-edit-token>
AUTHENTIK_TOKEN: <authentik-api-token>
AUTHENTIK_GITHUB_SECRET: <github-oauth-app-secret>
NB_DATA_STORE_KEY: <openssl rand -base64 32>
NB_RELAY_SECRET: <openssl rand -base64 32>
```

> `CLOUDFLARE_API_TOKEN` is used by both cert-manager (K8s) and Traefik (Bifrost VPS, DNS ACME challenge). Same token, automatically reused — no manual step.

---

## Phase 1 — Bootstrap K8s cluster

```bash
# Creates two k8s Secrets before ArgoCD runs:
#   infisical/infisical-secrets
#   cert-manager/cloudflare-api-token
just create-secrets

# Provision Proxmox VMs → bootstrap Talos → install Cilium + ArgoCD (~15 min)
just pulumi talos up

# Apply Cilium Gateway, IP pool, HTTPRoutes
just pulumi platform up
```

Verify the cluster is up:

```bash
talosctl health --nodes 192.168.1.211
kubectl get nodes
kubectl get applications -n argocd   # ArgoCD starts syncing apps
```

ArgoCD will deploy all apps from the `v0.1.5-manifests` branch. Most apps will be `Degraded` at this point — that's expected. They need Infisical secrets to become healthy.

---

## Phase 2 — Infisical first-time setup

Wait for Infisical to be running (~5 min after Phase 1):

```bash
kubectl get pods -n infisical   # wait for Running
```

Open `http://infisical.madhan.app` (LAN only).

1. Create an account → **Organization** → **Project** → slug: `homelab-prod`
2. **Environments** → Add → name: `prod`, slug: `prod`
3. Add secrets (navigate to **Secrets** → `prod` environment):

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
DB_PASSWORD     = <strong password>
N8N_ENCRYPTION_KEY = <32-char random string — record this, needed on every rebuild>
```

**Path `/rancher`**
```
BOOTSTRAP_PASSWORD = <strong password>
```

*(Skip `/netbird` for now — NetBird isn't running yet. The netbird-peer pod will stay Pending until Phase 5.)*

4. Create a Service Token: **Settings** → **Service Tokens** → **Add Token**
   - Name: `k8s-homelab`, Environment: `prod`, Path: `/`, No expiry, Read only
   - Copy the token (shown once only)

5. Store the token in K8s:

```bash
kubectl create secret generic infisical-service-token \
  --from-literal=infisicalToken=<paste-token> \
  -n infisical
```

Within 60 seconds, `InfisicalSecret` CRs sync and apps start becoming healthy. Check:

```bash
kubectl get applications -n argocd
# grafana, harbor, n8n, rancher should move to Healthy
```

---

## Phase 3 — Authentik (Pulumi)

```bash
just pulumi authentik up
```

This creates in Authentik:
- GitHub OAuth source
- NetBird OIDC application (client ID: `aumenijDycfG1cQURqH9BNJpV3KVUCoMHGPUVUlT`)
- Homelab ForwardAuth proxy provider + embedded outpost (covers `*.madhan.app`)
- Service account `sa-netbird` + API token

Note the token for Bifrost:
```bash
cd infra/pulumi && pulumi stack output NetbirdServiceToken --stack authentik
```

---

## Phase 4 — Bifrost VPS deploy

### 4a. Update config and DNS records

```bash
# Creates Cloudflare DNS records:
#   auth.madhan.app, netbird.madhan.app, proxy.madhan.app → 23.121.200.108
#   *.proxy.madhan.app → 23.121.200.108
#   grafana.madhan.app, harbor.madhan.app → 23.121.200.108
just pulumi cloudflare up

# Deploy/update Hetzner VPS + copy Bifrost config (docker-compose, traefik, netbird)
# This also auto-generates:
#   bifrost/traefik/dynamic/public-services.yml  (Traefik public routes)
#   bifrost/.secrets.env                          (CF_DNS_API_TOKEN from SOPS)
just pulumi hetzner up
```

### 4b. Populate `bifrost/.env` on the VPS

SSH into Bifrost and append the remaining secrets:

```bash
ssh root@23.121.200.108

cat >> /etc/bifrost/.env << 'EOF'
NB_DATA_STORE_KEY=<value from bootstrap SOPS>
NB_RELAY_SECRET=<value from bootstrap SOPS>
NB_IDP_MGMT_TOKEN=<NetbirdServiceToken output from Phase 3>
EOF
```

> `NB_DATA_STORE_KEY` and `NB_RELAY_SECRET` are in bootstrap SOPS (added in Phase 0). Fresh install: generate with `openssl rand -base64 32`. These must never change after first deploy — the NetBird peer database is encrypted with them.

### 4c. Start containers

```bash
# Still on the VPS
cd /etc/bifrost
docker compose pull
docker compose up -d

# Verify all 7 containers are Up
docker compose ps
```

Expected containers: `traefik`, `authentik-postgres`, `authentik-server`, `authentik-worker`, `netbird-server`, `netbird-dashboard`, `netbird-proxy`, `netbird-agent`

Wait ~2 minutes for Traefik to obtain the wildcard TLS cert via Cloudflare DNS challenge:
```bash
docker logs traefik --follow   # look for "Obtained certificate"
```

### 4d. Verify Bifrost is working

```bash
# From your laptop
curl -I https://auth.madhan.app          # → 200 (Authentik login page)
curl -I https://netbird.madhan.app       # → 200 (NetBird dashboard)
curl -I https://grafana.madhan.app       # → 302 to auth.madhan.app (ForwardAuth working)
```

---

## Phase 5 — NetBird setup

### 5a. Create setup keys

Open `https://netbird.madhan.app` → log in with GitHub (via Authentik).

**Setup Keys** → **Add Key**:

| Name | Type | Used for |
|------|------|----------|
| `bifrost-agent` | Reusable | netbird-agent container on Bifrost host |
| `k8s-routing-peer` | Reusable | netbird-peer pod in K8s |

### 5b. Add bifrost-agent key to Bifrost

```bash
ssh root@23.121.200.108

cat >> /etc/bifrost/.env << 'EOF'
NB_BIFROST_SETUP_KEY=<bifrost-agent key from UI>
EOF

# Generate proxy token (netbird-server must be running)
docker exec netbird-server netbird-server generate-proxy-token
# ↑ copy the output

cat >> /etc/bifrost/.env << 'EOF'
NB_PROXY_TOKEN=<proxy token from above>
EOF

# Restart containers that use the new env vars
cd /etc/bifrost
docker compose restart netbird-agent netbird-proxy
```

### 5c. Add k8s-routing-peer key to Infisical

Open `http://infisical.madhan.app` → Project `homelab-prod` → Env `prod` → Path `/netbird`:

```
NETBIRD_SETUP_KEY = <k8s-routing-peer key from UI>
```

Within 60 seconds the `InfisicalSecret` in the `netbird` namespace syncs. The `netbird-peer` pod starts and connects to the NetBird mesh.

Check:
```bash
kubectl get pods -n netbird   # → Running
```

### 5d. Configure the LAN route in NetBird UI

**Network Routes** → **Add Route**:
- Network: `192.168.1.0/24`
- Routing Peer: `k8s-routing-peer` (appears once pod connects)
- Enabled: Yes

Once the route is active, the Bifrost `netbird-agent` can reach `192.168.1.220` through the WireGuard mesh. This is how Traefik proxies public services to the K8s gateway.

---

## Phase 6 — Publish CDK8s manifests

The netbird-peer chart is new — it needs to be synthesized and pushed to the manifests branch:

```bash
# From repo root
just synth   # generates app/netbird/ directory

# Commit to the manifests branch
git add app/netbird/
git commit -m "feat: Add NetBird routing peer"
git push origin v0.1.5-manifests
```

ArgoCD auto-syncs within 3 minutes. Watch:
```bash
kubectl get application netbird -n argocd
```

---

## Phase 7 — End-to-end verification

```bash
# DNS resolves correctly
dig grafana.madhan.app    # → 23.121.200.108  (public, via Hetzner)
dig headlamp.madhan.app   # → 192.168.1.220   (LAN wildcard, private)

# Public service redirects to SSO when unauthenticated
curl -sI https://grafana.madhan.app | grep location
# → location: https://auth.madhan.app/...

# NetBird mesh
# NetBird UI → Peers: bifrost-agent Connected, k8s-routing-peer Connected
# NetBird UI → Routes: 192.168.1.0/24 Active

# Full end-to-end
# Open https://grafana.madhan.app in browser
# → Redirected to auth.madhan.app → Log in with GitHub → Land on Grafana
```

---

## Quick reference — what goes where

| Secret | Stored in | Set during |
|--------|-----------|-----------|
| `INFISICAL_DB_PASSWORD` etc | Bootstrap SOPS | Phase 0 |
| `CLOUDFLARE_API_TOKEN` | Bootstrap SOPS | Phase 0 |
| `AUTHENTIK_TOKEN`, `AUTHENTIK_GITHUB_SECRET` | Bootstrap SOPS | Phase 0 |
| `NB_DATA_STORE_KEY`, `NB_RELAY_SECRET` | Bootstrap SOPS → `bifrost/.env` on VPS | Phase 0 + 4b |
| `NB_IDP_MGMT_TOKEN` | `bifrost/.env` on VPS | Phase 4b (Pulumi output) |
| `NB_BIFROST_SETUP_KEY` | `bifrost/.env` on VPS | Phase 5b |
| `NB_PROXY_TOKEN` | `bifrost/.env` on VPS | Phase 5b |
| `NETBIRD_SETUP_KEY` | Infisical `/netbird` | Phase 5c |
| All app passwords | Infisical paths | Phase 2 |

---

## LAN access without hairpin routing

Public services (`grafana.madhan.app`, `harbor.madhan.app`) DNS-resolve to the Hetzner VPS IP for all clients. LAN users still route through Hetzner.

To bypass this on your laptop:

```bash
# Add to /etc/hosts — direct LAN access, no SSO
192.168.1.220  grafana.madhan.app harbor.madhan.app
```

For LAN-wide bypass, set the overrides in your router's DNS (Pi-hole, etc.).
