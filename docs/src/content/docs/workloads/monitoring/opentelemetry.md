---
title: OpenTelemetry
description: Two OpenTelemetry Collectors — a per-node agent for logs and node metrics, and a gateway for cluster metrics and events.
sidebar:
  order: 50
---

Two OpenTelemetry Collector releases ship metrics to VictoriaMetrics and logs to VictoriaLogs. [Upstream docs](https://opentelemetry.io/docs/collector/).

| | |
|---|---|
| Namespace | `opentelemetry` (privileged: hostPath and hostPorts) |
| Source | [`otel_collector.go`](src:workloads/observability/otel_collector.go) |
| Exposure | — |
| Storage | hostPath `/var/lib/otelcol` on each node (log read offsets) |
| Auth | — |

## How it fits

| Release | Mode | Receivers | Exports |
|---|---|---|---|
| `otel-agent` | DaemonSet, every node (`tolerations: Exists`) | `filelog` (`/var/log/pods`), `kubeletstats`, `hostmetrics` | Logs → VictoriaLogs, metrics → VictoriaMetrics |
| `otel-gateway` | Deployment, 1 replica | `k8s_cluster`, `k8sobjects` (events) | Metrics → VictoriaMetrics, events as logs → VictoriaLogs |

Falco writes JSON alerts to stdout, so the agent's `filelog` receiver picks them up with every other container log.

## Configuration

| Setting | Value | Why |
|---|---|---|
| Image | `otel/opentelemetry-collector-contrib` | Carries the Kubernetes receivers |
| `kubeletstats.endpoint` | `https://${env:K8S_NODE_IP}:10250`, `insecure_skip_verify` | The preset uses the node name, which does not resolve in-cluster on Talos |
| `file_storage` extension | `/var/lib/otelcol` on a hostPath | Restarts resume from saved offsets instead of re-reading old logs |
| `prometheusremotewrite` | `vmsingle-vm-stack…:8428/api/v1/write`, `resource_to_telemetry_conversion` | Resource attributes such as `k8s.pod.name` become labels |
| `k8sattributes.passthrough` (gateway) | `true` | Events have no pod IP or UID; the default drops all of them |
| `memory_limiter` | 80%, spike 25% | |
| `batch` | `10s` | |

## Troubleshooting

### Logs or metrics missing

Check every node has an agent and the exporters log no errors:

```bash frame="terminal"
kubectl get pods -n opentelemetry -o wide
kubectl logs -n opentelemetry -l app.kubernetes.io/instance=otel-agent --tail=50 | grep -i error
```

### Kubernetes events missing from VictoriaLogs

`k8sattributes.passthrough` is not `true` on the gateway, so every event is dropped before export.
