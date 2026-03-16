+++
title = "AI"
description = "GPU-accelerated AI workloads: Ollama, ComfyUI, Kubeflow, and NVIDIA GPU Operator."
weight = 10
sort_by = "weight"
+++

All AI workloads run on **k8s-worker4** (`192.168.1.224`), which has PCIe passthrough to an NVIDIA RTX 5070 Ti (16 GB VRAM, Blackwell sm_120).

GPU time-slicing advertises 5 virtual GPU resources from 1 physical GPU, allowing Ollama, ComfyUI, and Kubeflow notebooks/training jobs to run concurrently. VRAM is not isolated between processes.
