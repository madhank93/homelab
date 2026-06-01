+++
title = "Platform"
description = "core/platform/ Pulumi stacks: Proxmox/Talos cluster, Cilium CNI, ArgoCD GitOps, cert-manager, and secrets."
weight = 50
sort_by = "weight"
+++

Platform covers both cluster provisioning (`core/platform/`) and the GitOps delivery layer.

## Pulumi Stacks (`core/platform/`)

| Stack | Command | Manages |
|-------|---------|---------|
| `talos` | `just core talos up` | Proxmox VMs, Talos bootstrap, Cilium CNI, ArgoCD |
| `platform` | `just core platform up` | Gateway API, IP pool, HTTPRoutes, cert-manager |

## GitOps Layer

- **No secrets in git** — CDK8s generates zero `Secret` resources. Bootstrap secrets are created by `just create-secrets` from a laptop.
- **Manifests are generated, not hand-written** — The `v0.1.5-manifests` branch is machine-generated YAML. Never edit it by hand.
- **One CDK8s app per ArgoCD Application** — Each `main.go` entry writes to a separate `app/<name>/` directory, which becomes one ArgoCD Application.
- **ArgoCD is the single source of truth** — All drift is auto-corrected (`selfHeal: true`), all removed resources are pruned (`prune: true`).

See [GitOps Flow](/architecture/gitops-flow) for the end-to-end diagram.
