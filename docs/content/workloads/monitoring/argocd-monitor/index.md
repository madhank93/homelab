+++
title = "Argo CD Monitor"
description = "Metrics Services, ServiceMonitors, and the TLS certificate that Argo CD's own chart does not ship."
weight = 15
+++

A small CDK8s chart that supplies two things Argo CD's own Helm chart does not.

Argo CD is installed by Pulumi in the platform layer, and its chart ships no
metrics Services, so VMAgent has nothing to scrape. Adding them from the workloads
layer avoids forking the Helm values or touching the platform stack.

## How It's Used Here

Four Argo CD components are scraped:

| Component | Metrics Port |
|-----------|-------------|
| `argocd-application-controller` | 8082 |
| `argocd-server` | 8083 |
| `argocd-repo-server` | 8084 |
| `argocd-applicationset-controller` | 8085 |

For each component, a `Service` and a `ServiceMonitor` are created in the `argocd`
namespace. The Service selects Argo CD pods by `app.kubernetes.io/name`, and the
ServiceMonitor selects the Service by a `-metrics` label suffix.

## TLS certificate

The same chart requests a cert-manager `Certificate` named `argocd-server-tls`
(letsencrypt-prod, DNS-01 via Cloudflare, for `argocd.madhan.app`). Argo CD
auto-detects a Secret of that name and serves it instead of its self-signed
certificate, which clears the browser TLS warning without exposing Argo CD
publicly.

Source: {{ src(path="workloads/observability/argocd_monitor.go") }}

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

If endpoints are empty, the Service selector does not match any pods — check the Argo CD pod labels:

```bash
kubectl get pods -n argocd --show-labels
```
