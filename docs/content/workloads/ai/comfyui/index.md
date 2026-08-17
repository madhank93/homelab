+++
title = "ComfyUI"
description = "Node-based image generation UI for Stable Diffusion and Flux models."
weight = 30
+++

## What is ComfyUI?

[ComfyUI](https://github.com/comfyanonymous/ComfyUI) is the image-generation front end on the GPU node. Workflows are built as a node graph — samplers, checkpoints, VAEs, ControlNets, LoRAs wired together — which is what makes it worth running over a form-based UI like Automatic1111.

## Why ComfyUI?

ComfyUI is the most flexible and performance-oriented frontend for diffusion models. Unlike Automatic1111 (which uses a form-based interface), ComfyUI exposes every parameter of the generation pipeline as a connectable node, enabling complex workflows that would be impossible in simpler UIs.

## How It's Used Here

ComfyUI runs on k8s-worker4 (GPU node) as a standard Kubernetes Deployment. Model files are stored on a 50 Gi RWX Longhorn PVC shared across pod restarts.

**It is scaled to zero by default.** ComfyUI keeps its model resident in VRAM
after a run and never releases it, which leaves Ollama unable to load one on the
same 16 GB card. Nothing in Kubernetes arbitrates VRAM — the `nvidia.com/gpu`
limit counts devices, not memory — so the two do not share the card and ComfyUI
rests at zero replicas.

Source: {{ src(path="workloads/ai/comfyui.go") }}

## Turning it on and off

```bash
just comfyui on     # scale to 1
just comfyui off    # scale back to 0
```

`kubectl scale` does not work here: Argo CD runs `selfHeal`, so a manual scale is
reverted within minutes. The replica count has to change in git. The recipe
rewrites the tagged literal in `comfyui.go`, synthesizes, commits, and pushes:

```go
replicas := float64(0) // comfyui-replicas
```

Argo CD applies the change once CI republishes the manifests branch, so expect a
couple of minutes' lag. The first start pulls a 12.3 GB image:

```bash
kubectl get pods -n comfyui -w
```

Leaving it on is fine as long as nothing else needs the GPU — but Ollama will
fail to load a model while ComfyUI holds VRAM, and the failure looks like an
unrelated Ollama error, so turn it off when finished.

## Configuration

| Setting | Value | Why |
|---------|-------|-----|
| Namespace | `comfyui` | Isolated namespace |
| Image | `yanwk/comfyui-boot` | CUDA 13.0 megapak — tag in the [Software Inventory](@/architecture/software-inventory.md) |
| HTTPRoute | `comfyui.madhan.app` → `comfyui:8188` | Gateway API |
| Deploy strategy | `Recreate` | GPU workloads cannot have two pods claiming `nvidia.com/gpu` simultaneously |
| `runtimeClassName` | `nvidia` | Routes through nvidia-container-runtime |
| `NVIDIA_VISIBLE_DEVICES` | `all` | Make all GPU devices visible |
| `nvidia.com/gpu` limit | `1` | One time-sliced virtual GPU |
| Node selector | `nvidia.com/gpu.present: "true"` | Schedule on GPU node |
| Toleration | `dedicated=ai:NoSchedule` | **Required** — worker4 is tainted, see [GPU](@/hardware/gpu/index.md#the-dedicated-ai-taint) |
| CPU limit | `4000m` | CPU-hungry during inference |
| RAM request | `1Gi` | Host RAM for process |
| Replicas | `0` | Default off — see above |
| RAM limit | `6Gi` | With Ollama's 8Gi, stays under worker4's 15.1Gi allocatable |
| Data PVC | `50Gi` RWX Longhorn | Models, outputs, custom nodes |
| Data mount path | `/home/user/opt/ComfyUI` | Image's working directory |
| `CLI_ARGS` | `--listen 0.0.0.0 --port 8188` | Bind all interfaces for Service to reach |

## Image Note

The `megapak` builds ship PyTorch, xformers, and many common custom nodes
pre-installed, which is why the image is 12.3 GB.

Tags are dated and there is no `latest-*` — always pin a specific one. The card
is Blackwell (sm_120) and needs CUDA 12.8 or newer, so `cu126` builds will not
run. The `cu128-megapak` line stopped building in May 2026; `cu130-megapak-pt*`
is the current one, and the node's driver reports CUDA 13.2, so it is supported.

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

**Fix:** Free VRAM by unloading Ollama's model, or use a smaller quantized
checkpoint:

```bash
curl https://ollama.madhan.app/api/generate -d '{"model": "llama3.2", "keep_alive": 0}'
```

If Ollama is the one failing to load, the cause is usually the reverse — ComfyUI
still holds VRAM from an earlier run. `just comfyui off` releases it.

### Custom Nodes Not Installing

Custom nodes installed via ComfyUI Manager persist because they write to the 50 Gi PVC at `/home/user/opt/ComfyUI/custom_nodes/`. If the PVC is deleted, all custom nodes are lost.
