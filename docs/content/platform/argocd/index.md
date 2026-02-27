+++
title = "ArgoCD"
description = "ArgoCD Helm installation, ApplicationSet, and sync configuration."
weight = 10
+++

## Helm Installation

ArgoCD chart v9.4.2 is installed in the `argocd` namespace by Pulumi (`infra/pulumi/argocd.go`).

Key Helm values:

```yaml
server:
  service:
    type: LoadBalancer  # Cilium L2 assigns an IP from the pool
configs:
  params:
    server.insecure: false
```

ArgoCD is accessible via:
- `http://argocd.local` (HTTPRoute via Gateway API)
- `https://argocd.madhan.app` (TLSRoute passthrough — once wildcard cert exists)

## ApplicationSet

A single `ApplicationSet` (`cots-applications`) bootstraps all platform apps. It watches the `v0.1.5-manifests` branch of the homelab repo:

```yaml
spec:
  generators:
    - git:
        repoURL: https://github.com/madhank93/homelab.git
        revision: v0.1.5-manifests
        directories:
          - path: "*"   # Each top-level directory = one Application
  template:
    spec:
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

## Sync Options

### `ServerSideApply=true`

Required for `kube-prometheus-stack` CRDs which exceed the 262 KB `kubectl.kubernetes.io/last-applied-configuration` annotation limit. Without SSA, these CRDs fail to apply.

### `ServerSideApply=false` on InfisicalSecret

InfisicalSecret resources must carry the annotation:

```yaml
argocd.argoproj.io/sync-options: ServerSideApply=false
```

This is because the Infisical CRD schema omits `projectSlug` from `serviceToken.secretsScope`. ArgoCD's SSA validation rejects the field as "not declared in schema". Disabling SSA for these resources falls back to client-side apply.

> **Note:** `ignoreDifferences` skips drift detection but does NOT bypass apply failures. The SSA annotation must be on the resource itself.

### `ignoreDifferences` for InfisicalSecret

The ApplicationSet also configures `ignoreDifferences` to skip spec drift detection for `InfisicalSecret` resources:

```yaml
ignoreDifferences:
  - group: secrets.infisical.com
    kind: InfisicalSecret
    jsonPointers:
      - /spec
```

### `Prune=false` on Bootstrap Secrets

The two bootstrap Secrets (`infisical-secrets`, `cloudflare-api-token`) carry:

```yaml
argocd.argoproj.io/sync-options: Prune=false
```

These Secrets are not in the manifests branch. Without `Prune=false`, ArgoCD would delete them on the next sync.

## App-of-Apps Pattern

Each directory under `app/` in the manifests branch becomes an ArgoCD `Application` automatically. No manual Application creation is needed when adding a new CDK8s app — just add an entry in `platform/cdk8s/main.go` and the CI pipeline creates the directory.
