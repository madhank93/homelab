+++
title = "Version Upgrade Guide"
description = "End-to-end procedure for upgrading all homelab components: Talos, platform tools, cloud services, and workloads."
weight = 70
+++

Version upgrades follow a strict layer order — each layer depends on the one below it being stable first.

---

## Current Version Inventory

> **Branch:** `v0.1.6` — code targets below. Some require Pulumi/kubectl apply to take effect on the live cluster.

### Infrastructure & Platform

| Component | v0.1.6 Target | Live | File |
|-----------|--------------|------|------|
| Talos | `v1.13.3` | v1.12.4 (pending `just core talos up`) | `core/platform/talos.go:17` |
| Cilium | `1.19.4` | 1.16.6 (step upgrade required) | `core/platform/cilium.go:30` |
| ArgoCD chart | `9.5.15` | 9.4.2 (pending `just core platform up`) | `core/platform/argocd.go:24` |
| ArgoCD manifests branch | `v0.1.6-manifests` | patched live | `core/platform/argocd.go:175,190` |
| cert-manager | `v1.20.2` | v1.19.3 ✓ | `workloads/cdk8s.yaml` |
| k8s API target | `1.35.0` | 1.35.0 ✓ | `workloads/cdk8s.yaml` |

### Cloud (Bifrost Docker Compose)

| Component | v0.1.6 Target | File |
|-----------|--------------|------|
| Traefik | `v3.7.1` | `core/cloud/bifrost/docker-compose.yml:35` |
| NetBird server | `0.71.4` | `docker-compose.yml:59` |
| NetBird dashboard | `v0.71.4` | `docker-compose.yml:73` |
| NetBird reverse-proxy | `v0.71.4` | `docker-compose.yml:82` |
| NetBird agent (Bifrost) | `0.71.4` | `docker-compose.yml:94` |
| Authentik | `2026.5.2` | `docker-compose.yml:137,168` |
| PostgreSQL (Authentik) | `16.14-alpine` | `docker-compose.yml:115` |

### Secrets & Storage

| Component | v0.1.6 Target | File |
|-----------|--------------|------|
| OpenBao helm | `0.28.3` | `workloads/secrets/openbao.go:42` |
| OpenBao image | `2.5.4` | `workloads/secrets/openbao.go:98` |
| CSI Driver | `1.6.0` | `workloads/cdk8s.yaml` |
| Longhorn | `1.11.2` | `workloads/cdk8s.yaml` |
| CNPG operator | `0.28.2` | `workloads/databases/cnpg.go:31` |

### Workloads

| Component | v0.1.6 Target | File |
|-----------|--------------|------|
| VictoriaMetrics k8s-stack | `0.80.0` | `workloads/cdk8s.yaml` |
| VictoriaLogs | `0.12.5` | `workloads/cdk8s.yaml` |
| Grafana | `12.4.1` | `workloads/cdk8s.yaml` |
| Metrics Server | `3.13.0` | `workloads/cdk8s.yaml` |
| OTel Collector | `0.156.2` | `workloads/observability/otel_collector.go:155,213` |
| Harbor | `1.19.0` | `workloads/cdk8s.yaml` |
| NVIDIA GPU Operator (import) | `v26.3.1` | `workloads/cdk8s.yaml` |
| NVIDIA Device Plugin | `0.19.1` | `workloads/hardware/nvidia_gpu_operator.go:47` |
| DCGM Exporter | `4.8.2` | `workloads/hardware/nvidia_gpu_operator.go:99` |
| n8n (8gears OCI) | `2.0.1` | `workloads/automation/n8n.go:184` |
| Ollama chart | `1.57.0` | `workloads/cdk8s.yaml` |
| Ollama image | `0.24.0` | `workloads/ai/ollama.go:33` |
| ComfyUI image | `cu128-megapak-20260223` | `workloads/ai/comfyui.go:73` |
| Rancher | `2.14.1` | `workloads/cdk8s.yaml` |
| Headlamp | `0.42.0` | `workloads/cdk8s.yaml` |
| Fleet | removed | Managed by Rancher — do not add to cdk8s.yaml |
| Kyverno | `3.8.1` | `workloads/cdk8s.yaml` |
| Trivy Operator | `0.32.1` | `workloads/security/trivy.go:22` + `cdk8s.yaml` |
| Falco | `8.0.5` | `workloads/security/falco.go:84` |
| Reloader | `2.2.12` | `workloads/support/reloader.go:20` |
| NetBird peer (k8s) | `0.71.4` | `workloads/networking/netbird_peer.go:139,158` |

