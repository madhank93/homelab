+++
title = "ComfyUI"
description = "Node-based image generation UI for Stable Diffusion and Flux models."
weight = 30
+++

## What is ComfyUI?

[ComfyUI](https://github.com/comfyanonymous/ComfyUI) is a powerful, modular, node-based GUI for Stable Diffusion and Flux image generation models. It allows building complex image generation pipelines by connecting nodes (samplers, models, VAEs, ControlNets, LoRAs, etc.) in a visual graph editor.

## Why ComfyUI?

ComfyUI is the most flexible and performance-oriented frontend for diffusion models. Unlike Automatic1111 (which uses a form-based interface), ComfyUI exposes every parameter of the generation pipeline as a connectable node, enabling complex workflows that would be impossible in simpler UIs.

## How It's Used Here

ComfyUI runs on k8s-worker4 (GPU node) as a standard Kubernetes Deployment. Model files are stored on a 50 Gi RWX Longhorn PVC shared across pod restarts.

Source: [`workloads/ai/comfyui.go`](https://github.com/madhank93/homelab/blob/v0.1.5/workloads/ai/comfyui.go)

## Configuration

| Setting | Value | Why |
|---------|-------|-----|
| Namespace | `comfyui` | Isolated namespace |
| Image | `yanwk/comfyui-boot:cu128-megapak-20260223` | CUDA 12.8, matching Talos driver 570.x |
| HTTPRoute | `comfyui.madhan.app` → `comfyui:8188` | Gateway API |
| Deploy strategy | `Recreate` | GPU workloads cannot have two pods claiming `nvidia.com/gpu` simultaneously |
| `runtimeClassName` | `nvidia` | Routes through nvidia-container-runtime |
| `NVIDIA_VISIBLE_DEVICES` | `all` | Make all GPU devices visible |
| `nvidia.com/gpu` limit | `1` | One time-sliced virtual GPU |
| Node selector | `nvidia.com/gpu.present: "true"` | Schedule on GPU node |
| Toleration | `dedicated=ai:NoSchedule` | Allow scheduling on tainted GPU node |
| CPU limit | `4000m` | CPU-hungry during inference |
| RAM request | `1Gi` | Host RAM for process |
| RAM limit | `8Gi` | Allow headroom for large model loading |
| Data PVC | `50Gi` RWX Longhorn | Models, outputs, custom nodes |
| Data mount path | `/home/user/opt/ComfyUI` | Image's working directory |
| `CLI_ARGS` | `--listen 0.0.0.0 --port 8188` | Bind all interfaces for Service to reach |

## Image Note

The image tag `cu128-megapak-20260223` is a dated CUDA 12.8 "megapak" build that includes PyTorch, xformers, and many common custom nodes pre-installed. The `latest-cu128` tag does not exist on Docker Hub — always use a specific dated tag like `cu128-megapak-20260223`.

## Recreate Strategy

ComfyUI uses `strategy: Recreate` instead of the default `RollingUpdate`:

```go
Strategy: &k8s.DeploymentStrategy{Type: jsii.String("Recreate")},
```

This is required because GPU workloads cannot have two pods claiming `nvidia.com/gpu` simultaneously. With RollingUpdate, the new pod starts before the old pod terminates, causing the new pod to get stuck waiting for the GPU resource.

## Storage

A 50 Gi RWX Longhorn PVC is mounted at `/home/user/opt/ComfyUI`. This stores:
- Model checkpoints (`.safetensors`, `.ckpt`)
- LoRA files
- ControlNet models
- Custom nodes (installed via ComfyUI Manager)
- Generated output images

Using RWX allows future multi-replica scenarios and avoids attach conflicts.

## How It Connects

```
Browser → comfyui.madhan.app
  → homelab-gateway → comfyui:8188
  → ComfyUI pod on k8s-worker4
  → nvidia-container-runtime (GPU injection)
  → RTX 5070 Ti (VRAM for model inference)
  → 50Gi Longhorn RWX PVC (model files, outputs)
```

## Screenshots

![ComfyUI node graph showing a Stable Diffusion workflow with samplers, VAE, and ControlNet nodes](/assets/screenshots/comfyui/workflow.png)

## Troubleshooting

### Pod Stuck in ContainerCreating

**Symptoms:** New pod won't start, events show GPU resource not available.

**Fix:** Old pod may still be terminating. The `Recreate` strategy ensures the old pod terminates first, but if it gets stuck:

```bash
kubectl delete pod -n comfyui <old-pod> --grace-period=0 --force
```

### Out of VRAM

**Symptoms:** ComfyUI shows `CUDA out of memory` during generation.

**Fix:** Reduce model size (use smaller quantized model) or free VRAM by offloading models in the Ollama API:

```bash
# Unload Ollama models from VRAM
curl http://ollama.madhan.app/api/generate -d '{"model": "llama3.2", "keep_alive": 0}'
```

### Custom Nodes Not Installing

Custom nodes installed via ComfyUI Manager persist because they write to the 50 Gi PVC at `/home/user/opt/ComfyUI/custom_nodes/`. If the PVC is deleted, all custom nodes are lost.
