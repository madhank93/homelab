+++
title = "Version Upgrade Guide"
description = "End-to-end procedure for upgrading all homelab components: Talos, platform tools, cloud services, and workloads."
weight = 70
+++

Version upgrades follow a strict layer order — each layer depends on the one below it being stable first.

---

## What is pinned right now

Every current version lives in one place — the
[Software Inventory](/architecture/software-inventory). It is not repeated here,
so there is nothing to keep in sync.

---

## Upgrade Order

Components must be upgraded in layer order. Never skip layers.

```
Talos
  └─► Cilium  (CNI must be compatible with Talos k8s version)
        └─► Gateway API CRDs  (vendored in core/platform/manifests/)
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

# All workload chart repos, then the latest two versions of each
while read -r name url chart; do
  helm repo add "$name" "$url" >/dev/null
done <<'REPOS'
longhorn       https://charts.longhorn.io                                          longhorn/longhorn
vm             https://victoriametrics.github.io/helm-charts                       vm/victoria-metrics-k8s-stack
grafana        https://grafana-community.github.io/helm-charts                     grafana/grafana
harbor         https://helm.goharbor.io                                            harbor/harbor
openbao        https://openbao.github.io/openbao-helm                              openbao/openbao
csi-driver     https://kubernetes-sigs.github.io/secrets-store-csi-driver/charts   csi-driver/secrets-store-csi-driver
headlamp       https://kubernetes-sigs.github.io/headlamp                          headlamp/headlamp
kyverno        https://kyverno.github.io/kyverno                                   kyverno/kyverno
trivy          https://aquasecurity.github.io/helm-charts                          trivy/trivy-operator
otel           https://open-telemetry.github.io/opentelemetry-helm-charts          otel/opentelemetry-collector
ollama         https://otwld.github.io/ollama-helm                                 ollama/ollama
reloader       https://stakater.github.io/stakater-charts                          reloader/reloader
cnpg           https://cloudnative-pg.github.io/charts                             cnpg/cloudnative-pg
metrics-server https://kubernetes-sigs.github.io/metrics-server                    metrics-server/metrics-server
REPOS
helm repo update >/dev/null

for chart in longhorn/longhorn vm/victoria-metrics-k8s-stack vm/victoria-logs-single \
  grafana/grafana harbor/harbor openbao/openbao csi-driver/secrets-store-csi-driver \
  headlamp/headlamp kyverno/kyverno trivy/trivy-operator otel/opentelemetry-collector \
  ollama/ollama reloader/reloader cnpg/cloudnative-pg metrics-server/metrics-server; do
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
// core/platform/talos.go — the talosVersion const
talosVersion = "vX.Y.Z"
```

**Rules:**
- Upgrade one minor version at a time (1.12 → 1.13, not 1.12 → 1.15)
- Workers first, then GPU workers, then control planes — verify health between nodes
- Check the k8s version embedded in the new Talos release — if it bumps, update `k8s@X.Y.Z` in `workloads/cdk8s.yaml` and re-run `cdk8s import`

> **`just core talos up` does not upgrade a running node.** Images download to a
> fixed filename, so the Proxmox file ID never changes and existing disks keep
> their old Talos. Only newly created VMs get the new image.

Upgrade each node explicitly:

```bash
just talos-upgrade 192.168.1.222          # worker
just talos-upgrade 192.168.1.224 gpu      # GPU worker (different schematic)
just talos-upgrade 192.168.1.224 gpu false  # skip the drain
just talos-health                          # must be clean before the next node
```

Pass `drain=false` for a node whose pods cannot be evicted — Longhorn gives each
instance-manager a PDB allowing zero disruptions while a volume is attached, and
a single-instance CNPG cluster can never release its only pod. The drain then
burns its timeout and `talosctl` exits *before* rebooting, leaving the node
cordoned and un-upgraded.

Kubernetes does not ride along: `just talos-upgrade-k8s <version>` is separate.

### After each node

