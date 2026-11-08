---
title: GPU
description: The RTX 5070 Ti on k8s-worker4 — passthrough, the dedicated=ai taint, time-slicing, and what every GPU pod needs.
sidebar:
  order: 20
---

An NVIDIA RTX 5070 Ti (16 GB VRAM) is passed through to `k8s-worker4` and shared by Ollama, ComfyUI and Kubeflow. [Device plugin docs](https://github.com/NVIDIA/k8s-device-plugin).

| | |
|---|---|
| Node | `k8s-worker4` · `192.168.1.224` |
| PCIe ID | `0000:09:00.0` (`pcie_ids` in [`config.yml`](src:config/config.yml)) |
| Namespace | `nvidia-gpu-operator` |
| Source | [`nvidia_gpu_operator.go`](src:workloads/hardware/nvidia_gpu_operator.go), [`talos.go`](src:core/platform/talos.go) |

## How it fits

The driver and container toolkit come from Talos system extensions in the GPU image. In the cluster, the [NVIDIA device plugin](/workloads/hardware/nvidia-device-plugin/) advertises the card, time-slices it and exports DCGM metrics; its settings (RuntimeClass, NVML discovery, time-slicing) are documented there.

Proxmox passes the card with `pcie: true` and `xvga: false`: it is compute-only, and VGA BAR constraints conflict with large memory allocations. Passthrough changes need a cold boot of the VM.

## Configuration

| Setting | Value | Why |
|---|---|---|
| Node taint | `dedicated=ai:NoSchedule` | Keeps ordinary pods off the GPU node's CPU during inference |

## Scheduling a GPU pod

Every GPU pod needs all of these. Device plugin and DCGM Exporter already carry them.

```yaml
runtimeClassName: nvidia            # without it CUDA is invisible in the container
nodeSelector:
  nvidia.com/gpu.present: "true"    # set by GPU Feature Discovery
tolerations:                        # the selector alone will not pass the taint
  - key: dedicated
    operator: Equal
    value: ai
    effect: NoSchedule
resources:
  limits:
    nvidia.com/gpu: 1
env:
  - name: NVIDIA_VISIBLE_DEVICES
    value: all
```

:::caution[Missing runtimeClassName fails quietly]
The pod schedules and gets `nvidia.com/gpu`, yet CUDA sees no device. Kubeflow notebooks need it too.
:::

:::caution[Five slices, one 16 GB pool]
Time-slicing shares GPU time, not memory, and nothing in Kubernetes limits VRAM. ComfyUI keeps its model in VRAM after a run, so with both running Ollama fails to load a model and the error shows up on the Ollama side. ComfyUI ships scaled to zero; turn it on only when needed and off afterwards.
:::

## Operate

```sh frame="terminal"
just comfyui on     # commits the replica change; Argo CD selfHeal reverts kubectl scale
just comfyui off    # the only reliable way to free ComfyUI's VRAM

# Evict Ollama's loaded model without a restart
curl https://ollama.madhan.app/api/generate -d '{"model": "<model>", "keep_alive": 0}'

kubectl get pods -n nvidia-gpu-operator
kubectl describe node k8s-worker4 | grep nvidia.com/gpu
```

GPU utilisation, VRAM, temperature and power come from DCGM Exporter through its ServiceMonitor; look for the DCGM dashboard in Grafana.

## Troubleshooting

### GPU pod stuck Pending with no obvious reason

The pod lacks the `dedicated=ai` toleration, or `nvidia.com/gpu` is exhausted. Add the toleration; check `kubectl describe node k8s-worker4` for allocated GPUs.

### Ollama cannot load a model while ComfyUI is running

ComfyUI holds the VRAM. Run `just comfyui off`.

### Pod runs but `nvidia-smi` or CUDA finds no device

`runtimeClassName: nvidia` is missing. Add it and recreate the pod.
