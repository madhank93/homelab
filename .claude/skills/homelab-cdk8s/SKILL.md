---
name: homelab-cdk8s
description: "Homelab CDK8s (Golang) conventions for workloads/ — adding or changing a chart in the apps table of workloads/main.go, kit helpers (Namespace, Route, SecretProviderClass, SecretSync), Helm values via cdk8s.NewHelm or typed imports from workloads/cdk8s.yaml, raw cdk8s.NewApiObject for CRDs, OpenBao secrets through the Secrets Store CSI driver, HTTPRoutes on the shared Gateway, GPU scheduling on worker4, and verifying app/ output with just synth. Use whenever a task touches workloads/, app/, an Argo CD Application, a Helm chart version, an HTTPRoute, a SecretProviderClass, or asks to deploy a new service to the cluster. Do NOT use for Pulumi code in core/ — use homelab-pulumi-go. Do NOT use for general Go style — use golang-code-style, golang-naming or golang-error-handling."
---

# Homelab CDK8s

`workloads/` is one Go program: `go run .` synthesizes every chart into `app/<dir>/`, CI publishes `app/` to `<branch>-manifests`, and the Argo CD ApplicationSet (`core/platform/argocd.go`) creates one Application per directory, named after it and defaulting to a namespace of the same name.

## Add a workload

1. Create `New<Name>Chart(scope constructs.Construct, id, namespace string) cdk8s.Chart` in the matching domain package (`ai/`, `registry/`, `monitoring/`, ...). Start with `cdk8s.NewChart(scope, id, &cdk8s.ChartProps{Namespace: ...})` and `kit.Namespace(chart, namespace, privileged)`.
2. Append one entry to the `apps` slice in `workloads/main.go`: `{"<dir>", func(a cdk8s.App) { pkg.NewXChart(a, "<x>-app", "<ns>") }}`. Keep `<dir>` equal to the namespace — the ApplicationSet uses the directory name as the destination namespace.
3. Pin versions in `config/versions.yaml` (`charts:` / `images:`), read with `config.HelmChart("x")` and `config.ContainerImage("x")`. Both panic on an unknown key, so a typo fails synth instead of rendering an empty version.
4. Expose it with `kit.Route` (see below), never an Ingress.
5. Run the verification loop at the end of this file.

## kit helpers (`workloads/kit/kit.go`)

| Helper | Use when |
|---|---|
| `Namespace(chart, ns, privileged)` | Every chart. `privileged=true` only for host-access workloads (storage, eBPF, WireGuard, GPU: Longhorn, Falco, NetBird, GPU operator) |
| `Route(chart, name, sub, svc, port)` | Expose `svc:port` at `<sub>.<domain>` on the shared Gateway from `config/config.yml` |
| `RouteRules` / `Backend` | Route needs redirects, header filters or several backends |
| `SecretProviderClass(chart, name, role, objects, secretName, sync...)` | Any OpenBao secret; reads `secret/<role>` with the Kubernetes auth role of the same name |
| `SecretsVolume` / `SecretsMount` | Typed pod spec: CSI volume at `/mnt/secrets` |
| `SecretsVolumeValues` / `SecretsMountValues` | Same volume as Helm values, for charts with `extraVolumes` |
| `SecretSync(chart, spc)` | Chart cannot mount extra volumes (Harbor, Falco); runs a pause pod whose mount triggers the Secret mirror |

Reuse these instead of writing HTTPRoute or SecretProviderClass objects by hand.

## Pick the construct

| Situation | Construct | Example |
|---|---|---|
| Helm chart, values only | `cdk8s.NewHelm` with `Repo`/`Version` from `config.HelmChart` | `support/reloader.go` |
| OCI Helm chart | `cdk8s.NewHelm`, `Chart: repo + "/name"`, no `Repo` | `automation/n8n.go` |
| Helm chart with a typed values struct | Typed import package (`imports/<chart>`), version in `workloads/cdk8s.yaml` | `registry/harbor.go` |
| Core Kubernetes kinds | `imports/k8s` (`k8s.NewKubeDeployment`, ...) | `ai/comfyui.go` |
| CRD with no import (HTTPRoute, SecretProviderClass, CNPG `Cluster`) | `cdk8s.NewApiObject` + `AddJsonPatch(cdk8s.JsonPatch_Add("/spec", ...))` | `kit/kit.go`, `automation/n8n.go` |

- List a chart in `cdk8s.yaml` only if Go code imports its generated package; regenerate `imports/` with `cdk8s import` after a version change, never hand-edit it.
- Keep the `k8s@` import aligned with the cluster's Kubernetes minor (comment in `cdk8s.yaml`).
- Disable chart features Argo CD or the platform already covers: Ingress, bundled databases (n8n uses CloudNativePG), Helm test hooks.

## Secrets — never render a `Secret`

Rendered `app/` is a public branch; `scripts/manifest-secret-guard.sh` fails CI on any Secret payload outside `scripts/manifest-secret-allowlist.txt`, and that list may only shrink.

| Pattern | Use when | Example |
|---|---|---|
| File-only | App can read the value from a file (Grafana `$__file{}`) | Grafana `ADMIN_PASSWORD` |
| Synced Secret (`kit.Sync`) | App needs an env var or an `existingSecret` | n8n `N8N_ENCRYPTION_KEY`, Harbor admin password |
| Operator-owned | An operator generates the credential | CNPG `n8n-pg-app` |

Details, rotation and the re-mount rule: [references/secrets.md](references/secrets.md).

## GPU workloads (worker4)

Set all four, or the pod will not schedule or will not see CUDA:

- [ ] `nodeSelector: nvidia.com/gpu.present: "true"`
- [ ] toleration `dedicated=ai:NoSchedule` (`Equal`)
- [ ] `runtimeClassName: nvidia`
- [ ] resource limit `nvidia.com/gpu: 1`; use a `Recreate` strategy so two pods never claim the GPU (`ai/comfyui.go`)

Typed example: `ai/comfyui.go`; Helm-values example: `ai/ollama.go`.

## Deterministic output

Refactors are verified by diffing `app/`, so code must not add nondeterminism of its own.

- Never build strings or slices by ranging over a Go map; sort the keys first (`gaming/cs2.go` `serverCfg`). Maps passed whole as Helm values or JSON patches serialize stably.
- Never embed timestamps, random values or build metadata.
- Charts that template random values (`randAlphaNum`, self-signed certs) differ on every synth: Harbor `Secret.*` plus their `checksum/*` annotations, and the VictoriaMetrics operator webhook Secret and `caBundle`. Ignore those files when diffing.

## Verify

Run from the repo root and fix until clean:

```bash
cp -R app /tmp/app-before            # before editing, for refactors
just synth                           # go run . in workloads/ → app/
diff -r /tmp/app-before app          # expect only the intended change (plus the random-value files above)
scripts/manifest-secret-guard.sh app # no new Secret payloads
just check                           # golangci-lint + go test ./... (config, workloads) + core build + synth + docs-check
```

For Go style questions, load `golang-code-style`, `golang-error-handling` or `golang-lint`.
