+++
title = "Secrets Flow"
description = "How bootstrap and runtime secrets are encrypted, distributed, and consumed — SOPS for bootstrap, OpenBao + CSI Driver for runtime."
weight = 40
+++

## Two-Tier Model

All secrets are managed by exactly two systems. The split is intentional — bootstrap secrets are needed to bootstrap the cluster itself, and the cluster hosts the runtime secrets platform:

| Tier | Tool | When | What's stored |
|------|------|------|---------------|
| **Bootstrap** | SOPS + age | One-time setup | OpenBao unseal key, Cloudflare API token, Hetzner token, Authentik keys, NetBird keys |
| **Runtime** | OpenBao + CSI Driver | Continuously available | All app secrets (Grafana, Harbor, n8n, Rancher, …) |

**CDK8s generates zero `Secret` resources.** The CI pipeline needs zero GitHub Actions secrets.

---

## Architecture Diagram

{% mermaid() %}
flowchart TB
    subgraph SOPSFILE["secrets/bootstrap.sops.yaml<br/>age-encrypted · safe to commit to git"]
        S1["CLOUDFLARE_API_TOKEN<br/>HCLOUD_TOKEN · PROXMOX_PASSWORD"]
        S2["AUTHENTIK_TOKEN · AUTHENTIK_SECRET_KEY<br/>AUTHENTIK_POSTGRESQL_PASSWORD"]
        S3["NB_DATA_STORE_KEY · NB_RELAY_SECRET<br/>NB_PROXY_TOKEN · NB_BIFROST_SETUP_KEY"]
        S4["OPENBAO_UNSEAL_KEY"]
    end

    subgraph BOOT["Bootstrap — one-time laptop operations"]
        AGE["age private key<br/>~/.config/sops/age/keys.txt"]
        CS["just create-secrets<br/>bash scripts/create-bootstrap-secrets.sh"]
        KBS1["k8s Secret: openbao-unseal-key<br/>namespace: openbao"]
        KBS2["k8s Secret: cloudflare-api-token<br/>namespace: cert-manager"]
    end

    subgraph BIFROST["Bifrost VPS — auto-provisioned by Pulumi"]
        HG["hetzner.go<br/>generateBifrostSecretsEnv()<br/>generateBifrostDotEnv()"]
        SE[".secrets.env on VPS<br/>CF · NB_ · AUTHENTIK_BOOTSTRAP_TOKEN"]
        DE[".env on VPS<br/>AUTHENTIK_SECRET_KEY<br/>AUTHENTIK_POSTGRESQL_PASSWORD"]
        BS["bootstrap.sh<br/>starts services in order"]
        AK["ak shell<br/>Token.objects.get_or_create<br/>netbird-mgmt-token"]
        NC["sed substitution<br/>config.yaml ${VARS} → real values"]
        NBI["NB_IDP_MGMT_TOKEN<br/>appended to .secrets.env"]
    end

    subgraph RUNTIME["Runtime — OpenBao + CSI Driver"]
        OB["OpenBao<br/>openbao.madhan.app (KV v2)"]
        CSI["Secrets Store CSI Driver<br/>DaemonSet on all nodes"]
        SPC["SecretProviderClass per app<br/>defines: vault role, secret paths"]
        POD["Workload Pods<br/>files at /mnt/secrets<br/>or secretObjects → k8s Secret"]
    end

    AGE -->|decrypts| SOPSFILE
    SOPSFILE -->|sops exec-env| CS
    CS -->|kubectl create secret| KBS1
    CS -->|kubectl create secret| KBS2
    KBS1 -->|unseal sidecar reads key| OB

    SOPSFILE -->|sops exec-env| HG
    HG --> SE & DE
    SE & DE -->|CopyToRemote| BS
    BS --> AK
    AK -->|TOKEN:key stdout| NBI
    BS --> NC
    NBI & NC -->|env ready| BS

    OB -->|K8s auth| CSI
    CSI -->|fetches secrets| SPC
    SPC -->|mounts files or syncs Secret| POD
{% end %}

---

## Bootstrap Secrets (SOPS + age)

`secrets/bootstrap.sops.yaml` is age-encrypted and committed to git. It contains every secret needed to provision infrastructure. Decryption requires the age private key (`~/.config/sops/age/keys.txt`).

### Populating the secrets file

```bash
# Open the encrypted file in $EDITOR — sops decrypts, you edit, it re-encrypts on save
sops secrets/bootstrap.sops.yaml
```

Required secrets (see [Deployment Guide](/getting-started/deployment) for full list):

```yaml
# Infrastructure access
HCLOUD_TOKEN: <hetzner cloud API token>
PROXMOX_PASSWORD: <proxmox root password>
CLOUDFLARE_API_TOKEN: <zone-edit API token>

# Authentik — Bifrost VPS
AUTHENTIK_TOKEN: <60-char random string>        # bootstrap token → also becomes NB_IDP_MGMT_TOKEN
AUTHENTIK_SECRET_KEY: <openssl rand -base64 48> # Django secret key
AUTHENTIK_POSTGRESQL_PASSWORD: <openssl rand -hex 16>
AUTHENTIK_GITHUB_SECRET: <github oauth app secret>

# NetBird — Bifrost VPS
NB_DATA_STORE_KEY: <openssl rand -base64 32>    # SQLite encryption key — never change after first deploy
NB_RELAY_SECRET: <openssl rand -base64 32>      # relay auth shared secret
NB_PROXY_TOKEN: <netbird personal access token> # added after first login to NetBird
NB_BIFROST_SETUP_KEY: <netbird setup key>       # added after first login to NetBird

# OpenBao — K8s secrets platform
OPENBAO_UNSEAL_KEY: <openssl rand -base64 32>   # generated once on first deploy
```

