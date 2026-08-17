+++
title = "Metrics Server"
description = "Kubernetes Metrics Server — required for kubectl top, HPA, and Headlamp resource views."
weight = 60
+++

[Metrics Server](https://github.com/kubernetes-sigs/metrics-server) serves the
Kubernetes Metrics API (`metrics.k8s.io`) by scraping CPU and memory from the
kubelets. It keeps only the latest data point and stores no history — anything
historical comes from
[VictoriaMetrics](@/workloads/monitoring/victoria-metrics/index.md) instead.

Without it, `kubectl top` returns `error: Metrics API not available`, Headlamp
shows no resource usage, and any HPA or VPA has nothing to scale on.

It is deployed as a CDK8s chart into `kube-system` with two flags that Talos
specifically requires:

| Flag | Why |
|---|---|
| `--kubelet-insecure-tls` | Talos kubelets serve self-signed certificates, which the default config rejects |
| `--kubelet-preferred-address-types=InternalIP` | Node hostnames do not resolve in-cluster |

Source: {{ src(path="workloads/monitoring/metrics_server.go") }}

```bash
# Verify Metrics Server is running
kubectl get deployment metrics-server -n kube-system

# Test it works
kubectl top nodes
kubectl top pods -A
```

## Relationship to Observability Stack

Metrics Server and VictoriaMetrics serve different purposes and are complementary:

| | Metrics Server | VictoriaMetrics |
|--|----------------|-----------------|
| Data retention | Last value only | 30 days |
| Scrape interval | 60s | 30s |
| Use case | `kubectl top`, HPA | Grafana, alerting, trending |
| Storage | In-memory | Longhorn PVC (100 Gi) |
| Protocol | Kubernetes Metrics API | Prometheus remote-write |
