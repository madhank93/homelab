+++
title = "Apps"
description = "All platform applications deployed via CDK8s and ArgoCD."
weight = 60
sort_by = "weight"
+++

All applications are defined as CDK8s Go code in `platform/cdk8s/cots/`, synthesized to YAML by CI, and deployed by ArgoCD from the `v0.1.5-manifests` branch.

## App Catalog

| App | Namespace | URL | UI | Purpose |
|-----|-----------|-----|----|---------|
| ComfyUI | `comfyui` | http://comfyui.madhan.app | Yes | Stable Diffusion / Flux image generation |
| Ollama | `ollama` | http://ollama.madhan.app | No (REST) | LLM inference server |
| NVIDIA GPU Operator | `nvidia-gpu-operator` | — | No | GPU device plugin + NFD |
| n8n | `n8n` | http://n8n.madhan.app | Yes | Workflow automation |
| Grafana | `grafana` | http://grafana.madhan.app | Yes | Dashboards |
| VictoriaMetrics | `victoria-metrics` | — | No | Metrics storage |
| VictoriaLogs | `victoria-logs` | — | No | Log storage |
| AlertManager | `alertmanager` | — | No | Alert routing |
| OpenTelemetry | `opentelemetry` | — | No | Metrics + log collection |
| Falco | `falco` | — | No | Runtime syscall security |
| Trivy | `trivy` | — | No | Vulnerability scanning |
| Rancher | `cattle-system` | http://rancher.madhan.app | Yes | Multi-cluster management |
| Headlamp | `headlamp` | http://headlamp.madhan.app | Yes | Kubernetes dashboard |
| Infisical | `infisical` | http://infisical.madhan.app | Yes | Secrets management |
| Harbor | `harbor` | http://harbor.madhan.app | Yes | Container registry |
| n8n | `n8n` | http://n8n.madhan.app | Yes | Workflow automation |
| Longhorn | `longhorn-system` | — | No | Distributed block storage |

## Apps Requiring Infisical

These apps use `InfisicalSecret` CRDs. They will not start correctly until the `infisical-service-token` Secret exists in the `infisical` namespace:

| App | Infisical Path | k8s Secret | Keys |
|-----|---------------|------------|------|
| Grafana | `/grafana` | `grafana-admin` | `ADMIN_PASSWORD` |
| Harbor | `/harbor` | `harbor-admin` | `HARBOR_ADMIN_PASSWORD` |
| n8n | `/n8n` | `n8n-db` | `DB_PASSWORD` |
| Rancher | `/rancher` | `rancher-bootstrap` | `BOOTSTRAP_PASSWORD` |
