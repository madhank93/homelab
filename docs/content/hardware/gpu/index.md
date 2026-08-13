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

These extensions are baked into the Talos image at boot. No `machine.files` drop-ins are needed — the container toolkit extension configures containerd automatically. (Talos v1.10+ restricts `machine.files` writes to `/var`; `/etc/cri/conf.d/` is not writable.) <!-- docs-check: historical -->

## Time-Slicing

A single physical GPU is shared across multiple workloads (Ollama, ComfyUI, Kubeflow notebooks, training jobs) using NVIDIA GPU time-slicing. This is configured inline in the `nvidia-device-plugin` Helm values:

```yaml
sharing:
  timeSlicing:
    resources:
      - name: nvidia.com/gpu
        replicas: 5
```

Result: the node advertises **5 virtual `nvidia.com/gpu` resources** from 1 physical GPU. VRAM is shared (not partitioned), so all workloads compete for the 16 GB pool.

## Resource Requests

These are **host RAM** cgroup limits, not VRAM. Nothing in Kubernetes limits VRAM
at all. The two limits together stay under worker4's ~15.1 Gi allocatable, so a
runaway workload is OOM-killed on its own rather than taking its neighbours down.

| Workload | vCPU limit | RAM Request | RAM Limit | GPU |
|----------|------------|-------------|-----------|-----|
| Ollama | 4000m | 4 Gi | 8 Gi | 1 |
| ComfyUI | 4000m | 1 Gi | 6 Gi | 1 |

## Sharing the 16 GB of VRAM

Time-slicing shares GPU *time*, not memory. ComfyUI keeps its model resident in
VRAM after a run and does not release it, so with both running Ollama fails to
load a model — and the error surfaces on the Ollama side, which makes it look
like an Ollama problem.

ComfyUI therefore ships scaled to zero and is turned on only when needed:

```bash
just comfyui on
just comfyui off
```

Ollama's model can be evicted without a restart:

```bash
curl https://ollama.madhan.app/api/generate -d '{"model": "llama3.2", "keep_alive": 0}'
```

Kubeflow notebooks requesting `nvidia.com/gpu` compete for the same pool, and
**must** set `runtimeClassName: nvidia` — without it the NVIDIA container hook
never fires and CUDA is invisible even though the resource was granted.

## GPU Workload Configuration

All GPU workloads use `nodeSelector: nvidia.com/gpu.present: "true"` to pin to `k8s-worker4`. `runtimeClassName: nvidia` is required — without it the NVIDIA container hook does not fire and CUDA is inaccessible even with the `nvidia.com/gpu` resource requested.

```yaml
runtimeClassName: nvidia
nodeSelector:
  nvidia.com/gpu.present: "true"
resources:
  limits:
    nvidia.com/gpu: "1"
env:
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
