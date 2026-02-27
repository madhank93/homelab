+++
title = "Cluster Nodes"
description = "Talos Linux VM configuration, node IPs, and Proxmox specs."
weight = 10
+++

## Node Configuration

All nodes are QEMU VMs on a Proxmox host, defined in `infra/pulumi/talos.go`.

### Control Plane Nodes

| Name | IP | vCPUs | RAM | Disk | CPU Priority |
|------|----|-------|-----|------|-------------|
| k8s-controller1 | 192.168.1.211 | 4 | 6 GiB | 30 GiB | 1024 (high) |
| k8s-controller2 | 192.168.1.212 | 4 | 6 GiB | 30 GiB | 1024 (high) |
| k8s-controller3 | 192.168.1.213 | 4 | 6 GiB | 30 GiB | 1024 (high) |

Control plane nodes carry a floating VIP (`192.168.1.210`) on their primary interface. This VIP is configured as a Talos machine patch so the API server endpoint remains stable regardless of which node holds it.

### Worker Nodes

| Name | IP | vCPUs | RAM | Disk | CPU Priority |
|------|----|-------|-----|------|-------------|
| k8s-worker1 | 192.168.1.221 | 4 | 6 GiB | 125 GiB | 100 (standard) |
| k8s-worker2 | 192.168.1.222 | 4 | 6 GiB | 125 GiB | 100 (standard) |
| k8s-worker3 | 192.168.1.223 | 4 | 6 GiB | 125 GiB | 100 (standard) |
| k8s-worker4 | 192.168.1.224 | 4 | 6 GiB | 125 GiB | 100 (standard) |

Workers use 125 GiB disks to provide ~100 GiB of usable Longhorn storage per node.

## Talos Version

**Talos v1.12.4** — cluster name `talos-cluster`.

### Talos Images

Two Talos image variants are used:

**Base image** (schematic `88d1f7a5c4f1d3aba7df787c448c1d3d008ed29cfb34af53fa0df4336a56040b`):
- Extensions: `iscsi-tools`, `util-linux-tools`, `qemu-guest-agent`
- Used by: control plane nodes + k8s-worker1–3

**GPU image** (schematic `901b9afcf2f7eda57991690fc5ca00414740cc4ee4ad516109bcc58beff1b829`):
- Extensions: all base extensions + `nvidia-container-toolkit`, `nvidia-open-gpu-kernel-modules`
- Used by: k8s-worker4

## Worker Node Kernel Modules

All workers load the following kernel modules via Talos machine config patches:

```
nbd, iscsi_tcp, iscsi_generic, configfs
```

The GPU worker (k8s-worker4) additionally loads:

```
nvidia, nvidia_uvm, nvidia_drm, nvidia_modeset
```

## Network Configuration

Each node is assigned a static IP via Talos machine config (patched by Pulumi at provision time):

```yaml
machine:
  network:
    hostname: k8s-worker1
    interfaces:
      - deviceSelector:
          physical: true
        dhcp: false
        addresses:
          - 192.168.1.221/24
        routes:
          - gateway: 192.168.1.254
    nameservers:
      - 1.1.1.1
      - 192.168.1.254
```

## Longhorn Node Labels

All worker nodes carry the label `node.longhorn.io/create-default-disk: config`, applied via the Talos worker machine patch. This tells Longhorn to provision its default disk on every worker.
