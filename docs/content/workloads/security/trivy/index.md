+++
title = "Trivy"
description = "Continuous vulnerability scanning for running workloads."
weight = 20
+++

## Overview

| Property | Value |
|----------|-------|
| CDK8s file | `workloads/security/trivy.go` |
| Namespace | `trivy` |
| HTTPRoute | None |
| UI | No |

## Purpose

Trivy Operator continuously scans all running workloads for:
- Container image vulnerabilities (CVEs)
- Kubernetes configuration issues

Results are stored as custom resources queryable via `kubectl`.

## Querying Reports

```bash
# List vulnerability reports across all namespaces
kubectl get vulnerabilityreports -A

# Inspect a specific report
kubectl describe vulnerabilityreport <name> -n <namespace>

# List config audit reports
kubectl get configauditreports -A
```

## How It Works

Trivy Operator watches for new or updated Pods. When a workload changes, it spawns a scan Job that pulls the image and checks it against the Trivy vulnerability database. Results are stored as `VulnerabilityReport` CRDs in the same namespace as the workload.
