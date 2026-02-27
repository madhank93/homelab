+++
title = "CDK8s"
description = "CDK8s Go app: manifest synthesis and CI publish pipeline."
weight = 20
+++

## What is CDK8s?

[CDK8s](https://cdk8s.io/) (Cloud Development Kit for Kubernetes) generates Kubernetes YAML manifests from Go code. In this project it replaces hand-written YAML with typed, testable Go.

## Structure

```
platform/cdk8s/
├── main.go              # App entrypoint — one CDK8s App per platform application
├── go.mod / go.sum
└── cots/
    ├── ai/
    │   ├── comfyui.go
    │   ├── nvidia_gpu_operator.go
    │   └── ollama.go
    ├── automation/
    │   └── n8n.go
    ├── compliance/
    │   ├── falco.go
    │   ├── keyverno.go
    │   └── trivy.go
    ├── management/
    │   ├── fleet_device_manager.go
    │   ├── headlamp.go
    │   └── rancher.go
    ├── monitoring/
    │   ├── alert_manager.go
    │   ├── grafana.go
    │   ├── otel_collector.go
    │   ├── victoria_logs.go
    │   └── victoria_metrics.go
    ├── registry/
    │   └── harbor.go
    ├── security/
    │   └── infisical.go
    └── storage/
        └── longhorn.go
```

## How main.go Works

Each app is registered as a CDK8s `App` with its own output directory:

```go
longhornApp := cdk8s.NewApp(&cdk8s.AppProps{
    Outdir:         jsii.String("../../app/longhorn"),
    YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
})
storage.NewLonghornChart(longhornApp, "longhorn-app", "longhorn-system")
longhornApp.Synth()
```

Running `go run main.go` writes all manifests to `../../app/` (relative to `platform/cdk8s/`), which is `app/` at the repo root.

## CI Pipeline

`.github/workflows/cdk8s-seal-publish.yml` runs on push to `main` or any `v*` branch when `platform/cdk8s/**` changes:

1. Checkout source
2. Set up Go 1.23
3. `go run main.go` — synthesizes manifests to `app/`
4. `peaceiris/actions-gh-pages@v3` — publishes `app/` to `${branch}-manifests` branch

The manifests branch (e.g. `v0.1.5-manifests`) is the ArgoCD source.

## No Secrets in Generated Manifests

CDK8s never generates `Secret` resources. All secrets are:
- Bootstrap: created by `just create-secrets` from SOPS-encrypted file
- Runtime: synced by Infisical operator via `InfisicalSecret` CRDs

The CI pipeline requires **zero GitHub Actions secrets**.

## Adding a New App

1. Create `platform/cdk8s/cots/<folder>/<app>.go`
2. Add a registration block in `main.go`
3. Push — CI synthesizes and publishes the new directory
4. ArgoCD detects the new directory and creates an Application automatically
