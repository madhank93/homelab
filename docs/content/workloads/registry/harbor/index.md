+++
title = "Harbor"
description = "Enterprise container image registry with vulnerability scanning and pull-through proxy."
weight = 10
+++

## Overview

| Property | Value |
|----------|-------|
| CDK8s file | `workloads/registry/harbor.go` |
| Namespace | `harbor` |
| HTTPRoute | `harbor.madhan.app` → `harbor-core:80` |
| UI | Yes |
| Requires Infisical | Yes — `harbor-admin` Secret |

## Purpose

Harbor provides:
- **Private container registry** — push/pull images from `harbor.madhan.app`
- **Vulnerability scanning** — integrates with Trivy for image scanning on push
- **RBAC** — project-level access control
- **Pull-through proxy cache** — caches images from Docker Hub, GHCR, etc.

## Admin Credentials

The admin password is managed by Infisical at path `/harbor`, synced to the `harbor-admin` Secret (`HARBOR_ADMIN_PASSWORD` key).

## Known Issue: RWO PVC Multi-Attach

Harbor uses multiple RWO PVCs from Longhorn. During ArgoCD sync rolling updates, pods from the old ReplicaSet may still hold the PVC, causing `Multi-Attach error` for the new pod.

**Force-delete the stuck pod:**

```bash
kubectl delete pod -n harbor -l "app=harbor,component=jobservice" --grace-period=0 --force
```

This happens on every ArgoCD sync that updates the Harbor Deployment. The `Recreate` strategy would fix it but Harbor's chart uses `RollingUpdate` by default.

## Configuring Pull-Through Proxy

1. Log in to `http://harbor.madhan.app` as admin
2. Go to **Administration → Registries → New Endpoint**
3. Add Docker Hub, GHCR, or other registries
4. Create a proxy project pointing to the registry endpoint
