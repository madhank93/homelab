---
title: cert-manager
description: Bump cert-manager through Pulumi without spending the Let's Encrypt duplicate-certificate budget.
sidebar:
  order: 13
---

Phase 4. Risk: low, but every HTTPS route depends on it. Pin: `charts.cert-manager` in [`config/versions.yaml`](src:config/versions.yaml), read by [`cert_manager.go`](src:core/platform/cert_manager.go).

```bash frame="terminal"
just core platform preview
just core platform up
kubectl get certificate -A   # every one Ready=True
```

:::danger[Never add a second Certificate for madhan.app / *.madhan.app]
Let's Encrypt allows 5 certificates per exact identifier set per 168h. Cilium mirrors the one in `kube-system` into `cilium-secrets` itself — see [cert-manager](/platform/cert-manager/). A `429 rateLimited` error, or a Certificate whose `Revision:` is in the dozens, means duplicates are competing for that budget. Do not delete and recreate the Certificate: that spends more of it.
:::
