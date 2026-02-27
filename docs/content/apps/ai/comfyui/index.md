+++
title = "ComfyUI"
description = "Node-based image generation UI for Stable Diffusion and Flux models."
weight = 30
+++

## Overview

| Property | Value |
|----------|-------|
| CDK8s file | `platform/cdk8s/cots/ai/comfyui.go` |
| Namespace | `comfyui` |
| Image | `yanwk/comfyui-boot:cu128-megapak-20260223` |
| HTTPRoute | `comfyui.madhan.app` → `comfyui:8188` |
| UI | Yes |
| Storage | 100 Gi Longhorn PVC (models, outputs, custom nodes) |
| Node | k8s-worker4 (GPU) |

## Purpose

ComfyUI is a node-based GUI for Stable Diffusion and Flux image generation models. It downloads model checkpoints to a persistent Longhorn volume.

## GPU Configuration

```yaml
runtimeClassName: nvidia
resources:
  requests:
    memory: "1Gi"
  limits:
    memory: "8Gi"
    nvidia.com/gpu: "1"
env:
  - name: NVIDIA_VISIBLE_DEVICES
    value: all
```

## Deployment Strategy

Uses `strategy: Recreate` to avoid multi-attach conflicts on the Longhorn RWO PVC.

## Storage

A 100 Gi Longhorn PVC is mounted at `/home/runner` (or the image's working directory). This stores:
- Model checkpoints (`.safetensors`, `.ckpt`)
- LoRA files
- Custom nodes
- Generated output images

## Image Note

The image tag `cu128-megapak-20260223` (CUDA 12.8, megapak variant) includes PyTorch, xformers, and common custom nodes pre-installed. The tag `latest-cu128` does not exist on Docker Hub — use a dated tag.

## Accessing the UI

Navigate to `http://comfyui.madhan.app` in a browser. The ComfyUI node graph interface loads directly.
