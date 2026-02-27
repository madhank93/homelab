+++
title = "Longhorn"
description = "Distributed block storage for Kubernetes on all 4 worker nodes."
weight = 10
+++

## Overview

| Property | Value |
|----------|-------|
| CDK8s file | `platform/cdk8s/cots/storage/longhorn.go` |
| Namespace | `longhorn-system` |
| HTTPRoute | None |
| UI | No (internal Longhorn UI exists but not exposed) |
| Nodes | All 4 workers (k8s-worker1–4) |

## Purpose

Longhorn provides distributed block storage (RWO PVCs) for the cluster. It replicates volumes across worker nodes for resilience and provides the `longhorn` StorageClass.

## Storage Capacity

| Node | Disk | Usable |
|------|------|--------|
| k8s-worker1–4 | 125 GiB | ~100 GiB each |

**Total raw capacity:** ~400 GiB across 4 nodes (with replication factor 3, effective capacity ~133 GiB).

## Overprovisioning

```yaml
storageOverProvisioningPercentage: 200
```

This allows 240 GiB scheduled per 120 GiB physical disk (200% = 2x overprovisioning). The default of 100% only allowed ~5 GiB headroom per node — insufficient for large AI PVCs (ComfyUI uses 100 Gi).

## Talos Configuration

| Setting | Value | Reason |
|---------|-------|--------|
| `createDefaultDiskLabeledNodes: false` | (n/a — creates on all nodes) | All workers provision Longhorn disk |
| `preUpgradeChecker.jobEnabled: false` | Disabled | Avoids GitOps upgrade check conflicts |

The `longhorn-system` namespace has `pod-security.kubernetes.io/enforce: privileged` for the Longhorn DaemonSet (requires host mounts).

## Node Labels

All workers carry `node.longhorn.io/create-default-disk: config` applied via the Talos worker machine patch in Pulumi. This tells Longhorn to provision its disk on every worker node automatically.

## Checking Volume Health

```bash
# List all Longhorn volumes
kubectl get volumes.longhorn.io -n longhorn-system

# Check volume state and robustness
kubectl get volumes.longhorn.io -n longhorn-system \
  -o custom-columns="NAME:.metadata.name,STATE:.status.state,ROBUSTNESS:.status.robustness"
```

Healthy volumes show `STATE=attached` and `ROBUSTNESS=healthy`.
