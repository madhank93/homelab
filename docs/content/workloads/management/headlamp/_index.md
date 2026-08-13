+++
title = "Headlamp"
description = "Lightweight Kubernetes dashboard with cluster-admin access. LAN-accessible, external access via NetBird reverse proxy."
weight = 20
+++

## What is Headlamp?

[Headlamp](https://headlamp.dev/) is a lightweight, extensible Kubernetes dashboard providing a clean web UI for browsing workloads, events, nodes, storage, RBAC, and custom resources. It is plugin-based and actively maintained by the CNCF ecosystem.

## Why Headlamp?

Headlamp is faster and lighter than the standard Kubernetes Dashboard, and it is
the only cluster UI here. Anything it cannot do is done through Argo CD or
`kubectl` — the cluster is GitOps-managed, so changes belong in code rather than
in a console.

| Tool | Use case |
|------|----------|
| Headlamp | Quick browsing, logs, events, resource inspection |
| Argo CD | Sync state, diffs, rollbacks |

## How It's Used Here

Headlamp runs in the `headlamp` namespace, accessible at `https://headlamp.madhan.app`. It uses a long-lived ServiceAccount token with `cluster-admin` ClusterRoleBinding.

Source: {{ src(path="workloads/management/headlamp.go") }}

## Configuration

| Setting | Value | Why |
|---------|-------|-----|
| Namespace | `headlamp` | Isolated namespace |
| HTTPRoute | `headlamp.madhan.app` → `headlamp:80` | Gateway API (LAN only) |
| ServiceAccount | `headlamp-admin` | cluster-admin access |
| ClusterRoleBinding | `cluster-admin` | Full read/write access to all resources |
| Token Secret | `headlamp-admin-token` | Long-lived SA token (kubernetes.io/service-account-token) |
| Resources (limits) | `500m` / `512Mi` | Lightweight dashboard |

## Authentication

Headlamp uses token-based authentication. Retrieve the long-lived token:

```bash
kubectl get secret headlamp-admin-token -n headlamp \
  -o jsonpath='{.data.token}' | base64 -d
```

Paste the token into the Headlamp login screen.

## Resource Views

Headlamp shows pod CPU and memory usage if Metrics Server is installed. Without Metrics Server, the resource columns show `N/A`.

## External Access

Headlamp is LAN-only by default (no Cloudflare A record). For temporary external access, use the NetBird reverse proxy feature.

→ [Exposing Headlamp via NetBird Reverse Proxy](./netbird-expose)

## Screenshots

![Headlamp cluster overview showing namespace list, pod health, and node resources](/assets/screenshots/headlamp/cluster-overview.png)

## Troubleshooting

### Token Expired / Invalid

Long-lived SA tokens do not expire (unlike projected tokens), but if the ServiceAccount is deleted and recreated, the token must be re-retrieved:

```bash
kubectl get secret headlamp-admin-token -n headlamp \
  -o jsonpath='{.data.token}' | base64 -d
```

### Resource Metrics Not Showing

Ensure Metrics Server is running:

```bash
kubectl get deployment metrics-server -n kube-system
kubectl top nodes  # should return data
```
