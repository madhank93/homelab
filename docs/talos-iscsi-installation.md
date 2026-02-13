# Talos iSCSI Tools Installation Guide

## Problem

Longhorn manager pods are failing with error:
```
failed to check environment, please make sure you have iscsiadm/open-iscsi installed on the host
nsenter: failed to execute iscsiadm: No such file or directory
```

**Root Cause**: Talos Linux nodes are missing the `iscsi-tools` system extension, which provides the `iscsiadm` binary required by Longhorn.

## Solution

Install the `iscsi-tools` system extension on all Talos worker nodes.

## Installation Steps

### Step 1: Check Current Extensions

```bash
# Check installed extensions on a worker node
talosctl -n <worker-ip> get extensions

# Example:
talosctl -n 192.168.1.101 get extensions
```

### Step 2: Create Talos Image with iSCSI Tools

You need to create a custom Talos image that includes the `iscsi-tools` extension using the Talos Image Factory.

**Option A: Using Talos Image Factory (Recommended)**

1. Visit https://factory.talos.dev/
2. Select your Talos version (v1.9.3)
3. Add system extensions:
   - `siderolabs/iscsi-tools`
   - `siderolabs/util-linux-tools` (recommended)
4. Generate the schematic ID

**Option B: Using talosctl**

```bash
# Create a schematic file
cat > schematic.yaml <<EOF
customization:
  systemExtensions:
    officialExtensions:
      - siderolabs/iscsi-tools
      - siderolabs/util-linux-tools
EOF

# Generate schematic ID
curl -X POST --data-binary @schematic.yaml https://factory.talos.dev/schematics
```

This will return a schematic ID like: `376567988ad370138ad8b2698212367b8edcb69b5fd68c80be1f2ec7d603b4ba`

### Step 3: Upgrade Worker Nodes

Upgrade each worker node with the new image:

```bash
# Get the installer image URL with your schematic ID
SCHEMATIC_ID="<your-schematic-id>"
TALOS_VERSION="v1.9.3"
IMAGE="factory.talos.dev/installer/${SCHEMATIC_ID}:${TALOS_VERSION}"

# Upgrade each worker node (one at a time)
talosctl -n <worker-ip> upgrade --image="${IMAGE}" --preserve

# Example:
talosctl -n 192.168.1.101 upgrade --image="factory.talos.dev/installer/376567988ad370138ad8b2698212367b8edcb69b5fd68c80be1f2ec7d603b4ba:v1.9.3" --preserve
```

**Important**: 
- Nodes will reboot during upgrade
- Upgrade one node at a time to maintain cluster availability
- Wait for each node to be Ready before upgrading the next

### Step 4: Verify Installation

After each node reboots:

```bash
# Check if iscsid service is running
talosctl -n <worker-ip> service iscsid status

# Verify iscsiadm is available
talosctl -n <worker-ip> read /proc/modules | grep iscsi

# Check extensions
talosctl -n <worker-ip> get extensions
```

Expected output should show:
- `iscsid` service running
- `iscsi_tcp` kernel module loaded
- `iscsi-tools` extension installed

### Step 5: Restart Longhorn Pods

After all workers are upgraded:

```bash
# Delete Longhorn manager pods to trigger restart
kubectl delete pods -n longhorn-system -l app=longhorn-manager

# Watch pods come back up
kubectl get pods -n longhorn-system -w
```

## Alternative: Load Kernel Modules Only

If you can't upgrade nodes immediately, you can try loading the kernel modules manually (temporary workaround):

```bash
# On each worker node
talosctl -n <worker-ip> apply-config --patch '[{"op": "add", "path": "/machine/kernel/modules", "value": [{"name": "iscsi_tcp"}]}]'
```

**Note**: This is a temporary workaround and may not work if the `iscsiadm` binary is still missing.

## Verification Checklist

After installation:

- [ ] All worker nodes upgraded with `iscsi-tools` extension
- [ ] `iscsid` service running on all workers
- [ ] `iscsi_tcp` kernel module loaded
- [ ] Longhorn manager pods running (4/4)
- [ ] Longhorn driver-deployer running (1/1)
- [ ] Longhorn UI running (2/2)
- [ ] StorageClass `longhorn` created

## Troubleshooting

### Issue: Node won't upgrade

```bash
# Check node status
talosctl -n <worker-ip> health

# Check upgrade logs
talosctl -n <worker-ip> dmesg | tail -100
```

### Issue: iscsid service not starting

```bash
# Check service status
talosctl -n <worker-ip> service iscsid status

# View service logs
talosctl -n <worker-ip> logs iscsid
```

### Issue: Longhorn pods still failing

```bash
# Check pod logs
kubectl logs -n longhorn-system -l app=longhorn-manager --tail=50

# Describe pod for events
kubectl describe pod -n longhorn-system -l app=longhorn-manager
```

## References

- [Talos Image Factory](https://factory.talos.dev/)
- [Talos System Extensions](https://www.talos.dev/v1.9/talos-guides/configuration/system-extensions/)
- [Longhorn on Talos](https://www.talos.dev/v1.9/kubernetes-guides/configuration/storage/#longhorn)
- [Talos iSCSI Tools Extension](https://github.com/siderolabs/extensions/tree/main/storage/iscsi-tools)

## Next Steps

Once iSCSI tools are installed and Longhorn is running:

1. Verify StorageClass creation
2. Test Infisical PVC binding
3. Complete Infisical deployment
4. Update release notes with resolution
