+++
title = "CDK8s"
description = "CDK8s Go app: manifest synthesis pipeline and zero-secret GitOps."
weight = 20
+++

## What is CDK8s?

[CDK8s](https://cdk8s.io/) (Cloud Development Kit for Kubernetes) generates Kubernetes YAML manifests from Go (or TypeScript, Python, Java) code. In this project it replaces hand-written YAML with typed, testable Go — every Helm release, every CRD, every HTTPRoute is expressed as a Go struct.

## Why CDK8s?

| Approach | Type safety | Reuse | Tests | Complexity |
|----------|-------------|-------|-------|------------|
| Raw YAML | None | Copy-paste | Manual | Low |
| Helm | Partial (templates) | Good | Limited | Medium |
| Kustomize | None | Overlays | Limited | Medium |
| **CDK8s (Go)** | Full | Functions | Full Go testing | Medium |

CDK8s enables Go functions to generate manifests, making it easy to share patterns (like the OpenBao CSI volume + SecretProviderClass pattern) across all apps without duplicating YAML.

## How It's Used Here

All workloads — Helm releases, CRDs, HTTPRoutes, SecretProviderClasses — are defined as Go structs in `workloads/`. A CI pipeline runs `go run .` on every push to synthesize YAML into `app/` and force-pushes that output to the `v0.1.5-manifests` branch, which ArgoCD watches. No Kubernetes credentials are needed in CI because CDK8s generates zero `Secret` resources.

## Structure

```
workloads/
├── main.go              # Entrypoint — one CDK8s App per workload
├── go.mod / go.sum
├── cdk8s.yaml           # Import versions (update here + re-run cdk8s import)
├── imports/             # Auto-generated CDK8s Helm chart bindings
├── storage/             longhorn.go
├── secrets/             openbao.go, csi_driver.go
├── observability/       victoria_metrics.go, victoria_logs.go, otel_collector.go, alert_manager.go
├── monitoring/          grafana.go
├── security/            falco.go, trivy.go
├── hardware/            nvidia_gpu_operator.go
├── networking/          netbird_peer.go
├── registry/            harbor.go
├── automation/          n8n.go
├── databases/           cnpg.go
├── ai/                  ollama.go, comfyui.go
├── management/          headlamp.go, rancher.go
└── support/             reloader.go
```

## How main.go Works

Each workload is a separate CDK8s `App` with its own output directory:

```go
// workloads/main.go
longhornApp := cdk8s.NewApp(&cdk8s.AppProps{
    Outdir:         jsii.String("../app/longhorn"),
    YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
})
storage.NewLonghornChart(longhornApp, "longhorn-app", "longhorn-system")
longhornApp.Synth()

openBaoApp := cdk8s.NewApp(&cdk8s.AppProps{
    Outdir:         jsii.String("../app/openbao"),
    YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
})
secrets.NewOpenBaoChart(openBaoApp, "openbao-app", "openbao")
openBaoApp.Synth()
// ... one block per workload
```

Running `just synth` executes `go run .` in `workloads/`, which writes all manifests to `../app/` (the `app/` directory at the repo root).

## CI Pipeline

`.github/workflows/cdk8s-seal-publish.yml` runs on push to `main` or any `v*` branch when `workloads/**` changes:

1. Checkout source
2. Set up Go
3. `go run .` — synthesizes all manifests to `app/`
4. Force-pushes `app/` content to `${branch}-manifests` branch (e.g. `v0.1.5-manifests`)

The manifests branch is the ArgoCD source. ArgoCD's `ApplicationSet` directory generator watches every top-level directory in `v0.1.5-manifests` and creates an Application for each.

## Synthesis Flow

```
workloads/main.go (Go source)
  → cdk8s.Synth()
  → YAML files per resource in app/<workload>/
  → CI pushes to v0.1.5-manifests branch
  → ArgoCD detects new/changed directories
  → ArgoCD syncs to cluster
```

## No Secrets in Generated Manifests

CDK8s never generates `Secret` resources. The CI pipeline requires **zero GitHub Actions secrets**.

- **Bootstrap secrets**: created by `just create-secrets` from SOPS-encrypted `secrets/bootstrap.sops.yaml`, applied once manually
- **Runtime secrets**: fetched by the in-cluster CSI driver from OpenBao at pod startup

The entire manifests branch can be public (and is) without any security risk.

## Adding a New App

1. Create `workloads/<category>/<app>.go` with the appropriate package name
2. Add a registration block in `workloads/main.go`:
   ```go
   myApp := cdk8s.NewApp(&cdk8s.AppProps{
       Outdir:         jsii.String(fmt.Sprintf("%s/myapp", rootFolder)),
       YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
   })
   mycategory.NewMyAppChart(myApp, "myapp-app", "myapp-namespace")
   myApp.Synth()
   ```
3. Push — CI synthesizes and publishes the new directory to the manifests branch
4. ArgoCD detects the new directory and creates an Application automatically

## Updating Chart Versions

CDK8s chart bindings are imported via `cdk8s import`:

```bash
# Update version in workloads/cdk8s.yaml
# Then re-import to regenerate typed bindings:
cd workloads && cdk8s import

# Or for OCI charts:
cdk8s import helm:<chart-name>@<version>
```

The `imports/` directory is committed to the repo. Never edit files in `imports/` by hand.
