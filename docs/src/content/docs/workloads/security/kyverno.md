---
title: Kyverno
description: Fail-closed admission webhook, installed with no policies yet.
sidebar:
  order: 10
---

Kyverno admits, rejects or mutates every resource create and update, with policies written as Kubernetes CRDs. It is the cluster's only preventive control. [Upstream docs](https://kyverno.io/docs/).

| | |
|---|---|
| Namespace | `kyverno` |
| Source | [`keyverno.go`](src:workloads/security/keyverno.go) |
| Exposure | None |
| Storage | — |
| Auth | — |

:::note[Nothing is enforced yet]
No `ClusterPolicy` is deployed. Until one is added, the webhook approves everything.
:::

## Configuration

| Controller | Replicas | Limits |
|---|---|---|
| Admission | 3 | `1000m` / `512Mi` |
| Background, cleanup, reports | 2 each | `500m` / `256Mi` |

The admission controller gets three replicas because the webhook is fail-closed: if it is down, every write in the cluster fails.

| Setting | Value | Why |
|---|---|---|
| `webhooksCleanup.enabled` | `true` | Uninstalling otherwise leaves a fail-closed webhook that blocks all writes |
| `policyExceptions.enabled` | `true` | One workload can opt out without weakening a policy |
| `imageVerification.enabled` | `false` | No image signing here |
| `crds.migration.enabled` | `false` | The post-upgrade hook's ServiceAccount is pruned before the Job runs under Argo CD ([kyverno#17741](https://github.com/kyverno/kyverno/issues/17741)) |
| `grafana.enabled` | `true` | Ships the dashboard ConfigMap for Grafana's sidecar |
| `metricsService.port` | `8000` | Scraped through a ServiceMonitor |

:::caution[CRD storage-version changes]
With `crds.migration` off, a chart bump that changes a CRD storage version needs the migration re-enabled for that sync, or run by hand.
:::

## Troubleshooting

### Every create or update times out

The admission controller is unreachable. Restart its pods:

```bash frame="terminal"
kubectl get pods -n kyverno
kubectl logs -n kyverno -l app.kubernetes.io/component=admission-controller --tail=50
kubectl rollout restart deployment -n kyverno kyverno-admission-controller
```

### Find policy violations

```bash frame="terminal"
kubectl get policyreport -A
```
