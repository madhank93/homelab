---
title: Software inventory
description: Every pinned version in the homelab and the file that pins it. The only page that states versions.
sidebar:
  order: 2
---

Every version below is pinned in code; there are no floating tags. When a number here disagrees with the cluster, the code wins and the cluster is drifting.

This is the only page that carries version numbers. `just docs-check` re-reads every row, greps the file in the last column, and fails if the number is gone. Most Go code reads its pins from [`config/versions.yaml`](src:config/versions.yaml), which the check also accepts.

## Infrastructure

| Software | Version | Managed by | Pinned in |
|---|---|---|---|
| Talos Linux | v1.14.5 | Pulumi | [`versions.yaml`](src:config/versions.yaml) |
| Kubernetes | 1.36.6 | Talos machine config | [`versions.yaml`](src:config/versions.yaml) |
| Kubernetes API types | 1.36.0 | CDK8s | [`cdk8s.yaml`](src:workloads/cdk8s.yaml) |

Kubernetes is pinned separately from Talos, which would otherwise default to a newer minor than Cilium supports; `just talos-upgrade-k8s` rolls it out — see [Talos upgrades](/operate/upgrades/talos/). The CDK8s API types must stay on the cluster's minor. The NVIDIA driver comes from the Talos `-production` system extensions and moves with the Talos release.

## Bifrost (Hetzner VPS)

| Software | Version | Pinned in |
|---|---|---|
| Traefik | v3.7.16 | [`core/cloud/bifrost/docker-compose.yml`](src:core/cloud/bifrost/docker-compose.yml) |
| NetBird (server, agent, proxy) | 0.80.0 | same |
| NetBird dashboard | v2.94.0 | same |
| Authentik | 2026.8.5 | same |
| PostgreSQL (Authentik) | 16.15-alpine | same |
| Gatus (uptime) | v5.37.0 | same |

## Platform

Installed by Pulumi with `just core platform up`, not GitOps: these bootstrap the cluster that runs everything else.

| Software | Version | Pinned in |
|---|---|---|
| Cilium | 1.20.4 | [`core/platform/cilium.go`](src:core/platform/cilium.go) |
| Gateway API CRDs | v1.6.3 | [`core/platform/manifests/gateway-api-v1.6.3-experimental-install.yaml`](src:core/platform/manifests/gateway-api-v1.6.3-experimental-install.yaml) |
| Argo CD (chart) | 10.10.5 | [`core/platform/argocd.go`](src:core/platform/argocd.go) |
| Argo CD (image) | v3.5.6 | same — runs ahead of the chart, see [Argo CD upgrades](/operate/upgrades/argocd/) |
| cert-manager | v1.21.3 | [`core/platform/cert_manager.go`](src:core/platform/cert_manager.go) |

The Gateway API CRDs are vendored rather than fetched on every apply — see [Cilium](/platform/cilium/).

## Workloads

Synthesized by CDK8s and delivered by Argo CD. Charts with a generated typed package are pinned in [`cdk8s.yaml`](src:workloads/cdk8s.yaml); the rest are read from `config/versions.yaml` at their call site.

### Storage and databases

| Software | Chart | Pinned in |
|---|---|---|
| Longhorn | 1.13.0 | [`cdk8s.yaml`](src:workloads/cdk8s.yaml) |
| CloudNativePG | 0.29.1 | [`workloads/databases/cnpg.go`](src:workloads/databases/cnpg.go) |
| PostgreSQL (n8n-pg operand) | 18.6-system-trixie | [`versions.yaml`](src:config/versions.yaml) |

### Secrets

| Software | Version | Pinned in |
|---|---|---|
| OpenBao (chart) | 0.30.4 | [`workloads/secrets/openbao.go`](src:workloads/secrets/openbao.go) |
| OpenBao (image) | 2.6.5 | same — server, CSI agent and unseal sidecar; held on 2.6.x for the `file` storage backend |
| Secrets Store CSI Driver | 1.6.1 | [`workloads/secrets/csi_driver.go`](src:workloads/secrets/csi_driver.go) |

### Observability

