+++
title = "Trivy"
description = "Continuous vulnerability scanning for running workloads via Trivy Operator."
weight = 20
+++

## What is Trivy Operator?

[Trivy Operator](https://aquasecurity.github.io/trivy-operator/) is a Kubernetes operator by Aqua Security that continuously scans all running workloads for container image vulnerabilities (CVEs), Kubernetes configuration issues, and compliance violations. Results are stored as Kubernetes Custom Resources queryable with `kubectl`.

## Why Trivy?

Trivy is the most widely adopted open-source vulnerability scanner. The Operator pattern means scanning happens automatically whenever a workload changes — no manual scans required, no separate CI pipeline integration needed.

| Tool | Scan type | When |
|------|-----------|------|
| **Trivy Operator** | Running workloads in cluster | Continuous, on pod change |
| Falco | Syscall behavior | Runtime (live) |
| Harbor's Trivy | Images on push to registry | On registry push |

## How It's Used Here

Trivy Operator runs as a single-replica Deployment in the `trivy` namespace. It watches for new or updated Pods and spawns scan Jobs that pull the image and check it against the Trivy vulnerability database.

Source: [`workloads/security/trivy.go`](https://github.com/madhank93/homelab/blob/v0.1.5/workloads/security/trivy.go)

## Configuration

| Setting | Value | Why |
|---------|-------|-----|
| Namespace | `trivy` | Isolated namespace |
| Chart version | `trivy-operator` v0.32.0 | Pinned version |
| Replicas | `1` | Single operator instance |
| `scanJobsConcurrentLimit` | `3` | Max parallel scan Jobs |
| `scanJobsRetryDelay` | `30s` | Wait between retries on scan failure |
| `scanJobTimeout` | `5m` | Timeout for each scan Job |
| `serviceMonitor.enabled` | `true` | VMAgent scrapes Trivy metrics |
| Resources (limit) | `500m` / `512Mi` | Operator itself is lightweight; scan Jobs are separate |

## What It Scans

| Scan type | CRD | Description |
|-----------|-----|-------------|
| Container images | `VulnerabilityReport` | CVEs in container images of running pods |
| Configuration | `ConfigAuditReport` | Kubernetes security misconfigs (e.g., no resource limits, privileged containers) |
| Exposed secrets | `ExposedSecretReport` | Secrets found in image layers or environment variables |
| SBOM | `SbomReport` | Software Bill of Materials for image layers |

## Querying Reports

```bash
# List vulnerability reports across all namespaces
kubectl get vulnerabilityreports -A

# Summary: just namespace + name + critical/high counts
kubectl get vulnerabilityreports -A \
  -o custom-columns="NS:.metadata.namespace,NAME:.metadata.name,CRITICAL:.report.summary.criticalCount,HIGH:.report.summary.highCount"

# Inspect a specific report
kubectl describe vulnerabilityreport <name> -n <namespace>

# List config audit reports
kubectl get configauditreports -A

# List exposed secret reports
kubectl get exposedsecretreports -A
```

## How It Connects

```
New/updated Pod in any namespace
  → Trivy Operator detects change
  → Spawns scan Job (pulls image, checks against vulnerability DB)
  → Stores results as VulnerabilityReport CR in same namespace
  → VMAgent scrapes Trivy metrics via ServiceMonitor
  → Grafana dashboard (optional — CRD-based data, not metrics-based)
```

## Troubleshooting

### Scan Jobs Failing

**Symptoms:** VulnerabilityReport not created after deploying a new workload.

```bash
# Check scan jobs
kubectl get jobs -A | grep trivy

# Check job logs
kubectl logs -n <target-namespace> job/<trivy-scan-job>
```

**Common causes:** Image pull failure (private registry without credentials), scan timeout (large image).

### CRDs Not Present

Trivy Operator's CRDs are downloaded at CDK8s synthesis time from the Helm chart tarballs. If `TRIVY_OPERATOR_SKIP_CRDS=true` is set in the environment, CRDs are skipped.

```bash
# Verify CRDs are installed
kubectl get crd | grep aquasecurity
```
