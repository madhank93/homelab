+++
title = "NVIDIA Device Plugin"
description = "Standalone NVIDIA k8s-device-plugin with NFD, GFD, DCGM Exporter, and time-slicing for Talos Linux."
weight = 10
+++

## What is the NVIDIA Device Plugin?

The [NVIDIA k8s-device-plugin](https://github.com/NVIDIA/k8s-device-plugin) is a Kubernetes DaemonSet that advertises `nvidia.com/gpu` as a schedulable resource on GPU nodes. It enables pods to request GPUs in their resource limits, which triggers device assignment and NVIDIA library injection into the container.

## Why the Standalone Device Plugin (Not Full GPU Operator)?

The full NVIDIA GPU Operator includes a validator that runs init containers checking standard library paths. On Talos Linux, NVIDIA libraries live at `/usr/local/glibc/usr/lib/` (inside a squashfs filesystem, not bind-mountable via hostPath). The operator's validator fails on Talos.

The standalone device plugin has no validation init containers and works out of the box on Talos, as long as the `nvidia-open-gpu-kernel-modules-production` and `nvidia-container-toolkit-production` Talos system extensions are installed.

| Approach | Talos compatible | Setup complexity |
|----------|-----------------|-----------------|
| **Standalone device plugin** | Yes | Low |
| Full GPU Operator | No (validator fails) | High |

## How It's Used Here

Three Helm charts are deployed in the `nvidia-gpu-operator` namespace:

| Chart | Version | Purpose |
|-------|---------|---------|
| `nvidia-device-plugin` | v0.18.2 | GPU device advertisement + time-slicing config |
| `dcgm-exporter` | v3.4.2 | GPU metrics (util, VRAM, temp, power) |
| (NFD/GFD) | bundled with device plugin | Node/GPU feature labels |

A `RuntimeClass` named `nvidia` is also created (handler: `nvidia`), matching the containerd runtime configured by the `nvidia-container-toolkit-production` Talos extension.

Source: [`workloads/hardware/nvidia_gpu_operator.go`](https://github.com/madhank93/homelab/blob/v0.1.5/workloads/hardware/nvidia_gpu_operator.go)

## Configuration

| Setting | Value | Why |
|---------|-------|-----|
| Namespace | `nvidia-gpu-operator` | Privileged PSA required |
| `runtimeClassName` | `nvidia` | Routes containers through nvidia-container-runtime |
| `deviceDiscoveryStrategy` | `nvml` | Direct kernel module NVML access (not standard paths) |
| `deviceListStrategy` | `envvar` | Inject `NVIDIA_VISIBLE_DEVICES` (CDI hostPath fails on Talos) |
| Time-slicing replicas | `5` | Ollama + ComfyUI + Kubeflow notebook + training job + Katib trial |
| NFD | enabled | Labels GPU nodes with `feature.node.kubernetes.io/pci-10de.present=true` |
| GFD | enabled | Labels with `nvidia.com/gpu.present=true`, product, memory, count |

## Time-Slicing Config

The inline plugin configuration enables time-slicing:

```yaml
sharing:
  timeSlicing:
    resources:
      - name: nvidia.com/gpu
        replicas: 5
```

After applying, the GPU node advertises `nvidia.com/gpu: 5` (allocatable). Ollama, ComfyUI, and up to 3 Kubeflow workloads (notebooks, training jobs, Katib trials) can each request `nvidia.com/gpu: 1` simultaneously. VRAM is shared (not isolated) — typical concurrent load is 2–3 processes (~10–12 GB), well within the 16 GB pool. Running Flux.1 (~12 GB) + Ollama simultaneously will OOM.

## Talos-Specific Configuration

- **`deviceDiscoveryStrategy: nvml`** — `auto` probes standard paths that don't exist on Talos. `nvml` talks directly to the kernel module loaded by the `nvidia-open-gpu-kernel-modules-production` extension.
- **`deviceListStrategy: envvar`** — CDI mode generates hostPath mounts pointing to standard library paths that don't exist on Talos. `envvar` injects `NVIDIA_VISIBLE_DEVICES=<uuid>` into the container environment and lets the Talos `nvidia-container-runtime` extension handle GPU injection using its own Talos-aware paths.

## DCGM Exporter

The DCGM Exporter runs as a DaemonSet only on the GPU node (via node affinity on `nvidia.com/gpu.present=true` and toleration for `dedicated=ai:NoSchedule`). It exports GPU metrics to VMAgent via a ServiceMonitor.

Metrics available:
- GPU utilization (`DCGM_FI_DEV_GPU_UTIL`)
- VRAM usage (`DCGM_FI_DEV_FB_USED`)
- Temperature (`DCGM_FI_DEV_GPU_TEMP`)
- Power draw (`DCGM_FI_DEV_POWER_USAGE`)
- Clock speeds

The DCGM Exporter also creates a Grafana dashboard ConfigMap with pre-built GPU panels.

## How It Connects

```
Talos system extensions (boot time):
  nvidia-open-gpu-kernel-modules → loads nvidia.ko, nvidia_uvm.ko, etc.
  nvidia-container-toolkit → configures containerd nvidia runtime

NVIDIA Device Plugin DaemonSet:
  → NFD labels GPU nodes
  → GFD labels GPU capabilities
  → Advertises nvidia.com/gpu:2 on k8s-worker4

Pod with nvidia.com/gpu: 1 + runtimeClassName: nvidia:
  → containerd routes to nvidia-container-runtime
  → nvidia-container-runtime injects libnvidia-ml.so.1 into container
  → GPU accessible to application

DCGM Exporter → ServiceMonitor → VMAgent → VictoriaMetrics → Grafana
```

## Troubleshooting

### GPU Not Advertised

```bash
# Check device plugin is running on GPU node
kubectl get pods -n nvidia-gpu-operator -o wide | grep worker4

# Check node allocatable
kubectl describe node k8s-worker4 | grep -A5 Allocatable

# Should show:
# nvidia.com/gpu: 2
```

### NVML Initialization Failed

**Symptoms:** Device plugin logs show `Failed to initialize NVML`.

**Fix:** Check kernel modules are loaded:

```bash
talosctl --talosconfig ~/.talos/config dmesg --nodes 192.168.1.224 | grep nvidia
```

If modules are missing, the Talos GPU image may not have been used for this node. Re-provision with the GPU schematic image.

### Container Not Getting GPU

```bash
# Check runtimeClassName is set
kubectl get pod <pod-name> -n <namespace> -o yaml | grep runtimeClassName

# Check tolerations include dedicated=ai
kubectl get pod <pod-name> -n <namespace> -o yaml | grep -A5 tolerations
```
