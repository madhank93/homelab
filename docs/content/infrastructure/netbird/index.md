+++
title = "NetBird VPN"
description = "NetBird v0.66 combined server — WireGuard mesh for remote cluster access, with embedded Dex OIDC and Authentik as the upstream identity connector."
weight = 35
+++

## Overview

[NetBird](https://netbird.io) provides a WireGuard-based overlay mesh for secure remote access to the homelab. The **combined server** (`netbirdio/netbird-server:0.66.0`) runs on Bifrost (Hetzner VPS) and consolidates management, signal, relay, STUN, and an **embedded Dex OIDC provider** into a single container.

A **routing peer** pod runs inside Kubernetes and advertises the cluster subnet `192.168.1.0/24` into the mesh — making all cluster services reachable from any connected NetBird client.

```
Your laptop (NetBird client)
    │  WireGuard encrypted tunnel
    │
    ▼  netbird.madhan.app:443
Bifrost VPS
    ├─ Traefik  →  netbird-server:80  (management + signal + relay + embedded Dex)
    │
    └─ WireGuard mesh ──────────────────────────────────┐
                                                         ▼
                                          K8s: netbird-peer pod
                                               routes 192.168.1.0/24
                                                         │
                                                         ▼
                                          Cluster services  192.168.1.220–230
```

---

## Architecture Diagram

{% mermaid() %}
flowchart LR
    subgraph CLIENTS["NetBird Clients"]
        LAP["Laptop / Phone<br/>NetBird client app"]
    end

    subgraph VPS["Bifrost VPS"]
        TR["Traefik :443"]
        NBS["netbird-server<br/>management · signal · relay<br/>+ embedded Dex OIDC"]
        NBD["netbird-dashboard<br/>web UI"]
        NBP["netbird-proxy<br/>*.proxy.madhan.app"]
        NBP2["netbird-agent<br/>WireGuard peer"]
        AUT["Authentik<br/>GitHub SSO"]
    end

    subgraph K8S["Kubernetes Cluster"]
        NBPEER["netbird-peer pod<br/>network: host<br/>routes 192.168.1.0/24"]
        GW["Cilium Gateway<br/>192.168.1.220"]
        PODS["Service Pods"]
    end

    LAP -->|"WireGuard tunnel<br/>netbird.madhan.app:443"| TR
    TR -->|"gRPC paths"| NBS
    TR -->|"dashboard paths"| NBD
    NBD -->|"authenticate via<br/>embedded Dex"| NBS
    NBS -->|"Dex connector:<br/>OIDC upstream"| AUT
    AUT -->|"GitHub OAuth"| LAP
    NBS <-->|"mesh coordination"| NBPEER
    NBS <-->|"mesh coordination"| NBP2
    NBPEER --> GW
    GW --> PODS
    LAP -->|"192.168.1.x via mesh"| NBPEER
{% end %}

---

## Embedded Dex OIDC — Key Design Constraint

NetBird v0.66 combined server **always** runs an embedded [Dex](https://dexidp.io/) OIDC provider. This is hardcoded in the Go source (`Enabled: true` in `ToManagementConfig()`) and cannot be disabled via configuration.

**Consequence:** all JWT tokens that NetBird validates are issued by embedded Dex — not by Authentik directly. Pointing `auth.issuer` to Authentik's URL would cause Dex to claim to be Authentik while signing tokens with its own SQLite-stored keys, producing a JWKS mismatch and `unable to find appropriate key` errors.

### Correct OIDC flow

```
User browser
    │  1. login request
    ▼
Embedded Dex  (issuer: https://netbird.madhan.app/oauth2)
    │  2. redirect to upstream connector
    ▼
Authentik  (https://auth.madhan.app/application/o/netbird/)
    │  3. GitHub OAuth → authenticate user
    ▼
Embedded Dex  (receives callback at /oauth2/callback)
    │  4. issues JWT signed with Dex's own keys
    ▼
NetBird management  (validates JWT against Dex JWKS)
```

Authentik is a **connector inside Dex**, not the token issuer. JWTs are always issued and validated by Dex.

---

## Authentik OIDC App

`core/cloud/authentik.go` creates a confidential OIDC application in Authentik for the Dex→Authentik connector:

| Field | Value |
|-------|-------|
| Client ID | `aumenijDycfG1cQURqH9BNJpV3KVUCoMHGPUVUlT` |
| Client Type | `confidential` |
| Redirect URI | `https://netbird.madhan.app/oauth2/callback` (Dex's callback) |
| Launch URL | `https://netbird.madhan.app/` |
| Client Secret | `NETBIRD_CLIENT_SECRET` from SOPS |

This connector is registered in NetBird via the UI after first login (see [First-Time Setup](#first-time-setup-checklist)).

---

## Server Configuration

`core/cloud/bifrost/netbird/config.yaml` is a **template** — `bootstrap.sh` substitutes `${VAR}` placeholders before starting `netbird-server`.

| Config field | Template value | Final value |
|---|---|---|
| `server.authSecret` | `${NB_RELAY_SECRET}` | base64 secret from `.secrets.env` |
| `store.encryptionKey` | `${NB_DATA_STORE_KEY}` | base64 key from `.secrets.env` |
| `auth.owner.password` | `${NB_OWNER_HASH}` | bcrypt hash generated from `NB_OWNER_PASSWORD` |
| `auth.issuer` | `https://netbird.madhan.app/oauth2` | embedded Dex's own issuer URL |
| `server.exposedAddress` | `https://netbird.madhan.app:443` | static |
| `reverseProxy.trustedHTTPProxies` | `172.30.0.10/32` | Traefik IP in bifrost_net (static) |
| `store.engine` | `sqlite` | static |

> **Note:** The `auth.audience` field and `server.idp` section are silently ignored by NetBird v0.66's combined server config parser — the audience is hardcoded to `"netbird-dashboard"` in Go, and the idp section is not part of `ServerConfig`.

---

## Required Secrets

All secrets come from `secrets/bootstrap.sops.yaml` via `generateBifrostSecretsEnv()`:

| Variable | Purpose | Generate with | Rotate? |
|----------|---------|--------------|---------|
| `NB_DATA_STORE_KEY` | SQLite encryption key | `openssl rand -base64 32` | **No** — DB encrypted with it |
| `NB_RELAY_SECRET` | Relay auth shared secret | `openssl rand -base64 32` | Yes (all peers reconnect) |
| `NB_OWNER_PASSWORD` | Initial admin password for embedded Dex owner account | Any strong password | After Authentik connector confirmed |
| `NETBIRD_CLIENT_SECRET` | Dex→Authentik OIDC connector secret | Authentik UI or `openssl rand -hex 32` | Yes |
| `NB_PROXY_TOKEN` | Personal access token for netbird-proxy | NetBird UI → Settings → Access Tokens | Yes |
| `NB_BIFROST_SETUP_KEY` | Setup key for netbird-agent on Bifrost | NetBird UI → Setup Keys | Yes |

`NB_OWNER_PASSWORD` is only written to `.secrets.env` (and bcrypt-hashed at runtime). It enables the initial local admin login before an external identity provider is connected.

---

## Traefik Routing

NetBird traffic on `netbird.madhan.app` is split across three Traefik routers (defined in `core/cloud/bifrost/traefik/dynamic/services.yml`):

| Router | Rule | Backend | Protocol |
|--------|------|---------|----------|
| `netbird-grpc` | `/signalexchange*/`, `/management*/` (gRPC) | `netbird-server:80` | HTTP/2 cleartext (h2c) |
| `netbird-backend` | `/relay`, `/api`, `/oauth2`, `/ws-proxy/` | `netbird-server:80` | HTTP |
| `netbird-dashboard` | all other `netbird.madhan.app` paths | `netbird-dashboard:80` | HTTP |

The `/oauth2` prefix in `netbird-backend` routes embedded Dex's OIDC endpoints (discovery, token, keys, callback) to `netbird-server`. STUN (UDP/3478) bypasses Traefik entirely — port-forwarded directly to the host.

---

## NetBird Dashboard Environment

`core/cloud/bifrost/netbird/dashboard.env` configures the dashboard container to authenticate against **embedded Dex** (not Authentik directly):

| Variable | Value |
|----------|-------|
| `NETBIRD_MGMT_API_ENDPOINT` | `https://netbird.madhan.app` |
| `AUTH_AUTHORITY` | `https://netbird.madhan.app/oauth2` (embedded Dex) |
| `AUTH_CLIENT_ID` | `netbird-dashboard` (hardcoded static client in Dex) |
| `AUTH_CLIENT_SECRET` | _(empty — public client, no secret)_ |
| `AUTH_AUDIENCE` | `netbird-dashboard` |
| `USE_AUTH0` | `false` |
| `AUTH_SUPPORTED_SCOPES` | `openid profile email groups` |

The dashboard client `netbird-dashboard` is a public OIDC client registered statically inside embedded Dex. No secret is required.

---

## K8s Routing Peer

The `netbird-peer` Deployment in the `netbird` namespace connects to the WireGuard mesh and advertises `192.168.1.0/24` as a route. This makes all cluster services reachable from any NetBird-connected device.

| Setting | Value |
|---------|-------|
| Image | `netbirdio/netbird:latest` |
| Namespace | `netbird` (Pod Security Admission: `privileged`) |
| Setup key | From Infisical `/netbird` → `NETBIRD_SETUP_KEY` |
| Management URL | `https://netbird.madhan.app` |
| Capabilities | `NET_ADMIN`, `SYS_MODULE` |
| `hostNetwork` | `true` — required for kernel WireGuard interface |

The setup key is stored in Infisical (path `/netbird`, key `NETBIRD_SETUP_KEY`) and synced to the `netbird` namespace by an `InfisicalSecret` CR. See [Secrets Flow](/architecture/secrets-flow) for the InfisicalSecret pattern.

---

## NetBird Reverse Proxy (`*.proxy.madhan.app`)

The `netbird-proxy` container exposes NetBird peers as TCP endpoints via `*.proxy.madhan.app`. Traefik routes `*.proxy.madhan.app` TCP traffic to `netbird-proxy:8443` via a wildcard TCP router.

`core/cloud/bifrost/netbird/proxy.env`:
```bash
NB_PROXY_MANAGEMENT_ADDRESS=netbird-server:80
NB_PROXY_DOMAIN=proxy.madhan.app
NB_PROXY_ACME_CERTIFICATES=true
```

`NB_PROXY_TOKEN` is injected from `.secrets.env`. The proxy container is only started by `bootstrap.sh` if `NB_PROXY_TOKEN` is present. If missing, bootstrap prints setup instructions and skips it.

---

## First-Time Setup Checklist

After the initial `just core hetzner up` succeeds and all containers are running:

### Step 1 — Log in with local admin

- [ ] Open `https://netbird.madhan.app`
- [ ] Sign in with the local owner account:
  - Email: `admin@madhan.app`
  - Password: value of `NB_OWNER_PASSWORD` in `secrets/bootstrap.sops.yaml`
- This bypasses SSO and uses the embedded Dex owner account directly

### Step 2 — Connect Authentik as the identity provider

- [ ] Settings → Identity Providers → Add → **Authentik**
  - Client ID: `aumenijDycfG1cQURqH9BNJpV3KVUCoMHGPUVUlT`
  - Client Secret: value of `NETBIRD_CLIENT_SECRET` from SOPS
  - Issuer: `https://auth.madhan.app/application/o/netbird/`
- [ ] Verify redirect URI shown is `https://netbird.madhan.app/oauth2/callback` (pre-configured in Authentik)
- [ ] Test login via GitHub to confirm the connector works

### Step 3 — Create tokens and keys

- [ ] Settings → Access Tokens → Create Personal Access Token — copy it → used as `NB_PROXY_TOKEN`
- [ ] Setup Keys → Add key `bifrost-agent` (Reusable) → copy key value → used as `NB_BIFROST_SETUP_KEY`
- [ ] Setup Keys → Add key `k8s-routing-peer` (Reusable) → copy key value → used as `NETBIRD_SETUP_KEY` in Infisical (Step 5)

### Step 4 — Store tokens and re-deploy

```bash
sops edit secrets/bootstrap.sops.yaml
# Add:
#   NB_PROXY_TOKEN: <personal access token>
#   NB_BIFROST_SETUP_KEY: <bifrost-agent setup key>

just core hetzner up
```

Bootstrap.sh picks up the new tokens and starts `netbird-proxy` and `netbird-agent`.

### Step 5 — Configure K8s routing peer

- [ ] Open Infisical → Project `homelab-prod` → Env `prod` → Path `/netbird`
  - Add `NETBIRD_SETUP_KEY: <value of the k8s-routing-peer setup key created in Step 3>`
  - This is the key the `netbird-peer` pod uses to join the mesh at startup
- [ ] Network Routes → Add Route: network `192.168.1.0/24`, peer group `k8s-routing-peer`
- [ ] Verify both peers appear as **Connected** in the NetBird dashboard
