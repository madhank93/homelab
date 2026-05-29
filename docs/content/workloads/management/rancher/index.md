+++
title = "Rancher"
description = "Removed in v0.1.6."
weight = 10
+++

> **Removed in v0.1.6.** Rancher was removed from CDK8s workloads. See [v0.1.6 Release Notes](/releases/v0.1.6/) for details.

## Why It Was Removed

Rancher added ~1000m CPU overhead per replica and caused double-reconciliation conflicts with ArgoCD for Fleet-managed apps. ArgoCD + Headlamp cover all cluster management needs for this homelab without the overhead.

**Removed files:**
- `workloads/management/rancher.go`
- `workloads/imports/rancher/`
- `workloads/imports/fleet/`

Use [Headlamp](/workloads/management/headlamp/) for Kubernetes UI and [ArgoCD](/infrastructure/argocd/) for GitOps.
