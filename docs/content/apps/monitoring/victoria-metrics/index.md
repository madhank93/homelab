+++
title = "VictoriaMetrics"
description = "Time-series metrics storage, Prometheus-compatible, cluster mode."
weight = 20
+++

## What is VictoriaMetrics?

[VictoriaMetrics](https://victoriametrics.com/) is a fast, resource-efficient time-series database and monitoring solution that is fully compatible with the Prometheus remote-write protocol and PromQL query language. It offers significantly lower memory and storage usage compared to Prometheus, with better compression and faster ingestion.

## Why VictoriaMetrics?

| Feature | Prometheus | VictoriaMetrics |
|---------|------------|-----------------|
| Memory usage | High | Low (3–7x less) |
| Compression | Good | Excellent |
| Ingestion speed | Good | Faster |
| Query language | PromQL | MetricsQL (superset of PromQL) |
| Long-term storage | Requires Thanos/Cortex | Built-in |
| Cluster mode | Manual | Native |

VictoriaMetrics is a drop-in Prometheus replacement that works with any PromQL-compatible client (Grafana, etc.) while using fewer resources — important for a homelab running many workloads on limited hardware.

## How It's Used Here

VictoriaMetrics runs in **cluster mode** with three components:

| Component | Replicas | Role |
|-----------|---------|------|
| `vminsert` | 1 | Receives prometheus remote-write from VMAgent and OTel collectors |
| `vmselect` | 1 | Serves PromQL queries from Grafana |
| `vmstorage` | 1 | Stores time-series data to a 100 Gi Longhorn PVC |

**VMAgent** (a separate lightweight agent) scrapes metrics across the cluster and remote-writes to vminsert:
- Discovers `ServiceMonitor` resources across all namespaces
- Discovers `PodMonitor` resources across all namespaces (required for CNPG)
- Remote-writes to `vminsert:8480`

Source: [`workloads/observability/victoria_metrics.go`](https://github.com/madhank93/homelab/blob/v0.1.5/workloads/observability/victoria_metrics.go)

## Configuration

| Setting | Value | Why |
|---------|-------|-----|
| Namespace | `victoria-metrics` | Isolated namespace |
| Retention | `30d` | 30 days of metrics |
| vmstorage PVC | `100Gi` Longhorn | Long-term metrics storage |
| vminsert resources | `500m` / `512Mi` | Lightweight write path |
| vmselect resources | `500m` / `1Gi` | Query path needs more memory |
| vmstorage resources | `1000m` / `1Gi` | Storage is most resource-intensive |
| VMAgent chart version | `0.15.3` | victoria-metrics-agent |

## How VMAgent Discovers Metrics

VMAgent uses `serviceMonitorSelector: {}` and `podMonitorSelector: {}` (empty = match all), meaning it discovers every ServiceMonitor and PodMonitor across every namespace:

```yaml
# Every app with a ServiceMonitor is automatically scraped
# Examples:
# - OpenBao: /v1/sys/metrics (Prometheus format)
# - Falco sidekick: :2801/metrics
# - DCGM Exporter: GPU metrics
# - ArgoCD components: :8082-8085/metrics
# - Longhorn: via ServiceMonitor
```

## HTTPRoute

VictoriaMetrics vmselect is accessible at `http://vmselect.madhan.app`. Browsing to the root redirects to `/select/0/vmui/` — the vmui web interface.

## How It Connects

```
All cluster apps (ServiceMonitor/PodMonitor)
  → VMAgent (scrapes every 30s)
  → vminsert:8480 (prometheus remote-write)
  → vmstorage (100Gi Longhorn PVC)
  → vmselect:8481
  → Grafana (Prometheus datasource)
  → AlertManager (rules evaluation via Prometheus operator)
```

## Troubleshooting

### VMAgent Not Scraping

**Symptoms:** Missing metrics in Grafana.

**Diagnosis:**

```bash
# Check VMAgent targets
kubectl port-forward -n victoria-metrics svc/vmagent 8429:8429
# Open http://localhost:8429/targets
```

**Fix:** If a target is down, check the ServiceMonitor selector matches the service labels. If the ServiceMonitor itself is missing, the app chart may not deploy it.

### Storage Full

**Symptoms:** vminsert returns errors, no new data ingested.

```bash
kubectl exec -n victoria-metrics victoria-metrics-victoria-metrics-cluster-vmstorage-0 \
  -- df -h /storage
```

Expand the PVC via Longhorn UI or increase `vmstorage.persistentVolume.size` and re-sync.
