---
name: homelab-upgrade
description: "Homelab version upgrades — moving any pinned version (Talos, Kubernetes, Gateway API CRDs, Cilium, Argo CD, cert-manager, OpenBao, Longhorn, CNPG, workload charts and images, Bifrost Docker Compose images, the Authentik bridged provider, Pulumi provider SDKs): researching upstream releases and open issues with gh first, finding where the pin lives (config/versions.yaml, workloads/cdk8s.yaml plus cdk8s import, core/Pulumi.yaml packages, core/go.mod, docker-compose.yml, kustomization.yaml), updating the software inventory row in the same change, layer order bottom-up, and how Renovate PRs fit. Use whenever a task bumps, upgrades, downgrades or rolls back a version, reviews a Renovate PR, or asks which version to move to. Do NOT use for reconciling a Pulumi provider diff in detail — use provider-upgrade. Do NOT use for Go module mechanics — use golang-dependency-management. Do NOT use for Talos machine config edits — use homelab-talos. Do NOT use for deploying the change — use homelab-rollout."
---

# Homelab upgrades

Every version is pinned in code, and the [software inventory](../../../docs/src/content/docs/reference/software-inventory.md) is the only doc page that states one. An upgrade is: research, move the pin, update the inventory, verify, hand off to rollout. One layer per change, bottom-up.

## Hard rules

- Research upstream before writing any code (step 1). Summarize the findings to the user first.
- One minor at a time for Talos, Cilium and Longhorn. Never skip.
- Never lower `talos.version` and run `just core talos up` — see the danger aside in [Talos upgrades](../../../docs/src/content/docs/operate/upgrades/talos.mdx).
- Never run `just core`; the user runs `preview` then `up`. Upgrades go through Git, never `helm upgrade` or `kubectl apply` by hand (except the documented Cilium rollback fallback).
- No version numbers outside the inventory; `just docs-check` fails otherwise.

## Procedure

1. **Research upstream.** For the target version and every version between current and target:

   ```bash
   gh release list --repo <org>/<repo> --limit 10
   gh release view <tag> --repo <org>/<repo>              # breaking changes, upgrade notes
   gh issue list --repo <org>/<repo> --state open --search "<version>"
   gh issue list --repo <org>/<repo> --state open --search "upgrade <from-minor>"
   gh issue list --repo <org>/<repo> --state open --label kind/regression   # if the repo uses it
   ```

   Look for regressions, removed Helm values, CRD changes and "upgrade from X.Y" reports. For charts, compare the CHART and APP columns of `helm search repo <repo>/<chart> --versions` — a chart can lag its app. Check the [breaking changes table](../../../docs/src/content/docs/operate/upgrades/breaking-changes.md) and the [Phase 0 commands](../../../docs/src/content/docs/operate/upgrades/index.mdx).

2. **Find the pin** (table below), change it, and keep paired pins together.

3. **Update the inventory row** in `docs/src/content/docs/reference/software-inventory.md` in the same commit. Add a row to the breaking changes page if research found one that bites this cluster.

4. **Regenerate** what the pin feeds (imports, SDKs), then run Verify.

5. **Hand off**: workloads deploy through Git (homelab-rollout); `core/` changes need the user's `just core <stack> preview` then `up`.

## Where pins live

| Pin | File | After changing |
|---|---|---|
| Talos, Kubernetes, Helm charts and images read by Go | `config/versions.yaml` | `just synth` (workloads) or `cd core && go build ./...` |
| Charts with a typed package, `k8s@` API types | `workloads/cdk8s.yaml` | `cd workloads && cdk8s import && go mod tidy`, then `just synth` ([CDK8s imports](../../../docs/src/content/docs/operate/upgrades/cdk8s-imports.md)) |
| Authentik Terraform provider (bridged) | `core/Pulumi.yaml` `packages.authentik` | Regenerate `core/sdks/authentik/` on the laptop, then `cd core && go build ./...` |
| Pulumi provider SDKs (proxmoxve, talos, kubernetes, cloudflare, hcloud, command) | `core/go.mod` | provider-upgrade skill |
| Gateway API CRDs | vendored file in `core/platform/manifests/`, named in `core/platform/cilium.go` | Download the new `experimental-install.yaml`, delete the old one |
| Bifrost images | `core/cloud/bifrost/docker-compose.yml` | `just core hetzner preview` |
| Kubeflow | `workloads/ai/kubeflow/kustomization.yaml` (`ref=` on every base) | `cd workloads && kustomize build ai/kubeflow/ > /dev/null` |
| Tooling | `.devcontainer/Dockerfile`, `.mise.toml`, `.github/workflows/` | — |

