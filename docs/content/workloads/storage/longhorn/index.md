+++
title = "Longhorn"
description = "Distributed block storage for Kubernetes on all 4 worker nodes."
weight = 10
+++

## What is Longhorn?

[Longhorn](https://longhorn.io/) is a lightweight, cloud-native distributed block storage system for Kubernetes. It provides persistent volumes that are automatically replicated across multiple nodes, with a built-in UI for volume management, snapshots, and backups.

## Why Longhorn?

| Option | Pros | Cons |
|--------|------|------|
| **Longhorn** | Lightweight, no dedicated disks needed, RWX via NFS share-manager | Single vendor (Rancher/SUSE) |
| Rook-Ceph | Production-grade, feature-rich | Requires dedicated disks, complex setup, heavy resource usage |
| NFS | Simple | No HA, single point of failure |
| Local-path | Zero overhead | No replication, node-local only |

For a 4-worker homelab without dedicated storage disks, Longhorn's ability to use a portion of each worker's OS disk makes it the right choice. Rook-Ceph would require separate disk devices on each node.

## How It's Used Here

Longhorn provides the `longhorn` StorageClass used by virtually every stateful workload:

- Harbor (registry + jobservice — RWX)
- Grafana (dashboards — RWX)
- VictoriaMetrics (metrics storage — RWO)
- VictoriaLogs (log storage — RWO)
- Ollama (model storage — RWO, 100 Gi)
- ComfyUI (models + outputs — RWX, 50 Gi)
- n8n (workflow data — RWX)
- OpenBao (secrets data — RWO)
- NetBird (config — RWO, 100 Mi)

**RWX volumes** use Longhorn's built-in NFS share-manager — Longhorn automatically provisions an NFS server pod for each RWX volume. This eliminates rolling update deadlocks that occur with RWO volumes when a new pod starts before the old pod releases the volume.

Source: [`workloads/storage/longhorn.go`](https://github.com/madhank93/homelab/blob/v0.1.5/workloads/storage/longhorn.go)

## Configuration

| Setting | Value | Why |
|---------|-------|-----|
| `defaultReplicaCount` | `3` | Replicate every volume across 3 nodes for resilience |
| `storageOverProvisioningPercentage` | `300` | Allow 3x scheduling vs physical capacity |
| `createDefaultDiskLabeledNodes` | `false` | Create disk on ALL nodes, not just labelled ones |
| `preUpgradeChecker.jobEnabled` | `false` | Disable pre-upgrade hook for GitOps compatibility |
| Namespace PSA | `privileged` | Longhorn CSI driver requires host mounts |

**Why 300% overprovisioning?** Workers have ~90–95 GiB actually free. At 200% (the previous setting), the 240 GiB scheduling cap was exhausted — there was no room for new 50 GiB AI model replicas. At 300%, each node schedules up to ~285 GiB, providing ~150 GiB of headroom for AI volumes and future growth. Longhorn does not actually allocate all scheduled space immediately; volumes grow on demand.

## Storage Capacity

| Node | Disk | Usable for Longhorn |
|------|------|---------------------|
| k8s-worker1 | 200 GiB | ~90–95 GiB |
| k8s-worker2 | 200 GiB | ~90–95 GiB |
| k8s-worker3 | 200 GiB | ~90–95 GiB |
| k8s-worker4 (GPU) | 250 GiB | ~90–95 GiB |

With replication factor 3 and 4 nodes, each volume's data is replicated to 3 of the 4 workers.

## How It Connects

```
Pod (any namespace)
  → PVC with storageClassName: longhorn
  → Longhorn CSI Driver (DaemonSet on each node)
  → Longhorn Manager (coordinates replicas)
  → Volume replicated across 3 workers
  → (RWX) Longhorn NFS share-manager pod bridges multiple pod attachments
```

## Screenshots

![Longhorn volume list showing attached volumes and replica health](/assets/screenshots/longhorn/volume-list.png)

## Troubleshooting

### Volume Stuck Attaching

**Symptoms:** PVC stays in `Pending` or pod stuck in `ContainerCreating`.

**Diagnosis:**

```bash
kubectl get volumes.longhorn.io -n longhorn-system
kubectl describe pvc <pvc-name> -n <namespace>
kubectl get pods -n longhorn-system | grep share-manager  # for RWX
```

**Fix:**

```bash
# Force detach the volume in Longhorn UI, then delete the stuck pod
kubectl delete pod -n <namespace> <pod-name> --grace-period=0 --force
```

### Degraded Replicas

**Symptoms:** `kubectl get volumes.longhorn.io` shows `ROBUSTNESS=degraded`.

**Diagnosis:**

```bash
kubectl get volumes.longhorn.io -n longhorn-system \
  -o custom-columns="NAME:.metadata.name,STATE:.status.state,ROBUSTNESS:.status.robustness"
```

**Fix:** Check if a worker node is offline. If a node is down, Longhorn will rebuild the degraded replica on another available node automatically once it can schedule. This may take several minutes.

### RWO Multi-Attach Deadlock

**Symptoms:** Pod stuck with `Multi-Attach error for volume — volume is already used by pod`.

**Why this happens:** RWO (ReadWriteOnce) volumes can only be attached to one node at a time. During a rolling update, the new pod may start on a different node before the old pod fully terminates and releases the volume.

**Fix for Harbor (and similar):**

```bash
# 1. Find the old ReplicaSet
kubectl get replicasets -n harbor

# 2. Scale down the old RS
kubectl scale replicaset <old-rs-name> -n harbor --replicas=0

# 3. Force delete any stuck pod
kubectl delete pod -n harbor <stuck-pod> --grace-period=0 --force
```

**Long-term fix:** Switch the PVC to `ReadWriteMany` — Harbor's registry and jobservice PVCs are already set to RWX in this homelab for exactly this reason.

### Checking Volume Health

```bash
# List all Longhorn volumes
kubectl get volumes.longhorn.io -n longhorn-system

# Check volume state and robustness
kubectl get volumes.longhorn.io -n longhorn-system \
  -o custom-columns="NAME:.metadata.name,STATE:.status.state,ROBUSTNESS:.status.robustness"

# Healthy volumes show:
# STATE=attached   ROBUSTNESS=healthy
```
