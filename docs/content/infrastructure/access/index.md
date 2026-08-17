+++
title = "Service Access & Internet Exposure"
description = "Cloudflare DNS, the publicServices toggle, and how services get exposed to the internet or kept on the LAN"
weight = 50
aliases = ["/infrastructure/dns-tls/"]
+++

## What is Service Access?

Service access describes how homelab services are reachable — whether from the local network only, or from the internet. The routing strategy combines Cloudflare DNS, the Bifrost Hetzner VPS, and the cluster's Cilium Gateway to implement a clean LAN-vs-public split.

## Why This Split?

Keeping services LAN-only by default minimizes attack surface — a service is private until explicitly promoted to public with a single config change. Adding ForwardAuth via Authentik on the public path means no service is directly exposed without authentication, even if it lacks its own login.

## How It's Used Here

All services default to `*.madhan.app → 192.168.1.220` (LAN wildcard). Adding a service name to `publicServices` in `cloudflare.go` creates an explicit Cloudflare A record pointing to the Bifrost VPS and generates the corresponding Traefik router with ForwardAuth — one `pulumi up` call makes a service public or private.

## Architecture Overview

Homelab services are LAN-only by default. Selectively exposing a service to the internet is a single-line config change, controlled by `publicServices` in `core/cloud/cloudflare.go`.

```
*.madhan.app          → 192.168.1.220  (wildcard, LAN gateway — default for all services)
*.internal.madhan.app → 192.168.1.220  (explicit internal label)
auth.madhan.app       → 178.156.199.250 (Authentik — always public)
netbird.madhan.app    → 178.156.199.250 (NetBird — always public)
proxy.madhan.app      → 178.156.199.250 (NetBird expose base — always public)
*.proxy.madhan.app    → 178.156.199.250 (NetBird expose wildcard)
grafana.madhan.app    → 178.156.199.250 (overrides wildcard → internet accessible)
headlamp.madhan.app   → 192.168.1.220  (no explicit record → LAN wildcard → private)
```

Cloudflare resolves specific records before wildcards. A service with no explicit A record falls through to the `*.madhan.app` wildcard → private IP → unreachable from the internet.

## Network Flow Diagrams

### Public service (e.g. Grafana)

```
Internet Browser
  → Cloudflare: grafana.madhan.app → 178.156.199.250
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

**Step 1**: Add the service name to `publicServices` in `core/cloud/cloudflare.go`:

```go
// Before (only grafana is public by default)
var publicServices = []PublicService{
    {Name: "grafana", SkipAuth: true},
}

// After — adding Headlamp (requires Authentik forwardauth)
var publicServices = []PublicService{
    {Name: "grafana", SkipAuth: true},
    {Name: "headlamp", SkipAuth: false},
}
```

**Step 2**: Run `just core hetzner up` — Pulumi generates the Traefik router. DNS is a
separate stack; step 3 runs both. Alternatively, add a Traefik router manually to `core/cloud/bifrost/traefik/dynamic/services.yml`:

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

**Step 3**: Run `just core cloudflare up && just core hetzner up` — this:
- Creates the Cloudflare A record `headlamp.madhan.app → 178.156.199.250`
- Generates `traefik/dynamic/public-services.yml` with the new router
- Copies updated config to Hetzner VPS via CopyToRemote
- Traefik file-watcher hot-reloads the new route (no container restart needed)

## How to Revoke Internet Access

Remove the service from `publicServices` in `cloudflare.go` and remove its router from `services.yml`, then run `just core cloudflare up && just core hetzner up`.

The Cloudflare A record is deleted. DNS falls back to `*.madhan.app → 192.168.1.220` (private) — service becomes LAN-only again automatically.

## Cloudflare

Cloudflare hosts `madhan.app` and every record is provisioned by Pulumi
({{ src(path="core/cloud/cloudflare.go") }}, `just core cloudflare up`). Its API
is also what cert-manager's DNS-01 solver uses to issue the wildcard certificate
without exposing an HTTP endpoint — see [cert-manager](@/platform/cert-manager/index.md).

All records are **DNS-only** (Proxied: false, orange cloud off). Cloudflare
resolves specific records before wildcards, so a service with no explicit A
record falls through to `*.madhan.app` → private IP → unreachable from the
internet. That fallback is the mechanism behind revoking access below: delete
the record and the service is LAN-only again, with no other change.

## DNS Split Strategy

| Domain Pattern | Resolves To | Accessible From |
|----------------|-------------|-----------------|
| `*.madhan.app` (wildcard) | `192.168.1.220` | LAN only |
| `auth.madhan.app` | `178.156.199.250` | Internet |
| `netbird.madhan.app` | `178.156.199.250` | Internet |
| `proxy.madhan.app` | `178.156.199.250` | Internet |
| `*.proxy.madhan.app` | `178.156.199.250` | Internet (NetBird expose) |
| `grafana.madhan.app` | `178.156.199.250` | Internet + ForwardAuth |

### LAN user accessing a public service (hairpin routing)

When a service is added to `publicServices`, Cloudflare resolves it to `178.156.199.250` for
**everyone** — including devices already on the LAN. A LAN browser accessing `grafana.madhan.app`
will still route out to the Hetzner VPS and tunnel back through WireGuard to reach the pod, even
though both the client and the pod are on the same network.

```
LAN Browser
  → DNS (Cloudflare): grafana.madhan.app → 178.156.199.250   ← public IP, not LAN
  → Traefik on Hetzner VPS
  → WireGuard tunnel → 192.168.1.220 → Grafana pod
  (unnecessary internet round-trip for a LAN client)
```

To avoid this, run a local DNS resolver (e.g. Pi-hole) that overrides public service records to
the LAN gateway IP for devices on your network:

```
# Pi-hole custom DNS overrides
grafana.madhan.app  → 192.168.1.220
```

With Pi-hole, LAN clients resolve directly to `192.168.1.220` and bypass Hetzner entirely.
Without it, all public services hairpin through the VPS regardless of where the client is.

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
dig grafana.madhan.app     # → 178.156.199.250  (public, explicit A record)
dig headlamp.madhan.app    # → 192.168.1.220   (LAN wildcard, private)

# Unauthenticated access redirects to SSO
curl -I https://grafana.madhan.app   # → 302 to auth.madhan.app

# K8s routing peer connected
kubectl get pods -n netbird           # → Running

# NetBird UI: Peers → bifrost-agent Connected, k8s-routing-peer Connected
# NetBird UI: Routes → 192.168.1.0/24 → Active
```
