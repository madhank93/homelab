+++
title = "Kyverno"
description = "Kyverno policy engine — admission control, background scanning, and Grafana dashboard integration."
weight = 10
+++

[Kyverno](https://kyverno.io/) is the cluster's admission controller: it sits in
front of every resource create and update and can reject or mutate it. Policies
are ordinary Kubernetes CRDs, so they ship through the same CDK8s → Argo CD path
as everything else, with no Rego to learn.

It is the only **preventive** control here — Trivy and Falco both report on things
that are already running.

> **No policies are deployed yet.** Kyverno is installed and wired into monitoring,
> but the cluster currently enforces nothing. Adding a `ClusterPolicy` is what turns
> it on; until then it is an admission webhook that approves everything.

Because the webhook is fail-closed, Kyverno being down stops *all* resource
creates and updates cluster-wide — which is why the admission controller runs
three replicas while the rest run two.

Source: {{ src(path="workloads/security/keyverno.go") }}

## Configuration

| Component | Replicas | CPU Limit | Memory Limit |
|-----------|----------|-----------|--------------|
| Admission Controller | 3 | 1000m | 512Mi | 
| Background Controller | 2 | 500m | 256Mi |
| Cleanup Controller | 2 | 500m | 256Mi |
| Reports Controller | 2 | 500m | 256Mi |

| Setting | Value | Why |
|---------|-------|-----|
| `metricsService.port` | `8000` | VMAgent scrape target |
| `webhooksCleanup.enabled` | `true` | Removes the webhooks on uninstall — without this, deleting Kyverno leaves a fail-closed webhook that blocks every write to the cluster |
| `policyExceptions.enabled` | `true` | Lets a specific workload opt out without weakening the policy for everyone |
| `imageVerification.enabled` | `false` | No Cosign signing in this homelab |
| `grafana.enabled` | `true` | Ships the dashboard as a ConfigMap for Grafana's sidecar |

Metrics are scraped by VMAgent via a ServiceMonitor on port 8000.

## Troubleshooting

### Admission Webhook Timeout

```bash
kubectl get validatingwebhookconfigurations | grep kyverno
kubectl logs -n kyverno -l app.kubernetes.io/name=kyverno-admission-controller
```

The webhook is fail-closed: if the admission controller is unreachable, every
create and update in the cluster fails, including the ones that would fix it.
Restarting the pods usually clears it.

### Policy Violation Report

```bash
kubectl get policyreport -A
kubectl describe policyreport <name> -n <namespace>
```
