+++
title = "Workloads"
description = "All workloads deployed to the cluster via CDK8s and ArgoCD."
weight = 60
sort_by = "weight"
+++

All applications are defined as CDK8s Go code in `workloads/`, synthesized to YAML by CI, and deployed by ArgoCD from the `v0.1.5-manifests` branch.

## App Catalog

| App | Namespace | URL | UI | Purpose |
|-----|-----------|-----|----|---------|
| ComfyUI | `comfyui` | http://comfyui.madhan.app | Yes | Stable Diffusion / Flux image generation |
| Ollama | `ollama` | http://ollama.madhan.app | No (REST) | LLM inference server |
| NVIDIA Device Plugin | `nvidia-gpu-operator` | — | No | GPU device plugin + NFD |
| n8n | `n8n` | http://n8n.madhan.app | Yes | Workflow automation |
| Grafana | `grafana` | http://grafana.madhan.app | Yes | Dashboards |
| VictoriaMetrics | `victoria-metrics` | http://vmselect.madhan.app | Yes (vmui) | Metrics storage |
| VictoriaLogs | `victoria-logs` | http://victorialogs.madhan.app | Yes | Log storage |
| AlertManager | `alertmanager` | http://alertmanager.madhan.app | Yes | Alert routing |
| OpenTelemetry | `opentelemetry` | — | No | Metrics + log collection |
| Falco | `falco` | http://falco.madhan.app | Yes (sidekick-ui) | Runtime syscall security |
| Trivy | `trivy` | — | No | Vulnerability scanning |
| Rancher | `cattle-system` | http://rancher.madhan.app | Yes | Multi-cluster management |
| Headlamp | `headlamp` | http://headlamp.madhan.app | Yes | Kubernetes dashboard |
| OpenBao | `openbao` | http://openbao.madhan.app | Yes | Secrets management |
| Harbor | `harbor` | http://harbor.madhan.app | Yes | Container registry |
| Longhorn | `longhorn-system` | http://longhorn.madhan.app | Yes | Distributed block storage |
| Reloader | `reloader` | — | No | Auto-reload pods on ConfigMap/Secret changes |

## Runtime Secrets (OpenBao + CSI Driver)

All apps source their runtime secrets from OpenBao via the Secrets Store CSI Driver. CDK8s generates zero `Secret` resources.

| App | OpenBao Path | Pattern | k8s Secret created |
|-----|-------------|---------|-------------------|
| Grafana | `secret/data/grafana` | A (file-only) | none |
| Harbor | `secret/data/harbor` | B (secretObjects) | `harbor-admin` |
| n8n | `secret/data/n8n` | B (secretObjects) | `n8n-secrets` |
| Rancher | `secret/data/rancher` | B (secretObjects) | `rancher-bootstrap` |
| NetBird | `secret/data/netbird` | B (secretObjects) | `netbird-setup-key` |

> Pattern A mounts secrets as files only. Pattern B also syncs a k8s Secret so Helm charts that require `existingSecret` can reference it.
