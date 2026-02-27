+++
title = "AI"
description = "GPU-accelerated AI workloads: Ollama, ComfyUI, and NVIDIA GPU Operator."
weight = 10
sort_by = "weight"
+++

All AI workloads run on **k8s-worker4** (`192.168.1.224`), which has PCIe passthrough to an NVIDIA RTX 5070 Ti (16 GB VRAM).

GPU time-slicing allows Ollama and ComfyUI to run simultaneously by advertising 2 virtual GPU resources from 1 physical GPU.
