+++
title = "Kyverno"
description = "Kyverno policy engine — admission control, background scanning, and Grafana dashboard integration."
weight = 10
+++

## What is Kyverno?

[Kyverno](https://kyverno.io/) is a Kubernetes-native policy engine that validates, mutates, and generates resources using policies written as Kubernetes CRDs — no OPA/Rego required. It operates as a validating and mutating admission webhook.

## Why Kyverno?

Kyverno policies are Kubernetes resources (YAML), so they live in the same GitOps repo and follow the same ArgoCD sync workflow as everything else. The admission controller intercepts every resource creation/update, making it the right place to enforce security standards (e.g., require non-root containers, disallow `latest` tags) without custom admission webhooks.

## How It's Used Here

Kyverno runs in HA mode in its own namespace. It integrates with:

- **VMAgent** — a ServiceMonitor scrapes Kyverno metrics on port 8000 every 30s
- **Grafana** — the Helm chart creates a ConfigMap with the Kyverno dashboard JSON; Grafana's sidecar picks it up automatically

Source: [`workloads/security/keyverno.go`](https://github.com/madhank93/homelab/blob/v0.1.5/workloads/security/keyverno.go)

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
| `webhooksCleanup.enabled` | `true` | Cleans up webhooks on uninstall |
| `policyExceptions.enabled` | `true` | Allows per-resource policy exemptions |
| `imageVerification.enabled` | `false` | Not using Cosign image signing |
| `grafana.enabled` | `true` | Grafana dashboard ConfigMap |
| `podSecurityContext.runAsNonRoot` | `true` | Non-root containers |

## Troubleshooting

### Admission Webhook Timeout

```bash
kubectl get validatingwebhookconfigurations | grep kyverno
kubectl logs -n kyverno -l app.kubernetes.io/name=kyverno-admission-controller
```

If the admission controller is down, all resource creates/updates will fail (fail-closed webhook). Restarting the pods usually resolves this.

### Policy Violation Report

```bash
kubectl get policyreport -A
kubectl describe policyreport <name> -n <namespace>
```
