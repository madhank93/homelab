+++
title = "Hetzner Bifrost"
description = "Hetzner VPS running Traefik v3.3, NetBird v0.66, and Authentik — the automated public edge for the homelab."
weight = 30
+++

## Overview

**Bifrost** is a lightweight Hetzner Cloud VPS that acts as the public edge of the homelab. A single Pulumi command (`just core hetzner up`) provisions the server, copies all config files, and runs a bootstrap script that starts every service in dependency order — fully unattended. No manual SSH required.

```
just core hetzner up
    │
    ├─ generateBifrostSecretsEnv()   writes .secrets.env from SOPS
    ├─ generateBifrostDotEnv()       writes .env from SOPS
    ├─ CopyToRemote                  uploads /etc/bifrost/ to VPS
    └─ remote.Command → bootstrap.sh
           ├─ 1/6  traefik           TLS termination + routing
           ├─ 2/6  authentik-postgres
           ├─ 3/6  authentik-server + worker
           ├─      ak shell           auto-provision NB_IDP_MGMT_TOKEN
           ├─      sed               substitute ${VARS} in netbird/config.yaml
           ├─ 4/6  netbird-server    management + signal + relay + STUN
           ├─ 5/6  netbird-dashboard
           └─ 6/6  netbird-proxy     (only if NB_PROXY_TOKEN set)
```

---

## Services

All services run via `docker compose` from `/etc/bifrost/`:

| Container | Image | Role |
|-----------|-------|------|
| `traefik` | `traefik:v3.3` | TLS termination, ForwardAuth, routing |
| `authentik-server` | `ghcr.io/goauthentik/server:2025.10.4` | GitHub OAuth, OIDC, ForwardAuth provider |
| `authentik-worker` | `ghcr.io/goauthentik/server:2025.10.4` | Background tasks, email, jobs |
| `authentik-postgres` | `postgres:16.6-alpine` | Authentik database |
| `netbird-server` | `netbirdio/netbird-server:0.66.0` | Combined: management + signal + relay + STUN |
| `netbird-dashboard` | `netbirdio/dashboard:latest` | NetBird web UI |
| `netbird-proxy` | `netbirdio/reverse-proxy:latest` | `*.proxy.madhan.app` TCP passthrough |
| `netbird-agent` | `netbirdio/netbird:latest` | WireGuard peer, advertises `192.168.1.0/24` |

All containers share `bifrost_net` (172.30.0.0/24). Traefik is the only container with public ports 80/443.

---

## Pulumi Configuration

The Hetzner stack reads from `core/config.yml`:

```yaml
hetzner:
  server_name: bifrost-public-vps1
  image: ubuntu-24.04
  server_type: cpx21
  location: ash            # Ashburn, VA
  ssh_key: mac-ssh
  vps_ip: "178.156.199.250"
```

