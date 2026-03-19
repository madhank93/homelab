+++
title = "cert-manager"
description = "Wildcard TLS certificate for *.madhan.app via Let's Encrypt DNS-01 challenge and Cloudflare."
weight = 55
+++

cert-manager issues a wildcard TLS certificate for `*.madhan.app` using Let's Encrypt DNS-01 challenge via the Cloudflare API. Managed by Pulumi (`core/platform/cert_manager.go`, stack: `platform`):

```bash
just core platform up
```

## Chart

| Setting | Value |
|---------|-------|
| Chart | `cert-manager` |
| Repo | `https://charts.jetstack.io` |
| Version | `v1.19.3` |
| Namespace | `cert-manager` |
| CRDs | Bundled (`installCRDs: true`) |

## ClusterIssuers

| Name | Type | Used For |
|------|------|---------|
| `letsencrypt-prod` | ACME DNS-01 via Cloudflare | Wildcard `*.madhan.app` certificate |
| `homelab-ca` | Self-signed | Internal / testing certificates |

## Wildcard Certificate

| Setting | Value |
|---------|-------|
| Name | `wildcard-madhan-app` |
| Namespace | `kube-system` |
| Secret | `wildcard-madhan-app-tls` |
| DNS names | `madhan.app`, `*.madhan.app` |
| Issuer | `letsencrypt-prod` |

The certificate lives in `kube-system` so the `homelab-gateway` HTTPS listener can reference it across namespaces.

## Cloudflare API Token

The DNS-01 solver requires a Cloudflare API token scoped to `madhan.app`:

| Permission | Purpose |
|------------|---------|
| Zone → Zone → Read | Resolve domain to Cloudflare Zone ID |
| Zone → DNS → Edit | Create/delete `_acme-challenge` TXT records |

The token is stored in the `cert-manager/cloudflare-api-token` Secret (key: `CLOUDFLARE_API_TOKEN`), created by `just create-secrets` from SOPS. It carries `argocd.argoproj.io/sync-options: Prune=false` so ArgoCD never deletes it.

## Bootstrap Dependency

cert-manager needs the `cloudflare-api-token` Secret before it can issue certificates:

```bash
just create-secrets   # creates the Secret from SOPS
just core platform up # deploys cert-manager + ClusterIssuer
```

## Current Status

> The HTTPS Gateway listener is currently **disabled** pending `wildcard-madhan-app-tls` creation. All app URLs use HTTP.
>
> Once the certificate exists in `kube-system`, re-enable the HTTPS listener in `core/platform/cilium.go` and run `just core platform up`.

## Troubleshooting

```bash
# Watch Certificate status
kubectl describe certificate wildcard-madhan-app -n kube-system

# Watch CertificateRequest and Order objects
kubectl get certificaterequests,orders -n kube-system

# Check cert-manager logs for Cloudflare API errors
kubectl logs -n cert-manager deployment/cert-manager | grep -i cloudflare
```

A successful issuance shows: `Status: True  Type: Ready  Message: Certificate is up to date and has not expired`
