+++
title = "Ollama"
description = "LLM inference server on RTX 5070 Ti GPU."
weight = 20
+++

## What is Ollama?

[Ollama](https://ollama.com/) is an open-source LLM inference server that makes it easy to run large language models locally. It provides an OpenAI-compatible REST API and manages model downloads, GPU memory loading, and inference serving in a single binary.

## Why Ollama?

Ollama is the simplest way to run local LLMs on a GPU. It handles model quantization, GPU memory management, and serves an OpenAI-compatible API that works with any client that supports the OpenAI SDK (Python, TypeScript, etc.). The Helm chart provides native Kubernetes integration.

## How It's Used Here

Ollama runs on k8s-worker4 (the GPU node, `192.168.1.224`) using the RTX 5070 Ti for inference. It stores downloaded model files on a 100 Gi Longhorn PVC.

Source: [`workloads/ai/ollama.go`](https://github.com/madhank93/homelab/blob/v0.1.5/workloads/ai/ollama.go)

## Configuration

| Setting | Value | Why |
|---------|-------|-----|
| Namespace | `ollama` | Isolated namespace |
| Image | `ollama/ollama:0.17.0` | Pinned version |
| HTTPRoute | `ollama.madhan.app` → `ollama:11434` | Gateway API |
| `runtimeClassName` | `nvidia` | Routes through nvidia-container-runtime |
| `NVIDIA_VISIBLE_DEVICES` | `all` | Make all GPU devices visible |
| `nvidia.com/gpu` limit | `1` | One time-sliced virtual GPU |
| Node selector | `nvidia.com/gpu.present: "true"` | Schedule on GPU node |
| Toleration | `dedicated=ai:NoSchedule` | Allow scheduling on tainted GPU node |
| CPU limit | `4000m` | Ollama + ComfyUI both CPU-hungry at inference |
| RAM request | `2Gi` | Host RAM for model metadata + process |
| RAM limit | `4Gi` | Keep below worker4's 16 GiB total (shared with ComfyUI) |
| Model PVC | `100Gi` Longhorn | Stores downloaded model weights |

> **Note:** `memory: 4Gi` is the host RAM cgroup limit, NOT GPU VRAM. GPU VRAM (16 GB) is controlled by the `nvidia.com/gpu: 1` resource limit and is shared with ComfyUI via time-slicing.

## API Usage

```bash
# List available models
curl http://ollama.madhan.app/api/tags

# Run inference (streaming)
curl http://ollama.madhan.app/api/generate \
  -d '{"model": "llama3.2", "prompt": "Hello!"}'

# Pull a new model (stores to 100Gi PVC)
curl http://ollama.madhan.app/api/pull \
  -d '{"name": "mistral"}'

# OpenAI-compatible chat completions
curl http://ollama.madhan.app/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model": "llama3.2", "messages": [{"role": "user", "content": "Hello"}]}'
```

## Coexistence with ComfyUI

Ollama and ComfyUI share the RTX 5070 Ti via GPU time-slicing (2 virtual GPUs from 1 physical). VRAM is not isolated — approximately 4 GB (Ollama 7B model) + 6 GB (ComfyUI SDXL) fits within the 16 GB pool.

If running a large model (70B quantized, ~40 GB VRAM) alongside ComfyUI, VRAM exhaustion will occur. Use smaller quantized models or stop ComfyUI first.

## How It Connects

```
Browser / API client → ollama.madhan.app
  → homelab-gateway → ollama:11434
  → Ollama pod on k8s-worker4
  → nvidia-container-runtime (GPU injection)
  → RTX 5070 Ti (VRAM for model weights)
  → 100Gi Longhorn PVC (model file storage)
```

## Troubleshooting

### Model Pull Fails / OOM

**Symptoms:** Model pull command hangs or returns OOM error.

```bash
# Check disk space on PVC
kubectl exec -n ollama -l app.kubernetes.io/name=ollama -- df -h /root/.ollama

# Check GPU memory
kubectl exec -n ollama -l app.kubernetes.io/name=ollama -- nvidia-smi
```

### GPU Not Available in Pod

```bash
# Verify pod spec
kubectl get pod -n ollama -o yaml | grep -E "runtimeClass|nvidia|dedicated"

# Check node advertises GPU
kubectl describe node k8s-worker4 | grep "nvidia.com/gpu"
```

### Inference Too Slow

At idle, the RTX 5070 Ti uses 0dB fan mode and is essentially cold. First inference after idle warms up the GPU. Subsequent inferences are faster.
