# Talos In-Place Upgrade Guide - System Extensions

**Date**: February 8, 2026  
**Schematic ID**: `144f58860e456dda4f18038a2c7ebc91a4360f9a2b80458f03a6852f1ae12743`  
**Talos Version**: v1.9.3

## Overview

This guide provides step-by-step instructions for upgrading your existing Talos cluster to include system extensions (iSCSI tools, GPU support) **without destroying any data**.

## What's Being Added

### Schematic Configuration

The following schematic defines all system extensions:

```yaml
customization:
  systemExtensions:
    officialExtensions:
      - siderolabs/iscsi-tools
      - siderolabs/util-linux-tools
      - siderolabs/nvidia-container-toolkit
      - siderolabs/nonfree-kmod-nvidia
      - siderolabs/qemu-guest-agent
```

**Schematic ID**: `144f58860e456dda4f18038a2c7ebc91a4360f9a2b80458f03a6852f1ae12743`

To regenerate this schematic ID:
```bash
curl -X POST --data-binary @- https://factory.talos.dev/schematics <<EOF
customization:
  systemExtensions:
    officialExtensions:
      - siderolabs/iscsi-tools
      - siderolabs/util-linux-tools
      - siderolabs/nvidia-container-toolkit
      - siderolabs/nonfree-kmod-nvidia
      - siderolabs/qemu-guest-agent
EOF
```

### System Extensions
- `siderolabs/iscsi-tools` - Required for Longhorn storage
- `siderolabs/util-linux-tools` - Utilities for Longhorn
- `siderolabs/nvidia-container-toolkit` - NVIDIA GPU support
- `siderolabs/nonfree-kmod-nvidia` - NVIDIA kernel modules
- `siderolabs/qemu-guest-agent` - Proxmox integration

### Kernel Modules (Workers Only)
- `nbd` - Network Block Device
- `iscsi_tcp` - iSCSI over TCP
- `iscsi_generic` - Generic iSCSI
- `configfs` - Configuration filesystem

## Prerequisites

1. **Backup cluster state**:
   ```bash
   kubectl get all --all-namespaces -o yaml > cluster-backup-$(date +%Y%m%d).yaml
   kubectl get pv,pvc --all-namespaces -o yaml > storage-backup-$(date +%Y%m%d).yaml
   ```

2. **Verify cluster health**:
   ```bash
   kubectl get nodes
   kubectl get pods --all-namespaces | grep -v Running
   ```

3. **Have talosconfig ready**:
   ```bash
   export TALOSCONFIG=/Volumes/work/git-repos/homelab/infra/pulumi/talosconfig
   ```

## Upgrade Process

### Phase 1: Upgrade Worker Nodes (One at a Time)

**Why workers first?** Safer - control plane stays stable while workers upgrade.

#### Worker 1 (192.168.1.104)

```bash
# Set variables
export NEW_IMAGE="factory.talos.dev/installer/144f58860e456dda4f18038a2c7ebc91a4360f9a2b80458f03a6852f1ae12743:v1.9.3"

# Upgrade worker1
talosctl -n 192.168.1.104 upgrade --image="${NEW_IMAGE}" --preserve

# Monitor upgrade progress
talosctl -n 192.168.1.104 dmesg -f

# Wait for node to be Ready (usually 2-3 minutes)
kubectl wait --for=condition=Ready node/k8s-worker1 --timeout=5m

# Verify extensions installed
talosctl -n 192.168.1.104 get extensions

# Verify iscsid service running
talosctl -n 192.168.1.104 service iscsid status
```

**Expected output for extensions**:
```
NAMESPACE   TYPE              ID                               VERSION
runtime     ExtensionStatus   000.ghcr.io-siderolabs-iscsi-tools-...
runtime     ExtensionStatus   001.ghcr.io-siderolabs-util-linux-tools-...
runtime     ExtensionStatus   002.ghcr.io-siderolabs-nvidia-container-toolkit-...
runtime     ExtensionStatus   003.ghcr.io-siderolabs-nonfree-kmod-nvidia-...
runtime     ExtensionStatus   004.ghcr.io-siderolabs-qemu-guest-agent-...
```

