+++
title = "VictoriaMetrics"
description = "Time-series metrics storage for the cluster: VMSingle, VMAgent, VMAlert and Alertmanager from one chart."
weight = 20
+++

## What is VictoriaMetrics?

[VictoriaMetrics](https://victoriametrics.com/) is where every metric in the cluster ends up. It speaks Prometheus remote-write and PromQL, so anything that would have targeted Prometheus works unchanged — at noticeably lower memory and disk cost, which is the reason it is here rather than Prometheus.

## Why VictoriaMetrics?

| Feature | Prometheus | VictoriaMetrics |
|---------|------------|-----------------|
| Memory usage | High | Low (3–7x less) |
| Compression | Good | Excellent |
| Ingestion speed | Good | Faster |
| Query language | PromQL | MetricsQL (superset of PromQL) |
| Long-term storage | Requires Thanos/Cortex | Built-in |

VictoriaMetrics is a drop-in Prometheus replacement that works with any PromQL-compatible client (Grafana, etc.) while using fewer resources — important for a homelab running many workloads on limited hardware.

## How It's Used Here

One chart (`victoria-metrics-k8s-stack`) brings up the whole metrics pipeline in the
`victoria-metrics` namespace:

| Component | Role |
|-----------|------|
| `vmsingle` | Stores and serves everything — ingest, query, and a 100 Gi Longhorn PVC |
| `vmagent` | Scrapes the cluster and remote-writes into vmsingle |
| `vmalert` | Evaluates PrometheusRules and VMRules |
| `vmalertmanager` | Groups and routes the resulting alerts |

**Single-node, not cluster mode.** `vmcluster.enabled: false` — there is no
`vminsert`/`vmselect`/`vmstorage` split. One node's worth of metrics does not justify
the extra moving parts, and everything lands on one endpoint:

```
vmsingle-vm-stack.victoria-metrics.svc.cluster.local:8428
```

`fullnameOverride: "vm-stack"` keeps that name short on purpose: the default naming
pushes VMAlertmanager's generated pod labels past Kubernetes' 63-byte limit.

Source: {{ src(path="workloads/observability/victoria_metrics.go") }}

## Configuration

| Setting | Value | Why |
|---------|-------|-----|
| Namespace | `victoria-metrics` | Isolated namespace |
| Retention | `30d` | 30 days of metrics |
| vmsingle PVC | `100Gi` Longhorn | Metrics storage |
| vmsingle limits | `1000m` / `2Gi` | Ingest and query in one process |
| vmagent limits | `500m` / `512Mi` | Scrape path only |
| vmalert / alertmanager limits | `200m` / `256Mi` each | Both are light |
| Scrape interval | `30s` | |

## Alertmanager

VMAlertmanager ships inside the same k8s-stack chart, so there is no separate
deployment and no separate namespace — it runs as `vmalertmanager-vm-stack` in
`victoria-metrics` and is reachable at `https://alertmanager.madhan.app`.

Alerts currently go nowhere on purpose — the only receiver is a blackhole:

```yaml
route:
  group_by: [alertname, namespace]
  group_wait: 10s
  group_interval: 10m
  repeat_interval: 1h
  receiver: blackhole
receivers:
  - name: blackhole
```

Add a real receiver under `alertmanager.config.receivers` in
{{ src(path="workloads/observability/victoria_metrics.go") }} to get
notifications out.

```bash
curl https://alertmanager.madhan.app/api/v2/alerts | jq .
curl https://alertmanager.madhan.app/api/v2/silences | jq .
```

## How VMAgent Discovers Metrics

VMAgent runs with `selectAllByDefault: true`, so it picks up every ServiceMonitor and
PodMonitor in every namespace with no per-app configuration:

```yaml
# Every app with a ServiceMonitor is automatically scraped
# Examples:
# - OpenBao: /v1/sys/metrics (Prometheus format)
# - Falco sidekick: :2801/metrics
# - DCGM Exporter: GPU metrics
# - Argo CD components: :8082-8085/metrics
# - Longhorn: via ServiceMonitor
```

## HTTPRoute

Reachable at `https://vmselect.madhan.app` (the hostname predates the move to
single-node). The root redirects to `/vmui/` — **not** `/select/0/vmui/`, which is
the cluster-mode path and 404s here.

OpenBao has no ServiceMonitor, so it is scraped by an `inlineScrapeConfig` against
`/v1/sys/metrics` instead.

## How It Connects

```
All cluster apps (ServiceMonitor/PodMonitor)
  → VMAgent (scrapes every 30s)
  → vmsingle-vm-stack:8428 (remote-write, storage, query — 100Gi Longhorn PVC)
  → Grafana (Prometheus datasource)
  → VMAlert (rule evaluation) → VMAlertmanager:9093 (grouping, routing)
```

## Troubleshooting

### VMAgent Not Scraping

**Symptoms:** Missing metrics in Grafana.

**Diagnosis:**

```bash
# Check VMAgent targets
kubectl port-forward -n victoria-metrics svc/vmagent-vm-stack 8429:8429
# Open http://localhost:8429/targets
```

**Fix:** If a target is down, check the ServiceMonitor selector matches the service labels. If the ServiceMonitor itself is missing, the app chart may not deploy it.

### Storage Full

**Symptoms:** writes fail, no new data ingested.

```bash
kubectl exec -n victoria-metrics vmsingle-vm-stack-0 -- df -h /storage
```

Expand the PVC via the Longhorn UI, or raise the `vmsingle` storage request in
{{ src(path="workloads/observability/victoria_metrics.go") }} and re-sync.
