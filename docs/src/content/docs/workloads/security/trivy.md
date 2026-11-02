---
title: Trivy
description: Trivy Operator scans running workloads and stores the results as CRDs.
sidebar:
  order: 20
---

Trivy Operator scans each new or changed workload for CVEs, misconfiguration and exposed secrets, and writes reports as Kubernetes resources. [Upstream docs](https://aquasecurity.github.io/trivy-operator/latest/).

| | |
|---|---|
| Namespace | `trivy` (privileged PSA, for node-collector) |
| Source | [`trivy.go`](src:workloads/security/trivy.go) |
| Exposure | None |
| Storage | — |
| Auth | — |

## Configuration

| Setting | Value | Why |
|---|---|---|
| `operator.scanJobsConcurrentLimit` | `3` | Caps parallel scan Jobs |
| `operator.scanJobsRetryDelay` / `trivyOperator.scanJobTimeout` | `30s` / `5m` | Retry pacing; large images get five minutes |
| `nodeCollector.volumes` | etcd, kubelet, scheduler, controller-manager, `/etc/kubernetes`, CNI only | Talos has no `/etc/systemd` or `/lib/systemd`; the chart's default mounts fail |
| `nodeCollector.tolerations` | control-plane `NoSchedule` | Control planes get node checks too |
| `serviceMonitor.enabled` | `true` | VMAgent scrapes vulnerability metrics |
| CRDs | Downloaded from the chart tarball at synth time | Kept in step with the operator; set `TRIVY_OPERATOR_SKIP_CRDS=true` to skip them when CRDs already exist |

## Operate

| Report | Kind |
|---|---|
| Image CVEs | `vulnerabilityreports` |
| Misconfiguration | `configauditreports` |
| Secrets in images | `exposedsecretreports` |
| SBOM | `sbomreports` |

```bash frame="terminal"
kubectl get vulnerabilityreports -A -o custom-columns=\
NS:.metadata.namespace,NAME:.metadata.name,CRIT:.report.summary.criticalCount,HIGH:.report.summary.highCount
```

## Troubleshooting

### No report for a new workload

The scan Job failed, usually an image pull from a private registry or a timeout on a large image:

```bash frame="terminal"
kubectl get jobs -n trivy
kubectl logs -n trivy job/<scan-job>
```

### Reports missing entirely

Check the CRDs exist: `kubectl get crd | grep aquasecurity`. A synth with `TRIVY_OPERATOR_SKIP_CRDS=true` leaves them out.
