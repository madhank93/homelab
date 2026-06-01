+++
title = "Software Inventory"
description = "All software and services in the homelab, their versions, and where they are defined in code."
weight = 50
+++

Versions sourced directly from code — Helm chart versions, Docker image tags, or binary version constants.

<table style="width:100%; table-layout:fixed; word-break:break-word;">
<thead>
<tr>
  <th style="width:12%">Hosted</th>
  <th style="width:12%">Category</th>
  <th style="width:20%">Software</th>
  <th style="width:11%">Version</th>
  <th style="width:14%">Managed by</th>
  <th>Code link</th>
</tr>
</thead>
<tbody>

<tr>
  <td>Proxmox</td>
  <td></td>
  <td>Talos Linux</td>
  <td>v1.13.3</td>
  <td>Pulumi</td>
  <td><a href="https://github.com/madhank93/homelab/blob/main/core/platform/talos.go">core/platform/talos.go</a></td>
</tr>

<tr>
  <td rowspan="6">Cloud<br><small>(Hetzner VPS)</small></td>
  <td rowspan="6"></td>
  <td>Traefik</td><td>v3.7.1</td>
  <td rowspan="6">Pulumi</td>
  <td><a href="https://github.com/madhank93/homelab/blob/main/core/cloud/bifrost/docker-compose.yml">bifrost/docker-compose.yml</a></td>
</tr>
<tr>
  <td>NetBird (server + agent)</td><td>0.71.4</td>
  <td><a href="https://github.com/madhank93/homelab/blob/main/core/cloud/bifrost/docker-compose.yml">bifrost/docker-compose.yml</a></td>
</tr>
<tr>
  <td>Authentik</td><td>2026.5.2</td>
  <td><a href="https://github.com/madhank93/homelab/blob/main/core/cloud/bifrost/docker-compose.yml">bifrost/docker-compose.yml</a></td>
</tr>
<tr>
  <td>Gatus</td><td>v5.36.0</td>
  <td><a href="https://github.com/madhank93/homelab/blob/main/core/cloud/bifrost/docker-compose.yml">bifrost/docker-compose.yml</a></td>
</tr>
<tr>
  <td>Cloudflare DNS</td><td>—</td>
  <td><a href="https://github.com/madhank93/homelab/blob/main/core/cloud/cloudflare.go">core/cloud/cloudflare.go</a></td>
</tr>
<tr>
  <td>Hetzner VPS</td><td>—</td>
  <td><a href="https://github.com/madhank93/homelab/blob/main/core/cloud/hetzner.go">core/cloud/hetzner.go</a></td>
</tr>

<!-- Kubernetes: 26 data rows total -->
<tr>
  <td rowspan="26">Kubernetes</td>
  <td rowspan="4">Core</td>
  <td>Cilium CNI</td><td>1.18.10</td>
  <td rowspan="4">Pulumi</td>
  <td><a href="https://github.com/madhank93/homelab/blob/main/core/platform/cilium.go">core/platform/cilium.go</a></td>
</tr>
<tr>
  <td>Gateway API CRDs</td><td>v1.2.1</td>
  <td><a href="https://github.com/madhank93/homelab/blob/main/core/platform/cilium.go">core/platform/cilium.go</a></td>
</tr>
<tr>
  <td>ArgoCD</td><td>9.5.15</td>
  <td><a href="https://github.com/madhank93/homelab/blob/main/core/platform/argocd.go">core/platform/argocd.go</a></td>
</tr>
<tr>
  <td>cert-manager</td><td>v1.19.3</td>
  <td><a href="https://github.com/madhank93/homelab/blob/main/core/platform/cert_manager.go">core/platform/cert_manager.go</a></td>
</tr>

<tr>
  <td>Networking</td>
  <td>NetBird routing peer</td><td>0.71.4</td>
  <td rowspan="22">CDK8s + ArgoCD</td>
  <td><a href="https://github.com/madhank93/homelab/blob/main/workloads/networking/netbird_peer.go">workloads/networking/netbird_peer.go</a></td>
</tr>

<tr>
  <td rowspan="2">Storage</td>
  <td>Longhorn</td><td>1.11.2</td>
  <td><a href="https://github.com/madhank93/homelab/blob/main/workloads/storage/longhorn.go">workloads/storage/longhorn.go</a></td>
</tr>
<tr>
  <td>CloudNativePG</td><td>0.28.2</td>
  <td><a href="https://github.com/madhank93/homelab/blob/main/workloads/databases/cnpg.go">workloads/databases/cnpg.go</a></td>
</tr>

<tr>
  <td rowspan="2">Secrets</td>
  <td>OpenBao</td><td>0.28.3</td>
  <td><a href="https://github.com/madhank93/homelab/blob/main/workloads/secrets/openbao.go">workloads/secrets/openbao.go</a></td>
