+++
title = "VictoriaLogs"
description = "Log storage with Loki-compatible query API."
weight = 30
+++

## What is VictoriaLogs?

[VictoriaLogs](https://docs.victoriametrics.com/victorialogs/) is a log management solution from VictoriaMetrics that is compatible with the Loki API and ingestion protocol. It stores structured log data with efficient compression and provides LogQL-compatible querying.

## Why VictoriaLogs?

VictoriaLogs uses significantly less memory and storage than Loki while maintaining API compatibility. Since Grafana already uses VictoriaMetrics as its metrics backend, using VictoriaLogs for logs keeps the entire observability stack within a single vendor ecosystem.

## How It's Used Here

VictoriaLogs runs as a single-node server in the `victoria-logs` namespace. It receives logs from the OTel Collector via OTLP/HTTP and serves them to Grafana via the Loki-compatible API.

**Log sources:**
- Container stdout/stderr from every pod (via OTel Agent `filelog` receiver)
- Kubernetes events (via OTel Gateway `k8sobjects` receiver)
- Falco runtime security alerts (JSON to stdout, collected by OTel Agent)

Source: [`workloads/observability/victoria_logs.go`](https://github.com/madhank93/homelab/blob/v0.1.5/workloads/observability/victoria_logs.go)

## Configuration

| Setting | Value | Why |
|---------|-------|-----|
| Namespace | `victoria-logs` | Isolated namespace |
| Service port | `9428` | VictoriaLogs default |
| Retention | `30d` | 30 days of logs |
| PVC | `100Gi` Longhorn | Log storage |
| Resources (limits) | `1000m` / `1Gi` | Log ingestion is memory-intensive |
| Image pull policy | `Always` | Recover from corrupted cached image layers |

## OTLP Ingestion Endpoint

OTel Collector sends logs to:

```
http://victoria-logs-victoria-logs-single-server.victoria-logs.svc.cluster.local:9428/insert/opentelemetry
```

The collector appends `/v1/logs` automatically (per the OTLP spec). VictoriaLogs accepts OTLP JSON and proto at this path.

## Grafana Datasource URL (Important)

The VictoriaLogs URL in Grafana is configured as:

```
http://victoria-logs-victoria-logs-single-server.victoria-logs.svc.cluster.local:9428/select
```

> **Do NOT append `/loki` or `/loki/api/v1` to this URL.** Grafana's Loki datasource plugin appends the API path (`/loki/api/v1/query_range`, etc.) automatically. If you configure the URL as `/select/loki`, Grafana will construct double-path URLs like `/select/loki/loki/api/v1/query_range` and every query will return 404.

## HTTPRoute

VictoriaLogs UI is accessible at `http://victorialogs.madhan.app`. The web UI provides a log explorer interface.

## Querying in Grafana

Select the VictoriaLogs datasource and use LogQL:

```logql
# All logs from a namespace
{namespace="comfyui"}

# Filter by content
{namespace="ollama"} |= "error"

# Falco security alerts
{namespace="falco"} |= "Terminal shell in container"

# Kubernetes events
{k8s.resource.name="Event"} | json
```

## How It Connects

```
Every cluster node
  → OTel Agent DaemonSet (filelog receiver collects /var/log/pods/)
  → OTLP/HTTP POST to VictoriaLogs:9428/insert/opentelemetry

OTel Gateway (Deployment)
  → k8sobjects receiver (Kubernetes events)
  → OTLP/HTTP POST to VictoriaLogs

VictoriaLogs:9428/select
  → Grafana Loki datasource
  → LogQL queries
```

## Troubleshooting

### Logs Not Appearing in Grafana

**Symptoms:** No results in Grafana VictoriaLogs datasource.

**Diagnosis:**

```bash
# Check VictoriaLogs is running
kubectl get pods -n victoria-logs

# Check OTel Agent is sending logs
kubectl logs -n opentelemetry -l app.kubernetes.io/name=otel-agent --tail=50

# Check ingestion directly
curl "http://victorialogs.madhan.app/select/logsql/query?query=*&limit=5"
```

**Fix:** If OTel Agent is failing, check if the VictoriaLogs OTLP endpoint is reachable from the agent pods. The agent runs in the `opentelemetry` namespace with `privileged` pod security.

### Double-Path 404 Errors

**Symptoms:** Grafana shows `404 Not Found` for log queries.

**Fix:** Check the VictoriaLogs datasource URL in Grafana. It must end at `/select` — not `/select/loki`. See the note above about the Loki plugin appending paths automatically.
