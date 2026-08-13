+++
title = "Workloads"
description = "All workloads deployed to the cluster via CDK8s and ArgoCD."
weight = 60
sort_by = "weight"
+++

All applications are defined as CDK8s Go code in `workloads/`, synthesized to YAML by CI, and deployed by ArgoCD from the `v0.1.7-manifests` branch.

## App Catalog

One row per chart in {{ src(path="workloads/main.go") }} — 23 of them.

| App | Namespace | URL | UI | Purpose |
|-----|-----------|-----|----|---------|
| ComfyUI | `comfyui` | https://comfyui.madhan.app | Yes | Image generation — **off by default**, see [ComfyUI](@/workloads/ai/comfyui/index.md) |
| Ollama | `ollama` | https://ollama.madhan.app | No (REST) | LLM inference server |
| Kubeflow | `kubeflow` | https://kubeflow.madhan.app | Yes | Notebooks and pipelines |
| notebook-gateway-controller | `kubeflow` | — | No | Creates HTTPRoutes for Kubeflow notebooks |
| NVIDIA Device Plugin | `nvidia-gpu-operator` | — | No | GPU device plugin + NFD |
| n8n | `n8n` | https://n8n.madhan.app | Yes | Workflow automation |
| CloudNativePG | `cnpg-system` | — | No | PostgreSQL operator |
| Grafana | `grafana` | https://grafana.madhan.app | Yes | Dashboards |
| VictoriaMetrics | `victoria-metrics` | https://vmselect.madhan.app | Yes (vmui) | Metrics storage + Alertmanager |
| VictoriaLogs | `victoria-logs` | https://victorialogs.madhan.app | Yes | Log storage |
| OpenTelemetry | `opentelemetry` | — | No | Metrics + log collection |
| Metrics Server | `kube-system` | — | No | `kubectl top` and HPA |
| argocd-monitor | `argocd` | — | No | ServiceMonitors for Argo CD components |
| Falco | `falco` | https://falco.madhan.app | Yes (sidekick-ui) | Runtime syscall security |
| Trivy | `trivy` | — | No | Vulnerability scanning |
| Kyverno | `kyverno` | — | No | Admission control and background scans |
| Headlamp | `headlamp` | https://headlamp.madhan.app | Yes | Kubernetes dashboard |
| OpenBao | `openbao` | https://openbao.madhan.app | Yes | Secrets management |
| Secrets Store CSI Driver | `kube-system` | — | No | Mounts OpenBao secrets into pods |
| Harbor | `harbor` | https://harbor.madhan.app | Yes | Container registry |
| Longhorn | `longhorn-system` | https://longhorn.madhan.app | Yes | Distributed block storage |
| NetBird peer | `netbird` | — | No | In-cluster WireGuard routing peer |
| Reloader | `reloader` | — | No | Auto-reload pods on ConfigMap/Secret changes |

## Runtime Secrets (OpenBao + CSI Driver)

All apps source their runtime secrets from OpenBao via the Secrets Store CSI Driver. CDK8s generates zero `Secret` resources.

| App | OpenBao Path | Pattern | k8s Secret created |
|-----|-------------|---------|-------------------|
| Grafana | `secret/data/grafana` | B (secretObjects) | `grafana-oauth-secret` |
| Harbor | `secret/data/harbor` | B (secretObjects) | `harbor-admin` |
| n8n | `secret/data/n8n` | B (secretObjects) | `n8n-secrets` |
| NetBird | `secret/data/netbird` | B (secretObjects) | `netbird-setup-key` |

Grafana uses both halves: its admin password is file-only, while the OIDC client
secret has to be a k8s Secret because Grafana reads it as a `GF_` env var.

The two patterns are explained once in [Secrets](@/platform/secrets/index.md).