> **`NB_DATA_STORE_KEY` and `NB_RELAY_SECRET` must never change** after the first deploy — the NetBird SQLite database is encrypted with these values. Rotating them means losing all peer registrations.

### Bootstrap script

```bash
# Creates two k8s Secrets from the SOPS file:
just create-secrets
# → sops exec-env secrets/bootstrap.sops.yaml 'bash scripts/create-bootstrap-secrets.sh'
# → kubectl create secret generic openbao-unseal-key -n openbao
# → kubectl create secret generic cloudflare-api-token -n cert-manager
```

Both secrets carry `argocd.argoproj.io/sync-options: Prune=false` — ArgoCD never deletes them.

---

## Bifrost Auto-Provisioning

The Bifrost VPS secrets are fully automated by Pulumi. No manual SSH required.

**`just core hetzner up` does the following in sequence:**

1. **`generateBifrostSecretsEnv()`** reads SOPS env vars, writes `./cloud/bifrost/.secrets.env` on the laptop
2. **`generateBifrostDotEnv()`** reads SOPS env vars, writes `./cloud/bifrost/.env` on the laptop
3. **`CopyToRemote`** uploads the entire `./cloud/bifrost/` directory to `/etc/bifrost/` on the VPS
4. **`bootstrap.sh`** runs on the VPS:
   - Starts all containers in dependency order, waiting for health at each step
   - After Authentik is healthy, runs `docker exec authentik-server ak shell` to create `netbird-mgmt-token` via Django ORM
   - Writes `NB_IDP_MGMT_TOKEN` to `.secrets.env` on the VPS
   - Runs `sed` to substitute `${NB_RELAY_SECRET}`, `${NB_DATA_STORE_KEY}`, `${NB_IDP_MGMT_TOKEN}` in `netbird/config.yaml`
   - Starts `netbird-server` (now has all required secrets)

See [Hetzner Bifrost](/infrastructure/hetzner-bifrost) for the full bootstrap sequence.

---

## Runtime Secrets (OpenBao + CSI Driver)

After the cluster is up, all application secrets are managed by [OpenBao](/workloads/secrets/openbao). Apps consume secrets via the Secrets Store CSI Driver — secrets are mounted as files in pods, or synced to k8s Secrets via `secretObjects`.

### Patterns

**Pattern A — file-only** (no k8s Secret created):

```
Pod → CSI volume mount → /mnt/secrets/ADMIN_PASSWORD
env: GF_SECURITY_ADMIN_PASSWORD__FILE=/mnt/secrets/ADMIN_PASSWORD
```

Used by: **Grafana**

**Pattern B — secretObjects sync** (k8s Secret created and kept in sync):

```
SecretProviderClass.secretObjects → k8s Secret (e.g. harbor-admin)
Helm chart: existingSecret: harbor-admin
```

Used by: **Harbor**, **n8n**, **Rancher**, **NetBird peer**

### Apps and their OpenBao paths

| App | OpenBao path | Pattern | k8s Secret |
|-----|-------------|---------|------------|
| Grafana | `secret/data/grafana` | A (file) | — |
| Harbor | `secret/data/harbor` | B (sync) | `harbor-admin` |
| n8n | `secret/data/n8n` | B (sync) | `n8n-db` |
| Rancher | `secret/data/rancher` | B (sync) | `rancher-bootstrap` |
| NetBird peer | `secret/data/netbird` | B (sync) | `netbird-setup-key` |

### One-time setup

After first deploy, run `just openbao-setup` to configure K8s auth, policies, roles, and write initial secrets. See [OpenBao](/workloads/secrets/openbao) for details.

---

## What Lives Where — Quick Reference

| Secret | Lives in | Set during | Rotatable? |
|--------|---------|-----------|-----------|
| `NB_DATA_STORE_KEY` | SOPS → `.secrets.env` | Phase 0 | **No** — NetBird DB encrypted with it |
| `NB_RELAY_SECRET` | SOPS → `.secrets.env` | Phase 0 | Yes (all peers reconnect) |
| `AUTHENTIK_TOKEN` | SOPS → `.secrets.env` as `AUTHENTIK_BOOTSTRAP_TOKEN` | Phase 0 | Yes |
| `NB_IDP_MGMT_TOKEN` | Auto-written to `.secrets.env` by `bootstrap.sh` | Auto | Yes |
| `NB_PROXY_TOKEN` | SOPS → `.secrets.env` | After first NetBird login | Yes |
| `NB_BIFROST_SETUP_KEY` | SOPS → `.secrets.env` | After first NetBird login | Yes |
| `CLOUDFLARE_API_TOKEN` | SOPS → k8s Secret | Phase 0 | Yes |
| `OPENBAO_UNSEAL_KEY` | SOPS → k8s Secret | Phase 0 | Yes (redeploy OpenBao) |
| App passwords | OpenBao KV | Phase 2 (`just openbao-setup`) | Yes |
| `NETBIRD_SETUP_KEY` (k8s) | OpenBao `/netbird` | After first NetBird login | Yes |
