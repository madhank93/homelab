+++
title = "Software Inventory"
description = "Every pinned version in the homelab and the file that pins it."
weight = 50
+++

Every version below is pinned in code — there are no floating tags. When a
number here disagrees with the cluster, the code wins and the cluster is drifting.

This is the **only** page in these docs that carries version numbers. Every other
page describes what a component does and links here for which release of it is
running. `just docs-check` enforces both halves of that rule.

## Infrastructure

| Software | Version | Managed by | Pinned in |
|---|---|---|---|
| Talos Linux | v1.13.8 | Pulumi | {{ src(path="core/platform/talos.go") }} |
| Kubernetes API types | 1.36.0 | CDK8s | {{ src(path="workloads/cdk8s.yaml", label="cdk8s.yaml") }} |

Kubernetes itself has no pin of its own: the running version is whatever the
Talos release ships (v1.13.x ships 1.36.x), and it does **not** ride along with
a Talos upgrade — `talosctl upgrade-k8s` is a separate step, see the
[Upgrade Guide](@/upgrade-guide/index.md). The row above pins the API types
CDK8s generates Go structs from, which must stay aligned with the cluster.

## Bifrost (Hetzner VPS)

Docker Compose on a single public VPS. Everything reachable from the internet
terminates here.

| Software | Version | Pinned in |
|---|---|---|
| Traefik | v3.7.10 | {{ src(path="core/cloud/bifrost/docker-compose.yml") }} |
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
| Cilium | 1.18.12 | {{ src(path="core/platform/cilium.go") }} |
| Gateway API CRDs | v1.2.1 | {{ src(path="core/platform/manifests/gateway-api-v1.2.1-experimental-install.yaml") }} |
| Argo CD (chart) | 10.3.2 | {{ src(path="core/platform/argocd.go") }} |
| Argo CD (image) | v3.5.1 | same — runs ahead of the chart, see [Platform](@/platform/_index.md) |
| cert-manager | v1.21.1 | {{ src(path="core/platform/cert_manager.go") }} |

The Gateway API CRDs are vendored into the repo rather than fetched from GitHub
on every apply — see [Cilium](@/platform/networking/index.md#gateway-api-crds).

## Workloads

Synthesized by CDK8s and delivered by Argo CD. Charts with a generated typed
package are pinned in {{ src(path="workloads/cdk8s.yaml", label="cdk8s.yaml") }};
the rest are pinned at their call site.

### Storage and databases

| Software | Chart | Pinned in |
|---|---|---|
| Longhorn | 1.12.0 | {{ src(path="workloads/cdk8s.yaml", label="cdk8s.yaml") }} |
| CloudNativePG | 0.29.0 | {{ src(path="workloads/databases/cnpg.go") }} |

### Secrets

| Software | Version | Pinned in |
|---|---|---|
| OpenBao (chart) | 0.29.0 | {{ src(path="workloads/secrets/openbao.go") }} |
| OpenBao (image) | 2.6.1 | same — unseal sidecar |
| Secrets Store CSI Driver | 1.6.0 | {{ src(path="workloads/secrets/csi_driver.go") }} |

### Observability

| Software | Chart | Pinned in |
|---|---|---|
| VictoriaMetrics k8s-stack | 0.72.4 | {{ src(path="workloads/observability/victoria_metrics.go") }} |
| VictoriaLogs | 0.13.9 | {{ src(path="workloads/cdk8s.yaml", label="cdk8s.yaml") }} |
| Grafana | 12.10.4 | {{ src(path="workloads/cdk8s.yaml", label="cdk8s.yaml") }} |
| OpenTelemetry Collector | 0.169.0 | {{ src(path="workloads/observability/otel_collector.go") }} |
| Metrics Server | 3.13.1 | {{ src(path="workloads/cdk8s.yaml", label="cdk8s.yaml") }} |
| kube-prometheus-stack CRDs | 82.0.1 | {{ src(path="workloads/observability/victoria_metrics.go") }} |

VMAlertmanager ships inside the VictoriaMetrics k8s-stack chart; it has no
separate pin.

### Security

| Software | Chart | Pinned in |
|---|---|---|
| Falco | 8.0.5 | {{ src(path="workloads/security/falco.go") }} |
| Kyverno | 3.8.2 | {{ src(path="workloads/cdk8s.yaml", label="cdk8s.yaml") }} |
| Trivy Operator (chart) | 0.35.0 | {{ src(path="workloads/cdk8s.yaml", label="cdk8s.yaml") }} |
| Trivy Operator (CRD bundle) | 0.32.0 | {{ src(path="workloads/security/trivy.go") }} |

Trivy is the one component with two deliberately different pins: the chart comes
from the generated typed package, while the CRD bundle is downloaded separately
by URL. They do not have to match, but a chart bump usually wants a CRD bump too.

### Applications

| Software | Version | Pinned in |
|---|---|---|
| Harbor | 1.19.2 | {{ src(path="workloads/cdk8s.yaml", label="cdk8s.yaml") }} |
| Headlamp | 0.44.0 | {{ src(path="workloads/cdk8s.yaml", label="cdk8s.yaml") }} |
| n8n (chart) | 2.0.1 | {{ src(path="workloads/automation/n8n.go") }} |
| n8n (image) | 1.78.0 | same |
| Reloader | 2.2.16 | {{ src(path="workloads/support/reloader.go") }} |
| NetBird peer | 0.76.3 | {{ src(path="workloads/networking/netbird_peer.go") }} |
| Kubeflow | v1.11.0 | {{ src(path="workloads/ai/kubeflow/kustomization.yaml") }} |
| notebook-gateway-controller | v1 | {{ src(path="workloads/ai/notebook_gateway_controller.go") }} |

### AI and GPU

| Software | Version | Pinned in |
|---|---|---|
| Ollama (chart) | 1.74.0 | {{ src(path="workloads/cdk8s.yaml", label="cdk8s.yaml") }} |
| NVIDIA Device Plugin | 0.19.3 | {{ src(path="workloads/hardware/nvidia_gpu_operator.go") }} |
| DCGM Exporter | 4.8.3 | same |
| ComfyUI | cu130-megapak-pt211-20260812 | {{ src(path="workloads/ai/comfyui.go") }} |

Ollama's image tag is deliberately unset so it tracks the chart's appVersion —
the chart is the pin. ComfyUI's tag encodes its CUDA line: the card is Blackwell
(sm_120) and needs CUDA 12.8 or newer, so a `cu126` tag will not run.

## Checking for updates

`helm search repo` compares **chart** versions. A chart can lag the software it
ships, so check the app version too — Argo CD 3.5.1 was released while chart
10.3.2 still pinned 3.5.0.

```bash
helm repo update
helm search repo <repo>/<chart> --versions | head -3   # CHART and APP columns
```

Run `just docs-check` after any bump. It re-reads every row above, greps the file
named in the last column, and fails if the number here is no longer in that file.
