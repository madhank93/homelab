+++
title = "AI"
description = "GPU-accelerated AI workloads: Ollama, ComfyUI, Kubeflow, and the NVIDIA device plugin."
weight = 10
sort_by = "weight"
+++

All AI workloads run on **k8s-worker4** (`192.168.1.224`), which has PCIe passthrough to an NVIDIA RTX 5070 Ti (16 GB VRAM, Blackwell sm_120).

GPU time-slicing advertises 5 virtual GPU resources from 1 physical GPU, allowing Ollama, ComfyUI, and Kubeflow notebooks/training jobs to run concurrently. VRAM is not isolated between processes.

| Page | What it covers |
|---|---|
| [Ollama](@/workloads/ai/ollama/index.md) | LLM inference server |
| [ComfyUI](@/workloads/ai/comfyui/index.md) | Image generation — off by default |
| [Kubeflow](@/workloads/ai/kubeflow/index.md) | Notebooks and pipelines, without Istio |
| [Notebook Gateway Controller](@/workloads/ai/notebook-gateway-controller/index.md) | Creates HTTPRoutes for notebooks |
| [NVIDIA Device Plugin](@/workloads/ai/nvidia-gpu-operator/index.md) | GPU advertisement and time-slicing |

Hardware details and the VRAM constraint are on [GPU](@/hardware/gpu/index.md).
