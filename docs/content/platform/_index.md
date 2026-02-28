+++
title = "Platform"
description = "GitOps platform: CDK8s manifest generation and ArgoCD deployment."
weight = 50
sort_by = "weight"
+++

The platform layer sits above bare infrastructure and delivers all applications via GitOps.

- **No secrets in git** — CDK8s generates zero `Secret` resources. Bootstrap secrets are created by `just create-secrets` from a laptop.
- **Manifests are generated, not hand-written** — The `v0.1.5-manifests` branch is machine-generated YAML. Never edit it by hand.
- **One CDK8s app per ArgoCD Application** — Each `main.go` entry writes to a separate `app/<name>/` directory, which becomes one ArgoCD Application.
- **ArgoCD is the single source of truth** — All drift is auto-corrected (`selfHeal: true`), all removed resources are pruned (`prune: true`).

See [GitOps Flow](/architecture/gitops-flow) for the end-to-end diagram.
