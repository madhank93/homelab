+++
title = "Cloudflare"
description = "Cloudflare DNS records and the publicServices toggle — how services get exposed to the internet."
weight = 40
+++

## Cloudflare DNS

All cluster services use the `madhan.app` domain managed by Cloudflare. DNS records are provisioned by Pulumi (`core/cloud/cloudflare.go`, stack: `cloudflare`):

```bash
just core cloudflare up
```

### Record Layout

| Record | Resolves To | Purpose |
|--------|-------------|---------|
| `*.madhan.app` | `192.168.1.220` | Wildcard — all services default to LAN gateway |
| `auth.madhan.app` | `178.156.199.250` | Authentik on Bifrost — always public |
| `netbird.madhan.app` | `178.156.199.250` | NetBird dashboard+server — always public |
| `proxy.madhan.app` | `178.156.199.250` | NetBird expose base — always public |
| `*.proxy.madhan.app` | `178.156.199.250` | NetBird expose wildcard |
| `grafana.madhan.app` | `178.156.199.250` | Public via Bifrost (self-managed auth) |

All records are **DNS-only** (Proxied: false — orange cloud off). Cloudflare resolves specific records before wildcards — a service with no explicit A record falls through to the `*.madhan.app` wildcard → private IP → unreachable from the internet.

### Making a Service Public

Edit the `publicServices` slice in `core/cloud/cloudflare.go`:

```go
var publicServices = []PublicService{
    // SkipAuth: true  → no Authentik ForwardAuth (service handles its own auth)
    // SkipAuth: false → Authentik ForwardAuth required before reaching the service
    {Name: "grafana", SkipAuth: true},
    // Add a new public service:
    {Name: "headlamp", SkipAuth: false},
}
```

Then run both stacks — one creates the DNS A record, the other generates the Traefik router:

```bash
just core cloudflare up && just core hetzner up
```

Removing a service from `publicServices` deletes the A record. DNS falls back to the LAN wildcard automatically — the service becomes LAN-only with no further changes.

See [Service Access](/infrastructure/access) for the complete traffic flow diagram.

---

TLS certificates are managed by cert-manager in the Platform layer. See [cert-manager](/platform/cert-manager/).