| Software | Chart | Pinned in |
|---|---|---|
| VictoriaMetrics k8s-stack | 0.96.0 | [`workloads/observability/victoria_metrics.go`](src:workloads/observability/victoria_metrics.go) |
| VictoriaMetrics operator CRDs | 0.68.1 | same |
| kube-prometheus-stack CRDs | 82.0.1 | same |
| VictoriaLogs | 0.13.10 | [`cdk8s.yaml`](src:workloads/cdk8s.yaml) |
| Grafana | 14.0.1 | [`cdk8s.yaml`](src:workloads/cdk8s.yaml) |
| Grafana (image) | 13.2.5-distroless | [`versions.yaml`](src:config/versions.yaml) |
| OpenTelemetry Collector | 0.176.0 | [`workloads/observability/otel_collector.go`](src:workloads/observability/otel_collector.go) |
| OpenTelemetry Collector (image) | 0.162.0 | same |
| Metrics Server | 3.14.0 | [`cdk8s.yaml`](src:workloads/cdk8s.yaml) |

VMAlertmanager ships inside the VictoriaMetrics k8s-stack chart and has no separate pin.

### Security

| Software | Chart | Pinned in |
|---|---|---|
| Falco | 9.2.1 | [`workloads/security/falco.go`](src:workloads/security/falco.go) |
| Kyverno | 3.9.2 | [`cdk8s.yaml`](src:workloads/cdk8s.yaml) |
| Trivy Operator (chart) | 0.37.0 | [`cdk8s.yaml`](src:workloads/cdk8s.yaml) |
| Trivy Operator (CRD bundle) | 0.37.0 | [`workloads/security/trivy.go`](src:workloads/security/trivy.go) |

Trivy has two pins: the chart comes from the typed package, while the CRD bundle is downloaded by URL from the release named in `config/versions.yaml`. Bump both together.

### Applications

| Software | Version | Pinned in |
|---|---|---|
| Harbor | 1.19.4 | [`cdk8s.yaml`](src:workloads/cdk8s.yaml) |
| Headlamp | 0.45.0 | [`cdk8s.yaml`](src:workloads/cdk8s.yaml) |
| n8n (chart) | 2.1.1 | [`workloads/automation/n8n.go`](src:workloads/automation/n8n.go) |
| n8n (image) | 2.45.5 | same |
| Reloader | 2.2.19 | [`workloads/support/reloader.go`](src:workloads/support/reloader.go) |
| NetBird peer | 0.80.0 | [`workloads/networking/netbird_peer.go`](src:workloads/networking/netbird_peer.go) |
| Kubeflow | 26.03.1 | [`workloads/ai/kubeflow/kustomization.yaml`](src:workloads/ai/kubeflow/kustomization.yaml) |
| notebook-gateway-controller | v1 | [`workloads/ai/notebook_gateway_controller.go`](src:workloads/ai/notebook_gateway_controller.go) |

### AI and GPU

| Software | Version | Pinned in |
|---|---|---|
| Ollama (chart) | 1.88.0 | [`cdk8s.yaml`](src:workloads/cdk8s.yaml) |
| NVIDIA Device Plugin | 0.20.1 | [`workloads/hardware/nvidia_gpu_operator.go`](src:workloads/hardware/nvidia_gpu_operator.go) |
| DCGM Exporter | 4.8.4 | same |
| ComfyUI | cu130-megapak-pt211-20260812 | [`workloads/ai/comfyui.go`](src:workloads/ai/comfyui.go) |
| Open WebUI | v0.12.5 | [`config/versions.yaml`](src:config/versions.yaml) |

Ollama's image tag is unset so it tracks the chart's appVersion; the chart is the pin. ComfyUI's tag encodes its CUDA line: the card is Blackwell (sm_120) and needs CUDA 12.8 or newer.

### Gaming

| Software | Version | Pinned in |
|---|---|---|
| Counter-Strike 2 server | 5.0.1 | [`config/versions.yaml`](src:config/versions.yaml) |

## Checking for updates

`helm search repo` compares chart versions, and a chart can lag the software it ships, so check the APP column too — Argo CD 3.5.6 was released while chart 10.10.5 still pinned 3.5.5.

```bash frame="terminal"
helm repo update
helm search repo <repo>/<chart> --versions | head -3   # CHART and APP columns
```

Run `just docs-check` after any bump. The upgrade procedure is in [Upgrades](/operate/upgrades/).
