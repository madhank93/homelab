+++
title = "Grafana"
description = "Dashboards and data visualization for metrics and logs."
weight = 10
+++

## Overview

| Property | Value |
|----------|-------|
| CDK8s file | `workloads/monitoring/grafana.go` |
| Namespace | `grafana` |
| HTTPRoute | `grafana.madhan.app` → `grafana:80` |
| UI | Yes |
| Requires Infisical | Yes — `grafana-admin` Secret |

## Purpose

Grafana provides dashboards for metrics (VictoriaMetrics) and logs (VictoriaLogs). Datasources are auto-provisioned via Helm values.

## Datasources

| Name | Type | URL |
|------|------|-----|
| VictoriaMetrics | Prometheus | `http://victoria-metrics-victoria-metrics-single-server.victoria-metrics:8428` |
| VictoriaLogs | Loki | `http://victoria-logs-single-server.victoria-logs:9428` |

## Secret

Grafana's admin password is managed by Infisical:

```yaml
apiVersion: secrets.infisical.com/v1alpha1
kind: InfisicalSecret
metadata:
  name: grafana-admin
  namespace: grafana
  annotations:
    argocd.argoproj.io/sync-options: ServerSideApply=false
spec:
  authentication:
    serviceToken:
      secretsScope:
        secretsPath: /grafana
  managedSecretReference:
    secretName: grafana-admin
    secretNamespace: grafana
```

The `grafana-admin` Secret must exist before Grafana starts.
