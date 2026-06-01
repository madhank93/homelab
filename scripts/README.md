# Talos Upgrade Script

Automated script for upgrading Talos cluster nodes with system extensions.

## Location

`scripts/talos-upgrade.sh`

## Features

- ✅ **Intelligent Wait Logic** - Uses `kubectl wait` and `talosctl health` instead of hardcoded sleeps
- ✅ **Automatic Backup** - Creates cluster backup before upgrade
- ✅ **Pre-flight Checks** - Verifies tools and connectivity
- ✅ **Extension Verification** - Confirms all extensions installed
- ✅ **Error Handling** - Stops on failures, provides clear error messages
- ✅ **Colored Output** - Easy to read progress indicators
- ✅ **Safe Rollout** - Workers first, then control plane

## Usage

### Basic Usage

```bash
cd /Volumes/work/git-repos/homelab/infra/pulumi

# Run with bash (required)
bash ../scripts/talos-upgrade.sh
```

### With Custom Configs

```bash
export TALOSCONFIG=/path/to/talosconfig
export KUBECONFIG=/path/to/kubeconfig
bash ./scripts/talos-upgrade.sh
```

> **Note**: This script requires `bash`. Do not run with `sh` as it uses bash-specific features.

## What It Does

1. **Pre-flight Checks**
   - Verifies `kubectl` and `talosctl` are installed
   - Checks cluster connectivity
   - Validates config files exist

2. **Backup**
   - Creates timestamped backup in `./backups/`
   - Saves all resources, PVs/PVCs, and node state

3. **Upgrade Workers** (one at a time)
   - Initiates upgrade with `--preserve` flag
   - Waits for Talos health check
   - Waits for Kubernetes node Ready
   - Verifies extensions installed
   - Verifies iscsid service running

4. **Upgrade Controllers** (one at a time)
   - Same process as workers
   - Additionally verifies etcd health

5. **Final Verification**
   - Shows cluster node status
   - Provides next steps

## Wait Logic

The script uses **intelligent waiting** instead of hardcoded sleeps:

- `kubectl wait --for=condition=Ready` - Waits for node to be Ready (5min timeout)
- `talosctl health --wait-timeout=5s` - Polls Talos health (60 attempts = 5min)
- Extension verification with retries
- Service status checks

## Error Handling

- **Worker failure**: Aborts control plane upgrade for safety
- **Controller failure**: Reports error but completes remaining nodes
- **Timeout**: Clear error messages with troubleshooting hints
- **Pre-flight failure**: Exits before making any changes

## Output Example

```
[INFO] ==========================================
[INFO] Talos Cluster Upgrade - System Extensions
[INFO] ==========================================
[INFO] Schematic ID: 144f58860e456dda4f18038a2c7ebc91a4360f9a2b80458f03a6852f1ae12743
[INFO] Talos Version: v1.9.3

[INFO] Running pre-flight checks...
[SUCCESS] Pre-flight checks passed

[INFO] Creating backup in ./backups/talos-upgrade-20260208-194425...
[SUCCESS] Backup created

[INFO] ==========================================
[INFO] PHASE 1: Upgrading Worker Nodes
[INFO] ==========================================

[INFO] ==========================================
[INFO] Upgrading worker: k8s-worker1 (192.168.1.104)
[INFO] ==========================================
[INFO] Starting upgrade to image: factory.talos.dev/installer/...
[SUCCESS] Upgrade initiated for k8s-worker1
[INFO] Waiting for Talos health on 192.168.1.104...
[SUCCESS] Talos health check passed for 192.168.1.104
[INFO] Waiting for node k8s-worker1 to be Ready (timeout: 300s)...
[SUCCESS] Node k8s-worker1 is Ready
[INFO] Verifying extensions on 192.168.1.104...
[SUCCESS]   ✓ iscsi-tools
[SUCCESS]   ✓ util-linux-tools
[SUCCESS]   ✓ nvidia-container-toolkit
[SUCCESS]   ✓ nonfree-kmod-nvidia
[SUCCESS]   ✓ qemu-guest-agent
[SUCCESS] All extensions verified on 192.168.1.104
[SUCCESS] Successfully upgraded k8s-worker1
```

## Troubleshooting

### Script fails to find configs

```bash
# Run from infra/pulumi directory
cd /Volumes/work/git-repos/homelab/infra/pulumi
../scripts/talos-upgrade.sh

# Or set environment variables
export TALOSCONFIG=/path/to/talosconfig
export KUBECONFIG=/path/to/kubeconfig
```

### Node fails to become Ready

The script will show the error and stop. Check:
```bash
kubectl get nodes
talosctl -n <node-ip> dmesg -f
```

### Extensions not verified

The script will warn but continue. Manually verify:
```bash
talosctl -n <node-ip> get extensions
```

## Rollback

If you need to rollback, use the old schematic:

```bash
OLD_IMAGE="factory.talos.dev/installer/ce4c980550dd2ab1b17bbf2b08801c7eb59418eafe8f279833297925d67c7515:v1.9.3"
talosctl -n <node-ip> upgrade --image="${OLD_IMAGE}" --preserve
```

## Timeline

- **Total time**: ~30-40 minutes for all 7 nodes
- **Per node**: ~3-5 minutes (including wait time)
- **No hardcoded sleeps**: Uses actual health checks

## See Also

- [Talos Upgrade Guide](../docs/talos-upgrade-guide.md) - Manual upgrade procedure
- [Talos Installation Guide](../docs/talos-iscsi-installation.md) - Installation details
