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

{% mermaid() %}
flowchart TB
    subgraph CARD["RTX 5070 Ti · 16 GB VRAM · not partitioned"]
        VRAM["one shared pool"]
    end

    subgraph SLICES["Device plugin advertises nvidia.com/gpu: 5"]
        S["5 time-slices of GPU *compute*"]
    end

    OLL["Ollama<br/>evictable: keep_alive 0"]
    CFY["ComfyUI<br/>holds VRAM after a run"]
    NB["Kubeflow notebooks<br/>hold until deleted"]

    S --> OLL & CFY & NB
    OLL --> VRAM
    CFY --> VRAM
    NB --> VRAM
{% end %}

Time-slicing divides GPU *time*, and the `nvidia.com/gpu: 5` count is a count of
those slices — it says nothing about memory. Three pods can each hold a slice and
still collectively exhaust the 16 GB, at which point the next allocation fails.

ComfyUI therefore ships scaled to zero and is turned on only when needed:

```bash
just comfyui on
just comfyui off
```

Ollama's model can be evicted without a restart:

```bash
curl https://ollama.madhan.app/api/generate -d '{"model": "llama3.2", "keep_alive": 0}'
```

## The `dedicated=ai` taint

worker4 carries `dedicated=ai:NoSchedule`, set in the GPU machine patch
({{ src(path="core/platform/talos.go") }}) so that ordinary workloads do not drift
onto the GPU node and compete for its CPU during inference.

Every pod that belongs on worker4 must therefore carry the matching toleration —
Ollama, ComfyUI, the NVIDIA device plugin, DCGM exporter, and any Kubeflow notebook
requesting a GPU. Without it the pod stays `Pending` with no obvious cause:

```yaml
tolerations:
  - key: dedicated
    operator: Equal
    value: ai
    effect: NoSchedule
```

The taint controls *scheduling*; `nodeSelector: nvidia.com/gpu.present: "true"`
(a label from GPU Feature Discovery) is what selects the node. Both are needed —
the selector alone will not get a pod past the taint.

## GPU Workload Configuration

Every GPU pod needs all four of these. Miss `runtimeClassName` and the NVIDIA
container hook never fires — the pod schedules, the resource is granted, and CUDA
is still invisible inside the container, which is a confusing failure to debug.

```yaml
runtimeClassName: nvidia            # or CUDA is not visible in the container
nodeSelector:
  nvidia.com/gpu.present: "true"    # GPU Feature Discovery label
tolerations:                        # worker4 is tainted — see above
  - key: dedicated
    operator: Equal
    value: ai
    effect: NoSchedule
resources:
  limits:
    nvidia.com/gpu: "1"
env:
  - name: NVIDIA_VISIBLE_DEVICES
    value: all
```

This applies to Kubeflow notebooks too — they request the same pool.

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
kubectl get pods -n nvidia-gpu-operator -l app.kubernetes.io/name=dcgm-exporter
```
