+++
title = "Cloudflare"
description = "Cloudflare DNS records and the publicServices toggle — how services get exposed to the internet."
weight = 40
+++

## What is Cloudflare?

[Cloudflare](https://www.cloudflare.com/) is a DNS provider (and optional CDN/proxy) that manages the `madhan.app` domain for this homelab. All DNS records are provisioned as code via the Cloudflare Terraform/Pulumi provider, keeping DNS configuration in sync with the cluster.

## Why Cloudflare?

Cloudflare's API is required by cert-manager's DNS-01 ACME solver to issue wildcard TLS certificates without a public HTTP endpoint. It also provides a clean separation between LAN-only services (pointing to the private wildcard `192.168.1.220`) and internet-exposed services (pointing to the Bifrost VPS), controlled by a single `publicServices` list in code.

## How It's Used Here

The `madhan.app` domain uses a wildcard A record (`*.madhan.app → 192.168.1.220`) as the default for all LAN services. Public services override the wildcard with an explicit A record pointing to the Bifrost VPS (`178.156.199.250`). Everything is managed by Pulumi (`core/cloud/cloudflare.go`, `just core cloudflare up`).

## Record Layout

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
