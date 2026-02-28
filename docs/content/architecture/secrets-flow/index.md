+++
title = "Secrets Flow"
description = "How bootstrap and runtime secrets are encrypted, distributed, and consumed — including automated Bifrost VPS secret provisioning."
weight = 40
+++

## Two-Tier Model

All secrets are managed by exactly two systems. The split is intentional — bootstrap secrets are needed to bootstrap the cluster itself, and the cluster hosts the runtime secrets platform:

| Tier | Tool | When | What's stored |
|------|------|------|---------------|
| **Bootstrap** | SOPS + age | One-time setup | Infisical DB credentials, Cloudflare API token, Hetzner token, Authentik keys, NetBird keys |
| **Runtime** | Infisical operator | Continuously synced | All app secrets (Grafana, Harbor, n8n, Rancher, …) |

**CDK8s generates zero `Secret` resources.** The CI pipeline needs zero GitHub Actions secrets.

---

## Architecture Diagram

{% mermaid() %}
flowchart TB
    subgraph SOPSFILE["secrets/bootstrap.sops.yaml<br/>age-encrypted · safe to commit to git"]
        S1["CLOUDFLARE_API_TOKEN<br/>HCLOUD_TOKEN · PROXMOX_PASSWORD"]
        S2["AUTHENTIK_TOKEN · AUTHENTIK_SECRET_KEY<br/>AUTHENTIK_POSTGRESQL_PASSWORD"]
        S3["NB_DATA_STORE_KEY · NB_RELAY_SECRET<br/>NB_PROXY_TOKEN · NB_BIFROST_SETUP_KEY"]
        S4["INFISICAL_DB_PASSWORD · INFISICAL_AUTH_SECRET<br/>INFISICAL_ENCRYPTION_KEY · REDIS_PASSWORD"]
    end

    subgraph BOOT["Bootstrap — one-time laptop operations"]
        AGE["age private key<br/>~/.config/sops/age/keys.txt"]
        CS["just create-secrets<br/>bash scripts/create-bootstrap-secrets.sh"]
        KBS1["k8s Secret: infisical-secrets<br/>namespace: infisical"]
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

    subgraph RUNTIME["Runtime — Infisical continuous sync"]
        ISVC["Infisical Platform<br/>infisical.madhan.app"]
        IOP["Infisical Operator<br/>InfisicalSecret CRs"]
        APPSEC["App k8s Secrets<br/>grafana-admin · harbor-admin<br/>n8n-db · rancher-bootstrap"]
        POD["Workload Pods<br/>envFrom / volumeMount"]
    end

    AGE -->|decrypts| SOPSFILE
    SOPSFILE -->|sops exec-env| CS
    CS -->|kubectl create secret| KBS1
    CS -->|kubectl create secret| KBS2
    KBS1 -->|mounts service token| ISVC

    SOPSFILE -->|sops exec-env| HG
    HG --> SE & DE
    SE & DE -->|CopyToRemote| BS
    BS --> AK
    AK -->|TOKEN:key stdout| NBI
    BS --> NC
    NBI & NC -->|env ready| BS

    ISVC -->|operator polls API| IOP
    IOP -->|creates/updates| APPSEC
    APPSEC -->|envFrom| POD
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

# Infisical — K8s secrets platform
INFISICAL_DB_PASSWORD: <strong random>
INFISICAL_ENCRYPTION_KEY: <openssl rand -hex 16>
INFISICAL_AUTH_SECRET: <openssl rand -hex 24>
REDIS_PASSWORD: <strong random>
```

> **`NB_DATA_STORE_KEY` and `NB_RELAY_SECRET` must never change** after the first deploy — the NetBird SQLite database is encrypted with these values. Rotating them means losing all peer registrations.

### Bootstrap script

```bash
# Creates two k8s Secrets from the SOPS file:
just create-secrets
# → sops exec-env secrets/bootstrap.sops.yaml 'bash scripts/create-bootstrap-secrets.sh'
# → kubectl create secret generic infisical-secrets -n infisical
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

## Runtime Secrets (Infisical)

After the cluster is up, all application secrets are managed by the [Infisical operator](/workloads/management/infisical). Each app defines an `InfisicalSecret` CR that points to a path in the Infisical project.

### Apps using Infisical

| App | Infisical path | k8s Secret | Keys |
|-----|---------------|------------|------|
| Grafana | `/grafana` | `grafana-admin` | `ADMIN_PASSWORD` |
| Harbor | `/harbor` | `harbor-admin` | `HARBOR_ADMIN_PASSWORD` |
| n8n | `/n8n` | `n8n-db` | `DB_PASSWORD`, `N8N_ENCRYPTION_KEY` |
| Rancher | `/rancher` | `rancher-bootstrap` | `BOOTSTRAP_PASSWORD` |
| NetBird peer | `/netbird` | `netbird-setup-key` | `NETBIRD_SETUP_KEY` |

### InfisicalSecret pattern

```yaml
apiVersion: secrets.infisical.com/v1alpha1
kind: InfisicalSecret
metadata:
  name: grafana-admin
  namespace: grafana
  annotations:
    # Required — Infisical CRD missing projectSlug breaks SSA validation
    argocd.argoproj.io/sync-options: ServerSideApply=false
spec:
  hostAPI: https://infisical.madhan.app/api
  resyncInterval: 60
  authentication:
    serviceToken:
      serviceTokenSecretReference:
        secretName: infisical-service-token
        secretNamespace: infisical
      secretsScope:
        envSlug: prod
        secretsPath: /grafana
  managedSecretReference:
    secretName: grafana-admin
    secretNamespace: grafana
```

The operator polls Infisical every 60 seconds. Secrets are updated in-place without pod restarts (unless using [Reloader](/workloads/support/reloader)).

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
| `CLOUDFLARE_API_TOKEN` | SOPS → `.secrets.env` + k8s Secret | Phase 0 | Yes |
| App passwords | Infisical | Phase 2 | Yes |
| `NETBIRD_SETUP_KEY` (k8s) | Infisical `/netbird` | After first NetBird login | Yes |
