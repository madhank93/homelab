+++
title = "OpenTelemetry"
description = "Two-tier OTel collection pipeline: DaemonSet agent per node and Gateway deployment."
weight = 50
+++

## What is OpenTelemetry?

[OpenTelemetry](https://opentelemetry.io/) (OTel) is a vendor-neutral observability framework for collecting metrics, logs, and traces. The [OpenTelemetry Collector](https://opentelemetry.io/docs/collector/) is a configurable pipeline that receives telemetry data, processes it, and exports it to backends.

## Why OpenTelemetry?

OTel provides a unified collection pipeline for both metrics and logs, replacing separate tools like Fluent Bit and Prometheus exporters. Using the `otel/opentelemetry-collector-contrib` image provides access to all Kubernetes-specific receivers (kubeletstats, k8s_cluster, filelog, k8sobjects) in a single binary.

## How It's Used Here

Two OTel Collector instances work together:

| Instance | Mode | Purpose |
|----------|------|---------|
| `otel-agent` | DaemonSet | Runs on every node; collects container logs, kubelet metrics, host metrics |
| `otel-gateway` | Deployment | Cluster-scoped; collects k8s resource metrics and events |

Both export:
- **Metrics** → VictoriaMetrics via Prometheus remote-write
- **Logs** → VictoriaLogs via OTLP/HTTP

Source: [`workloads/observability/otel_collector.go`](https://github.com/madhank93/homelab/blob/v0.1.5/workloads/observability/otel_collector.go)

## Agent DaemonSet Configuration

Runs on **every node** (including control plane, via `tolerations: [{operator: Exists}]`).

| Preset | What it configures |
|--------|-------------------|
| `logsCollection` | `filelog` receiver watching `/var/log/pods/`; enriches with k8s metadata |
| `kubeletMetrics` | Node, pod, container, volume metrics from kubelet `/stats/summary` |
| `hostMetrics` | CPU, memory, disk, network from the host OS |
| `kubernetesAttributes` | Enriches all telemetry with `k8s.pod.name`, `k8s.namespace.name`, `k8s.node.name`, etc. |

**Pipelines:**

```yaml
pipelines:
  logs:
    receivers:  [filelog]
    processors: [memory_limiter, k8sattributes, batch]
    exporters:  [otlphttp/logs]     # → VictoriaLogs
  metrics:
    receivers:  [kubeletstats, hostmetrics]
    processors: [memory_limiter, k8sattributes, batch]
    exporters:  [prometheusremotewrite]   # → VictoriaMetrics
```

**Namespace pod security:** `privileged` — required for host metric collection and `/var/log/pods` access on Talos.

## Gateway Deployment Configuration

Runs as a single-replica Deployment (cluster-scoped receivers don't need multiple instances).

| Preset | What it configures |
|--------|-------------------|
| `clusterMetrics` | `k8s_cluster` receiver: node/pod/deployment/namespace resource metrics |
| `kubernetesEvents` | `k8sobjects` receiver: Kubernetes events forwarded as log records |
| `kubernetesAttributes` | Enriches with k8s metadata |

**Pipelines:**

```yaml
pipelines:
  metrics:
    receivers:  [k8s_cluster]
    processors: [memory_limiter, k8sattributes, batch]
    exporters:  [prometheusremotewrite]
  logs:
    receivers:  [k8sobjects]
    processors: [memory_limiter, batch]
    exporters:  [otlphttp/logs]
```

## Exporter Endpoints

| Exporter | Endpoint | Data |
|----------|----------|------|
| `prometheusremotewrite` | `http://victoria-metrics-victoria-metrics-cluster-vminsert.victoria-metrics.svc.cluster.local:8480/insert/0/prometheus/api/v1/write` | Metrics |
| `otlphttp/logs` | `http://victoria-logs-victoria-logs-single-server.victoria-logs.svc.cluster.local:9428/insert/opentelemetry` | Logs |

## Common Processors

| Processor | Config | Purpose |
|-----------|--------|---------|
| `batch` | `timeout: 10s` | Batch telemetry to reduce write pressure |
| `memory_limiter` | `limit: 80%`, `spike: 25%` | Prevent OOM |

## Falco Integration

Falco outputs JSON alerts to stdout. The OTel Agent's `filelog` receiver collects these from the Falco DaemonSet pod logs (`/var/log/pods/falco_*/**/*.log`) and forwards them to VictoriaLogs. This allows querying Falco security events in Grafana alongside application logs.

## How It Connects

```
Every node (Agent DaemonSet)
  ← Container logs from /var/log/pods/
  ← kubelet metrics from :10250/stats/summary
  ← Host metrics (CPU, memory, disk, network)
  → VictoriaMetrics (metrics)
  → VictoriaLogs (logs)

One node (Gateway Deployment)
  ← Kubernetes API (cluster metrics, events)
  → VictoriaMetrics (k8s resource metrics)
  → VictoriaLogs (k8s events as logs)
```

## Troubleshooting

### Agent Not Collecting Logs

```bash
# Check agent is running on all nodes
kubectl get pods -n opentelemetry -o wide

# Check for file permission errors
kubectl logs -n opentelemetry -l app.kubernetes.io/name=otel-agent --tail=50
```

### Missing Metrics in VictoriaMetrics

```bash
# Check remote-write errors
kubectl logs -n opentelemetry -l app.kubernetes.io/name=otel-agent | grep "remote_write\|error"

# Verify vminsert is reachable
kubectl exec -n opentelemetry <agent-pod> -- \
  curl -s -o /dev/null -w "%{http_code}" \
  http://victoria-metrics-victoria-metrics-cluster-vminsert.victoria-metrics.svc.cluster.local:8480
```
