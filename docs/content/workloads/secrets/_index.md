+++
title = "Secrets Management"
description = "OpenBao secrets store and Secrets Store CSI Driver for zero-secret GitOps."
weight = 40
sort_by = "weight"
+++

Runtime secrets are stored in OpenBao (a Vault-compatible fork) and mounted into pods via the Secrets Store CSI Driver. CDK8s generates zero `Secret` manifests — all runtime secrets are fetched in-cluster at pod startup.

| Page | What it covers |
|---|---|
| [OpenBao](@/workloads/secrets/openbao/index.md) | The KV store, unseal sidecar, and day-to-day commands |
| [Secrets Store CSI Driver](@/workloads/secrets/csi-driver/index.md) | How the values reach a pod |

The two consumption patterns are explained in [Secrets](@/platform/secrets/index.md).
