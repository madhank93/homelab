---
title: VictoriaMetrics and VictoriaLogs
description: Why metrics and logs go to single-node VictoriaMetrics and VictoriaLogs rather than Prometheus and Loki.
---
**Decision:** VictoriaMetrics single-node (VMSingle) for metrics and VictoriaLogs single-node for logs, both read by Grafana.

**Context:** The cluster runs many workloads on limited RAM and on Longhorn disks shared by every volume. The stores must accept Prometheus remote write and OTLP, and work with Grafana's stock Prometheus and Loki datasources.

| Option | For | Against |
|---|---|---|
| **VictoriaMetrics + VictoriaLogs** | Several times less memory and disk than Prometheus; long retention built in; PromQL-compatible (MetricsQL); one vendor's operator, CRDs and naming | Grafana reads logs through the Loki plugin, which VictoriaLogs implements only partly (no volume API) |
| Prometheus + Loki | Reference implementations; widest ecosystem | Higher memory; long-term storage needs Thanos or Cortex; Loki adds object storage or more components |
| VMCluster (insert/select/storage) | Horizontal scale | Three components to run for one cluster's worth of metrics |

**Consequences:**
- The Grafana VictoriaLogs datasource needs `logsVolumeEnabled: false` and a URL ending at `/select`.
- ServiceMonitor and PrometheusRule CRDs are installed by hand so app charts' monitors still work; VMOperator converts them.
- One VMSingle is a single point of failure for metrics.
