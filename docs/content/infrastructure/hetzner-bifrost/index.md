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
    ├─ bifrostEnvFiles               writes env/<consumer>.env from SOPS (mode 0600)
    ├─ CopyToRemote                  uploads /etc/bifrost/ to VPS
    └─ remote.Command → bootstrap.sh
           ├─ pre-flight             validate secrets, wait for cloud-init, delete legacy .env/.secrets.env
           ├─ harden_host()          key-only SSH, fail2ban, unattended-upgrades
           ├─ 1/8  traefik           TLS termination + routing
           ├─ 2/8  authentik-postgres
           ├─ 3/8  authentik         DB migrations (one-off, before the servers)
           ├─ 4/8  authentik-server + authentik-worker
           ├─      process_netbird_config()
           │         sed: substitute ${NB_RELAY_SECRET}, ${NB_DATA_STORE_KEY}
           │         python: bcrypt hash NB_OWNER_PASSWORD → ${NB_OWNER_HASH}
           ├─ 5/8  netbird-server + netbird-dashboard
           ├─ 6/8  netbird-agent     WireGuard peer (only if NB_SETUP_KEY set)
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
| `${NB_RELAY_SECRET}` | relay auth secret from `env/bootstrap.env` | `sed` (base64 — safe) |
| `${NB_DATA_STORE_KEY}` | SQLite encryption key from `env/bootstrap.env` | `sed` (base64 — safe) |
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

### netbird-agent setup key

`netbird-agent` reads `NB_SETUP_KEY` from its own `env/netbird-agent.env` via `env_file:`, so a plain `$COMPOSE up -d netbird-agent` picks it up. `bootstrap.sh` skips the agent until that key is set.

---

## Generated Files

Pulumi writes these files on the **laptop** before uploading:

Pulumi writes one env file per consumer under `core/cloud/bifrost/env/` (mode `0600`), from the `bifrostEnvFiles` list in `hetzner.go`. Each file has exactly one consumer, so no container sees another service's secrets:

| File | Consumer | Variables (SOPS key) |
|------|----------|----------------------|
| `traefik.env` | `traefik` | `CF_DNS_API_TOKEN` (`CLOUDFLARE_API_TOKEN`) |
| `authentik-postgres.env` | `authentik-postgres` | `POSTGRES_PASSWORD` (`AUTHENTIK_POSTGRESQL_PASSWORD`) |
| `authentik-server.env` | `authentik-server` | `AUTHENTIK_SECRET_KEY`, `AUTHENTIK_POSTGRESQL__PASSWORD` (`AUTHENTIK_POSTGRESQL_PASSWORD`), `AUTHENTIK_BOOTSTRAP_TOKEN` (`AUTHENTIK_TOKEN`) |
| `authentik-worker.env` | `authentik-worker` | `AUTHENTIK_SECRET_KEY`, `AUTHENTIK_POSTGRESQL__PASSWORD` |
| `netbird-proxy.env` | `netbird-proxy` | `NB_PROXY_TOKEN` (optional) |
| `netbird-agent.env` | `netbird-agent` | `NB_SETUP_KEY` (`NB_BIFROST_SETUP_KEY`) |
| `bootstrap.env` | `bootstrap.sh` | `NB_DATA_STORE_KEY`, `NB_RELAY_SECRET`, `NB_OWNER_PASSWORD` |

`netbird-server` has no env file: its secrets are substituted into `netbird/config.yaml` (see above).

All files are gitignored and regenerated on every `just core hetzner up`.

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
| UDP | 3478 | STUN (NetBird NAT traversal) |

No TURN ports are open: the NetBird relay runs over WebSocket on 443.

---

## Host Hardening

`harden_host()` in `bootstrap.sh` runs before any container starts:

- **SSH** — drop-in `/etc/ssh/sshd_config.d/10-bifrost-hardening.conf` sets `PasswordAuthentication no`, `KbdInteractiveAuthentication no` and `PermitRootLogin prohibit-password` (Pulumi still connects as root with a key). It is validated with `sshd -t` before the reload; an invalid drop-in is removed and the bootstrap stops.
- **fail2ban** — installed and enabled with Debian's default `sshd` jail.
- **unattended-upgrades** — daily security updates.

Traefik and Gatus are hardened too:

- **TLS 1.2 minimum** (`tls.options.default`) and an `hsts` middleware (1 year; no `includeSubDomains`, no preload) on the `websecure` entrypoint, so it covers every HTTPS router.
- **ForwardAuth `trustForwardHeader: false`** — Traefik is the edge, so client-sent `X-Forwarded-*` headers are never trusted.
- **Gatus** — the status page is public, so every endpoint hides connection errors (they embed resolved IPs); the API VIP endpoint also hides its hostname, port and URL.

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

> **Adding a new public service:** add the service name to `publicServices` in `core/cloud/cloudflare.go`, then run `just core cloudflare up` and `just core hetzner up`. The DNS record and Traefik route are created automatically; the route gets Authentik ForwardAuth unless the entry sets `SkipAuth: true` (as Grafana does, since it runs its own OIDC login).

---

## After the First Deploy

After `just core hetzner up` succeeds, NetBird setup must be completed manually. See [NetBird VPN — First-Time Setup Checklist](/infrastructure/netbird/#first-time-setup-checklist) for the full step-by-step sequence (connecting Authentik, creating setup keys, configuring the K8s routing peer, and verifying end-to-end connectivity).
