---
title: ComfyUI
description: Node-graph image generation on the GPU node. Ships scaled to zero so it does not hold VRAM Ollama needs.
sidebar:
  order: 30
  badge: { text: Off, variant: caution }
---

ComfyUI generates images from Stable Diffusion and Flux workflows on the RTX 5070 Ti. [Upstream docs](https://docs.comfy.org/).

| | |
|---|---|
| Namespace | `comfyui` |
| Source | [`comfyui.go`](src:workloads/ai/comfyui.go) |
| Exposure | LAN · `comfyui.madhan.app` |
| Storage | 50Gi RWX on `longhorn`, mounted at `/home/user/opt/ComfyUI` |
| Auth | None |

It ships with `replicas: 0`: ComfyUI keeps its model in VRAM after a run, which leaves Ollama unable to load one. See [Scheduling a GPU pod](/hardware/gpu/#scheduling-a-gpu-pod) for the shared-VRAM limits.

## Configuration

| Setting | Value | Why |
|---|---|---|
| Image | `yanwk/comfyui-boot` megapak, [pinned version](/reference/software-inventory/) | Ships PyTorch and common custom nodes (12.3 GB). Blackwell needs a CUDA 12.8+ build |
| Replicas | `0` (marked `// comfyui-replicas`) | Default off; `just comfyui` rewrites the literal |
| Strategy | `Recreate` | Two pods cannot both start against the GPU during a rollout |
| Memory limit | `6Gi` (request `1Gi`) | A runaway workflow is OOM-killed alone instead of starving worker4 |
| `CLI_ARGS` | `--listen 0.0.0.0 --port 8188` | Binds all interfaces so the Service can reach it |
| PVC | 50Gi RWX | Checkpoints, LoRAs, custom nodes and outputs survive restarts |

## Operate

Argo CD self-heals, so `kubectl scale` is reverted within minutes. Change the replica count in git:

```bash frame="terminal"
just comfyui on    # rewrite replicas to 1, synth, commit, push
just comfyui off   # back to 0 — frees the VRAM
```

Argo CD applies it once CI republishes the manifests branch. The first start pulls a 12.3 GB image: `kubectl get pods -n comfyui -w`.

## Troubleshooting

### CUDA out of memory during generation

Ollama holds VRAM. Unload its model: `curl https://ollama.madhan.app/api/generate -d '{"model":"executor","keep_alive":0}'`.

### Custom nodes vanish

They live on the PVC under `custom_nodes/`. Deleting the PVC deletes them.
