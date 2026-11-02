---
title: VictoriaLogs
description: Single-node log store fed by the OpenTelemetry collectors over OTLP and queried through Grafana's Loki datasource.
sidebar:
  order: 30
  badge: { text: LAN, variant: note }
---

VictoriaLogs stores container logs, Kubernetes events and Falco alerts. Grafana reads it through its Loki plugin. [Upstream docs](https://docs.victoriametrics.com/victorialogs/).

| | |
|---|---|
| Namespace | `victoria-logs` |
| Source | [`victoria_logs.go`](src:workloads/observability/victoria_logs.go) |
| Exposure | LAN · `victorialogs.madhan.app` (built-in UI) |
| Storage | 100Gi on the default `longhorn` class |
| Auth | None |

## Configuration

| Setting | Value | Why |
|---|---|---|
| `retentionPeriod` | `30d` | |
| `retentionDiskSpaceUsage` | `40` GiB | Each replica shares a worker disk with every other Longhorn volume, so cap by size as well as age |
| Limits | `1000m` / `1Gi` | |
| `image.pullPolicy` | `Always` | Recovers from a corrupted cached image layer |

| Endpoint | Used by |
|---|---|
| `http://victoria-logs-victoria-logs-single-server.victoria-logs.svc.cluster.local:9428/insert/opentelemetry` | OTel `otlp_http/logs` exporter, which appends `/v1/logs` |
| `…:9428/select` | Grafana Loki datasource |

:::caution[Do not append /loki to the Grafana URL]
The Loki plugin adds `/loki/api/v1/…` itself. A URL ending in `/select/loki` doubles the path and every query returns 404.
:::

## Operate

Query in Grafana's Explore with the VictoriaLogs datasource (main org only):

```txt
{namespace="ollama"} |= "error"
{namespace="falco"} |= "Terminal shell in container"
```

Or directly: `curl "https://victorialogs.madhan.app/select/logsql/query?query=*&limit=5"`.

## Troubleshooting

### No logs in Grafana

Check VictoriaLogs runs and the agents export without errors:

```bash frame="terminal"
kubectl get pods -n victoria-logs
kubectl logs -n opentelemetry -l app.kubernetes.io/instance=otel-agent --tail=50
```

### Every log query returns 404

The datasource URL has `/loki` appended. It must end at `/select`.
