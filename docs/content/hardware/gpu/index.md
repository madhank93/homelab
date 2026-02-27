+++
title = "GPU"
description = "NVIDIA RTX 5070 Ti setup: PCIe passthrough, Talos extensions, time-slicing."
weight = 20
+++

## Hardware

**Node:** k8s-worker4 (`192.168.1.224`)
**GPU:** NVIDIA RTX 5070 Ti — 16 GB GDDR7 VRAM
**PCIe ID:** `0000:28:00.0` (passthrough to VM)
**Allocatable RAM on node:** ~5.4 GiB

## Talos GPU Extensions

The GPU worker uses a custom Talos image with two additional system extensions:

| Extension | Purpose |
|-----------|---------|
| `nvidia-open-gpu-kernel-modules-production` | Open-source NVIDIA kernel driver (loaded as kernel modules, not compiled) |
| `nvidia-container-toolkit-production` | Container runtime hook — configures containerd CDI automatically |

These extensions are baked into the Talos image at boot. No `machine.files` drop-ins are needed — the container toolkit extension configures containerd automatically. (Talos v1.10+ restricts `machine.files` writes to `/var`; `/etc/cri/conf.d/` is not writable.)

## Time-Slicing

A single physical GPU is shared between **Ollama** and **ComfyUI** using NVIDIA GPU time-slicing. This is configured via:

1. A `time-slicing-config` ConfigMap in the `nvidia-gpu-operator` namespace with `replicas: 2`
2. The `devicePlugin.config` reference in the ClusterPolicy

Result: the node advertises **2 virtual `nvidia.com/gpu` resources** from 1 physical GPU. VRAM is shared (not partitioned), so both workloads compete for the 16 GB pool.

## Resource Requests

| Workload | CPU | RAM Request | RAM Limit | GPU |
|----------|-----|-------------|-----------|-----|
| Ollama | — | 2 Gi | — | 1 |
| ComfyUI | — | 1 Gi | 8 Gi | 1 |

The low RAM requests (3 Gi total) fit within the 5.4 Gi allocatable on the node.

## GPU Workload Configuration

Both GPU workloads use:

```yaml
runtimeClassName: nvidia
resources:
  limits:
    nvidia.com/gpu: "1"
```

And the environment variable:
```yaml
- name: NVIDIA_VISIBLE_DEVICES
  value: all
```

## Kernel Modules

The GPU worker's Talos machine patch loads these modules at boot:

```
nvidia
nvidia_uvm
nvidia_drm
nvidia_modeset
```