Pairs that move together: `images.netbird` + the Bifrost NetBird images; `charts.trivy-operator` + the `trivy-operator` import in `cdk8s.yaml`; Talos + `talosctl`; `kubernetes.version` minor + `k8s@` in `cdk8s.yaml`; `charts.argo-cd` + `images.argocd` (drop the image override once the chart's appVersion catches up).

## Layer order

Bottom-up, each layer stable before the next: Talos and Kubernetes → Gateway API CRDs → Cilium → Argo CD → cert-manager → OpenBao and CSI driver → Longhorn and CNPG → workloads. Bifrost is independent. The phase table and per-layer runbooks are in [Upgrades](../../../docs/src/content/docs/operate/upgrades/index.mdx); follow the page for the layer being moved.

## Sharpest gotchas

| Component | Gotcha | Page |
|---|---|---|
| Talos | `just core talos up` never upgrades a running node; use `just talos-upgrade <ip> [base\|gpu] [drain]`. Nodes need `install.disk`. Pass `drain=false` for the GPU worker and the node hosting `n8n-pg-1`, or the drain times out and leaves the node cordoned. Never `--preserve` | [Talos](../../../docs/src/content/docs/operate/upgrades/talos.mdx) |
| Kubernetes | Pinned separately from Talos and held inside Cilium's supported range (Renovate `allowedVersions`). Rolled out by `just talos-upgrade-k8s`. A minor bump also moves `k8s@` in `cdk8s.yaml` | [Talos](../../../docs/src/content/docs/operate/upgrades/talos.mdx) |
| Cilium | Gateway API CRDs first, as their own apply. An agent restart can leave a stale L2 lease: in-cluster works, LAN times out → `just gateway-repair-l2`. Abort on any replace in preview | [Cilium](../../../docs/src/content/docs/operate/upgrades/cilium.mdx) |
| Authentik | Bridged provider version is the second `parameters` entry in `core/Pulumi.yaml`, not `version` (that is the bridge). The server image refuses to skip a minor | [Bifrost](../../../docs/src/content/docs/operate/upgrades/bifrost.md) |
| Charts | A changed Deployment `spec.selector` is immutable: Argo CD cannot sync until the old Deployment is deleted (pod/Deployment recreation, not a config fix) | [Workloads](../../../docs/src/content/docs/operate/upgrades/workloads.mdx) |
| OpenBao | Image held on its current minor for the `file` storage backend; the StatefulSet is OnDelete | [Secrets](../../../docs/src/content/docs/operate/upgrades/secrets.md) |
| n8n | One-way DB migration: `pg_dump` before pushing | [Workloads](../../../docs/src/content/docs/operate/upgrades/workloads.mdx) |

## Renovate

`renovate.json` drives most bumps: regex managers read `config/versions.yaml`, `workloads/cdk8s.yaml`, `core/platform/cilium.go` (Gateway API file name), the Kubeflow kustomization and `.devcontainer/Dockerfile`; native managers cover `go.mod`, docker-compose, GitHub Actions, mise and npm. Minor/patch updates are batched into one "non-major" PR; majors wait for dashboard approval.

Treat a Renovate PR as step 2 done, not the upgrade done:

- Still run step 1 for every package in the PR; split out anything risky (Talos, Cilium, Longhorn, anything in `core/`).
- The PR does not touch the inventory, regenerate `cdk8s import`, download the Gateway API manifest, or update checksums — its `prBodyNotes` say which. Do those on the branch.
- `core/Pulumi.yaml` is not managed by Renovate; bump the Authentik provider by hand.

## Verify

```bash
just check          # lint + tests + core build + synth + docs-check
just docs-check     # inventory row matches the pin; no stray versions
git diff --stat     # pin, inventory row and regenerated files in one change
```

For `core/` changes, ask the user for `just core <stack> preview` and read it before `up`. After rollout, run the post-upgrade checks in [Upgrades](../../../docs/src/content/docs/operate/upgrades/index.mdx).
