+++
title = "Ollama"
description = "LLM inference server on RTX 5070 Ti GPU."
weight = 20
+++

## Overview

| Property | Value |
|----------|-------|
| CDK8s file | `platform/cdk8s/cots/ai/ollama.go` |
| Namespace | `ollama` |
| Helm chart | `ollama` v1.41.0 (otwld.github.io/ollama-helm) |
| HTTPRoute | `ollama.madhan.app` → `ollama:11434` |
| UI | No (REST API only) |
| Node | k8s-worker4 (GPU) |

## Purpose

Runs open-source LLMs locally using the NVIDIA RTX 5070 Ti. Exposes an OpenAI-compatible REST API at `http://ollama.madhan.app`.

Supported models include: llama3.2, mistral, deepseek-r1, and any model in the [Ollama library](https://ollama.com/library).

## GPU Configuration

```yaml
runtimeClassName: nvidia
resources:
  requests:
    memory: "2Gi"
  limits:
    nvidia.com/gpu: "1"
env:
  - name: NVIDIA_VISIBLE_DEVICES
    value: all
```

## API Usage

```bash
# List available models
curl http://ollama.madhan.app/api/tags

# Run inference (streaming)
curl http://ollama.madhan.app/api/generate \
  -d '{"model": "llama3.2", "prompt": "Hello!"}'

# Pull a new model
curl http://ollama.madhan.app/api/pull \
  -d '{"name": "mistral"}'
```

## Node Scheduling

The pod is scheduled to k8s-worker4 via a `nodeSelector` or `nodeName` constraint targeting the GPU node.

## Coexistence with ComfyUI

Ollama (2 Gi RAM) and ComfyUI (1 Gi RAM request) run simultaneously on the same node via GPU time-slicing. Total RAM on node is ~5.4 Gi allocatable — the combined 3 Gi fits within limits.
