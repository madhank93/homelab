+++
title = "Monitoring"
description = "Observability stack: VictoriaMetrics, VictoriaLogs, Grafana, AlertManager, OpenTelemetry."
weight = 20
sort_by = "weight"
+++

The monitoring stack provides full observability for the cluster:

- **VictoriaMetrics** — time-series metrics storage (Prometheus-compatible)
- **VictoriaLogs** — log storage (Loki-compatible)
- **Grafana** — dashboards and visualization
- **AlertManager** — alert routing and deduplication
- **OpenTelemetry Collector** — metrics and log collection pipeline

## Data Flow

```
Nodes/Pods
    │
    ├── Container logs ──→ OTel Agent DaemonSet ──→ VictoriaLogs
    ├── Host metrics ────→ OTel Agent DaemonSet ──→ VictoriaMetrics
    ├── kubelet metrics ──→ OTel Agent DaemonSet ──→ VictoriaMetrics
    └── k8s events ─────→ OTel Gateway ──────────→ VictoriaMetrics
                                                         │
                                               Grafana queries ◄──── User
```
