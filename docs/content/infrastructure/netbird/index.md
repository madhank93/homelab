+++
title = "NetBird VPN"
description = "NetBird v0.66 combined server — WireGuard mesh for remote cluster access, with Authentik SSO and automated IDP token provisioning."
weight = 35
+++

## Overview

[NetBird](https://netbird.io) provides a WireGuard-based overlay mesh for secure remote access to the homelab. The **combined server** (`netbirdio/netbird-server:0.66.0`) runs on Bifrost (Hetzner VPS) and consolidates management, signal, relay, and STUN into a single container.

A **routing peer** pod runs inside Kubernetes and advertises the cluster subnet `192.168.1.0/24` into the mesh — making all cluster services reachable from any connected NetBird client.

```
Your laptop (NetBird client)
    │  WireGuard encrypted tunnel
    │
    ▼  netbird.madhan.app:443
Bifrost VPS
    ├─ Traefik  →  netbird-server:80  (management + signal + relay)
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
        NBS["netbird-server<br/>management · signal · relay"]
        NBD["netbird-dashboard<br/>web UI"]
        NBP["netbird-proxy<br/>*.proxy.madhan.app"]
        NBP2["netbird-agent<br/>WireGuard peer"]
    end

    subgraph K8S["Kubernetes Cluster"]
        NBPEER["netbird-peer pod<br/>network: host<br/>routes 192.168.1.0/24"]
        GW["Cilium Gateway<br/>192.168.1.220"]
        PODS["Service Pods"]
    end

    LAP -->|"WireGuard tunnel<br/>netbird.madhan.app:443"| TR
    TR -->|"gRPC paths"| NBS
    TR -->|"dashboard paths"| NBD
    NBS <-->|"mesh coordination"| NBPEER
    NBS <-->|"mesh coordination"| NBP2
    NBPEER --> GW
    GW --> PODS
    LAP -->|"192.168.1.x via mesh"| NBPEER
{% end %}

---

## Authentik Integration

NetBird uses Authentik as its OIDC identity provider. Users log in to the NetBird dashboard via GitHub (proxied through Authentik). The NetBird server also calls the Authentik management API to sync user groups.

### How NB_IDP_MGMT_TOKEN is provisioned

This token is the Authentik API key used by NetBird for user sync. It cannot exist until Authentik is running, creating a chicken-and-egg dependency. The [bootstrap.sh](/infrastructure/hetzner-bifrost) script resolves this automatically:

1. Waits for `authentik-server` to report healthy
2. Runs `docker exec authentik-server ak shell` — a Python script in Authentik's Django context
3. Calls `Token.objects.get_or_create(identifier='netbird-mgmt-token', key=AUTHENTIK_BOOTSTRAP_TOKEN)`
4. Appends `NB_IDP_MGMT_TOKEN=<key>` to `/etc/bifrost/.secrets.env`
5. Starts `netbird-server` with the token now available

**No manual steps required.** On subsequent `pulumi up` runs, if the token is already in `.secrets.env`, provisioning is skipped.

---

## Server Configuration

`core/cloud/bifrost/netbird/config.yaml` is a **template** — `bootstrap.sh` substitutes `${VAR}` placeholders before starting `netbird-server` (NetBird v0.66 does not expand env vars in its config file natively).

| Config field | Template value | Substituted from |
|---|---|---|
| `server.authSecret` | `${NB_RELAY_SECRET}` | `.secrets.env` |
| `store.encryptionKey` | `${NB_DATA_STORE_KEY}` | `.secrets.env` |
| `idp.authentik.managementToken` | `${NB_IDP_MGMT_TOKEN}` | `.secrets.env` (auto-provisioned) |
| `server.exposedAddress` | `https://netbird.madhan.app:443` | static |
| `auth.issuer` | `https://auth.madhan.app/application/o/netbird/` | static |
| `reverseProxy.trustedHTTPProxies` | `172.30.0.10/32` | static (Traefik IP in bifrost_net) |
| `store.engine` | `sqlite` | static |

---

## Required Secrets

All secrets come from `secrets/bootstrap.sops.yaml` via Pulumi's `generateBifrostSecretsEnv()`:

| Variable | Purpose | Generate with | Rotate? |
|----------|---------|--------------|---------|
| `NB_DATA_STORE_KEY` | SQLite encryption key | `openssl rand -base64 32` | **No** — DB encrypted with it |
| `NB_RELAY_SECRET` | Relay auth shared secret | `openssl rand -base64 32` | Yes (all peers reconnect) |
| `NB_IDP_MGMT_TOKEN` | Authentik API token (user sync) | **Auto-provisioned** by bootstrap.sh | Yes |
| `NB_PROXY_TOKEN` | Personal access token for netbird-proxy | NetBird UI → Settings → Access Tokens | Yes |
| `NB_BIFROST_SETUP_KEY` | Setup key for netbird-agent on Bifrost | NetBird UI → Setup Keys | Yes |

---

## Traefik Routing

NetBird traffic on `netbird.madhan.app` is split across three Traefik routers (defined in `core/cloud/bifrost/traefik/dynamic/services.yml`):

| Router | Rule | Backend | Protocol |
|--------|------|---------|----------|
| `netbird-grpc` | `/signalexchange*/`, `/management*/` (gRPC) | `netbird-server:80` | HTTP/2 cleartext (h2c) |
| `netbird-backend` | `/relay`, `/api`, `/oauth2`, `/ws-proxy/` | `netbird-server:80` | HTTP |
| `netbird-dashboard` | all other `netbird.madhan.app` paths | `netbird-dashboard:80` | HTTP |

STUN (UDP/3478) bypasses Traefik entirely — port-forwarded directly to the host.

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

## NetBird Dashboard Environment

`core/cloud/bifrost/netbird/dashboard.env` configures the dashboard container:

| Variable | Value |
|----------|-------|
| `NETBIRD_MGMT_API_ENDPOINT` | `https://netbird.madhan.app` |
| `AUTH_AUTHORITY` | `https://auth.madhan.app/application/o/netbird/` |
| `AUTH_CLIENT_ID` | Authentik OAuth2 client ID for NetBird app |
| `USE_AUTH0` | `false` |
| `AUTH_SUPPORTED_SCOPES` | `openid profile email groups` |

---

## NetBird Reverse Proxy (`*.proxy.madhan.app`)

The `netbird-proxy` container exposes NetBird peers as TCP endpoints via `*.proxy.madhan.app`. Traefik routes `*.proxy.madhan.app` TCP traffic to `netbird-proxy:8443` via a wildcard TCP router.

`core/cloud/bifrost/netbird/proxy.env`:
```bash
NB_PROXY_MANAGEMENT_ADDRESS=netbird-server:80
NB_PROXY_DOMAIN=proxy.madhan.app
NB_PROXY_ACME_CERTIFICATES=true
```

`NB_PROXY_TOKEN` is injected from `.secrets.env` (not in `proxy.env` — Docker Compose env_file does not expand variables).

The proxy container is only started by `bootstrap.sh` if `NB_PROXY_TOKEN` is present in `.secrets.env`. If it's missing, the bootstrap prints setup instructions and skips it.

---

## First-Time Setup Checklist

After the initial `just core hetzner up` succeeds and all containers are running:

- [ ] Open `https://netbird.madhan.app` → Log in with GitHub (via Authentik)
- [ ] **Setup Keys** → Add key `bifrost-agent` (Reusable) for the VPS WireGuard agent
- [ ] **Setup Keys** → Add key `k8s-routing-peer` (Reusable) for the K8s pod
- [ ] **Settings → Access Tokens** → Create a Personal Access Token — copy it
- [ ] `sops edit secrets/bootstrap.sops.yaml`:
  - Add `NB_PROXY_TOKEN: <personal access token>`
  - Add `NB_BIFROST_SETUP_KEY: <bifrost-agent key>`
- [ ] `just core hetzner up` — bootstrap.sh starts `netbird-proxy` and `netbird-agent` with the new tokens
- [ ] Open Infisical → Project `homelab-prod` → Env `prod` → Path `/netbird`:
  - Add `NETBIRD_SETUP_KEY: <k8s-routing-peer key>`
- [ ] **Network Routes** → Add Route: network `192.168.1.0/24`, peer `k8s-routing-peer`
- [ ] Verify peers are connected in the NetBird dashboard
