+++
title = "Headlamp"
description = "Lightweight Kubernetes dashboard with cluster-admin access. Exposed externally via NetBird reverse proxy + Authentik SSO."
weight = 20
+++

## Overview

| Property | Value |
|----------|-------|
| CDK8s file | `workloads/management/headlamp.go` |
| Namespace | `headlamp` |
| HTTPRoute | `headlamp.madhan.app` → `headlamp:80` (LAN only) |
| External access | `headlamp.proxy.madhan.app` via NetBird expose (SSO gated) |
| UI | Yes |

## Purpose

Headlamp is a lightweight, extensible Kubernetes dashboard providing a clean UI for browsing workloads, events, nodes, and resources across the cluster.

## Service Account

CDK8s creates a `headlamp-admin` ServiceAccount with `cluster-admin` ClusterRoleBinding. The token is stored in `headlamp-admin-token` Secret in the `headlamp` namespace.

## Authentication

Use the long-lived token for login:

```bash
kubectl get secret headlamp-admin-token -n headlamp -o jsonpath='{.data.token}' | base64 -d
```

Paste the token into the Headlamp login screen.

## External Access (Internet)

Headlamp is exposed to the internet via the **NetBird v0.66 reverse proxy** with **Authentik SSO group-based access control**. Only members of the `homelab-admins` group can reach the UI at `https://headlamp.proxy.madhan.app`.

→ [Exposing Headlamp via NetBird Reverse Proxy](./netbird-expose)
