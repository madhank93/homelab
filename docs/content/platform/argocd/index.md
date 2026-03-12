+++
title = "ArgoCD"
description = "ArgoCD Helm bootstrap, ApplicationSet directory generator, and sync configuration."
weight = 10
+++

## Screenshots

![ArgoCD application list showing all workloads with sync status and health indicators](/assets/screenshots/argocd/app-list.png)

## Overview

ArgoCD is installed once by Pulumi (`core/platform/argocd.go`) and then manages all workloads via GitOps from the `v0.1.5-manifests` branch. It self-manages after the initial bootstrap.

**Code:** [`core/platform/argocd.go`](https://github.com/madhank93/homelab/blob/v0.1.5/core/platform/argocd.go) · **Namespace:** `argocd` · **Chart version:** `9.4.2`

## Helm Installation

```go
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

One `ApplicationSet` named `cots-applications` watches the `v0.1.5-manifests` branch. Every top-level directory automatically becomes an ArgoCD Application:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: cots-applications
  namespace: argocd
spec:
  generators:
    - git:
        repoURL: https://github.com/madhank93/homelab.git
        revision: v0.1.5-manifests
        directories:
          - path: "*"      # Every top-level directory = one Application
  template:
    metadata:
      name: "{{path.basename}}"
    spec:
      project: default
      source:
        repoURL: https://github.com/madhank93/homelab.git
        targetRevision: v0.1.5-manifests
        path: "{{path}}"
      destination:
        server: https://kubernetes.default.svc
        namespace: "{{path.basename}}"
      syncPolicy:
        automated:
          prune: true
          selfHeal: true
        syncOptions:
          - CreateNamespace=true
          - ServerSideApply=true
```

## Key Settings

| Setting | Value | Reason |
|---------|-------|--------|
| `ServerSideApply=true` | ApplicationSet-level | kube-prometheus-stack CRDs exceed the 262 KB `kubectl.kubernetes.io/last-applied-configuration` annotation limit |
| `automated.prune=true` | ApplicationSet-level | Resources removed from manifests are deleted from cluster |
| `automated.selfHeal=true` | ApplicationSet-level | Manual kubectl changes are reverted |
| `Prune=false` on bootstrap Secrets | Per-Secret annotation | Prevents ArgoCD deleting `openbao-unseal-key` and `cloudflare-api-token` |

## HTTPRoutes

ArgoCD is exposed via two routes:

| Route | URL | Purpose |
|-------|-----|---------|
| HTTPRoute | `argocd.local` | LAN HTTP access |
| TLSRoute | `argocd.madhan.app` | TLS passthrough to argocd-server:443 |

## `ignoreDifferences`

The ApplicationSet ignores fields that change dynamically and would otherwise cause permanent OutOfSync:

| Resource | Ignored fields | Why |
|----------|---------------|-----|
| `Secret` (VM operator validation) | `/data` | Operator regenerates TLS cert on every restart |
| `ValidatingWebhookConfiguration` (VM operator) | `.webhooks[].clientConfig.caBundle` | CA bundle updated dynamically by operator |
| All `StatefulSet` resources | `.spec.volumeClaimTemplates[].apiVersion`, `.spec.volumeClaimTemplates[].kind` | CDK8s generates these fields; Kubernetes strips them on admission |

## Bootstrap Secrets (`Prune=false`)

Two Secrets are created by `just create-secrets` and must never be deleted by ArgoCD:

| Secret | Namespace | Purpose |
|--------|-----------|---------|
| `openbao-unseal-key` | `openbao` | Auto-unseal sidecar reads this |
| `cloudflare-api-token` | `cert-manager` | DNS-01 challenge for wildcard TLS cert |

Both carry `argocd.argoproj.io/sync-options: Prune=false`.

## Operations

```bash
# Apply ArgoCD config changes
just core platform up

# Manual sync of a specific app
argocd app sync <app-name>

# List all apps and health
argocd app list

# Force a sync even if not out-of-sync
argocd app sync <app-name> --force

# Check sync waves (for ordered deploys)
argocd app get <app-name>
```

## Troubleshooting

### App Stuck Syncing

**Symptoms:** App shows `OutOfSync` but sync keeps failing.

```bash
# Get detailed sync status
argocd app get <app-name>

# Check events
kubectl get events -n argocd | grep <app-name>
```

**Common causes:**
- CRD not yet installed (wrong sync wave ordering)
- Hook job failed (check job logs in the app namespace)
- `ServerSideApply=true` conflict with existing resource ownership

### Prune=false Secret Deleted

If a bootstrap Secret is accidentally deleted:

```bash
just create-secrets
```

This recreates both `openbao-unseal-key` and `cloudflare-api-token` from `secrets/bootstrap.sops.yaml`.

### Out of Sync After kubectl Change

ArgoCD's `selfHeal=true` will revert any manual `kubectl apply/patch/delete` within 3 minutes. This is intentional — all changes must go through Git.
