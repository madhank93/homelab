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
           ├─ 1/5  traefik              TLS termination + routing
           ├─ 2/5  authentik-postgres
           ├─ 3/5  authentik-server + worker
           ├─      process_netbird_config()
           │         sed: substitute ${NB_RELAY_SECRET}, ${NB_DATA_STORE_KEY}
           │         python: bcrypt hash NB_OWNER_PASSWORD → ${NB_OWNER_HASH}
           ├─ 4/5  netbird-server     management + signal + relay + STUN + embedded Dex
           ├─      netbird-dashboard  (started in same step)
           └─ 5/5  netbird-proxy      (only if NB_PROXY_TOKEN set)
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
| `netbird-server` | `netbirdio/netbird-server:0.66.0` | Combined: management + signal + relay + STUN + embedded Dex OIDC |
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
    PF["Preflight<br/>validate 5 required secrets<br/>wait for cloud-init<br/>check docker compose"]

    subgraph S1["Step 1/5"]
        T["docker compose up -d traefik<br/>wait_healthy 60s"]
    end
    subgraph S2["Step 2/5"]
        AP["docker compose up -d authentik-postgres<br/>wait_healthy 120s"]
    end
    subgraph S3["Step 3/5"]
        AS["docker compose up -d authentik-server authentik-worker<br/>wait_healthy 300s"]
    end

    subgraph CFG["process_netbird_config()"]
        SED["sed: replace base64 placeholders<br/>\${NB_RELAY_SECRET}<br/>\${NB_DATA_STORE_KEY}"]
        PY1["python3: bcrypt.hashpw(NB_OWNER_PASSWORD)<br/>→ owner_hash"]
        PY2["python3: replace \${NB_OWNER_HASH}<br/>in netbird/config.yaml"]
        SED --> PY1 --> PY2
    end

    subgraph S4["Step 4/5"]
        NS["docker compose up -d netbird-server netbird-dashboard<br/>wait_healthy 120s / 60s"]
    end
    subgraph S5["Step 5/5"]
        NP{"NB_PROXY_TOKEN set?"}
        NPY["docker compose up -d netbird-proxy<br/>wait_healthy 60s"]
        NPN["skip — show setup instructions"]
    end

    PF --> S1 --> S2 --> S3 --> CFG --> S4 --> S5
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

### netbird/config.yaml template substitution

NetBird v0.66 does not expand `${VAR}` in its config file — the YAML is read verbatim. `bootstrap.sh` substitutes three placeholders before starting `netbird-server`:

| Placeholder | Substituted with | Method |
|-------------|-----------------|--------|
| `${NB_RELAY_SECRET}` | relay auth secret from `.secrets.env` | `sed` (base64 — safe) |
| `${NB_DATA_STORE_KEY}` | SQLite encryption key from `.secrets.env` | `sed` (base64 — safe) |
| `${NB_OWNER_HASH}` | bcrypt hash of `NB_OWNER_PASSWORD` | Python (bcrypt contains `$` and `/` — breaks sed) |

The bcrypt hash is generated at runtime inside `process_netbird_config()`:

```bash
owner_hash=$(_OWNER_PASS="$owner_pass" python3 - <<'PYEOF'
import bcrypt, os
p = os.environ['_OWNER_PASS'].encode()
print(bcrypt.hashpw(p, bcrypt.gensalt(10)).decode())
PYEOF
)
```

The password is passed via an environment variable, never via command-line arguments, and never logged.

This substitution is idempotent: on re-runs `CopyToRemote` restores the original template from the laptop, then the placeholders are substituted again.

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
| `NB_OWNER_PASSWORD` | `NB_OWNER_PASSWORD` | Yes |
| `NB_PROXY_TOKEN` | `NB_PROXY_TOKEN` | No (optional) |
| `NB_BIFROST_SETUP_KEY` | `NB_BIFROST_SETUP_KEY` | No (optional) |

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
- `sed` and Python substitutions in `config.yaml` are a no-op if placeholders are already replaced (re-deploy always restores the template via `CopyToRemote` first)

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

After `just core hetzner up` succeeds, complete the one-time setup in this order:

1. **Log in to NetBird with the local admin account**
   - Open `https://netbird.madhan.app`
   - Sign in with email: `admin@madhan.app`, password: the value of `NB_OWNER_PASSWORD` in SOPS
   - This uses the embedded Dex owner account, which only works before an external IdP is configured

2. **Connect Authentik as the Identity Provider**
   - Settings → Identity Providers → Add → **Authentik**
   - Client ID: `aumenijDycfG1cQURqH9BNJpV3KVUCoMHGPUVUlT`
   - Client Secret: value of `NETBIRD_CLIENT_SECRET` from SOPS
   - Issuer: `https://auth.madhan.app/application/o/netbird/`
   - This configures Authentik as a connector inside embedded Dex — users will then log in via GitHub (Authentik) through Dex

3. **Create a Personal Access Token**
   - Settings → Access Tokens → Create — copy the token

4. **Create Setup Keys**
   - Setup Keys → Add key `bifrost-agent` (Reusable) for the VPS WireGuard agent
   - Setup Keys → Add key `k8s-routing-peer` (Reusable) for the K8s pod

5. **Add tokens to SOPS and re-deploy**
   ```bash
   sops edit secrets/bootstrap.sops.yaml
   # Add:
   #   NB_PROXY_TOKEN: <personal access token>
   #   NB_BIFROST_SETUP_KEY: <bifrost-agent setup key>
   just core hetzner up
   ```
   Bootstrap.sh picks up the new tokens and starts `netbird-proxy` and `netbird-agent`.

6. **Add K8s routing peer setup key to Infisical**
   - Project `homelab-prod` → Env `prod` → Path `/netbird`
   - Add `NETBIRD_SETUP_KEY: <k8s-routing-peer key>`

7. **Create the cluster route in NetBird**
   - Network Routes → Add Route: network `192.168.1.0/24`, peer `k8s-routing-peer`

See [NetBird VPN](/infrastructure/netbird) and the [Deployment Guide](/getting-started/deployment) for the complete setup sequence.
