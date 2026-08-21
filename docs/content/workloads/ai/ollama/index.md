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

Ollama runs on k8s-worker4 (the GPU node, `192.168.1.224`) using the RTX 5070 Ti for inference. It stores downloaded model files on a 40 Gi Longhorn PVC.

Source: {{ src(path="workloads/ai/ollama.go") }}

## Configuration

| Setting | Value | Why |
|---------|-------|-----|
| Namespace | `ollama` | Isolated namespace |
| Image | `ollama/ollama` | Tag deliberately unset — tracks the chart's appVersion |
| HTTPRoute | `ollama.madhan.app` → `ollama:11434` | Gateway API |
| Service type | `LoadBalancer` | Also gets an IP from the Cilium L2 pool, for clients that cannot use the hostname |
| Model PVC | `40Gi` on `longhorn-model-cache` | Pulled models survive pod restarts |
| PVC replicas | `1`, `dataLocality: strict-local` | Model blobs are a re-pullable cache; the single replica sits on the GPU node's own disk |
| `runtimeClassName` | `nvidia` | Routes through nvidia-container-runtime |
| `NVIDIA_VISIBLE_DEVICES` | `all` | Make all GPU devices visible |
| `nvidia.com/gpu` limit | `1` | One time-sliced virtual GPU |
| Node selector | `nvidia.com/gpu.present: "true"` | Schedule on GPU node |
| Toleration | `dedicated=ai:NoSchedule` | **Required** — worker4 is tainted, see [GPU](@/hardware/gpu/index.md#the-dedicated-ai-taint) |
| CPU limit | `4000m` | Ollama + ComfyUI both CPU-hungry at inference |
| RAM request | `4Gi` | Host RAM for model metadata + process |
| RAM limit | `12Gi` | The executor's 13.8 GB GGUF is mmap'd on load and charged to this cgroup. ComfyUI rests at 0 replicas, so worker4's ~15.1Gi allocatable absorbs it |

> **Note:** `memory` here is the host RAM cgroup limit, **not** GPU VRAM. Nothing in
> Kubernetes limits VRAM — `nvidia.com/gpu: 1` grants one time-sliced share of the
> device, not a slice of its memory. See [GPU](@/hardware/gpu/index.md#sharing-the-16-gb-of-vram).

## API Usage

```bash
# List available models
curl https://ollama.madhan.app/api/tags

# Run inference (streaming)
curl https://ollama.madhan.app/api/generate \
  -d '{"model": "llama3.2", "prompt": "Hello!"}'

# Pull a new model (stores to 100Gi PVC)
curl https://ollama.madhan.app/api/pull \
  -d '{"name": "mistral"}'

# OpenAI-compatible chat completions
curl https://ollama.madhan.app/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model": "llama3.2", "messages": [{"role": "user", "content": "Hello"}]}'
```

## Coexistence with ComfyUI

Ollama and ComfyUI share the RTX 5070 Ti via time-slicing (5 virtual GPUs from 1
physical). VRAM is not partitioned, and ComfyUI does not release it after a run —
which is why ComfyUI ships scaled to zero. See
[GPU](@/hardware/gpu/index.md#sharing-the-16-gb-of-vram).

If running a large model (70B quantized, ~40 GB VRAM) alongside ComfyUI, VRAM exhaustion will occur. Use smaller quantized models or stop ComfyUI first.

## Declarative model pull

The otwld chart pulls models on boot, so a fresh pod arrives with the model
already present rather than waiting on a manual `/api/pull`:

```go
"ollama": map[string]any{
    "models": map[string]any{"pull": []string{executorBaseModel}},
},
```

It is idempotent and writes into the PVC, so it survives pod restarts. Add
models by extending that list rather than pulling them by hand — a hand pulled
model is lost the moment the PVC is recreated.

The list can only *pull* published models. The `executor` model that the aider
offload workflow targets is an `ollama create` layered on these base weights,
and its Modelfile lives in the `local-ai-setup` repo — see
[Local model offload](#local-model-offload) below.

Ollama will fail to load a model while ComfyUI is holding VRAM. If a pull or a
first inference fails for no clear reason, check whether ComfyUI is running —
see [GPU](@/hardware/gpu/index.md#sharing-the-16-gb-of-vram).

## How It Connects

```
Browser / API client → ollama.madhan.app
  → homelab-gateway → ollama:11434
  → Ollama pod on k8s-worker4
  → nvidia-container-runtime (GPU injection)
  → RTX 5070 Ti (VRAM for model weights)
  → 40Gi Longhorn PVC (model file storage)
```

## Local model offload

`executor` is the model that aider targets from a laptop (`ollama/executor`,
configured in `local-ai-setup`). It is not on the chart's pull list because it
is built, not pulled:

```bash
# from the local-ai-setup repo, with OLLAMA_API_BASE pointing at this cluster
./ollama/sync-models.sh --force
```

That runs `ollama create executor` against the base weights already in the PVC.
It is a one-time step per volume — the built model persists like any other.

## Troubleshooting

### Models Disappear After a Pod Restart

**Symptoms:** `/api/tags` returns an empty list, or only the models named in the
chart's `pull:` list. Anything built with `ollama create` is gone.

The model directory is not on a PVC. The chart's key is `persistentVolume`, and
an unrecognised key such as `persistence` is silently dropped — the chart then
falls back to its default `emptyDir`, which is discarded with the pod.

```bash
# Must return a bound PVC, not "No resources found"
kubectl -n ollama get pvc

# Must be a PVC, not emptyDir
kubectl -n ollama get deploy ollama -o jsonpath='{.spec.template.spec.volumes}'
```

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
