+++
title = "Proxmox + Talos"
description = "Proxmox VM provisioning and Talos Linux cluster bootstrap via Pulumi."
weight = 10
+++

## Overview

Pulumi provisions 7 QEMU VMs on Proxmox and bootstraps a 3-control-plane Talos Linux cluster. The entry point is `infra/pulumi/talos.go`.

## Talos Images

Two image variants are downloaded from the Talos image factory:

**Base image** (for controllers and standard workers):
- Schematic: `88d1f7a5c4f1d3aba7df787c448c1d3d008ed29cfb34af53fa0df4336a56040b`
- Extensions: `iscsi-tools`, `util-linux-tools`, `qemu-guest-agent`

**GPU image** (for k8s-worker4):
- Schematic: `901b9afcf2f7eda57991690fc5ca00414740cc4ee4ad516109bcc58beff1b829`
- Extensions: all base extensions + `nvidia-container-toolkit-production`, `nvidia-open-gpu-kernel-modules-production`

## Cluster Configuration

**Cluster name:** `talos-cluster`
**Talos version:** v1.12.4
**API endpoint (VIP):** `https://192.168.1.210:6443`

### Cluster-Level Patch

Applied to all nodes:

```yaml
cluster:
  network:
    cni:
      name: none   # Cilium is installed separately
  proxy:
    disabled: true  # kube-proxy replaced by Cilium eBPF
```

### Control Plane Patch

Adds VIP to the primary interface:

```yaml
machine:
  network:
    interfaces:
      - deviceSelector:
          physical: true
        dhcp: true
        vip:
          ip: 192.168.1.210
```

### Worker Patch

Adds Longhorn label and loads storage/iSCSI kernel modules:

```yaml
machine:
  nodeLabels:
    "node.longhorn.io/create-default-disk": "config"
  network:
    interfaces:
      - deviceSelector:
          physical: true
        dhcp: true
  kernel:
    modules:
      - name: nbd
      - name: iscsi_tcp
      - name: iscsi_generic
      - name: configfs
```

### GPU Worker Patch

Extends the worker patch with NVIDIA kernel modules:

```yaml
  kernel:
    modules:
      - name: nbd
      - name: iscsi_tcp
      - name: iscsi_generic
      - name: configfs
      - name: nvidia
      - name: nvidia_uvm
      - name: nvidia_drm
      - name: nvidia_modeset
```

## Bootstrap Flow

Pulumi performs the full cluster bootstrap in dependency order:

1. Download Talos images to Proxmox
2. Create VMs (Proxmox QEMU resources)
3. Generate Talos machine secrets (PKI, tokens)
4. Patch per-node configs (hostname, static IP, gateway, DNS)
5. Apply configs via Talos API (in-band, using QEMU guest agent IP)
6. Bootstrap etcd on `k8s-controller1`
7. Wait for the cluster to become healthy
8. Retrieve kubeconfig → write to `infra/pulumi/kubeconfig`

The kubeconfig is used by downstream platform Pulumi code (Cilium, ArgoCD) via the Kubernetes provider.

## Exported Outputs

```bash
pulumi stack output talosconfig   # Talos client config
pulumi stack output kubeconfig    # Kubernetes admin config
pulumi stack output k8s-worker4-ip  # etc.
```

## PCIe GPU Passthrough

k8s-worker4 is created with `PcieIDs: ["0000:28:00.0"]` which instructs Proxmox to pass the NVIDIA RTX 5070 Ti directly to the VM.
