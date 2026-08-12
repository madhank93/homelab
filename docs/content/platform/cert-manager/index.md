+++
title = "cert-manager"
description = "One wildcard certificate for *.madhan.app, issued via Let's Encrypt DNS-01 and mirrored to Cilium for TLS termination."
weight = 55
+++

[cert-manager](https://cert-manager.io/) issues and renews the TLS certificate
that terminates HTTPS for every `*.madhan.app` service. A DNS-01 challenge
through Cloudflare means the cluster needs no public HTTP endpoint — only write
access to the DNS zone.

Managed by Pulumi (`core/platform/cert_manager.go`, stack `platform`):

```bash
just core platform up
```

## Chart

| Setting | Value |
|---|---|
| Chart | `cert-manager` (`https://charts.jetstack.io`) |
| Version | `v1.21.1` |
| Namespace | `cert-manager` |
| CRDs | Bundled (`installCRDs: true`) |

## ClusterIssuers

| Name | Type | Used for |
|---|---|---|
| `letsencrypt-prod` | ACME DNS-01 via Cloudflare | The wildcard certificate |
| `homelab-ca` | Self-signed | Internal / testing certificates |

## One certificate, mirrored

There is exactly **one** Certificate for these names:

| Setting | Value |
|---|---|
| Name | `wildcard-madhan-app` |
| Namespace | `kube-system` |
| Secret | `wildcard-madhan-app-tls` |
| DNS names | `madhan.app`, `*.madhan.app` |

Cilium's Envoy reads TLS material from its own namespace, `cilium-secrets`, so
it mirrors that secret there itself — `gatewayAPI.secretsNamespace.sync: true`
in `core/platform/cilium.go`. The copy appears as
`cilium-sync-secret-<hash>` and Cilium owns it.

```
cert-manager                Cilium                      Envoy
─────────────               ──────                      ─────
kube-system/                cilium-secrets/             HTTPS listener
wildcard-madhan-app-tls ──> cilium-sync-secret-<hash> ─> homelab-gateway
```

> **Never add a second Certificate for the same DNS names.** Let's Encrypt caps
> issuance at 5 per *exact set of identifiers* per 168h. A second Certificate
> writing into `cilium-secrets` shares that budget, and because the namespace
> belongs to Cilium the secret gets removed and reissued repeatedly until the
> limit is hit. When that happens there is no certificate at all and **every**
> `*.madhan.app` HTTPS endpoint resets, while plain HTTP through the same
> Gateway keeps returning 200. Let Cilium do the copying.

## Cloudflare API token

The DNS-01 solver needs a token scoped to `madhan.app`:

| Permission | Purpose |
|---|---|
| Zone → Zone → Read | Resolve the domain to a Zone ID |
| Zone → DNS → Edit | Create and delete `_acme-challenge` TXT records |

Stored in the `cert-manager/cloudflare-api-token` Secret (key
`CLOUDFLARE_API_TOKEN`), created by `just create-secrets` from SOPS. It carries
`argocd.argoproj.io/sync-options: Prune=false` so Argo CD never deletes it.

cert-manager cannot issue anything until that Secret exists:

```bash
just create-secrets    # from SOPS
just core platform up  # cert-manager + ClusterIssuer
```

## Troubleshooting

```bash
# Is the certificate healthy? Expect Ready=True.
kubectl get certificate -A

# Why not? Look at the Issuing condition, not just Ready.
kubectl describe certificate wildcard-madhan-app -n kube-system

# Did Cilium mirror it?
kubectl get secret -n cilium-secrets

# Cloudflare API errors
kubectl logs -n cert-manager deployment/cert-manager | grep -i cloudflare
```

**`429 ... too many certificates already issued for this exact set of
identifiers`** means the weekly budget is gone; the message includes the retry
time. Do not delete and recreate the Certificate — that consumes more budget.
Find what is issuing duplicates first.

**A high `Revision:` on a Certificate** (dozens) means something keeps deleting
its Secret and cert-manager keeps reissuing. That is the shape of the failure
above.