#### Worker 2 (192.168.1.105)

```bash
# Upgrade worker2
talosctl -n 192.168.1.105 upgrade --image="${NEW_IMAGE}" --preserve

# Wait for Ready
kubectl wait --for=condition=Ready node/k8s-worker2 --timeout=5m

# Verify
talosctl -n 192.168.1.105 get extensions
talosctl -n 192.168.1.105 service iscsid status
```

#### Worker 3 (192.168.1.106)

```bash
# Upgrade worker3
talosctl -n 192.168.1.106 upgrade --image="${NEW_IMAGE}" --preserve

# Wait for Ready
kubectl wait --for=condition=Ready node/k8s-worker3 --timeout=5m

# Verify
talosctl -n 192.168.1.106 get extensions
talosctl -n 192.168.1.106 service iscsid status
```

#### Worker 4 (192.168.1.107 - GPU Node)

```bash
# Upgrade worker4 (has GPU passthrough)
talosctl -n 192.168.1.107 upgrade --image="${NEW_IMAGE}" --preserve

# Wait for Ready
kubectl wait --for=condition=Ready node/k8s-worker4 --timeout=5m

# Verify extensions (should include NVIDIA)
talosctl -n 192.168.1.107 get extensions

# Verify GPU modules loaded
talosctl -n 192.168.1.107 read /proc/modules | grep nvidia
```

### Phase 2: Upgrade Control Plane Nodes (One at a Time)

**Important**: Upgrade control plane nodes **after** all workers are upgraded and healthy.

#### Controller 1 (192.168.1.101)

```bash
# Upgrade controller1
talosctl -n 192.168.1.101 upgrade --image="${NEW_IMAGE}" --preserve

# Wait for Ready
kubectl wait --for=condition=Ready node/k8s-controller1 --timeout=5m

# Verify etcd health
talosctl -n 192.168.1.101 service etcd status
```

#### Controller 2 (192.168.1.102)

```bash
# Upgrade controller2
talosctl -n 192.168.1.102 upgrade --image="${NEW_IMAGE}" --preserve

# Wait for Ready
kubectl wait --for=condition=Ready node/k8s-controller2 --timeout=5m

# Verify etcd health
talosctl -n 192.168.1.102 service etcd status
```

#### Controller 3 (192.168.1.103)

```bash
# Upgrade controller3
talosctl -n 192.168.1.103 upgrade --image="${NEW_IMAGE}" --preserve

# Wait for Ready
kubectl wait --for=condition=Ready node/k8s-controller3 --timeout=5m

# Verify etcd health
talosctl -n 192.168.1.103 service etcd status
```

## Post-Upgrade Verification

### 1. Verify All Nodes

```bash
# Check all nodes are Ready
kubectl get nodes

# Expected output: All nodes STATUS=Ready
```

### 2. Verify Extensions on All Workers

```bash
for ip in 192.168.1.104 192.168.1.105 192.168.1.106 192.168.1.107; do
  echo "=== Node $ip ==="
  talosctl -n $ip get extensions | grep -E "iscsi|util-linux|nvidia|qemu"
done
```

### 3. Verify iSCSI Service

```bash
for ip in 192.168.1.104 192.168.1.105 192.168.1.106 192.168.1.107; do
  echo "=== Node $ip ==="
  talosctl -n $ip service iscsid status
done
```

**Expected**: `STATE: Running`

### 4. Verify Kernel Modules

```bash
for ip in 192.168.1.104 192.168.1.105 192.168.1.106 192.168.1.107; do
  echo "=== Node $ip ==="
  talosctl -n $ip read /proc/modules | grep -E "nbd|iscsi_tcp|iscsi_generic|configfs"
done
```

### 5. Deploy Longhorn