</tr>
<tr>
  <td>Secrets Store CSI Driver</td><td>1.5.6</td>
  <td><a href="https://github.com/madhank93/homelab/blob/main/workloads/secrets/csi_driver.go">workloads/secrets/csi_driver.go</a></td>
</tr>

<tr>
  <td rowspan="5">Observability</td>
  <td>VictoriaMetrics k8s-stack</td><td>0.72.4</td>
  <td><a href="https://github.com/madhank93/homelab/blob/main/workloads/observability/victoria_metrics.go">workloads/observability/victoria_metrics.go</a></td>
</tr>
<tr>
  <td>VictoriaLogs</td><td>0.12.5</td>
  <td><a href="https://github.com/madhank93/homelab/blob/main/workloads/observability/victoria_logs.go">workloads/observability/victoria_logs.go</a></td>
</tr>
<tr>
  <td>Grafana</td><td>12.4.1</td>
  <td><a href="https://github.com/madhank93/homelab/blob/main/workloads/monitoring/grafana.go">workloads/monitoring/grafana.go</a></td>
</tr>
<tr>
  <td>OpenTelemetry Collector</td><td>0.156.2</td>
  <td><a href="https://github.com/madhank93/homelab/blob/main/workloads/observability/otel_collector.go">workloads/observability/otel_collector.go</a></td>
</tr>
<tr>
  <td>Metrics Server</td><td>3.13.0</td>
  <td><a href="https://github.com/madhank93/homelab/blob/main/workloads/monitoring/metrics_server.go">workloads/monitoring/metrics_server.go</a></td>
</tr>

<tr>
  <td rowspan="3">Security</td>
  <td>Falco</td><td>8.0.5</td>
  <td><a href="https://github.com/madhank93/homelab/blob/main/workloads/security/falco.go">workloads/security/falco.go</a></td>
</tr>
<tr>
  <td>Trivy Operator</td><td>0.32.0</td>
  <td><a href="https://github.com/madhank93/homelab/blob/main/workloads/security/trivy.go">workloads/security/trivy.go</a></td>
</tr>
<tr>
  <td>Kyverno</td><td>3.8.1</td>
  <td><a href="https://github.com/madhank93/homelab/blob/main/workloads/security/kyverno.go">workloads/security/kyverno.go</a></td>
</tr>

<tr>
  <td rowspan="3">Management</td>
  <td>Headlamp</td><td>0.42.0</td>
  <td><a href="https://github.com/madhank93/homelab/blob/main/workloads/management/headlamp.go">workloads/management/headlamp.go</a></td>
</tr>
<tr>
  <td>Harbor</td><td>1.19.0</td>
  <td><a href="https://github.com/madhank93/homelab/blob/main/workloads/registry/harbor.go">workloads/registry/harbor.go</a></td>
</tr>
<tr>
  <td>Reloader</td><td>2.2.12</td>
  <td><a href="https://github.com/madhank93/homelab/blob/main/workloads/support/reloader.go">workloads/support/reloader.go</a></td>
</tr>

<tr>
  <td rowspan="5">AI / ML</td>
  <td>Ollama</td><td>1.57.0</td>
  <td><a href="https://github.com/madhank93/homelab/blob/main/workloads/ai/ollama.go">workloads/ai/ollama.go</a></td>
</tr>
<tr>
  <td>ComfyUI</td><td>cu128-megapak-20260223</td>
  <td><a href="https://github.com/madhank93/homelab/blob/main/workloads/ai/comfyui.go">workloads/ai/comfyui.go</a></td>
</tr>
<tr>
  <td>Kubeflow</td><td>v1.11.0</td>
  <td><a href="https://github.com/madhank93/homelab/blob/main/workloads/ai/kubeflow/kustomization.yaml">workloads/ai/kubeflow/kustomization.yaml</a></td>
</tr>
<tr>
  <td>NVIDIA Device Plugin</td><td>0.19.1</td>
  <td><a href="https://github.com/madhank93/homelab/blob/main/workloads/hardware/nvidia_gpu_operator.go">workloads/hardware/nvidia_gpu_operator.go</a></td>
</tr>
<tr>
  <td>NVIDIA DCGM Exporter</td><td>4.8.2</td>
  <td><a href="https://github.com/madhank93/homelab/blob/main/workloads/hardware/nvidia_gpu_operator.go">workloads/hardware/nvidia_gpu_operator.go</a></td>
</tr>

<tr>
  <td>Automation</td>
  <td>n8n</td><td>2.0.1</td>
  <td><a href="https://github.com/madhank93/homelab/blob/main/workloads/automation/n8n.go">workloads/automation/n8n.go</a></td>
</tr>

</tbody>
</table>

**Notes:** Cloudflare and Hetzner have no version — provider/account configs. ComfyUI uses a build tag, not semver. All Kubernetes chart versions are Helm chart versions.
