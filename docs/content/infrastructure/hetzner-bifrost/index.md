+++
title = "Hetzner Bifrost"
description = "Hetzner VPS running Traefik, NetBird, and Authentik — the automated public edge for the homelab."
weight = 10
+++

## What is Hetzner Bifrost?

[Hetzner Cloud](https://www.hetzner.com/cloud) is a European cloud provider offering low-cost VPS instances. **Bifrost** is this homelab's Hetzner VPS — named after the Norse rainbow bridge connecting worlds — acting as the public edge that bridges the internet to the private cluster.

## Why a Separate Edge VPS?

Running a public-facing edge on a cheap VPS (rather than exposing the home cluster directly) keeps the cluster's private IP range off the internet entirely. Traefik on Bifrost handles TLS, ForwardAuth, and routing; the cluster itself only receives traffic that has passed through the VPS's WireGuard tunnel.

## How It's Used Here

A single Pulumi command (`just core hetzner up`) provisions the Hetzner VPS, uploads all config, and runs `bootstrap.sh` which starts Traefik, Authentik, NetBird server, and a WireGuard peer in dependency order — fully unattended. The VPS sits in front of every public `madhan.app` service and routes traffic through the NetBird WireGuard mesh to the cluster's Cilium gateway at `192.168.1.220`.

```
just core hetzner up
    │
    ├─ generateBifrostSecretsEnv()   writes .secrets.env from SOPS
    ├─ generateBifrostDotEnv()       writes .env from SOPS
    ├─ CopyToRemote                  uploads /etc/bifrost/ to VPS
    └─ remote.Command → bootstrap.sh
           ├─ pre-flight             validate secrets, wait for cloud-init
           ├─ 1/8  traefik           TLS termination + routing
           ├─ 2/8  authentik-postgres
           ├─ 3/8  authentik         DB migrations (one-off, before the servers)
           ├─ 4/8  authentik-server + authentik-worker
           ├─      process_netbird_config()
           │         sed: substitute ${NB_RELAY_SECRET}, ${NB_DATA_STORE_KEY}
           │         python: bcrypt hash NB_OWNER_PASSWORD → ${NB_OWNER_HASH}
           ├─ 5/8  netbird-server + netbird-dashboard
           ├─ 6/8  netbird-agent     WireGuard peer (only if NB_BIFROST_SETUP_KEY set)
           ├─ 7/8  netbird-proxy     (only if NB_PROXY_TOKEN set)
           └─ 8/8  gatus             uptime monitoring
```

> **netbird-server vs netbird-agent on the same host:** These are two distinct roles.
> `netbird-server` is the coordination plane — it manages the mesh, assigns WireGuard keys,
> and distributes routes to peers. It does **not** create a WireGuard interface on the host.
> `netbird-agent` is a WireGuard peer that joins the mesh, establishes tunnels, and receives
> routes advertised by other peers (e.g. `192.168.1.0/24` from `k8s-routing-peer`). Without
> the agent, Traefik cannot reach `192.168.1.220` and all public service proxying returns 504.

---

## Services

All services run via `docker compose` from `/etc/bifrost/`:

| Container | Image | Role |
|-----------|-------|------|
| `traefik` | `traefik` | TLS termination, ForwardAuth, routing |
| `authentik-server` | `ghcr.io/goauthentik/server` | GitHub OAuth, OIDC, ForwardAuth provider |
| `authentik-worker` | `ghcr.io/goauthentik/server` | Background tasks, email, jobs |
| `authentik-postgres` | `postgres` | Authentik database |
| `netbird-server` | `netbirdio/netbird-server` | Combined: management + signal + relay + STUN + embedded Dex OIDC |
| `netbird-dashboard` | `netbirdio/dashboard` | NetBird web UI |
| `netbird-proxy` | `netbirdio/reverse-proxy` | `*.proxy.madhan.app` TCP passthrough |
| `netbird-agent` | `netbirdio/netbird` | WireGuard peer, advertises `192.168.1.0/24` |
| `gatus` | `ghcr.io/twin/gatus` | Uptime monitoring at `uptime.madhan.app` |

Tags are pinned in {{ src(path="core/cloud/bifrost/docker-compose.yml", label="docker-compose.yml") }}
and listed in the [Software Inventory](@/architecture/software-inventory.md). The four
NetBird images must move together — mixing versions breaks the management protocol.

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

The sequence above is the whole of it; each step waits for the previous one to
report healthy before starting.

### Health polling

Each `wait_healthy <container>` call polls `docker inspect` every 5 seconds:

```
state=running + health=healthy  →  ready ✓
state=running + health=none     →  ready ✓ (no healthcheck configured)
timeout exceeded                →  prints last 30 log lines, exits 1
```

Authentik has an explicit `healthcheck: test: ["CMD-SHELL", "ak healthcheck"]` and a 60s start period. The script waits up to 300s for it.

### netbird/config.yaml template substitution

NetBird does not expand `${VAR}` in its config file — the YAML is read verbatim. `bootstrap.sh` substitutes three placeholders before starting `netbird-server`:

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

### netbird-agent setup key — Docker Compose env var caveat

`docker-compose.yml` uses Compose-level interpolation for `netbird-agent`'s setup key:

```yaml
netbird-agent:
  network_mode: host
  environment:
    - NB_SETUP_KEY=${NB_BIFROST_SETUP_KEY}
```

Docker Compose `${VAR}` interpolation reads from the OS environment or `.env` file — **not** from `env_file:`. Since `NB_BIFROST_SETUP_KEY` is only in `.secrets.env` (container-level), it was always blank when using `$COMPOSE up -d netbird-agent` directly.

`bootstrap.sh` works around this by exporting the secret to the OS environment for that specific command:

```bash
NB_BIFROST_SETUP_KEY=$(read_secret NB_BIFROST_SETUP_KEY) $COMPOSE up -d netbird-agent
```

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

> **Adding a new public service:** add the service name to `publicServices` in `core/cloud/cloudflare.go`, then run `just core cloudflare up` and `just core hetzner up`. The DNS record and Traefik route (with ForwardAuth) are created automatically.

---

## After the First Deploy

After `just core hetzner up` succeeds, NetBird setup must be completed manually. See [NetBird VPN — First-Time Setup Checklist](/infrastructure/netbird/#first-time-setup-checklist) for the full step-by-step sequence (connecting Authentik, creating setup keys, configuring the K8s routing peer, and verifying end-to-end connectivity).
