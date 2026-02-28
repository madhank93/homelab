+++
title = "Headlamp"
description = "Lightweight Kubernetes dashboard with cluster-admin access."
weight = 20
+++

## Overview

| Property | Value |
|----------|-------|
| CDK8s file | `workloads/management/headlamp.go` |
| Namespace | `headlamp` |
| HTTPRoute | `headlamp.madhan.app` → `headlamp:80` |
| UI | Yes |

## Purpose

Headlamp is a lightweight, extensible Kubernetes dashboard. It provides a clean UI for browsing workloads, events, nodes, and resources.

## Service Account

CDK8s creates a `headlamp-admin` ServiceAccount with `cluster-admin` ClusterRoleBinding. The associated token is stored in a Secret (`headlamp-admin-token`) in the `headlamp` namespace.

## Authentication

Use the token from `headlamp-admin-token` to authenticate in the Headlamp UI:

```bash
kubectl get secret headlamp-admin-token -n headlamp -o jsonpath='{.data.token}' | base64 -d
```

Paste the token into the Headlamp login screen at `http://headlamp.madhan.app`.