`HCLOUD_TOKEN` comes from SOPS. All other secrets are injected automatically (see [Generated Files](#generated-files) below).

---

## Bootstrap Automation

### What `bootstrap.sh` does

The bootstrap script runs on the VPS after every config or secret change. It is idempotent — safe to re-run.

{% mermaid() %}
flowchart TB
    PF["Preflight<br/>validate 4 required secrets<br/>wait for cloud-init<br/>check docker compose"]

    subgraph S1["Step 1/6"]
        T["docker compose up -d traefik<br/>wait_healthy 60s"]
    end
    subgraph S2["Step 2/6"]
        AP["docker compose up -d authentik-postgres<br/>wait_healthy 120s"]
    end
    subgraph S3["Step 3/6"]
        AS["docker compose up -d authentik-server authentik-worker<br/>wait_healthy 300s"]
    end

    subgraph PROV["NB_IDP_MGMT_TOKEN Provisioning"]
        CHK{"NB_IDP_MGMT_TOKEN<br/>already in .secrets.env?"}
        AKS["docker exec authentik-server ak shell<br/>Token.objects.get_or_create<br/>identifier='netbird-mgmt-token'<br/>key=AUTHENTIK_BOOTSTRAP_TOKEN"]
        WRT["echo NB_IDP_MGMT_TOKEN=TOKEN >> .secrets.env"]
        SKIP["skip — already provisioned"]
    end

    subgraph CFG["Config Template Substitution"]
        SED["sed s|\$\{NB_RELAY_SECRET\}|...|g<br/>sed s|\$\{NB_DATA_STORE_KEY\}|...|g<br/>sed s|\$\{NB_IDP_MGMT_TOKEN\}|...|g<br/>netbird/config.yaml overwritten"]
    end

    subgraph S4["Step 4/6"]
        NS["docker compose up -d netbird-server<br/>wait_healthy 120s"]
    end
    subgraph S5["Step 5/6"]
        ND["docker compose up -d netbird-dashboard<br/>wait_healthy 60s"]
    end
    subgraph S6["Step 6/6"]
        NP{"NB_PROXY_TOKEN set?"}
        NPY["docker compose up -d netbird-proxy<br/>wait_healthy 60s"]
        NPN["skip — show setup instructions"]
    end

    PF --> S1 --> S2 --> S3
    S3 --> CHK
    CHK -->|No| AKS --> WRT --> CFG
    CHK -->|Yes| SKIP --> CFG
    CFG --> S4 --> S5 --> S6
    NP -->|Yes| NPY
    NP -->|No| NPN
{% end %}

### Health polling

Each `wait_healthy <container>` call polls `docker inspect` every 5 seconds:

```
state=running + health=healthy  →  ready ✓
state=running + health=none     →  ready ✓ (no healthcheck configured)
timeout exceeded                →  prints last 30 log lines, exits 1
```

Authentik has an explicit `healthcheck: test: ["CMD-SHELL", "ak healthcheck"]` and a 60s start period. The script waits up to 300s for it.

### NB_IDP_MGMT_TOKEN provisioning

NetBird requires an Authentik API token to sync users via the management API. This token can't be created until Authentik is running — a chicken-and-egg problem that `bootstrap.sh` solves automatically.

After Authentik reports healthy, `bootstrap.sh` runs a Python script inside the container via `ak shell` (Authentik's Django management shell):

```python
# Runs inside the authentik-server container
from authentik.core.models import Token, TokenIntents, User

key = os.environ.get('AUTHENTIK_BOOTSTRAP_TOKEN', '')
user = User.objects.filter(username='akadmin').first()

t, created = Token.objects.get_or_create(
    identifier='netbird-mgmt-token',
    defaults={
        'user': user,
        'intent': TokenIntents.INTENT_API,
        'key': key,          # same value as SOPS AUTHENTIK_TOKEN
        'expiring': False,
        'description': 'NetBird IDP management token (auto-provisioned)',
    }
)
print(f'TOKEN:{t.key}')      # only this line goes to stdout
```

The bash wrapper greps for `TOKEN:` prefix and writes the value to `.secrets.env`:

```
NB_IDP_MGMT_TOKEN=<token>    (appended to /etc/bifrost/.secrets.env)
```

On re-runs, if `NB_IDP_MGMT_TOKEN` is already in `.secrets.env`, provisioning is skipped entirely. The token identifier `netbird-mgmt-token` is stable — `get_or_create` never duplicates it.

### netbird/config.yaml template substitution

NetBird v0.66 does not expand `${VAR}` in its config file. The config is read verbatim. `bootstrap.sh` uses `sed` to substitute three placeholders before starting `netbird-server`:

| Placeholder in `config.yaml` | Replaced with |
|-------------------------------|---------------|
| `${NB_RELAY_SECRET}` | relay auth secret from `.secrets.env` |
| `${NB_DATA_STORE_KEY}` | SQLite encryption key from `.secrets.env` |
| `${NB_IDP_MGMT_TOKEN}` | Authentik API token (just provisioned) |

This substitution is idempotent: on re-runs `CopyToRemote` restores the template, and `sed` replaces the placeholders again.

---

## Generated Files

Pulumi writes these files on the **laptop** before uploading:

### `core/cloud/bifrost/.secrets.env`

Generated by `generateBifrostSecretsEnv()` from SOPS env vars:

| Variable | SOPS key | Required |
|----------|----------|---------|
| `CF_DNS_API_TOKEN` | `CLOUDFLARE_API_TOKEN` | Yes |
| `NB_DATA_STORE_KEY` | `NB_DATA_STORE_KEY` | Yes |
| `NB_RELAY_SECRET` | `NB_RELAY_SECRET` | Yes |
| `AUTHENTIK_BOOTSTRAP_TOKEN` | `AUTHENTIK_TOKEN` | Yes |
| `NB_PROXY_TOKEN` | `NB_PROXY_TOKEN` | No (optional) |
| `NB_BIFROST_SETUP_KEY` | `NB_BIFROST_SETUP_KEY` | No (optional) |

`NB_IDP_MGMT_TOKEN` is **not** written here by Pulumi — it is appended by `bootstrap.sh` at runtime.

### `core/cloud/bifrost/.env`

Generated by `generateBifrostDotEnv()`. Docker Compose reads this file for `${VAR}` interpolation in `docker-compose.yml` (required for `POSTGRES_PASSWORD=${AUTHENTIK_POSTGRESQL_PASSWORD}`):

| Variable | SOPS key |
|----------|----------|
| `AUTHENTIK_SECRET_KEY` | `AUTHENTIK_SECRET_KEY` |
| `AUTHENTIK_POSTGRESQL_PASSWORD` | `AUTHENTIK_POSTGRESQL_PASSWORD` |

Both files are gitignored and regenerated on every `just core hetzner up`.

---

## Re-running / Updating

`bootstrap.sh` is triggered by a SHA-256 hash of all bifrost config files plus all secret values. Pulumi re-runs the script automatically when:

- Any file in `core/cloud/bifrost/` changes
- Any SOPS secret value changes (detected by hashing)

The script is idempotent:
- `docker compose up -d` is a no-op for already-running containers with the same image
- `NB_IDP_MGMT_TOKEN` provisioning is skipped if already in `.secrets.env`
- `sed` substitution in `config.yaml` is a no-op if placeholders are already replaced

---

## Firewall

Hetzner Cloud firewall (`bifrost-fw`) applied to the VPS:

| Protocol | Port(s) | Purpose |
|----------|---------|---------|
| TCP | 22 | SSH (management) |
| TCP | 80 | HTTP → Traefik (redirects to HTTPS) |
| TCP | 443 | HTTPS → Traefik |
| TCP + UDP | 3478 | STUN (NetBird NAT traversal) |
| TCP + UDP | 5349 | TURNS (TLS TURN relay) |
| UDP | 50000–50500 | TURN ephemeral relay range |

---

## DNS Records

Managed by `core/cloud/cloudflare.go`:

| Hostname | Points to | Traffic handled by |
|----------|-----------|-------------------|
| `auth.madhan.app` | `178.156.199.250` | Authentik (on VPS) |
| `netbird.madhan.app` | `178.156.199.250` | NetBird dashboard + server (on VPS) |
| `proxy.madhan.app` | `178.156.199.250` | NetBird reverse proxy |
| `*.proxy.madhan.app` | `178.156.199.250` | NetBird reverse proxy wildcard |
| `grafana.madhan.app` | `178.156.199.250` | Traefik → WireGuard → cluster |
| `harbor.madhan.app` | `178.156.199.250` | Traefik → WireGuard → cluster |

> **Adding a new public service:** add the service name to `publicServices` in `core/cloud/cloudflare.go`, then run `just core cloudflare up` and `just core hetzner up`. The DNS record and Traefik route (with ForwardAuth) are created automatically.

---

## After the First Deploy

After `just core hetzner up` succeeds, log in to https://netbird.madhan.app with GitHub (via Authentik) to complete the remaining one-time steps:

1. Create a **Personal Access Token** (Settings → Access Tokens)
2. Create two **Setup Keys** (one for the Bifrost agent, one for the K8s routing peer)
3. Add `NB_PROXY_TOKEN` and `NB_BIFROST_SETUP_KEY` to SOPS: `sops edit secrets/bootstrap.sops.yaml`
4. Run `just core hetzner up` again — bootstrap.sh picks up the new tokens and starts `netbird-proxy` and `netbird-agent`

See [NetBird VPN](/infrastructure/netbird) and the [Deployment Guide](/getting-started/deployment) for the complete setup sequence.
