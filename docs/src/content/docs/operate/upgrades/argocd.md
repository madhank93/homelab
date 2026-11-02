---
title: Argo CD
description: Bump the Argo CD chart and the image pinned ahead of it, applied by Pulumi.
sidebar:
  order: 12
---

Phase 3. Risk: medium, GitOps pauses while the server restarts. Pins: `charts.argo-cd` and `images.argocd` in [`config/versions.yaml`](src:config/versions.yaml), read by [`argocd.go`](src:core/platform/argocd.go).

```yaml title="config/versions.yaml"
charts:
  argo-cd: {repo: https://argoproj.github.io/argo-helm, version: X.Y.Z}
images:
  argocd: {image: quay.io/argoproj/argocd, tag: vX.Y.Z}
```

**Image ahead of chart.** The Release overrides `image.tag` to run a server newer than the chart pins. On every chart bump, check whether the chart's appVersion has caught up and drop the override once it has.

```bash frame="terminal"
just core platform preview
just core platform up
kubectl rollout status deployment argocd-server -n argocd
kubectl get applications -n argocd   # all Synced + Healthy
```

To roll out a large change one app at a time, set `argocd_manual_sync: true` in `config/config.yml`, apply, sync with `just argocd-sync <app>`, then set it back to `false` and apply again.
