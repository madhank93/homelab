# AI/ML Stack

The homelab AI stack runs on an NVIDIA RTX 5070 Ti GPU on `k8s-worker4`. It consists of Ollama for LLM inference and ComfyUI for image generation.

For full details on how the GPU was set up on Talos Linux, see `docs/nvidia-gpu-talos.md`.

---

## Table of Contents

- [Hardware](#hardware)
- [Ollama](#ollama)
- [ComfyUI](#comfyui)
- [Ollama vs ComfyUI](#ollama-vs-comfyui)
- [Note on Sharing the GPU](#note-on-sharing-the-gpu)

---

## Hardware

| Property | Value |
|----------|-------|
| GPU | NVIDIA RTX 5070 Ti |
| VRAM | 16 GB GDDR7 |
| Driver | 570.211.01 |
| CUDA | 12.8 |
| Node | k8s-worker4 (`192.168.1.224`) |
| NFD label | `nvidia.com/gpu.present=true` |

The GPU is detected by the NVIDIA Node Feature Discovery (NFD) component bundled with the GPU operator. NFD scans hardware and applies labels like `nvidia.com/gpu.present=true` and `nvidia.com/gpu.product=NVIDIA-GeForce-RTX-5070-Ti` to the node. All AI workloads use `nodeSelector: nvidia.com/gpu.present: "true"` to schedule on this node.

---

## Ollama

| Property | Value |
|----------|-------|
| File | `platform/cdk8s/cots/ai/ollama.go` |
| Namespace | `ollama` |
| Helm chart | `ollama` v1.41.0 (otwld.github.io/ollama-helm) |
| Image | `ollama/ollama:0.17.1` |
| External URL | http://ollama.madhan.app |
| Cluster service | `ollama.ollama.svc.cluster.local:11434` |
| Storage | 100 Gi Longhorn PVC (model cache) |

### Resource Limits

| Resource | Request | Limit |
|----------|---------|-------|
| CPU | 1000m | 4000m |
| RAM (host) | 2 Gi | 4 Gi |
| GPU (VRAM) | — | All 16 GB (via `nvidia.com/gpu: 1`) |

Note: The RAM limit is for host memory (cgroup). GPU VRAM is separate and fully available via `nvidia.com/gpu: 1`.

### GPU Configuration

```yaml
runtimeClassName: nvidia
nodeSelector:
  nvidia.com/gpu.present: "true"
env:
  - name: NVIDIA_VISIBLE_DEVICES
    value: all
resources:
  limits:
    nvidia.com/gpu: 1
```

### Ollama API Usage

```bash
# Health check
curl http://ollama.madhan.app/

# List loaded/available models
curl http://ollama.madhan.app/api/tags

# Pull a model
curl http://ollama.madhan.app/api/pull -d '{"name":"llama3.2"}'

# Generate (non-streaming)
curl http://ollama.madhan.app/api/generate -d '{
  "model": "llama3.2",
  "prompt": "Hello!",
  "stream": false
}'

# Chat
curl http://ollama.madhan.app/api/chat -d '{
  "model": "llama3.2",
  "messages": [{"role":"user","content":"Hello!"}]
}'

# OpenAI-compatible endpoint (for compatible clients)
curl http://ollama.madhan.app/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "llama3.2",
    "messages": [{"role":"user","content":"Hello!"}]
  }'
```

### Vision / Multimodal Models (Image Understanding)

Ollama supports multimodal models that can describe and analyze images provided as input:

```bash
ollama pull llama3.2-vision

curl http://ollama.madhan.app/api/chat -d '{
  "model": "llama3.2-vision",
  "messages": [{
    "role": "user",
    "content": "What is in this image?",
    "images": ["<base64-encoded-image>"]
  }]
}'
```

This is image *understanding* (the model reads an image), not image *generation*.

### Ollama Image Generation (Early 2026, Experimental)

As of early 2026, Ollama has experimental support for image generation models on Linux:

- Requires specific image models (e.g., `x/flux2-klein`, `x/z-image-turbo`)
- Pull and run the same as text models (`ollama pull x/flux2-klein`)
- Linux support is rolling out alongside the initial macOS-first release
- For production image generation workflows, ComfyUI (below) is the more mature option

---

## ComfyUI

| Property | Value |
|----------|-------|
| File | `platform/cdk8s/cots/ai/comfyui.go` |
| Namespace | `comfyui` |
| Image | `yanwk/comfyui-boot:latest-cu128` |
| External URL | http://comfyui.madhan.app |
| Storage | 100 Gi Longhorn PVC (models, outputs, custom nodes) |
| Start flags | `--listen 0.0.0.0 --port 8188` |

### Resource Limits

| Resource | Request | Limit |
|----------|---------|-------|
| CPU | 1000m | 4000m |
| RAM (host) | 4 Gi | 8 Gi |
| GPU (VRAM) | — | All 16 GB (via `nvidia.com/gpu: 1`) |

### GPU Configuration

```yaml
runtimeClassName: nvidia
nodeSelector:
  nvidia.com/gpu.present: "true"
strategy:
  type: Recreate   # ensures only one GPU pod at a time
env:
  - name: NVIDIA_VISIBLE_DEVICES
    value: all
resources:
  limits:
    nvidia.com/gpu: 1
```

The `Recreate` deployment strategy ensures the old pod is fully terminated before the new one starts. This is important for GPU workloads since only one pod can hold `nvidia.com/gpu: 1` at a time on a given node.

### ComfyUI Usage

1. Navigate to http://comfyui.madhan.app
2. Use the node-based workflow editor to build image generation pipelines
3. Download models to `/home/user/opt/ComfyUI/models/` (persisted on the Longhorn PVC)
4. Custom nodes can be installed to `/home/user/opt/ComfyUI/custom_nodes/`

The `cu128` image tag matches CUDA 12.8, which aligns with the RTX 5070 Ti driver (570.x series).

### Persistent Storage Layout

The 100 Gi Longhorn PVC is mounted at `/home/user/opt/ComfyUI/`:

```
/home/user/opt/ComfyUI/
├── models/
│   ├── checkpoints/      # SD/SDXL/Flux model files (.safetensors)
│   ├── vae/
│   ├── loras/
│   └── ...
├── custom_nodes/         # Community extensions
├── output/               # Generated images
└── input/                # Input images for img2img workflows
```

---

## Ollama vs ComfyUI

| | Ollama | ComfyUI |
|---|---|---|
| Purpose | LLM text generation, vision understanding | Image generation (Stable Diffusion, Flux) |
| Model types | Transformer LLMs (llama, mistral, deepseek, etc.) | Diffusion models (SD 1.5, SDXL, Flux) |
| Image input | Yes (llama3.2-vision, llava, etc.) | No |
| Image output | Experimental (early 2026) | Yes |
| API | REST at port 11434 | Web UI + REST at port 8188 |
| External URL | http://ollama.madhan.app | http://comfyui.madhan.app |

---

## Note on Sharing the GPU

Both Ollama and ComfyUI request `nvidia.com/gpu: 1`. Kubernetes tracks GPU allocation as a device count — only one pod can hold the single RTX 5070 Ti at a time on k8s-worker4.

**Scheduling behavior**:
- If Ollama is running, ComfyUI's pod will be `Pending` (insufficient GPU resources)
- If ComfyUI is running, Ollama's pod will be `Pending`
- Scale the idle deployment to 0 to free the GPU: `kubectl scale deploy/ollama -n ollama --replicas=0`

**VRAM capacity**:
- 16 GB is sufficient for most LLM models up to ~13B parameters (4-bit quantized) and most Flux/SDXL models
- Running both simultaneously is not possible without MIG (Multi-Instance GPU), which the RTX 5070 Ti does not support (consumer GPU)
