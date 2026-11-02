---
title: OpenBao and CSI Driver
description: Bump OpenBao and the Secrets Store CSI Driver, then prove the unseal sidecar and secret mounts still work.
sidebar:
  order: 14
---

Phase 5. Risk: medium, every app reads its secrets through this layer. Pins in [`config/versions.yaml`](src:config/versions.yaml): `charts.openbao`, `images.openbao` (server, CSI agent and unseal sidecar), `charts.secrets-store-csi-driver`.

```yaml title="config/versions.yaml"
charts:
  openbao: {repo: https://openbao.github.io/openbao-helm, version: 0.X.Y}
  secrets-store-csi-driver: {repo: https://kubernetes-sigs.github.io/secrets-store-csi-driver/charts, version: 1.X.Y}
images:
  openbao: {image: openbao/openbao, tag: 2.X.Y}
```

The image is pinned, not the chart appVersion: OpenBao 2.7 removes the `file` storage backend this cluster uses, so the image stays on 2.6.x while the chart moves. Check chart/image compatibility against the [OpenBao tags](https://github.com/openbao/openbao/tags).

```bash frame="terminal"
just synth && git push
kubectl logs -n openbao openbao-0 -c unseal
kubectl exec -n openbao openbao-0 -c openbao -- bao status        # Sealed: false
kubectl exec -n harbor deploy/secret-sync -- ls /mnt/secrets/      # CSI mount still populated
```

:::caution[OnDelete StatefulSet]
OpenBao does not restart on sync. After Argo CD applies the new version, run `kubectl delete pod -n openbao openbao-0` and wait for the sidecar to unseal it.
:::
