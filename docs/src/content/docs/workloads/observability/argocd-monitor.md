---
title: Argo CD monitor
description: Metrics Services, ServiceMonitors and the TLS certificate that the Pulumi-installed Argo CD does not ship.
sidebar:
  order: 15
---

A CDK8s chart that adds what the Pulumi-installed Argo CD lacks: metrics Services for VMAgent to scrape and a trusted TLS certificate. Adding them here avoids changing the platform stack.

| | |
|---|---|
| Namespace | `argocd` |
| Source | [`argocd_monitor.go`](src:workloads/observability/argocd_monitor.go) |
| Exposure | — |
| Storage | — |
| Auth | — |

## Configuration

Each component gets a `<name>-metrics` Service and a ServiceMonitor scraping `/metrics` every `30s`.

| Component | Metrics port |
|---|---|
| `argocd-application-controller` | 8082 |
| `argocd-server` | 8083 |
| `argocd-repo-server` | 8084 |
| `argocd-applicationset-controller` | 8085 |

The cert-manager `Certificate` `argocd-server-tls` (issuer `letsencrypt-prod`, `argocd.madhan.app`) writes the Secret Argo CD looks for by name, replacing its self-signed certificate.

## Troubleshooting

### Argo CD metrics missing

Empty endpoints mean the Service selector matches no pod; compare with the pod labels.

```bash frame="terminal"
kubectl get servicemonitor,endpoints -n argocd | grep metrics
kubectl get pods -n argocd --show-labels
```
