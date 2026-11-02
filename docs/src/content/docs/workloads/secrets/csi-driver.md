---
title: Secrets Store CSI driver
description: Mounts OpenBao values into pods as files and mirrors them into Secrets when asked.
sidebar:
  order: 20
---

The driver mounts values from OpenBao into pods and, for apps that need an `existingSecret`, mirrors them into a Kubernetes Secret. Which apps use which pattern: [Secrets](/concepts/secrets/). [Upstream docs](https://secrets-store-csi-driver.sigs.k8s.io/).

| | |
|---|---|
| Namespace | `kube-system` (DaemonSet on every node) |
| Source | [`csi_driver.go`](src:workloads/secrets/csi_driver.go), [`kit.go`](src:workloads/kit/kit.go) |
| Exposure | None |
| Storage | — |
| Auth | OpenBao Kubernetes auth with the pod's ServiceAccount (via the OpenBao CSI provider) |

## Configuration

| Setting | Value | Why |
|---|---|---|
| `syncSecret.enabled` | `true` | Mirrors `secretObjects` into Secrets |
| `enableSecretRotation` / `rotationPollInterval` | `true` / `2m` | Picks up changed values within two minutes |
| `linux.tolerations` | `operator: Exists` | Any node, control planes included, can serve mounts |

Charts never hand-write a `SecretProviderClass`. They call the helpers in [`kit.go`](src:workloads/kit/kit.go):

| Helper | Does |
|---|---|
| `kit.SecretProviderClass` | Reads `secret/data/<role>` with role `<role>`, mirrors listed keys into a Secret |
| `kit.SecretsVolume` / `kit.SecretsMount` | CSI volume `openbao-secrets` at `/mnt/secrets` |
| `kit.SecretSync` | `secret-sync` `pause` Deployment for charts that cannot add volumes (Harbor, Falco UI) |

:::caution[No mount, no Secret]
The mirror only happens while some pod mounts the CSI volume. A correct `SecretProviderClass` with no mounting pod produces nothing.
:::

## Troubleshooting

### `MountVolume.SetUp failed … failed to get secretproviderclass`

The `SecretProviderClass` is missing from the pod's namespace; it must live alongside the pod. Check Argo CD sync for that app and `kubectl get secretproviderclass -n <namespace>`.

### App still uses the old value

Rotation updates the file and the mirrored Secret, not the running process. Pods restart only with the [Reloader](/workloads/support/reloader/) annotation. Check `kubectl get secretproviderclasspodstatus -n <namespace>`.

### Permission denied

The OpenBao role does not match the pod's ServiceAccount; see [OpenBao](/workloads/secrets/openbao/#403--permission-denied-on-a-csi-mount).
