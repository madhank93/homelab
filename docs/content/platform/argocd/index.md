+++
title = "ArgoCD"
description = "ArgoCD Helm bootstrap, ApplicationSet, and sync configuration."
weight = 10
+++

## Overview

ArgoCD is installed once by Pulumi (`core/platform/argocd.go`) then manages itself and all workloads via GitOps from the `v*-manifests` branch.

**Code:** `core/platform/argocd.go` · **Namespace:** `argocd`

## Helm installation

```go
// core/platform/argocd.go
helm.NewRelease(ctx, "argo-cd", &helm.ReleaseArgs{
    Chart:   pulumi.String("argo-cd"),
    Version: pulumi.String("9.4.2"),
    RepositoryOpts: &helm.RepositoryOptsArgs{
        Repo: pulumi.String("https://argoproj.github.io/argo-helm"),
    },
    Namespace:       pulumi.String("argocd"),
    CreateNamespace: pulumi.Bool(true),
    Values: pulumi.Map{
        "server": pulumi.Map{
            "service": pulumi.Map{
                "type": pulumi.String("LoadBalancer"), // Cilium L2 assigns IP
            },
        },
    },
})
```

## ApplicationSet

One `ApplicationSet` watches the `v*-manifests` branch. Every top-level directory under `app/` automatically becomes an ArgoCD Application.

```yaml
apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: homelab-apps
  namespace: argocd
spec:
  goTemplate: true
  generators:
    - git:
        repoURL: https://github.com/madhank93/homelab.git
        revision: v0.1.5-manifests
        directories:
          - path: "*"
  template:
    metadata:
      name: "{{ .path.basename }}"
    spec:
      project: default
      source:
        repoURL: https://github.com/madhank93/homelab.git
        targetRevision: v0.1.5-manifests
        path: "{{ .path.path }}"
      destination:
        server: https://kubernetes.default.svc
        namespace: "{{ .path.basename }}"
      syncPolicy:
        automated:
          prune: true
          selfHeal: true
        syncOptions:
          - ServerSideApply=true
```

## Key settings

| Setting | Value | Reason |
|---|---|---|
| `ServerSideApply=true` | ApplicationSet level | kube-prometheus-stack CRDs exceed the 262 KB annotation limit |
| `ignoreDifferences` on `InfisicalSecret` | per-app | Infisical CRD schema missing `projectSlug` breaks SSA validation |
| `Prune=false` annotation on bootstrap Secrets | per-Secret | Prevents ArgoCD deleting hand-created `infisical-service-token` / `cloudflare-api-token` |

## HTTPRoute (Gateway API)

ArgoCD is exposed via a Gateway API `HTTPRoute` to `argocd.local` inside the LAN.

## Operations

```bash
# Apply ArgoCD config changes
just core platform up

# Manual sync
argocd app sync <app-name>

# List all apps and health
argocd app list

# Re-run bootstrap (after cluster rebuild)
just core platform up
```
