+++
title = "CDK8s"
description = "CDK8s Go app: manifest synthesis and CI publish pipeline."
weight = 20
+++

## What is CDK8s?

[CDK8s](https://cdk8s.io/) (Cloud Development Kit for Kubernetes) generates Kubernetes YAML manifests from Go code. In this project it replaces hand-written YAML with typed, testable Go.

## Structure

```
workloads/
├── main.go              # App entrypoint — one CDK8s App per platform application
├── go.mod / go.sum
├── imports/             # Auto-generated CDK8s Helm chart bindings
├── storage/             longhorn.go
├── secrets/             infisical.go
├── observability/       victoria_metrics.go, victoria_logs.go, otel_collector.go, alert_manager.go
├── monitoring/          grafana.go
├── security/            falco.go, keyverno.go, trivy.go
├── hardware/            nvidia_gpu_operator.go
├── networking/          netbird_peer.go
├── registry/            harbor.go
├── automation/          n8n.go
├── ai/                  ollama.go, comfyui.go
├── management/          headlamp.go, fleet_device_manager.go, rancher.go
└── support/             reloader.go
```

## How main.go Works

Each app is registered as a CDK8s `App` with its own output directory under `../app/`:

```go
longhornApp := cdk8s.NewApp(&cdk8s.AppProps{
    Outdir:         jsii.String("../app/longhorn"),
    YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
})
storage.NewLonghornChart(longhornApp, "longhorn-app", "longhorn-system")
longhornApp.Synth()
```

Running `just synth` (from the repo root) writes all manifests to `app/`.

## CI Pipeline

`.github/workflows/cdk8s-seal-publish.yml` runs on push to `main` or any `v*` branch when `workloads/**` changes:

1. Checkout source
2. Set up Go
3. `go run .` — synthesizes manifests to `app/`
4. Publishes `app/` to `${branch}-manifests` branch

The manifests branch (e.g. `v0.1.5-manifests`) is the ArgoCD source.

## No Secrets in Generated Manifests

CDK8s never generates `Secret` resources. All secrets are:
- Bootstrap: created by `just create-secrets` from SOPS-encrypted `secrets/bootstrap.sops.yaml`
- Runtime: synced by Infisical operator via `InfisicalSecret` CRDs

The CI pipeline requires **zero GitHub Actions secrets**.

## Adding a New App

1. Create `workloads/<category>/<app>.go` with the appropriate package name
2. Add a registration block in `workloads/main.go`
3. Push — CI synthesizes and publishes the new directory to the manifests branch
4. ArgoCD detects the new directory and creates an Application automatically
