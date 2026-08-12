+++
title = "Software Inventory"
description = "Every pinned version in the homelab and the file that pins it."
weight = 50
+++

Every version below is pinned in code — there are no floating tags. When a
number here disagrees with the cluster, the code wins and the cluster is drifting.

## Infrastructure

| Software | Version | Managed by | Pinned in |
|---|---|---|---|
| Talos Linux | v1.13.8 | Pulumi | [`core/platform/talos.go`](https://github.com/madhank93/homelab/blob/v0.1.7/core/platform/talos.go) |
| Kubernetes | v1.36.0 | Talos | [`core/platform/talos.go`](https://github.com/madhank93/homelab/blob/v0.1.7/core/platform/talos.go) |

Kubernetes does not ride along with a Talos upgrade — see the
[Upgrade Guide](/upgrade-guide).

## Bifrost (Hetzner VPS)

Docker Compose on a single public VPS. Everything reachable from the internet
terminates here.

| Software | Version | Pinned in |
|---|---|---|
| Traefik | v3.7.10 | [`docker-compose.yml`](https://github.com/madhank93/homelab/blob/v0.1.7/core/cloud/bifrost/docker-compose.yml) |
| NetBird (server, agent, proxy) | 0.76.3 | same |
| NetBird dashboard | v2.90.10 | same |
| Authentik | 2026.5.6 | same |
| PostgreSQL (Authentik) | 16.14-alpine | same |
| Gatus (uptime) | v5.36.0 | same |

## Platform

Installed by Pulumi, not GitOps — these bootstrap the cluster that runs
everything else. Applied with `just core platform up`.

| Software | Version | Pinned in |
|---|---|---|
| Cilium | 1.18.12 | [`core/platform/cilium.go`](https://github.com/madhank93/homelab/blob/v0.1.7/core/platform/cilium.go) |
| Argo CD (chart) | 10.3.2 | [`core/platform/argocd.go`](https://github.com/madhank93/homelab/blob/v0.1.7/core/platform/argocd.go) |
| Argo CD (image) | v3.5.1 | same — runs ahead of the chart, see [Platform](/platform) |
| cert-manager | v1.21.1 | [`core/platform/cert_manager.go`](https://github.com/madhank93/homelab/blob/v0.1.7/core/platform/cert_manager.go) |

## Workloads

Synthesized by CDK8s and delivered by Argo CD. Charts with a generated typed
package are pinned in [`workloads/cdk8s.yaml`](https://github.com/madhank93/homelab/blob/v0.1.7/workloads/cdk8s.yaml);
the rest are pinned at their call site.

### Storage and databases

| Software | Chart | Pinned in |
|---|---|---|
| Longhorn | 1.12.0 | `cdk8s.yaml` |
| CloudNativePG | 0.29.0 | `workloads/databases/cnpg.go` |

### Secrets

| Software | Version | Pinned in |
|---|---|---|
| OpenBao (chart) | 0.29.0 | `workloads/secrets/openbao.go` |
| OpenBao (image) | 2.6.1 | same — unseal sidecar |
| Secrets Store CSI Driver | 1.6.0 | `workloads/secrets/csi_driver.go` |

### Observability

| Software | Chart | Pinned in |
|---|---|---|
| VictoriaMetrics k8s-stack | 0.72.4 | `workloads/observability/victoria_metrics.go` |
| VictoriaLogs | 0.13.9 | `cdk8s.yaml` |
| Grafana | 12.10.4 | `cdk8s.yaml` |
| OpenTelemetry Collector | 0.169.0 | `workloads/observability/otel_collector.go` |
| Metrics Server | 3.13.1 | `cdk8s.yaml` |

### Security

| Software | Chart | Pinned in |
|---|---|---|
| Falco | 8.0.5 | `workloads/security/falco.go` |
| Kyverno | 3.8.2 | `cdk8s.yaml` |
| Trivy Operator | 0.35.0 | `cdk8s.yaml` |

### Applications

| Software | Version | Pinned in |
|---|---|---|
| Harbor | 1.19.2 | `cdk8s.yaml` |
| Headlamp | 0.44.0 | `cdk8s.yaml` |
| n8n (chart) | 2.0.1 | `workloads/automation/n8n.go` |
| n8n (image) | 1.78.0 | same |
| Reloader | 2.2.16 | `workloads/support/reloader.go` |
| NetBird peer | 0.76.3 | `workloads/networking/netbird_peer.go` |

### AI and GPU

| Software | Version | Pinned in |
|---|---|---|
| Ollama | 1.74.0 | `cdk8s.yaml` |
| NVIDIA Device Plugin | 0.19.3 | `workloads/hardware/nvidia_gpu_operator.go` |
| DCGM Exporter | 4.8.3 | same |
| ComfyUI | cu128-megapak-20260223 | `workloads/ai/comfyui.go` |

## Checking for updates

`helm search repo` compares **chart** versions. A chart can lag the software it
ships, so check the app version too — Argo CD 3.5.1 was released while chart
10.3.2 still pinned 3.5.0.

```bash
helm repo update
helm search repo <repo>/<chart> --versions | head -3   # CHART and APP columns
```
