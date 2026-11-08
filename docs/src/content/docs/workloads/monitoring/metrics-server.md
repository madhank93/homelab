---
title: Metrics Server
description: Serves the Kubernetes Metrics API for kubectl top, HPAs and Headlamp, with the flags Talos needs.
sidebar:
  order: 60
---

Metrics Server serves `metrics.k8s.io` from kubelet CPU and memory readings. It keeps only the latest value; history lives in [VictoriaMetrics](/workloads/observability/victoria-metrics/). [Upstream docs](https://github.com/kubernetes-sigs/metrics-server).

| | |
|---|---|
| Namespace | `kube-system` |
| Source | [`metrics_server.go`](src:workloads/monitoring/metrics_server.go) |
| Exposure | — |
| Storage | — |
| Auth | — |

## Configuration

| Setting | Value | Why |
|---|---|---|
| `--kubelet-insecure-tls` | set | Talos kubelets serve self-signed certificates |
| `--kubelet-preferred-address-types` | `InternalIP` | Node hostnames do not resolve in-cluster |
| Toleration | `node-role.kubernetes.io/control-plane` | Can run on control-plane nodes too |

## Troubleshooting

### `kubectl top` says `Metrics API not available`

Metrics Server is down or not yet ready: `kubectl get deploy metrics-server -n kube-system`, then `kubectl top nodes`.
