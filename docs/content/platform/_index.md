+++
title = "Platform"
description = "GitOps platform: CDK8s manifest generation and ArgoCD deployment."
weight = 50
sort_by = "weight"
+++

The platform layer sits above bare infrastructure and delivers all applications via GitOps.

## GitOps Flow

{% mermaid() %}
flowchart TD
    DEV["Developer\nedits platform/cdk8s/cots/app.go"]
    PUSH["git push to v0.1.5"]
    CI["GitHub Actions CI\n.github/workflows/cdk8s-seal-publish.yml"]
    SYNTH["go run main.go\ncdk8s synth"]
    MANIFESTS["v0.1.5-manifests branch\napp/appname/*.yaml"]
    ARGOCD["ArgoCD\nApplicationSet watches branch"]
    CLUSTER["Kubernetes Cluster\nresources created/updated"]

    DEV --> PUSH
    PUSH --> CI
    CI --> SYNTH
    SYNTH --> MANIFESTS
    MANIFESTS --> ARGOCD
    ARGOCD --> CLUSTER
{% end %}

## Key Properties

- **No secrets in git** — CDK8s generates zero `Secret` resources. Bootstrap secrets are created by `just create-secrets` from a laptop.
- **Manifests are generated, not hand-written** — The `v0.1.5-manifests` branch is machine-generated YAML. Never edit it by hand.
- **One CDK8s app per ArgoCD Application** — Each `main.go` entry writes to a separate `app/<name>/` directory, which becomes one ArgoCD Application.
- **ArgoCD is the single source of truth** — All drift is auto-corrected (`selfHeal: true`), all removed resources are pruned (`prune: true`).

## CDK8s Apps

| Directory | Namespace |
|-----------|-----------|
| `app/longhorn` | `longhorn-system` |
| `app/infisical` | `infisical` |
| `app/grafana` | `grafana` |
| `app/victoria-metrics` | `victoria-metrics` |
| `app/victoria-logs` | `victoria-logs` |
| `app/alertmanager` | `alertmanager` |
| `app/harbor` | `harbor` |
| `app/n8n` | `n8n` |
| `app/nvidia-gpu-operator` | `nvidia-gpu-operator` |
| `app/ollama` | `ollama` |
| `app/comfyui` | `comfyui` |
| `app/trivy` | `trivy` |
| `app/falco` | `falco` |
| `app/opentelemetry` | `opentelemetry` |
| `app/headlamp` | `headlamp` |
| `app/fleet` | `fleet` |
| `app/rancher` | `cattle-system` |
