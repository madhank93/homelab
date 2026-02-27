+++
title = "VictoriaMetrics"
description = "Time-series metrics storage, Prometheus-compatible."
weight = 20
+++

## Overview

| Property | Value |
|----------|-------|
| CDK8s file | `platform/cdk8s/cots/monitoring/victoria_metrics.go` |
| Namespace | `victoria-metrics` |
| HTTPRoute | None |
| UI | No |

## Purpose

VictoriaMetrics is a high-performance, horizontally scalable time-series database. It serves as the cluster's metrics backend, compatible with the Prometheus remote-write protocol.

The OTel Gateway collector sends metrics via Prometheus remote-write to VictoriaMetrics. Grafana queries it using the Prometheus datasource plugin.

## Deployment

Deployed as a VictoriaMetrics Single Server (not cluster mode) — sufficient for a homelab-scale deployment.

## Retention

Default retention is configured via Helm values. Adjust `retentionPeriod` in the CDK8s code as needed.
