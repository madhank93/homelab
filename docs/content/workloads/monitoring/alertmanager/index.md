+++
title = "AlertManager"
description = "Alert routing, grouping, and deduplication via kube-prometheus-stack (AlertManager-only mode)."
weight = 40
+++

## What is AlertManager?

[AlertManager](https://prometheus.io/docs/alerting/latest/alertmanager/) is the Prometheus ecosystem's alert routing and notification system. It receives alerts from Prometheus-compatible rule engines, deduplicates them, groups related alerts, and routes them to notification channels (email, Slack, PagerDuty, webhooks, etc.).

## Why AlertManager?

AlertManager is the standard in the Prometheus/VictoriaMetrics ecosystem. The Prometheus Operator CRDs (`PrometheusRule`, `ServiceMonitor`, `PodMonitor`, `AlertmanagerConfig`) are widely used by Helm charts to define both scraping targets and alert rules — using AlertManager means these CRDs work out of the box.

## How It's Used Here

AlertManager is deployed from `kube-prometheus-stack` (chart `82.0.1`) in **AlertManager-only mode** — all other kube-prometheus-stack components (Prometheus, Grafana, node-exporter, kube-state-metrics) are disabled. Only the Prometheus Operator (for CRDs) and AlertManager itself are enabled.

This pattern avoids running a full Prometheus stack when VictoriaMetrics is already handling metrics storage and VMAgent is handling scraping.

Source: [`workloads/observability/alert_manager.go`](https://github.com/madhank93/homelab/blob/v0.1.5/workloads/observability/alert_manager.go)

## Configuration

| Setting | Value | Why |
|---------|-------|-----|
| Namespace | `alertmanager` | Isolated namespace |
| HTTPRoute | `alertmanager.madhan.app` → port 9093 | Gateway API |
| Replicas | `1` | Single instance for homelab |
| PVC | `10Gi` Longhorn | Silence and notification state persistence |
| Resources (limits) | `200m` / `256Mi` | AlertManager is lightweight |
| `prometheus.enabled` | `false` | Using VictoriaMetrics instead |
| `grafana.enabled` | `false` | Separate Grafana deployment |
| `kubeStateMetrics.enabled` | `false` | OTel Gateway handles cluster metrics |
| `nodeExporter.enabled` | `false` | OTel Agent handles host metrics |
| `prometheusOperator.enabled` | `true` | Required for CRDs |

## Alert Routing Configuration

The current routing is a placeholder (webhook receiver with no target):

```yaml
route:
  group_by: [alertname]
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: web.hook
receivers:
  - name: web.hook
```

To add real notifications, add receiver configurations:

```yaml
receivers:
  - name: slack-notifications
    slack_configs:
      - api_url: 'https://hooks.slack.com/services/...'
        channel: '#alerts'
```

## CRDs

The AlertManager chart installs these CRDs (pinned to kube-prometheus-stack `82.0.1`):

- `PrometheusRule` — alert and recording rules
- `ServiceMonitor` — scraping targets for services
- `PodMonitor` — scraping targets for pods
- `Alertmanager` — AlertManager configuration

These CRDs are used by VMAgent for service discovery and by apps to declare their own alert rules.

## How It Connects

```
VictoriaMetrics (VMAgent evaluates PrometheusRules)
  → Fires alerts to AlertManager:9093
  → AlertManager groups and routes alerts
  → Notification receivers (Slack, email, webhook)

Grafana → AlertManager UI (via Unified Alerting datasource)
```

## Troubleshooting

### Accessing the UI

```bash
# Port-forward for direct access
kubectl port-forward -n alertmanager svc/alertmanager-kube-promethe-alertmanager 9093:9093
# Open http://localhost:9093
```

Or browse to `http://alertmanager.madhan.app`.

### Checking Active Alerts

```bash
# List active alerts via API
curl http://alertmanager.madhan.app/api/v2/alerts | jq .

# Check silences
curl http://alertmanager.madhan.app/api/v2/silences | jq .
```
