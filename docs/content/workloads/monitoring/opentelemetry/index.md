+++
title = "OpenTelemetry"
description = "Two-tier OTel collection pipeline: DaemonSet agent per node and Gateway deployment."
weight = 50
+++

## Overview

| Property | Value |
|----------|-------|
| CDK8s file | `workloads/observability/otel_collector.go` |
| Namespace | `opentelemetry` |
| HTTPRoute | None |
| UI | No |

## Architecture

Two OTel Collector deployments form a collection pipeline:

```
┌──────────────────────────────────────────┐
│  Each Node                                │
│  OTel Agent (DaemonSet)                  │
│   ├── filelog receiver → container logs  │
│   ├── kubeletstats receiver → pod metrics│
│   └── hostmetrics receiver → CPU/RAM/etc │
└──────────────┬───────────────────────────┘
               │ OTLP/gRPC
               ▼
┌──────────────────────────────────────────┐
│  OTel Gateway (Deployment)               │
│   ├── k8s_cluster receiver → k8s events  │
│   └── Exports:                           │
│       ├── metrics → VictoriaMetrics       │
│       │   (Prometheus remote-write)       │
│       └── logs → VictoriaLogs             │
│           (OTLP/HTTP)                    │
└──────────────────────────────────────────┘
```

## Agent DaemonSet

Runs on every node. Collects:
- Container stdout/stderr logs (via `filelog` receiver watching `/var/log/pods/`)
- kubelet metrics (node, pod, container)
- Host metrics (CPU, memory, disk, network)

Pod security is `privileged` (required for host metrics and log file access on Talos).

## Gateway Deployment

Cluster-level collection:
- Kubernetes resource attributes enrichment
- Kubernetes events
- Forwards batched telemetry to VictoriaMetrics and VictoriaLogs

## Falco Integration

Falco outputs runtime alerts as JSON to stdout. The OTel Agent's filelog receiver collects these from the Falco pod logs and forwards them to VictoriaLogs for alerting and long-term storage.