---

## Upgrade Order

Components must be upgraded in layer order. Never skip layers.

```
Talos
  └─► Cilium  (CNI must be compatible with Talos k8s version)
        └─► Gateway API CRDs  (bundled with Cilium chart)
              └─► ArgoCD  (GitOps engine)
                    └─► cert-manager
                          └─► OpenBao + CSI Driver  (secrets layer; all apps depend on this)
                                └─► Longhorn + CNPG  (storage; stateful apps depend on this)
                                      └─► Workloads  (batched by risk)

Bifrost docker-compose  (independent — upgrade anytime)
```

---

## Phase 0 — Research Latest Versions

Run before touching anything. Record results and substitute into the phases below.

```bash
# Infrastructure
curl -s https://api.github.com/repos/siderolabs/talos/releases/latest | jq -r .tag_name
helm repo add cilium https://helm.cilium.io && helm search repo cilium/cilium --versions | head -3
helm repo add argo https://argoproj.github.io/argo-helm && helm search repo argo/argo-cd --versions | head -3

# Add all workload chart repos
helm repo add longhorn    https://charts.longhorn.io
helm repo add vm          https://victoriametrics.github.io/helm-charts
helm repo add grafana     https://grafana-community.github.io/helm-charts
helm repo add harbor      https://helm.goharbor.io
helm repo add openbao     https://openbao.github.io/openbao-helm
helm repo add csi-driver  https://kubernetes-sigs.github.io/secrets-store-csi-driver/charts
helm repo add rancher     https://releases.rancher.com/server-charts/stable
helm repo add headlamp    https://kubernetes-sigs.github.io/headlamp
helm repo add kyverno     https://kyverno.github.io/kyverno
helm repo add trivy       https://aquasecurity.github.io/helm-charts
helm repo add otel        https://open-telemetry.github.io/opentelemetry-helm-charts
helm repo add ollama      https://otwld.github.io/ollama-helm
helm repo add reloader    https://stakater.github.io/stakater-charts
helm repo add cnpg        https://cloudnative-pg.github.io/charts
helm repo add metrics-server https://kubernetes-sigs.github.io/metrics-server
helm repo update

for chart in longhorn/longhorn \
  vm/victoria-metrics-k8s-stack vm/victoria-logs-single \
  grafana/grafana harbor/harbor openbao/openbao \
  csi-driver/secrets-store-csi-driver rancher/rancher \
  headlamp/headlamp kyverno/kyverno trivy/trivy-operator \
  otel/opentelemetry-collector ollama/ollama reloader/reloader \
  cnpg/cloudnative-pg metrics-server/metrics-server; do
  echo "=== $chart ===" && helm search repo "$chart" --versions | head -2
done

# Container images
curl -s https://api.github.com/repos/netbirdio/netbird/releases/latest | jq -r .tag_name
curl -s https://api.github.com/repos/goauthentik/authentik/releases/latest | jq -r .tag_name
curl -s https://api.github.com/repos/traefik/traefik/releases/latest | jq -r .tag_name
curl -s https://hub.docker.com/v2/repositories/ollama/ollama/tags?page_size=3 | jq -r '.results[].name'
```

---

## Phase 1 — Talos

**Risk:** High — rolling node restart.

```go
// core/platform/talos.go:17
talosVersion = "vX.Y.Z"
```

**Rules:**
- Upgrade one minor version at a time (1.12 → 1.13, not 1.12 → 1.15)
- Control planes upgrade first, workers after all CPs are healthy
- Check k8s version embedded in the new Talos release — if it bumps (e.g., 1.30 → 1.31), update `k8s@1.30.0` in `workloads/cdk8s.yaml` and re-run `cdk8s import`

```bash
just core talos up
talosctl --talosconfig ~/.talos/config health --nodes 192.168.1.210
```

---

## Phase 2 — Cilium

**Risk:** High — CNI restart interrupts pod networking briefly.

```go
// core/platform/cilium.go:30
Version: pulumi.String("1.X.Y"),
```

