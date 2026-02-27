+++
title = "Service Access & Internet Exposure"
description = "DNS split strategy, public vs internal service routing, and how to expose or restrict services"
weight = 25
+++

## Architecture Overview

Homelab services are LAN-only by default. Selectively exposing a service to the internet is a single-line config change, controlled by `publicServices` in `infra/pulumi/cloudflare.go`.

```
*.madhan.app          → 192.168.1.220  (wildcard, LAN gateway — default for all services)
*.internal.madhan.app → 192.168.1.220  (explicit internal label)
auth.madhan.app       → 23.121.200.108 (Authentik — always public)
netbird.madhan.app    → 23.121.200.108 (NetBird — always public)
proxy.madhan.app      → 23.121.200.108 (NetBird expose base — always public)
*.proxy.madhan.app    → 23.121.200.108 (NetBird expose wildcard)
grafana.madhan.app    → 23.121.200.108 (overrides wildcard → internet accessible)
harbor.madhan.app     → 23.121.200.108 (overrides wildcard → internet accessible)
headlamp.madhan.app   → 192.168.1.220  (no explicit record → LAN wildcard → private)
```

Cloudflare resolves specific records before wildcards. A service with no explicit A record falls through to the `*.madhan.app` wildcard → private IP → unreachable from the internet.

## Network Flow Diagrams

### Public service (e.g. Grafana)

```
Internet Browser
  → Cloudflare: grafana.madhan.app → 23.121.200.108
  → Traefik (TLS termination on Hetzner VPS)
  → [no session] → Authentik ForwardAuth
      → auth.madhan.app → GitHub OAuth → redirect back
  → [session cookie set] → k8s-gateway (http://192.168.1.220)
       ↑ reachable via WireGuard (netbird-agent on Hetzner routes 192.168.1.0/24)
  → Cilium HTTPRoute: grafana.madhan.app → Grafana pod
```

### Internal service (e.g. Headlamp)

```
LAN Browser
  → DNS: headlamp.madhan.app → 192.168.1.220 (LAN wildcard)
  → Cilium Gateway → HTTPRoute → Headlamp pod
  (No Traefik, no Hetzner, no SSO — direct LAN access only)
```

## How to Expose a Service to the Internet

**Step 1**: Add the service name to `publicServices` in `infra/pulumi/cloudflare.go`:

```go
// Before
var publicServices = []string{"grafana", "harbor"}

// After — adding Headlamp
var publicServices = []string{"grafana", "harbor", "headlamp"}
```

**Step 2**: Add a Traefik router to `infra/pulumi/bifrost/traefik/dynamic/services.yml`:

```yaml
http:
  routers:
    headlamp:
      rule: Host(`headlamp.madhan.app`)
      middlewares: [authentik-forwardauth]
      service: k8s-gateway
      tls:
        certResolver: cloudflare-dns
```

**Step 3**: Run `just pulumi platform up` — this:
- Creates the Cloudflare A record `headlamp.madhan.app → 23.121.200.108`
- Generates `traefik/dynamic/public-services.yml` with the new router
- Copies updated config to Hetzner VPS via CopyToRemote
- Traefik file-watcher hot-reloads the new route (no container restart needed)

## How to Revoke Internet Access

Remove the service from `publicServices` in `cloudflare.go` and remove its router from `services.yml`, then run `just pulumi platform up`.

The Cloudflare A record is deleted. DNS falls back to `*.madhan.app → 192.168.1.220` (private) — service becomes LAN-only again automatically.

## DNS Split Strategy

| Domain Pattern | Resolves To | Accessible From |
|----------------|-------------|-----------------|
| `*.madhan.app` (wildcard) | `192.168.1.220` | LAN only |
| `*.internal.madhan.app` | `192.168.1.220` | LAN only (explicit label) |
| `auth.madhan.app` | `23.121.200.108` | Internet |
| `netbird.madhan.app` | `23.121.200.108` | Internet |
| `proxy.madhan.app` | `23.121.200.108` | Internet |
| `*.proxy.madhan.app` | `23.121.200.108` | Internet (NetBird expose) |
| `grafana.madhan.app` | `23.121.200.108` | Internet + ForwardAuth |
| `harbor.madhan.app` | `23.121.200.108` | Internet + ForwardAuth |

## Public Services Config (Single Source of Truth)

The `publicServices` slice in `cloudflare.go` is the authoritative list. When you run `pulumi up`:

1. **Cloudflare** receives A records for each entry (overriding the LAN wildcard)
2. **`public-services.yml`** is generated with Traefik routers (one per service)
3. Both are applied atomically in one pulumi run

## NetBird Expose Feature (Temporary Sharing)

For temporary access without DNS changes, use the NetBird expose feature:

```bash
# From any machine on the NetBird mesh
netbird expose 8080

# Output: foo.proxy.madhan.app → your local port 8080
```

This routes via `*.proxy.madhan.app → Traefik TCP passthrough → netbird-proxy → WireGuard mesh → your machine`. TLS is provided by the reverse-proxy container's ACME.

This is useful for:
- Sharing a dev server temporarily with a colleague
- Exposing a local service for webhook testing
- Access that doesn't need SSO (the TCP passthrough bypasses Authentik)

## Verification

```bash
# DNS resolves correctly
dig grafana.madhan.app     # → 23.121.200.108  (public, explicit A record)
dig headlamp.madhan.app    # → 192.168.1.220   (LAN wildcard, private)

# Unauthenticated access redirects to SSO
curl -I https://grafana.madhan.app   # → 302 to auth.madhan.app

# K8s routing peer connected
kubectl get pods -n netbird           # → Running

# NetBird UI: Peers → bifrost-agent Connected, k8s-routing-peer Connected
# NetBird UI: Routes → 192.168.1.0/24 → Active
```
