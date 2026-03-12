+++
title = "GPU"
description = "NVIDIA RTX 5070 Ti setup: PCIe passthrough, Talos extensions, time-slicing."
weight = 20
+++

## Hardware

**Node:** k8s-worker4 (`192.168.1.224`)
**GPU:** NVIDIA RTX 5070 Ti — 16 GB GDDR7 VRAM
**PCIe ID:** `0000:09:00.0` (passthrough to VM)
**vCPUs:** 8 cores (dedicated AI node)
**RAM:** 16 GiB (16384 MB)
**Disk:** 250 GiB (extra space for AI model volumes)

## Talos GPU Extensions

The GPU worker uses a custom Talos image with two additional system extensions:

| Extension | Purpose |
|-----------|---------|
| `nvidia-open-gpu-kernel-modules-production` | Open-source NVIDIA kernel driver (loaded as kernel modules, not compiled) |
| `nvidia-container-toolkit-production` | Container runtime hook — configures containerd CDI automatically |

These extensions are baked into the Talos image at boot. No `machine.files` drop-ins are needed — the container toolkit extension configures containerd automatically. (Talos v1.10+ restricts `machine.files` writes to `/var`; `/etc/cri/conf.d/` is not writable.)

## GPU Node Taint

The GPU worker carries a `dedicated=ai:NoSchedule` taint, applied via the Talos machine patch:

```yaml
machine:
  nodeTaints:
    dedicated: "ai:NoSchedule"
```

Only workloads with a matching toleration (`key: dedicated, value: ai`) are scheduled on this node.

## Time-Slicing

A single physical GPU is shared between **Ollama** and **ComfyUI** using NVIDIA GPU time-slicing. This is configured inline in the `nvidia-device-plugin` Helm values:

```yaml
sharing:
  timeSlicing:
    resources:
      - name: nvidia.com/gpu
        replicas: 2
```

Result: the node advertises **2 virtual `nvidia.com/gpu` resources** from 1 physical GPU. VRAM is shared (not partitioned), so both workloads compete for the 16 GB pool.

## Resource Requests

| Workload | vCPU limit | RAM Request | RAM Limit | GPU |
|----------|------------|-------------|-----------|-----|
| Ollama | 4000m | 2 Gi | 4 Gi | 1 |
| ComfyUI | 4000m | 1 Gi | 8 Gi | 1 |

## GPU Workload Configuration

Both GPU workloads use:

```yaml
runtimeClassName: nvidia
nodeSelector:
  nvidia.com/gpu.present: "true"
tolerations:
  - key: dedicated
    operator: Equal
    value: ai
    effect: NoSchedule
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

## DCGM Exporter

A DCGM Exporter DaemonSet runs on the GPU node to export GPU metrics (utilisation, VRAM usage, temperature, power draw) to VictoriaMetrics via VMAgent. It also creates a Grafana dashboard ConfigMap.

```bash
# Check GPU metrics in Grafana: look for the DCGM dashboard
# Or query directly:
kubectl get pods -n nvidia-gpu-operator -l app=dcgm-exporter
```
