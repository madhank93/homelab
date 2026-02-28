+++
title = "Rancher"
description = "Multi-cluster Kubernetes management UI."
weight = 10
+++

## Overview

| Property | Value |
|----------|-------|
| CDK8s file | `workloads/management/rancher.go` |
| Namespace | `cattle-system` |
| HTTPRoute | `rancher.madhan.app` → `rancher:80` |
| UI | Yes |
| Requires Infisical | Yes — `rancher-bootstrap` Secret |

## Purpose

Rancher provides fleet-level visibility into workloads, nodes, and cluster events. It includes the Rancher Fleet GitOps agent for multi-cluster management.

## Bootstrap Password

The initial admin password is managed by Infisical at path `/rancher`, synced to the `rancher-bootstrap` Secret. Rancher reads `BOOTSTRAP_PASSWORD` from this Secret on first startup.

## Known Issue: Rolling Update Deadlock

Harbor and Rancher both use RWO (ReadWriteOnce) PVCs from Longhorn. During ArgoCD syncs, a rolling update can cause Multi-Attach errors when the old pod's PVC is still mounted.

**Workaround:**

```bash
kubectl delete pod -n cattle-system -l "app=rancher" --grace-period=0 --force
```
