---
title: Headlamp
description: Kubernetes dashboard on the LAN, logged in with short-lived cluster-admin tokens.
sidebar:
  order: 10
  badge: { text: LAN, variant: note }
---

Headlamp is the cluster's only UI, for browsing resources, logs and events; changes still go through code and [Argo CD](/platform/argocd/). [Upstream docs](https://headlamp.dev/docs/latest/).

| | |
|---|---|
| Namespace | `headlamp` |
| Source | [`headlamp.go`](src:workloads/management/headlamp.go) |
| Exposure | LAN · `headlamp.madhan.app` → `headlamp:80` |
| Storage | — |
| Auth | Token for ServiceAccount `headlamp-admin` (bound to `cluster-admin`) |

## Configuration

| Setting | Value | Why |
|---|---|---|
| `ingress.enabled` | `false` | Exposed through the Gateway instead |
| ServiceAccount + ClusterRoleBinding | `headlamp-admin` → `cluster-admin` | Login identity; no static token Secret exists |
| Resources | limits `500m` / `512Mi` | Small dashboard |

## Operate

Mint a token (default lifetime `24h`) and paste it into the login screen:

```bash frame="terminal"
just headlamp-token        # or: just headlamp-token 8h
```

CPU and memory columns need [Metrics Server](/workloads/monitoring/metrics-server/). From outside the LAN, connect through the [NetBird VPN](/cloud/netbird/).

![Headlamp cluster overview showing namespace list, pod health, and node resources](/assets/screenshots/headlamp/cluster-overview.png)

## Troubleshooting

### Token rejected

It expired, or `headlamp-admin` was recreated. Run `just headlamp-token` again.

### Resource columns show N/A

Metrics Server is down. Check `kubectl top nodes`.
