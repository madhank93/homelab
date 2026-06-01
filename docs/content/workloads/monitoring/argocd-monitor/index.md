+++
title = "ArgoCD Monitor"
description = "Headless metrics Services and ServiceMonitors for ArgoCD component scraping by VMAgent."
weight = 15
+++

## What is ArgoCD Monitor?

ArgoCD Monitor is a small CDK8s chart that creates the missing metrics plumbing for ArgoCD. ArgoCD's Helm chart does not expose metrics Services by default, so VMAgent cannot scrape any ArgoCD component. This chart fills that gap.

## Why ArgoCD Monitor?

ArgoCD is deployed by Pulumi in the platform layer. Its Helm chart disables metrics services by default. Rather than forking the Helm values or patching the platform layer, this workloads-layer chart adds the four headless metrics Services and matching ServiceMonitors so VMAgent discovers and scrapes them without any changes to the ArgoCD deployment itself.

## How It's Used Here

Four ArgoCD components are scraped:

| Component | Metrics Port |
|-----------|-------------|
| `argocd-application-controller` | 8082 |
| `argocd-server` | 8083 |
| `argocd-repo-server` | 8084 |
| `argocd-applicationset-controller` | 8085 |

For each component, a `Service` and a `ServiceMonitor` are created in the `argocd` namespace. The Service selects ArgoCD pods by `app.kubernetes.io/name`, and the ServiceMonitor selects the Service by a `-metrics` label suffix.

Source: [`workloads/observability/argocd_monitor.go`](https://github.com/madhank93/homelab/blob/v0.1.5/workloads/observability/argocd_monitor.go)

## Configuration

| Setting | Value |
|---------|-------|
| Namespace | `argocd` |
| Scrape interval | `30s` |
| Path | `/metrics` |

## Troubleshooting

### Metrics Not Appearing in VMAgent

```bash
# Check ServiceMonitors exist
kubectl get servicemonitor -n argocd

# Check Services exist and have endpoints
kubectl get svc -n argocd | grep metrics
kubectl get endpoints -n argocd | grep metrics
```

If endpoints are empty, the Service selector does not match any pods — check the ArgoCD pod labels:

```bash
kubectl get pods -n argocd --show-labels
```
