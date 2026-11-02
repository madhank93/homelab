---
title: VictoriaMetrics
description: Single-node metrics store with VMAgent, VMAlert and Alertmanager, all from the victoria-metrics-k8s-stack chart.
sidebar:
  order: 20
  badge: { text: LAN, variant: note }
---

VictoriaMetrics stores every metric in the cluster and answers PromQL for Grafana. One chart brings VMSingle, VMAgent, VMAlert, Alertmanager, node-exporter and kube-state-metrics. [Upstream docs](https://docs.victoriametrics.com/).

| | |
|---|---|
| Namespace | `victoria-metrics` (privileged, for node-exporter) |
| Source | [`victoria_metrics.go`](src:workloads/observability/victoria_metrics.go) |
| Exposure | LAN · `vmselect.madhan.app` (VMUI), `alertmanager.madhan.app` |
| Storage | 100Gi on the default `longhorn` class, 30-day retention |
| Auth | None |

Why VictoriaMetrics over Prometheus: [VictoriaMetrics and VictoriaLogs](/reference/decisions/victoria-stack/).

## Configuration

| Setting | Value | Why |
|---|---|---|
| Mode | VMSingle, `vmcluster.enabled: false` | One cluster's metrics do not justify the insert/select/storage split |
| `fullnameOverride` | `vm-stack` | The default name pushes Alertmanager pod labels past 63 bytes. Endpoint: `vmsingle-vm-stack.victoria-metrics.svc:8428` |
| CRDs | VM operator CRDs, plus ServiceMonitor and PrometheusRule | `cdk8s.NewHelm` never renders a chart's `crds/`; the operator converts ServiceMonitors |
| `vmagent.selectAllByDefault` | `true` | Scrapes every ServiceMonitor and PodMonitor cluster-wide, every `30s` |
| OpenBao scrape | `inlineScrapeConfig` on `/v1/sys/metrics` | OpenBao has no ServiceMonitor |
| Operator CRD cleanup hook | disabled | Argo CD would run the pre-delete hook on every sync |
| Alertmanager receiver | `blackhole` | No notification channel yet |
| `kubeProxy` | disabled | Cilium replaces kube-proxy |
| Default dashboards and rules | enabled | Created by the chart's sync job (an Argo CD PostSync hook), not rendered |
| Grafana | disabled | Deployed separately as [Grafana](/workloads/monitoring/grafana/) |

The `vmselect` route redirects `/` to `/vmui/`. The hostname predates the move to single-node.

## Operate

Add a real receiver under `alertmanager.config.receivers` in `victoria_metrics.go` to send notifications.

```bash frame="terminal"
curl -s https://alertmanager.madhan.app/api/v2/alerts | jq .
kubectl port-forward -n victoria-metrics svc/vmagent-vm-stack 8429:8429   # then open /targets
```

## Troubleshooting

### A target's metrics are missing

The ServiceMonitor is absent or its selector misses the Service's labels. Check `/targets` on VMAgent (above).

### Writes fail, no new data

The volume is full. Raise the `vmsingle` storage request in `victoria_metrics.go`, or shorten `retentionPeriod`.
