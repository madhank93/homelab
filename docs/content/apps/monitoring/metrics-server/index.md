+++
title = "Metrics Server"
description = "Kubernetes Metrics Server — required for kubectl top, HPA, and Headlamp resource views."
weight = 60
+++

## What is Metrics Server?

[Metrics Server](https://github.com/kubernetes-sigs/metrics-server) is a cluster-wide aggregator of resource usage data. It collects CPU and memory metrics from kubelets and exposes them via the Kubernetes Metrics API (`metrics.k8s.io`).

Metrics Server is **not** a monitoring solution — it only keeps the most recent data point per pod/node and does not persist history. For historical metrics, VictoriaMetrics is used.

## Why It's Needed

Several Kubernetes features depend on the Metrics API:

| Feature | Needs Metrics Server |
|---------|---------------------|
| `kubectl top nodes` | Yes |
| `kubectl top pods` | Yes |
| HPA (Horizontal Pod Autoscaler) | Yes |
| Headlamp resource views | Yes |
| VPA (Vertical Pod Autoscaler) | Yes |

Without Metrics Server, `kubectl top` returns `error: Metrics API not available` and Headlamp cannot show pod CPU/memory usage.

## How It's Deployed

Metrics Server is typically deployed via the Talos cluster bootstrap or as a separate workload. On Talos Linux, the default installation may need the `--kubelet-insecure-tls` flag because Talos kubelets use self-signed certificates.

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
