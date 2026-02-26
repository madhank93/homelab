# Observability Stack

The homelab observability stack is built on OpenTelemetry, VictoriaMetrics, VictoriaLogs, and Grafana. It is fully GitOps-managed via CDK8s in `platform/cdk8s/cots/monitoring/`.

---

## Table of Contents

- [Architecture](#architecture)
- [Components](#components)
- [What OTel Collects](#what-otel-collects)
- [Grafana Datasources](#grafana-datasources)
- [Falco Integration](#falco-integration)
- [Service Endpoints](#service-endpoints)
- [Recommended Grafana Dashboards](#recommended-grafana-dashboards)

---

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│  OTel Agent (DaemonSet — every node)                    │
│   filelog      → container logs from /var/log/pods      │
│   kubeletstats → pod/node/container metrics             │
│   hostmetrics  → CPU/memory/disk/network per host       │
└────────────────────┬────────────────────────────────────┘
                     │
┌─────────────────────────────────────────────────────────┐
│  OTel Gateway (Deployment)                              │
│   k8s_cluster → node/pod/deployment resource metrics   │
│   k8sobjects  → Kubernetes events as logs              │
└────────────────────┬────────────────────────────────────┘
                     │
          ┌──────────┴──────────┐
          ▼                     ▼
  VictoriaMetrics          VictoriaLogs
  :8481/select/...         :9428/select/loki
  (prometheus compat)      (loki compatible)
          │                     │
          └──────────┬──────────┘
                     ▼
                  Grafana
             grafana.madhan.app
```

Both the OTel Agent and OTel Gateway export:
- **Metrics** → VictoriaMetrics via Prometheus remote-write (`/insert/0/prometheus/api/v1/write`)
- **Logs** → VictoriaLogs via OTLP/HTTP (`/insert/opentelemetry`)

---

## Components

### OTel Agent (DaemonSet)

| Property | Value |
|----------|-------|
| File | `platform/cdk8s/cots/monitoring/otel_collector.go` |
| Namespace | `opentelemetry` |
| Helm release | `otel-agent` (opentelemetry-collector chart v0.108.0) |
| Image | `otel/opentelemetry-collector-contrib` |
| Mode | DaemonSet (one pod per node, including control plane) |

The agent uses Helm chart presets which auto-configure the necessary RBAC, volume mounts, and receiver configuration:

- **`logsCollection`** preset: Configures the `filelog` receiver to tail container logs from `/var/log/pods/` on each node. Uses the `k8sattributes` processor to enrich logs with Kubernetes metadata (namespace, pod name, container name, labels).
- **`kubeletMetrics`** preset: Configures the `kubeletstats` receiver to scrape node/pod/container/volume metrics from the kubelet `/stats/summary` endpoint.
- **`hostMetrics`** preset: Configures the `hostmetrics` receiver for CPU, memory, disk I/O, and network metrics from the host OS.
- **`kubernetesAttributes`** preset: Enriches all telemetry with `k8s.pod.name`, `k8s.namespace.name`, `k8s.deployment.name`, and all pod labels.

The agent toleration `operator: Exists` ensures it runs on control plane nodes as well.

---

### OTel Gateway (Deployment)

| Property | Value |
|----------|-------|
| File | `platform/cdk8s/cots/monitoring/otel_collector.go` |
| Namespace | `opentelemetry` |
| Helm release | `otel-gateway` (opentelemetry-collector chart v0.108.0) |
| Image | `otel/opentelemetry-collector-contrib` |
| Mode | Deployment (single replica) |

The gateway uses presets for cluster-level collection:

- **`clusterMetrics`** preset: Configures the `k8s_cluster` receiver for cluster-scoped resource metrics — node counts, pod phase distributions, deployment replica counts, namespace counts.
- **`kubernetesEvents`** preset: Configures the `k8sobjects` receiver to watch Kubernetes events (Pod scheduling, container restart reasons, PVC binding, etc.) and forward them as log records to VictoriaLogs.
- **`kubernetesAttributes`** preset: Enriches telemetry with cluster-scoped metadata.

---

### VictoriaMetrics Cluster

| Property | Value |
|----------|-------|
| File | `platform/cdk8s/cots/monitoring/victoria_metrics.go` |
| Namespace | `victoria-metrics` |
| Helm chart | `victoria-metrics-cluster` |

Deployed as a cluster (three components):
- **vminsert** (port 8480): accepts remote-write from OTel collectors and other Prometheus-compatible sources.
- **vmselect** (port 8481): serves PromQL queries; Grafana connects here.
- **vmstorage**: stores time-series data.

---

### VictoriaLogs

| Property | Value |
|----------|-------|
| File | `platform/cdk8s/cots/monitoring/victoria_logs.go` |
| Namespace | `victoria-logs` |
| Helm chart | `victoria-logs-single` |
| Port | 9428 |

Single-server deployment. Receives logs via OTLP/HTTP on port 9428. Exposes a Loki-compatible query API at `:9428/select/loki` which Grafana uses for log queries.

---

### AlertManager

| Property | Value |
|----------|-------|
| File | `platform/cdk8s/cots/monitoring/alert_manager.go` |
| Namespace | `alertmanager` |
| Helm chart | `kube-prometheus-stack` |

Handles alert routing, grouping, inhibition, and deduplication. Routes alerts to configured receivers (Slack, PagerDuty, webhooks).

---

### Grafana

| Property | Value |
|----------|-------|
| File | `platform/cdk8s/cots/monitoring/grafana.go` |
| Namespace | `grafana` |
| Helm chart | `grafana` v10.5.15 |
| HTTPRoute | `grafana.madhan.app` → `grafana:80` |
| UI | Yes |

Visualization platform. Datasources are provisioned automatically via Helm values — no manual configuration needed after deployment.

---

## What OTel Collects

| Source | Receiver | Data Type | Collector Tier | Destination |
|--------|----------|-----------|----------------|-------------|
| Container logs (`/var/log/pods/`) | `filelog` | Logs | Agent (DaemonSet) | VictoriaLogs |
| Pod/node/container metrics | `kubeletstats` | Metrics | Agent (DaemonSet) | VictoriaMetrics |
| Host OS metrics (CPU/mem/disk/net) | `hostmetrics` | Metrics | Agent (DaemonSet) | VictoriaMetrics |
| Kubernetes resource metrics | `k8s_cluster` | Metrics | Gateway (Deployment) | VictoriaMetrics |
| Kubernetes events | `k8sobjects` | Logs | Gateway (Deployment) | VictoriaLogs |

All telemetry is enriched with Kubernetes metadata by the `k8sattributes` processor before export.

---

## Grafana Datasources

Both datasources are provisioned automatically by Grafana's Helm chart values. No manual setup is needed after ArgoCD syncs the Grafana app.

### VictoriaMetrics (Prometheus-compatible)

| Property | Value |
|----------|-------|
| Datasource type | Prometheus |
| URL | `http://victoria-metrics-victoria-metrics-cluster-vmselect.victoria-metrics.svc.cluster.local:8481/select/0/prometheus` |
| Default | Yes |

Used for all metrics queries (PromQL).

### VictoriaLogs (Loki-compatible)

| Property | Value |
|----------|-------|
| Datasource type | Loki |
| URL | `http://victoria-logs-victoria-logs-single-server.victoria-logs.svc.cluster.local:9428/select/loki` |
| Default | No |

Used for all log queries (LogQL).

---

## Falco Integration

Falco (in the `compliance/` folder) outputs security alerts as JSON to stdout. These flow through the observability stack as follows:

```
Falco DaemonSet pod
  └── JSON alerts written to container stdout
        │
        ▼
  OTel Agent filelog receiver
  (collects all container logs from /var/log/pods/)
        │
        ▼
  VictoriaLogs
        │
        ▼
  Grafana (Loki datasource)
  Filter: {k8s_namespace_name="falco"}
```

No special Falco output plugin is required. The existing `filelog` receiver on the OTel agent automatically captures Falco's stdout as log records, enriches them with the `k8s_namespace_name=falco` and `k8s_container_name=falco` attributes, and forwards them to VictoriaLogs.

To query Falco alerts in Grafana:
```logql
{k8s_namespace_name="falco"} | json | priority="WARNING" or priority="ERROR" or priority="CRITICAL"
```

---

## Service Endpoints

All endpoints are cluster-internal (not exposed via HTTPRoute).

### VictoriaMetrics

| Use | Endpoint |
|-----|----------|
| Prometheus remote-write (write) | `http://victoria-metrics-victoria-metrics-cluster-vminsert.victoria-metrics.svc.cluster.local:8480/insert/0/prometheus/api/v1/write` |
| PromQL query (read) | `http://victoria-metrics-victoria-metrics-cluster-vmselect.victoria-metrics.svc.cluster.local:8481/select/0/prometheus` |

### VictoriaLogs

| Use | Endpoint |
|-----|----------|
| OTLP/HTTP log ingestion | `http://victoria-logs-victoria-logs-single-server.victoria-logs.svc.cluster.local:9428/insert/opentelemetry` |
| Loki-compatible query API | `http://victoria-logs-victoria-logs-single-server.victoria-logs.svc.cluster.local:9428/select/loki` |

---

## Recommended Grafana Dashboards

Import these dashboard IDs from grafana.com to get immediate visibility:

| Dashboard ID | Name | Data Source |
|-------------|------|-------------|
| 15661 | Kubernetes Cluster Overview | VictoriaMetrics |
| 1860 | Node Exporter Full (host metrics) | VictoriaMetrics |
| 11914 | Falco Alerts | VictoriaLogs |
| 17900 | VictoriaLogs Explorer | VictoriaLogs |

To import: Grafana UI → Dashboards → Import → enter ID → select datasource.
