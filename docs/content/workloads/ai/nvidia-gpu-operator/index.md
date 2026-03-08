+++
title = "NVIDIA Device Plugin"
description = "Standalone NVIDIA k8s-device-plugin with NFD, GFD, and time-slicing for Talos Linux."
weight = 10
+++

## Overview

| Property | Value |
|----------|-------|
| CDK8s file | `workloads/hardware/nvidia_gpu_operator.go` |
| Namespace | `nvidia-gpu-operator` |
| Helm chart | `nvidia-device-plugin` v0.18.2 (nvidia.github.io/k8s-device-plugin) |
| HTTPRoute | None |
| UI | No |

## Purpose

Advertises `nvidia.com/gpu` as a schedulable Kubernetes resource on nodes with an NVIDIA GPU.

On Talos Linux, the GPU driver and container toolkit are provided by Talos system extensions — the full GPU operator is not needed and its validator is incompatible with Talos's non-standard library paths. The standalone device plugin is a drop-in replacement with no validation init containers.

Sub-charts included:

- **NFD (Node Feature Discovery):** labels GPU nodes with `feature.node.kubernetes.io/pci-10de.present=true`
- **GFD (GPU Feature Discovery):** adds `nvidia.com/gpu.present=true`, product, memory, and count labels

## Talos-Specific Configuration

`deviceDiscoveryStrategy: nvml` and `deviceListStrategy: envvar` are set via the inline plugin config. CDI mode generates spec entries with hostPath mounts that fail on Talos because NVIDIA libs live at `/usr/local/glibc/usr/lib/` (squashfs, not bind-mountable). With `envvar` mode the device plugin injects `NVIDIA_VISIBLE_DEVICES=<uuid>` into the container environment, and the Talos `nvidia-container-runtime` extension handles GPU device injection using its own Talos-aware library paths.

A `RuntimeClass` named `nvidia` (handler: `nvidia`) is deployed alongside the chart. The device plugin and GFD pods use this RuntimeClass so containerd routes them through `nvidia-container-runtime`, which injects `libnvidia-ml.so.1` at startup. Without this, NVML is not accessible and both pods crash with `ERROR_LIBRARY_NOT_FOUND`.

## GPU Time-Slicing

Time-slicing is configured inline in the Helm values `config.map.default`:

```yaml
version: v1
flags:
  migStrategy: none
  deviceDiscoveryStrategy: nvml
  plugin:
    deviceListStrategy: envvar
sharing:
  timeSlicing:
    resources:
      - name: nvidia.com/gpu
        replicas: 2
```

After applying, `kubectl describe node k8s-worker4` shows:
```
Allocatable:
  nvidia.com/gpu: 2
```

Both Ollama and ComfyUI can request `nvidia.com/gpu: 1` simultaneously. VRAM is shared (not partitioned) — ~4 GB (Ollama 7B) + ~6 GB (ComfyUI SDXL) fits within the 16 GB RTX 5070 Ti pool.
