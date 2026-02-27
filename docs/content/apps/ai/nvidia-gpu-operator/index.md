+++
title = "NVIDIA GPU Operator"
description = "GPU Operator ClusterPolicy, Node Feature Discovery, and time-slicing configuration."
weight = 10
+++

## Overview

| Property | Value |
|----------|-------|
| CDK8s file | `platform/cdk8s/cots/ai/nvidia_gpu_operator.go` |
| Namespace | `nvidia-gpu-operator` |
| Helm chart | `gpu-operator` v25.10.1 (helm.ngc.nvidia.com/nvidia) |
| HTTPRoute | None |
| UI | No |

## Purpose

Manages NVIDIA GPU access within Kubernetes. On Talos Linux, the GPU driver and container toolkit are provided by Talos system extensions — the operator's role is:

- **Node Feature Discovery (NFD):** labels GPU nodes with `nvidia.com/gpu.present=true`
- **Device plugin:** advertises `nvidia.com/gpu` as a schedulable Kubernetes resource
- **DCGM exporter:** exports GPU metrics (temperature, utilization, memory) to VictoriaMetrics

## Talos-Specific Configuration

```yaml
driver:
  enabled: true   # operator detects Talos extensions → sets DESIRED=0 (driver pre-installed)
toolkit:
  enabled: false  # Talos extension provides the container toolkit
```

`DEVICE_LIST_STRATEGY=envvar` is set to avoid CDI hostPath issues on Talos (see `docs/nvidia-gpu-talos.md`).

A **Talos Validation Bridge DaemonSet** writes marker files that unblock GPU operator device plugin startup on Talos.

## GPU Time-Slicing

A `time-slicing-config` ConfigMap configures 2 virtual GPU replicas per physical GPU:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: time-slicing-config
  namespace: nvidia-gpu-operator
data:
  any: |-
    version: v1
    flags:
      migStrategy: none
    sharing:
      timeSlicing:
        resources:
          - name: nvidia.com/gpu
            replicas: 2
```

This is referenced in the ClusterPolicy via `devicePlugin.config.name: time-slicing-config`.

After applying, `kubectl describe node k8s-worker4` shows:
```
Allocatable:
  nvidia.com/gpu: 2
```

Both Ollama and ComfyUI can now request `nvidia.com/gpu: 1` simultaneously. VRAM is shared (not partitioned).
