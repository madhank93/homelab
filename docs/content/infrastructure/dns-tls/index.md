+++
title = "DNS & TLS"
description = "Cloudflare DNS, cert-manager, and wildcard TLS certificate for *.madhan.app."
weight = 40
+++

## DNS

All cluster services use the `madhan.app` domain managed by Cloudflare DNS. Each app has a DNS record pointing to the Cilium Gateway's LoadBalancer IP (`192.168.1.220`).

## TLS Certificate

A wildcard certificate `*.madhan.app` is provisioned by **cert-manager** using Let's Encrypt DNS-01 challenge via the Cloudflare API.

**Target Secret:** `wildcard-madhan-app-tls` in `kube-system`
**Used by:** the `homelab-gateway` HTTPS listener

## Bootstrap Dependency

The TLS setup has a circular dependency that must be broken manually on a fresh cluster:

```
cert-manager needs cloudflare-api-token Secret
  → cloudflare-api-token is created by: just create-secrets
    → just create-secrets reads from: secrets/bootstrap.sops.yaml (SOPS-encrypted)
```

Run `just create-secrets` before `just pulumi platform up` on a fresh cluster.

## Cloudflare API Token

The token requires two permissions scoped to `madhan.app` only:

| Permission | Purpose |
|------------|---------|
| Zone → Zone → Read | Resolve domain to Cloudflare Zone ID |
| Zone → DNS → Edit | Create/delete `_acme-challenge` TXT records |

The token is stored in the `cert-manager/cloudflare-api-token` Secret (key: `CLOUDFLARE_API_TOKEN`). This Secret is created by the bootstrap script and carries `argocd.argoproj.io/sync-options: Prune=false` so ArgoCD never deletes it.

## Verifying Certificate Issuance

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

## Current Status

> The HTTPS Gateway listener is currently **disabled** pending `wildcard-madhan-app-tls` creation. All app URLs use HTTP (`http://app.madhan.app`).
>
> Once the certificate exists in `kube-system`, re-enable the HTTPS listener in `infra/pulumi/cilium.go` and run `just pulumi platform up`.