`just talos-upgrade` runs `just longhorn-repair-iscsi` for you. It exists
because an upgrade can ship an open-iscsi that rejects a parameter its
predecessor wrote into `/var/lib/iscsi/nodes/`. `iscsiadm` reads *every* record,
so one unparseable file fails the whole call and no Longhorn volume on that node
can attach — the engine's frontend never starts, Longhorn marks the volume
faulted, and it retries forever. Those records live on persistent `/var`, so
rebooting does not clear them.

### If a node will not come back

A node whose Cilium datapath does not recover cannot reboot gracefully: the
unmount of a Longhorn RWX volume blocks on an NFS share-manager ClusterIP it can
no longer reach, so `talosctl reboot` hangs at `unmountPodMounts` forever. `cri`
and `kubelet` read `Finished/Fail` while the Talos API stays up, because `apid`
survives the shutdown sequence. Skip the graceful path:

```bash
talosctl --talosconfig ~/.talos/config reboot --nodes <ip> --mode powercycle
```

---

## Phase 2 — Cilium

**Risk:** High — CNI restart interrupts pod networking briefly.

```go
// core/platform/cilium.go — Version on the cilium Release
Version: pulumi.String("1.X.Y"),
```

**Rules:**
- Verify compatibility with new Talos k8s version: https://docs.cilium.io/en/stable/network/kubernetes/compatibility/
- `wt0` must **not** be added to Cilium devices — keep only `eth0` in `cilium.go`. See [NetBird routing notes](../infrastructure/#netbird).
- After upgrade check the `CiliumLoadBalancerIPPool` and `CiliumL2AnnouncementPolicy` CR specs in `cilium.go` for field renames

```bash
just core platform up
kubectl -n kube-system rollout status daemonset/cilium
kubectl get gateway -n kube-system homelab-gateway
```

### If every LoadBalancer IP goes dark afterwards

An agent restart can leave the L2 announcement lease holder renewing normally,
still listing the IP in `db/show l2-announce`, while it has quietly stopped
answering ARP. Nothing reports unhealthy — the Gateway stays `Programmed=True`
and Envoy stays ready on every node. The fingerprint is that **in-cluster
requests to the Gateway return 200 while the LAN times out**:

```bash
just gateway-repair-l2
```

It checks that discriminator before touching anything, then re-elects onto
another node. Ignore `Proxy Status: 0 redirects active` while debugging — with
`Envoy: external` the Gateway listeners live in the `cilium-envoy` DaemonSet and
are never counted there.

Pulumi reports the Cilium release as failed (`failed to become available within
allocated timeout`) if a single agent never goes ready, even when the DaemonSet
reached every node on the new image. Fix the node, then re-run
`just core platform up` — it is idempotent.

---

## Phase 3 — ArgoCD

**Risk:** Medium — GitOps engine downtime during pod restart.

```go
// core/platform/argocd.go — Version on the argo-cd Release
Version: pulumi.String("X.Y.Z"),
```

**Image ahead of chart:** the Release overrides `image.tag` to run a server
newer than the chart pins. Check on every chart bump whether the override is
still needed and drop it once the chart catches up.

```bash
just core platform up
kubectl rollout status deployment argocd-server -n argocd
```

---

## Phase 4 — cert-manager

**Risk:** Low — but high blast radius if broken (TLS for all services).

```go
// core/platform/cert_manager.go — Pulumi, not cdk8s
Version: pulumi.String("vX.Y.Z"),
```

```bash
just core platform up
kubectl get certificate -A   # must all show Ready=True
```

> **Never add a second Certificate for `madhan.app` / `*.madhan.app`.** Let's
> Encrypt allows 5 per exact identifier set per 168h. Cilium mirrors the one in
> `kube-system` into `cilium-secrets` itself — see
> [cert-manager](/platform/cert-manager). A `429 rateLimited` error, or a
> Certificate whose `Revision:` is in the dozens, means duplicates are competing
> for that budget. Do not delete and recreate the Certificate; that spends more
> of it.

---

## Phase 5 — OpenBao + CSI Driver

**Risk:** Medium — all running apps read secrets through this layer.

```go
// workloads/secrets/openbao.go — chart Version
Version: jsii.String("0.X.Y"),
// workloads/secrets/openbao.go — unseal sidecar image
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
// workloads/databases/cnpg.go
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

**NetBird rule:** All four NetBird components (`server`, `dashboard`, `reverse-proxy`, agent on Bifrost) **must be on the same version**. Also update both image tags in `workloads/networking/netbird_peer.go` — the
`setup-iptables` init container and the main container.

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
| Falco | `security/falco.go` |
| Metrics Server | `cdk8s.yaml` |
| Reloader | `support/reloader.go` |
| OTel Collector | `observability/otel_collector.go` — two releases, `otel-agent` and `otel-gateway`, both must move |
| Headlamp | `cdk8s.yaml` |
| Trivy | `cdk8s.yaml` (chart) **and** `security/trivy.go` (`trivyVersion`, builds the CRD fetch URL) — deliberately independent, but a chart bump usually wants a CRD bump |

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

Also update the chart Version in `workloads/observability/victoria_metrics.go`. Run `helm diff upgrade` first — the values schema changes frequently between minor versions.

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

Also update the device plugin and DCGM exporter versions in `workloads/hardware/nvidia_gpu_operator.go`. Verify RTX 5070 Ti (sm_120, Blackwell) still supported in the new operator release — Blackwell support was added in 570.x driver series.

```bash
kubectl exec -n ollama deploy/ollama -- nvidia-smi
```

---

## Phase 10 — Workloads: Complex


### n8n

```go
// workloads/automation/n8n.go — chart Version
Version: jsii.String("2.X.Y"),
```

Check latest: `helm show chart oci://8gears.container-registry.com/library/n8n`. DB schema migrations run automatically on pod start via CNPG. Verify n8n healthy after upgrade:

```bash
kubectl rollout status deployment n8n -n n8n
```

### Ollama

Bump the chart in `cdk8s.yaml`. There is no image tag to change: it is left
unset so Ollama tracks the chart's appVersion. Downloaded models stay in the
PVC — no re-pull needed.

```bash
kubectl exec -n ollama deploy/ollama -- ollama list
```

### ComfyUI

```go
// workloads/ai/comfyui.go
Image: jsii.String("yanwk/comfyui-boot:cu130-megapak-ptXXX-YYYYMMDD"),
```

Check [yanwk/comfyui-boot tags](https://hub.docker.com/r/yanwk/comfyui-boot/tags).
The RTX 5070 Ti is Blackwell (sm_120) and needs CUDA 12.8 or newer, so `cu126`
tags will not run. The `cu128-megapak` line stopped building in May 2026;
`cu130-megapak-pt*` is the current one. There is no `latest-*` tag.

ComfyUI is scaled to zero by default — `just comfyui on` after the bump, then
`just comfyui off` when finished. See [ComfyUI](@/workloads/ai/comfyui/index.md).

### NetBird Peer (k8s)

```go
// workloads/networking/netbird_peer.go  (both init + main containers)
Image: jsii.String("netbirdio/netbird:0.X.Y"),
```

Must match Bifrost docker-compose NetBird version exactly. Upgrade Bifrost and k8s peer in the same commit.

### Kubeflow

Kubeflow is deployed via Kustomize from `kubeflow/manifests`, not a Helm chart. Upgrade by updating the `ref` on each base in `workloads/ai/kubeflow/kustomization.yaml`. Check the [kubeflow/manifests releases](https://github.com/kubeflow/manifests/releases) for the version compatible with the current k8s version.

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
| Cilium any | `CiliumLoadBalancerIPPool` / `CiliumL2AnnouncementPolicy` field renames — check `cilium.go` |
| Grafana 10 → 11 | Angular plugins removed; datasource API changed |
| Longhorn minor | Upgrade one minor at a time only |
| NetBird any | All components (server/dashboard/proxy/agent/peer) must be on identical version |
| Authentik any | Migration race: `bootstrap.sh` runs `ak migrate` first; if health check times out SSH + force-migrate (see Phase 7) |
| Authentik any | After `netbird-agent` restart, verify `ip rule show \| grep 7120` — missing rules = restart agent again |
| Trivy any | `trivyVersion` const in `trivy.go` drives the CRD bundle; the chart comes from `cdk8s.yaml` |
| ComfyUI any | Needs a CUDA 12.8+ tag for sm_120; the `cu128` line is discontinued |