Now that iSCSI tools are installed, Longhorn should deploy successfully:

```bash
# Sync Longhorn in ArgoCD
kubectl patch application longhorn -n argocd --type merge -p '{"operation":{"initiatedBy":{"username":"admin"},"sync":{}}}'

# Watch Longhorn pods
kubectl get pods -n longhorn-system -w
```

**Expected**:
- `longhorn-manager` DaemonSet: 4/4 (one per worker)
- `longhorn-driver-deployer`: 1/1
- `longhorn-ui`: 2/2

### 6. Verify StorageClass

```bash
kubectl get storageclass

# Expected: longhorn (default)
```

### 7. Test Longhorn with Infisical

```bash
# Check Infisical PVCs
kubectl get pvc -n infisical

# Expected: Both PVCs should be Bound
# - data-postgresql-0: Bound
# - redis-data-redis-master-0: Bound

# Check Infisical pods
kubectl get pods -n infisical

# Expected: All Running
```

## Troubleshooting

### Node Stuck During Upgrade

```bash
# Check upgrade status
talosctl -n <node-ip> dmesg -f

# If stuck, check service status
talosctl -n <node-ip> services

# Force reboot if necessary (last resort)
talosctl -n <node-ip> reboot
```

### Extensions Not Showing

```bash
# Verify image was applied
talosctl -n <node-ip> version

# Should show the new schematic in the image URL

# If wrong image, re-run upgrade
talosctl -n <node-ip> upgrade --image="${NEW_IMAGE}" --preserve
```

### iscsid Service Not Running

```bash
# Check service logs
talosctl -n <node-ip> logs iscsid

# Restart service
talosctl -n <node-ip> service iscsid restart
```

### Longhorn Manager Pods Still Failing

```bash
# Check pod logs
kubectl logs -n longhorn-system -l app=longhorn-manager --tail=50

# Verify iscsiadm is accessible
talosctl -n <worker-ip> read /usr/bin/nsenter
```

## Rollback Plan

If you need to rollback to the old image:

```bash
# Old schematic ID
OLD_IMAGE="factory.talos.dev/installer/ce4c980550dd2ab1b17bbf2b08801c7eb59418eafe8f279833297925d67c7515:v1.9.3"

# Rollback a node
talosctl -n <node-ip> upgrade --image="${OLD_IMAGE}" --preserve
```

**Note**: Rollback will remove the extensions, so Longhorn will stop working.

## Timeline

**Total time**: ~30-40 minutes for all 7 nodes

- Worker 1: ~3 minutes
- Worker 2: ~3 minutes
- Worker 3: ~3 minutes
- Worker 4: ~3 minutes
- Controller 1: ~3 minutes
- Controller 2: ~3 minutes
- Controller 3: ~3 minutes
- Verification: ~10 minutes

## Success Criteria

- ✅ All 7 nodes showing `Ready`
- ✅ All workers have 5 extensions installed
- ✅ `iscsid` service running on all workers
- ✅ Kernel modules loaded on all workers
- ✅ Longhorn manager DaemonSet: 4/4
- ✅ Longhorn StorageClass created
- ✅ Infisical PVCs bound
- ✅ Infisical pods running

## Next Steps After Upgrade

1. **Verify Infisical UI**: Access `https://infisical.madhan.app`
2. **Complete Infisical setup**: Create admin account
3. **Update release notes**: Document the upgrade
4. **Monitor cluster**: Watch for any issues over next 24 hours

## References

- [Talos Upgrades Documentation](https://www.talos.dev/v1.9/talos-guides/upgrading-talos/)
- [Talos Image Factory](https://factory.talos.dev/)
- [Longhorn on Talos](https://www.talos.dev/v1.9/kubernetes-guides/configuration/storage/#longhorn)

---

**Created**: February 8, 2026  
**Last Updated**: February 8, 2026  
**Schematic**: `144f58860e456dda4f18038a2c7ebc91a4360f9a2b80458f03a6852f1ae12743`
