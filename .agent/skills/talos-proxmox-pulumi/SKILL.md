---
name: talos-proxmox-pulumiverse
description: The Gold Standard for deploying Talos Linux on Proxmox VE using Pulumi (Go).
---

# Talos on Proxmox (Pulumi Go)

## Goal
To deploy a production-ready, High Availability (HA) Kubernetes cluster using Talos Linux on Proxmox VE, entirely managed by Pulumi with NO external bootstrapping scripts.

## Core Technologies
- **Provider**: `muhlba91/pulumi-proxmoxve` (VMs, ISOs).
- **Provisions**: `pulumiverse/pulumi-talos` (Config, Secrets, Bootstrap).
- **Language**: Go.

## Critical Implementation Details

### 1. In-Band Bootstrapping (The "Pure Pulumi" Way)
Do NOT use `local-exec` or external shell scripts to bootstrap.
Use the Pulumi Resources:
- `talos_machine.NewSecrets` (Generate secrets once).
- `talos_machine.NewConfigurationApply` (Apply config to nodes).
- `talos_machine.NewBootstrap` (Bootstrap the cluster).
- `talos_cluster.NewKubeconfig` (Retrieve access).

### 2. High Availability (HA) & Race Conditions
When deploying 3+ control plane nodes:
- **Seed Node**: Target **ONLY** the first controller (e.g., `k8s-controller1`) for the `NewBootstrap` resource.
- **Why**: If you target the VIP or a random node, the bootstrap command might reach a non-leader or time out waiting for the VIP to float.
- **Pattern**:
  ```go
  bootstrap, err := talos_machine.NewBootstrap(ctx, "bootstrap", &talos_machine.BootstrapArgs{
      Node: controllerIP1, // Direct IP of Node 1, NOT VIP
      // ...
  })
  ```

### 3. Proxmox VM Configuration
Talos requires specific Proxmox settings to run correctly and support features like Graceful Shutdown and Longhorn/Storage.
- **CPU**: `Type: "host"` (Critical for performance/AES-NI).
- **Machine**: `q35`.
- **BIOS**: `ovmf` (UEFI).
- **Agent**: `Enabled: true` (QEMU Agent).
- **Disk**: `VirtIO` (scsi0), `SSD Emulation` (optional but good).
- **Network**: `VirtIO` bridge.

### 4. Handling DHCP Lag (The "Justfile" Pattern)
Proxmox QEMU Agent takes time to report the IP address. Pulumi `ApplyT` might fail on the first run if the IP is empty.
**Do NOT** use complex Go retry loops inside Pulumi (they read stale state).
**DO** use the `deploy-proxmox` workflow in `Justfile`:
1.  `pulumi up` (Might fail on "waiting for IP").
2.  `pulumi refresh` (Captures the IPs now that VMs are up).
3.  `pulumi up` (Completes the bootstrap).

### 5. Hostname Consistency
Proxmox VM Name must match Kubernetes Node Name.
- **Problem**: Talos defaults hostname to `talos-<random-id>`.
- **Fix**: Inject a patch into `MachineConfig`:
  ```yaml
  machine:
    network:
      hostname: <node-name> # e.g., k8s-worker1
  ```

### 6. App Verification
Keep Pulumi focused on **Infrastructure**. Use a script for **App Validation**.
- **Pattern**: `verify_cluster.sh`.
- **Steps**:
  1.  Wait for API (VIP).
  2.  Install/Check CNI (Cilium) via Helm.
  3.  install/Check Ingress/App (Nginx) via Helm/Kubectl.
  4.  Verify `kubectl get nodes` are `Ready`.

## Common Pitfalls
- **VIP Deadlock**: Configuring `machine.network.interfaces.vip` is good, but DO NOT bootstrap against it. Bootstrap against the specific IP of Controller 1.
- **Generics**: Do not use generic `local-exec` to run `talosctl`. Use the Provider.
- **GPU**: For Passthrough, set `HasGPU: true` logic to map `hostpci0`.
