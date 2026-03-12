+++
title = "DNS & TLS"
description = "Cloudflare DNS records, the publicServices toggle, cert-manager wildcard TLS, and ClusterIssuers."
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

## cert-manager

cert-manager issues a wildcard TLS certificate for `*.madhan.app` using Let's Encrypt DNS-01 challenge via the Cloudflare API. Managed by Pulumi (`core/platform/cert_manager.go`, stack: `platform`):

```bash
just core platform up
```

### Chart

| Setting | Value |
|---------|-------|
| Chart | `cert-manager` |
| Repo | `https://charts.jetstack.io` |
| Version | `v1.19.3` |
| Namespace | `cert-manager` |
| CRDs | Bundled (`installCRDs: true`) |

### ClusterIssuers

| Name | Type | Used For |
|------|------|---------|
| `letsencrypt-prod` | ACME DNS-01 via Cloudflare | Wildcard `*.madhan.app` certificate |
| `homelab-ca` | Self-signed | Internal / testing certificates |

### Wildcard Certificate

| Setting | Value |
|---------|-------|
| Name | `wildcard-madhan-app` |
| Namespace | `kube-system` |
| Secret | `wildcard-madhan-app-tls` |
| DNS names | `madhan.app`, `*.madhan.app` |
| Issuer | `letsencrypt-prod` |

The certificate is stored in `kube-system` so the `homelab-gateway` HTTPS listener can reference it across namespaces.

### Cloudflare API Token

The DNS-01 solver requires a Cloudflare API token scoped to `madhan.app`:

| Permission | Purpose |
|------------|---------|
| Zone → Zone → Read | Resolve domain to Cloudflare Zone ID |
| Zone → DNS → Edit | Create/delete `_acme-challenge` TXT records |

The token is stored in `cert-manager/cloudflare-api-token` Secret (key: `CLOUDFLARE_API_TOKEN`). Created by `just create-secrets` from SOPS and carries `argocd.argoproj.io/sync-options: Prune=false` so ArgoCD never deletes it.

### Bootstrap Dependency

On a fresh cluster, cert-manager needs the `cloudflare-api-token` Secret before it can issue certificates:

```bash
# 1. Create bootstrap secrets (reads from SOPS)
just create-secrets

# 2. Then deploy cert-manager + ClusterIssuer
just core platform up
```

### Current Status

> The HTTPS Gateway listener is currently **disabled** pending `wildcard-madhan-app-tls` creation. All app URLs use HTTP (`http://app.madhan.app`).
>
> Once the certificate exists in `kube-system`, re-enable the HTTPS listener in `core/platform/cilium.go` and run `just core platform up`.

### Verifying Certificate Issuance

```bash
# Watch Certificate status
kubectl describe certificate wildcard-madhan-app -n kube-system

# Watch CertificateRequest and Order objects
kubectl get certificaterequests,orders -n kube-system

# Check cert-manager logs for Cloudflare API errors
kubectl logs -n cert-manager deployment/cert-manager | grep -i cloudflare
```

A successful issuance shows:
```
Status: True  Type: Ready
Message: Certificate is up to date and has not expired
```