**Rules:**
- Verify compatibility with new Talos k8s version: https://docs.cilium.io/en/stable/network/kubernetes/compatibility/
- `wt0` must **not** be added to Cilium devices — keep only `eth0` in `cilium.go`. See [NetBird routing notes](../infrastructure/#netbird).
- After upgrade check `CiliumLoadBalancerIPPool` and `BGPCiliumPeeringPolicy` CR specs in `cilium.go:208,225` for field renames

```bash
just core platform up
kubectl -n kube-system rollout status daemonset/cilium
kubectl get gateway -n kube-system homelab-gateway
```

---

## Phase 3 — ArgoCD

**Risk:** Medium — GitOps engine downtime during pod restart.

```go
// core/platform/argocd.go:24
Version: pulumi.String("9.X.Y"),
```

**TLSRoute service name — critical:** The TLSRoute backend at `argocd.go:148` hardcodes a Helm-generated service name `argo-cd-964152f1-argocd-server`. The hash suffix may change on chart upgrade. After `just core platform up`, verify and update if needed:

```bash
kubectl get svc -n argocd | grep argocd-server
# Update argocd.go:148 if the name changed, then re-run just core platform up
```

**ApplicationSet UI:** Already enabled. ArgoCD chart 9.x (ArgoCD 2.14.x) shows ApplicationSets under the top-level **ApplicationSets** tab in the UI. No config changes needed.

```bash
just core platform up
kubectl rollout status deployment argocd-server -n argocd
```

---

## Phase 4 — cert-manager

**Risk:** Low — but high blast radius if broken (TLS for all services).

```yaml
# workloads/cdk8s.yaml
- helm:https://charts.jetstack.io/cert-manager@1.X.Y
```

```bash
just synth && git push
kubectl get certificaterequests -A  # must all show Ready=True
```

---

## Phase 5 — OpenBao + CSI Driver

**Risk:** Medium — all running apps read secrets through this layer.

```go
// workloads/secrets/openbao.go:42
Version: jsii.String("0.X.Y"),
// workloads/secrets/openbao.go:98
"image": "openbao/openbao:2.X.Y",
```

```yaml
# workloads/cdk8s.yaml
- helm:https://openbao.github.io/openbao-helm/openbao@0.X.Y
- helm:https://kubernetes-sigs.github.io/secrets-store-csi-driver/charts/secrets-store-csi-driver@1.X.Y
```

Helm chart version and image tag must be compatible — check OpenBao [release notes](https://github.com/openbao/openbao/releases).

```bash
just synth && git push
kubectl logs -n openbao -l app.kubernetes.io/name=openbao -c unseal
kubectl exec -n openbao openbao-0 -- bao status  # sealed: false
kubectl exec -n grafana deploy/grafana -- cat /mnt/secrets/ADMIN_PASSWORD  # smoke test
```

---

## Phase 6 — Longhorn + CNPG

**Risk:** Medium — storage disruption possible.

### Longhorn

```yaml
# workloads/cdk8s.yaml
- helm:https://charts.longhorn.io/longhorn@1.X.Y
```

One minor version at a time only (e.g., 1.10 → 1.11, not 1.10 → 1.12 directly).

```bash
just synth && git push
kubectl get nodes.longhorn.io -n longhorn-system
kubectl get volume.longhorn.io -n longhorn-system  # all must be healthy
```

### CNPG

```go
// workloads/databases/cnpg.go:31
Version: jsii.String("0.X.Y"),
```

CNPG operator upgrades are backwards-compatible with existing `Cluster` CRs.

```bash
kubectl get cluster -A  # all clusters healthy
```

---

## Phase 7 — Bifrost Docker Compose

**Risk:** Low — independent of k8s cluster.

```yaml
# core/cloud/bifrost/docker-compose.yml
traefik:vX.Y                               # line 35
netbirdio/netbird-server:0.X.Y             # line 59
netbirdio/dashboard:0.X.Y                  # line 73  (was: latest)
netbirdio/reverse-proxy:0.X.Y             # line 82  (was: latest)
netbirdio/netbird:0.X.Y                    # line 94
ghcr.io/goauthentik/server:20XX.X.X        # lines 137, 168
postgres:16.X-alpine                       # line 115  (minor bumps only)
```

**NetBird rule:** All four NetBird components (`server`, `dashboard`, `reverse-proxy`, agent on Bifrost) **must be on the same version**. Also update `workloads/networking/netbird_peer.go:139,158` to match.

**Authentik migration race (fixed in bootstrap.sh):** On Authentik upgrades, `bootstrap.sh` now runs `ak migrate` explicitly (via a one-off server container) before starting `authentik-server` and `authentik-worker`. This prevents a crash-loop where the server queries new ORM columns that haven't been added yet. If `just core hetzner up` fails at the Authentik health check step, SSH in and run:
```bash
docker exec authentik-server ak migrate
docker restart authentik-server authentik-worker
```

**NetBird ip rules lost on container restart:** After restarting `netbird-agent`, verify ip rules are restored — `ip rule show` must show rules for table `7120`. If missing, restart the container again: `docker restart netbird-agent`. This restores the `192.168.1.0/24` policy route needed for Traefik → k8s traffic.

```bash
just core hetzner up
# Verify NetBird mesh:
curl -sk https://netbird.madhan.app/api/v1/peers -H "Authorization: Bearer $NB_TOKEN"
# Verify ip rules on bifrost host:
ssh root@178.156.199.250 "ip rule show | grep 7120"
```

---

## Phase 8 — Workloads: Low Risk

No inter-dependencies. Update all in one commit, run `just synth`, push.

| Component | Change |
|-----------|--------|
| Kyverno | `cdk8s.yaml` |
| Falco | `security/falco.go:84` |
| Metrics Server | `cdk8s.yaml` |
| Reloader | `support/reloader.go:20` |
| Fleet | `cdk8s.yaml` |
| OTel Collector | `observability/otel_collector.go:168,227` — both agent + gateway releases |
| Headlamp | `cdk8s.yaml` |
| Trivy | `cdk8s.yaml` **and** `security/trivy.go:22` — both must match; CRD fetch URL is built from this version |

```bash
just synth && git push
kubectl get applications -n argocd  # all Synced + Healthy
```

---

## Phase 9 — Workloads: Medium Risk

### VictoriaMetrics + VictoriaLogs

```yaml
# workloads/cdk8s.yaml
- helm:https://victoriametrics.github.io/helm-charts/victoria-metrics-k8s-stack@0.X.Y
- helm:https://victoriametrics.github.io/helm-charts/victoria-logs-single@0.X.Y
```

Also update `workloads/observability/victoria_metrics.go:69`. Run `helm diff upgrade` first — the values schema changes frequently between minor versions.

### Grafana

```yaml
# workloads/cdk8s.yaml
- helm:https://grafana-community.github.io/helm-charts/grafana@X.Y.Z
```

**Grafana 10 → 11 warning:** Major version break. Angular plugin support removed. Datasource plugin API changed. Read the [migration guide](https://grafana.com/docs/grafana/latest/upgrade-guide/) before upgrading past 11.0. Verify all dashboards load after upgrade.

### Harbor

```yaml
# workloads/cdk8s.yaml
- helm:https://helm.goharbor.io/harbor@1.X.Y
```

Minor bumps only. After upgrade verify `harbor:80` routing still works (Harbor nginx proxy quirk — route must target `harbor:80`, not `harbor-core:80`).

### NVIDIA GPU Operator

```yaml
# workloads/cdk8s.yaml
- helm:https://helm.ngc.nvidia.com/nvidia/gpu-operator@X.Y.Z
```

Also update device plugin and DCGM exporter versions in `workloads/hardware/nvidia_gpu_operator.go:47,99`. Verify RTX 5070 Ti (sm_120, Blackwell) still supported in the new operator release — Blackwell support was added in 570.x driver series.

```bash
kubectl exec -n ollama deploy/ollama -- nvidia-smi
```

---

## Phase 10 — Workloads: Complex

### Rancher

```yaml
# workloads/cdk8s.yaml
- helm:https://releases.rancher.com/server-charts/stable/rancher@2.X.Y
```

Update `--kube-version` flag in `workloads/management/rancher.go:117` to match the actual cluster k8s version after the Talos upgrade:

```go
HelmFlags: &[]*string{jsii.String("--kube-version"), jsii.String("1.3X.0")},
```

### n8n

```go
// workloads/automation/n8n.go:184
Version: jsii.String("2.X.Y"),
```

Check latest: `helm show chart oci://8gears.container-registry.com/library/n8n`. DB schema migrations run automatically on pod start via CNPG. Verify n8n healthy after upgrade:

```bash
kubectl rollout status deployment n8n -n n8n
```

### Ollama

```go
// workloads/ai/ollama.go:33
"tag": "0.X.Y",
```

Also update chart in `cdk8s.yaml`. Downloaded models stay in the PVC — no re-pull needed.

```bash
kubectl exec -n ollama deploy/ollama -- ollama list
```

### ComfyUI

```go
// workloads/ai/comfyui.go:73
Image: jsii.String("yanwk/comfyui-boot:cu128-megapak-YYYYMMDD"),
```

Check [yanwk/comfyui-boot tags](https://hub.docker.com/r/yanwk/comfyui-boot/tags) — use `cu128-megapak-*` variants only (CUDA 12.8 required for RTX 5070 Ti / sm_120). Do **not** use `latest-cu128` — that tag does not exist.

### NetBird Peer (k8s)

```go
// workloads/networking/netbird_peer.go:139,158  (both init + main containers)
Image: jsii.String("netbirdio/netbird:0.X.Y"),
```

Must match Bifrost docker-compose NetBird version exactly. Upgrade Bifrost and k8s peer in the same commit.

### Kubeflow

Kubeflow is deployed via Kustomize from `kubeflow/manifests`, not a Helm chart. Upgrade by updating `targetRevision` in the ApplicationSet at `core/platform/argocd.go:175`. Check the [kubeflow/manifests releases](https://github.com/kubeflow/manifests/releases) for the version compatible with the current k8s version.

---

## cdk8s Import Regeneration

After changing any chart version in `cdk8s.yaml`, regenerate typed Go bindings:

```bash
cd workloads
cdk8s import        # regenerates workloads/imports/*
go mod tidy
just synth          # verify no compile errors before pushing
```

If Talos upgrades the embedded k8s version, also update the `k8s@1.30.0` import at the top of `cdk8s.yaml`.

---

## Post-Upgrade Verification

```bash
# Cluster health
talosctl --talosconfig ~/.talos/config health --nodes 192.168.1.210

# All ArgoCD apps synced
kubectl get applications -n argocd

# Secrets layer
kubectl exec -n openbao openbao-0 -- bao status
kubectl exec -n grafana deploy/grafana -- cat /mnt/secrets/ADMIN_PASSWORD
kubectl exec -n harbor deploy/harbor-secret-sync -- ls /mnt/secrets/

# Storage
kubectl get nodes.longhorn.io -n longhorn-system
kubectl get volume.longhorn.io -n longhorn-system
kubectl get cluster -A

# GPU
kubectl exec -n ollama deploy/ollama -- nvidia-smi

# Bifrost routing
curl -I https://auth.madhan.app
curl -I https://grafana.madhan.app

# NetBird tunnel (from Bifrost VPS)
docker exec netbird-agent netbird status
```

---

## Breaking Change Reference

| Upgrade | Breaking change |
|---------|----------------|
| Talos minor bump | k8s version embedded — check `cdk8s.yaml` `k8s@` version |
| Cilium any | `CiliumLoadBalancerIPPool` / `BGPCiliumPeeringPolicy` field renames — check `cilium.go:208,225` |
| ArgoCD chart major | TLSRoute backend service name hash changes — check `argocd.go:148` |
| Grafana 10 → 11 | Angular plugins removed; datasource API changed |
| Longhorn minor | Upgrade one minor at a time only |
| Rancher any | `--kube-version` flag must match cluster k8s version |
| NetBird any | All components (server/dashboard/proxy/agent/peer) must be on identical version |
| Authentik any | Migration race: `bootstrap.sh` runs `ak migrate` first; if health check times out SSH + force-migrate (see Phase 7) |
| Authentik any | After `netbird-agent` restart, verify `ip rule show \| grep 7120` — missing rules = restart agent again |
| Trivy any | `trivyVersion` const in `trivy.go:22` and `cdk8s.yaml` import must match |
| ComfyUI any | Only `cu128-megapak-*` tags work on RTX 5070 Ti — do not use `latest-cu128` |
