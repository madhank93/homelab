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
| HTTPRoute | `harbor.madhan.app` → `harbor:80` (nginx proxy) |
| UI | Yes |
| Secrets | OpenBao `secret/data/harbor` → CSI → `harbor-admin` Secret |

## Purpose

Harbor provides:
- **Private container registry** — push/pull images from `harbor.madhan.app`
- **Vulnerability scanning** — integrates with Trivy for image scanning on push
- **RBAC** — project-level access control
- **Pull-through proxy cache** — caches images from Docker Hub, GHCR, etc.

## Admin Credentials

The admin password is stored in OpenBao at `secret/data/harbor` (`HARBOR_ADMIN_PASSWORD` key) and synced to the `harbor-admin` k8s Secret via the Secrets Store CSI Driver and a `secret-sync` Deployment (Harbor's Helm chart does not support `extraVolumes`, so a dedicated sync pod triggers the secretObjects sync).

## Service Routing

Traffic enters via the `harbor` ClusterIP service (port 80) which is Harbor's built-in nginx proxy. The nginx service routes internally to portal, core, registry, and jobservice. The HTTPRoute backend must point to `harbor:80`, **not** `harbor-core:80`.

## Known Issue: RWO PVC Multi-Attach

Harbor's `jobservice` and `registry` components use RWO PVCs (Longhorn). During ArgoCD sync rolling updates, pods from the old ReplicaSet may still hold the PVC, blocking the new pod with `Multi-Attach error`.

**Fix — force-delete stuck pods and scale down old ReplicaSet:**

```bash
# Force-delete all jobservice and registry pods
kubectl delete pod -n harbor -l "app=harbor,component=jobservice" --grace-period=0 --force
kubectl delete pod -n harbor -l "app=harbor,component=registry" --grace-period=0 --force

# If two ReplicaSets both have DESIRED=1, scale down the old one
kubectl get replicasets -n harbor -l "app=harbor,component=jobservice"
kubectl scale replicaset <old-rs-name> -n harbor --replicas=0
```

This happens on every ArgoCD sync that updates the Harbor Deployment. The `Recreate` strategy would fix it permanently but Harbor's chart uses `RollingUpdate` by default.

## Configuring Pull-Through Proxy

1. Log in to `http://harbor.madhan.app` as admin
2. Go to **Administration → Registries → New Endpoint**
3. Add Docker Hub, GHCR, or other registries
4. Create a proxy project pointing to the registry endpoint
