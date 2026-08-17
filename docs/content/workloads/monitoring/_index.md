+++
title = "Monitoring"
description = "Observability stack: VictoriaMetrics, VictoriaLogs, Grafana, OpenTelemetry, Metrics Server."
weight = 20
sort_by = "weight"
+++

The monitoring stack provides full observability for the cluster:

- **VictoriaMetrics** — time-series metrics storage (Prometheus-compatible)
- **VictoriaLogs** — log storage (Loki-compatible)
- **Grafana** — dashboards and visualization
- **OpenTelemetry Collector** — metrics and log collection pipeline
- **Metrics Server** — resource metrics for `kubectl top` and HPAs
- **argocd-monitor** — ServiceMonitors for the Argo CD components

Alertmanager is not separate: it ships inside the VictoriaMetrics k8s-stack chart.

## Data Flow

```
Nodes/Pods
    │
    ├── Container logs ──→ OTel Agent DaemonSet ──→ VictoriaLogs
    ├── Host metrics ────→ OTel Agent DaemonSet ──→ VictoriaMetrics
    ├── kubelet metrics ──→ OTel Agent DaemonSet ──→ VictoriaMetrics
    └── k8s events ─────→ OTel Gateway ──────────→ VictoriaLogs
                                                         │
                                               Grafana queries ◄──── User
```

| Page | What it covers |
|---|---|
| [VictoriaMetrics](@/workloads/monitoring/victoria-metrics/index.md) | Metrics storage, VMAgent scraping, and Alertmanager |
| [VictoriaLogs](@/workloads/monitoring/victoria-logs/index.md) | Log storage and LogQL queries |
| [Grafana](@/workloads/monitoring/grafana/index.md) | Dashboards, datasources, and Authentik OIDC |
| [OpenTelemetry](@/workloads/monitoring/opentelemetry/index.md) | The agent/gateway collector pipelines |
| [Metrics Server](@/workloads/monitoring/metrics-server/index.md) | `kubectl top` and HPA support |
| [Argo CD Monitor](@/workloads/monitoring/argocd-monitor/index.md) | ServiceMonitors for Argo CD's components |
