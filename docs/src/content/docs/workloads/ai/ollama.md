---
title: Ollama
description: LLM inference on the shared GPU, and the endpoint the laptop's aider offload workflow talks to.
sidebar:
  order: 20
  badge: { text: GPU, variant: tip }
---

Ollama serves LLMs from the RTX 5070 Ti; its main client is aider on the laptop (`ollama/executor`). [Upstream docs](https://github.com/ollama/ollama/tree/main/docs).

| | |
|---|---|
| Namespace | `ollama` |
| Source | [`ollama.go`](src:workloads/ai/ollama.go) |
| Exposure | LAN · `ollama.madhan.app`, plus LoadBalancer `192.168.1.225:11434` |
| Storage | 40Gi on `longhorn-model-cache` (1 replica, strict-local) |
| Auth | None |

## Configuration

| Setting | Value | Why |
|---|---|---|
| Image tag | unset | Tracks the chart's appVersion, so a chart bump carries the runtime |
| Service | `LoadBalancer`, `lbipam.cilium.io/ips: 192.168.1.225` | The laptop's `OLLAMA_API_BASE` points at this fixed IP |
| `persistentVolume` | 40Gi, `longhorn-model-cache` | Model blobs are a re-pullable cache: one local replica keeps ~20 GB off other disks and loads off the iSCSI path |
| Memory limit | `12Gi` (request `4Gi`) | Host RAM, not VRAM: the 13.8 GB executor GGUF is mmap'd and charged to this cgroup |
| CPU limit | `4000m` | |
| `OLLAMA_FLASH_ATTENTION` | `1` | Required for a quantised KV cache |
| `OLLAMA_KV_CACHE_TYPE` | `q8_0` | Halves the KV cache, which sets the largest prompt that fits in 16 GB |
| `ollama.models.pull` | the executor's base weights | Pulled into the PVC on boot, so a fresh pod has them |
| GPU scheduling | `runtimeClassName: nvidia`, `nvidia.com/gpu: 1`, GPU node selector, `dedicated=ai` toleration | See [AI](/workloads/ai/) |

:::caution[The chart key is persistentVolume]
An unrecognised key such as `persistence` is silently dropped, and the model directory falls back to an `emptyDir` that loses every model on restart.
:::

## Operate

The `executor` model is built, not pulled: an `ollama create` over the base weights, with its Modelfile in the `local-ai-setup` repo. Build it once per volume:

```bash frame="terminal"
# from local-ai-setup, with OLLAMA_API_BASE pointing at 192.168.1.225
./ollama/sync-models.sh --force
```

Add other published models to the `pull` list in `ollama.go`; a model pulled by hand is lost when the PVC is recreated.

```bash frame="terminal"
curl https://ollama.madhan.app/api/tags                                       # list models
curl https://ollama.madhan.app/api/generate -d '{"model":"executor","keep_alive":0}'  # unload, freeing VRAM
```

## Troubleshooting

### Models disappear after a pod restart

The model directory is not on a PVC. Check that the PVC is bound and mounted:

```bash frame="terminal"
kubectl -n ollama get pvc
kubectl -n ollama get deploy ollama -o jsonpath='{.spec.template.spec.volumes}'
```

### A model fails to load for no clear reason

ComfyUI is probably holding VRAM. Run `just comfyui off`, then retry. Check with `kubectl exec -n ollama deploy/ollama -- nvidia-smi`.

### Pod has no GPU

Check the pod has `runtimeClassName: nvidia` and the toleration, and that the node advertises GPUs:

```bash frame="terminal"
kubectl describe node k8s-worker4 | grep nvidia.com/gpu
```
