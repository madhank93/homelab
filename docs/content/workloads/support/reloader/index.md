+++
title = "Reloader"
description = "Automatic pod restarting when ConfigMaps or Secrets change."
weight = 10
+++

## What is Reloader?

[Reloader](https://github.com/stakater/Reloader) (by Stakater) is a Kubernetes controller that watches ConfigMaps and Secrets for changes and triggers rolling restarts of Deployments, StatefulSets, and DaemonSets that reference them. This is the "reactive" side of GitOps secret rotation.

## Why Reloader?

When a ConfigMap or Secret changes in Kubernetes (e.g., when the CSI driver rotates a secret from OpenBao), the running pods are **not** automatically restarted — they keep using the stale values from their environment variables or volume mounts. Reloader solves this by watching for changes and triggering a rollout.

Without Reloader, secret rotation requires either:
- Manual `kubectl rollout restart deployment/<name>`
- Application code that watches the file for changes
- A custom operator

## How It's Used Here

Reloader is deployed from the Stakater Helm chart in the `reloader` namespace. Any Deployment, StatefulSet, or DaemonSet with the annotation `reloader.stakater.com/auto: "true"` is automatically restarted when any ConfigMap or Secret it references changes.

Source: [`workloads/support/reloader.go`](https://github.com/madhank93/homelab/blob/v0.1.5/workloads/support/reloader.go)

## Configuration

| Setting | Value | Why |
|---------|-------|-----|
| Namespace | `reloader` | Isolated namespace |
| Helm chart | `reloader` v2.2.8 | stakater.github.io/stakater-charts |

## Usage Pattern

Add the annotation to any workload that should restart on secret/configmap changes:

```yaml
metadata:
  annotations:
    reloader.stakater.com/auto: "true"
```

In CDK8s Go code (example from `grafana.go`):

```go
"podAnnotations": map[string]any{
    "reloader.stakater.com/auto": "true",
},
```

Apps that use this annotation in this homelab:

| App | Why |
|-----|-----|
| Grafana | Restart when OAuth secret rotates |
| VictoriaMetrics | Restart when config changes |
| VictoriaLogs | Restart when config changes |
| AlertManager | Restart when routing config changes |
| OpenBao | Restart when unseal key secret updates |
| Harbor secret-sync | Restart when OpenBao password rotates |
| Rancher secret-sync | Restart when bootstrap password rotates |
| NetBird | Restart when setup key changes |
| OTel collectors | Restart when pipeline config changes |

## How It Connects

```
OpenBao rotates a secret value
  → CSI driver detects change (rotationPollInterval: 2m)
  → CSI driver updates file in pod mount
  → CSI driver updates k8s Secret (Pattern B only)
  → Reloader detects Secret change
  → Reloader triggers rolling restart of annotated Deployment
  → New pod starts, mounts fresh secret from CSI
```

## Troubleshooting

### Deployment Not Restarting After Secret Change

**Diagnosis:**

```bash
# Check Reloader is running
kubectl get pods -n reloader

# Check the annotation is present on the Deployment/StatefulSet
kubectl get deployment <name> -n <namespace> -o yaml | grep reloader

# Check Reloader logs
kubectl logs -n reloader -l app=reloader --tail=50
```

**Fix:** Ensure the `reloader.stakater.com/auto: "true"` annotation is on the **pod template** spec (`spec.template.metadata.annotations`), not just the Deployment metadata. Helm chart values typically set this via `podAnnotations`.
