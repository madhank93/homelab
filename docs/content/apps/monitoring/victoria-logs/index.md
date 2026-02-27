+++
title = "VictoriaLogs"
description = "Log storage with Loki-compatible query API."
weight = 30
+++

## Overview

| Property | Value |
|----------|-------|
| CDK8s file | `platform/cdk8s/cots/monitoring/victoria_logs.go` |
| Namespace | `victoria-logs` |
| HTTPRoute | None |
| UI | No |

## Purpose

VictoriaLogs stores container and host logs from the OTel Agent DaemonSet. It exposes a Loki-compatible API, enabling Grafana to query logs using the Loki datasource plugin and LogQL.

## Ingestion

The OTel Agent on each node collects container logs via the `filelog` receiver and forwards them to VictoriaLogs via OTLP/HTTP.

Falco runtime alerts are also collected by the OTel Agent (from stdout) and forwarded here.

## Querying

From Grafana, select the VictoriaLogs datasource and use LogQL:

```
{namespace="comfyui"}
{namespace="ollama"} |= "error"
```
