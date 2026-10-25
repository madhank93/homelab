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
  └─► Gateway API CRDs  (vendored in core/platform/manifests/; at least what the next Cilium needs)
        └─► Cilium  (CNI must be compatible with Talos k8s version)
              └─► Argo CD  (GitOps engine)
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

```yaml
# config/versions.yaml
talos: {version: vX.Y.Z, base_schematic: ..., gpu_schematic: ...}
kubernetes: {version: X.Y.Z}
```

**Rules:**
- Upgrade one minor version at a time (1.12 → 1.13, not 1.12 → 1.15)
- Control planes first, one at a time, then workers, then the GPU worker last — verify health between nodes. Talos upgrades both roles the same way and leaves Kubernetes alone; `just talos-upgrade-k8s` orders the Kubernetes rollout itself
- Schematic IDs hash only the extension list, so they survive a version bump; check that every extension still exists for the new release: `curl -s https://factory.talos.dev/version/vX.Y.Z/extensions/official | jq -r '.[].name'`
- Kubernetes is pinned on its own line and must stay inside Cilium's supported range; when it moves a minor, update `k8s@X.Y.Z` in `workloads/cdk8s.yaml` and re-run `cdk8s import`
- Never lower `talos.version` and run `just core talos up`: the provider regenerates the cluster secrets on a downgrade (the resource is protected, so the run fails instead)

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

Kubernetes does not ride along: `just talos-upgrade-k8s` rolls out the pinned
version as a separate step. Run it before `just core talos up`, which would
otherwise push the new images to every node at once.

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

### Talos 1.13 → 1.14

Talos 1.14 defaults to Kubernetes 1.37, which Cilium does not support yet, so
`kubernetes.version` in `config/versions.yaml` holds the cluster on 1.36. An
upgraded node keeps its v1alpha1 config: the deprecated fields this repo sets
(`machine.install`, `machine.kernel`, `machine.nodeLabels`, `machine.nodeTaints`,
`machine.kubelet.nodeIP`) still work, and 1.14-only defaults such as workload
isolation (`SecurityProfileConfig`) and periodic trim stay off. Leave workload
isolation off: it moves the kubelet and pods into a sandbox PID namespace, and
Longhorn's iSCSI frontend and `just longhorn-repair-iscsi` both reach into the
host `iscsid` namespaces.

Use a `talosctl` client of 1.14.1 or later (1.14.0 can hang the full timeout
after a node is already back). Then, one node at a time:

```bash
just talos-versions                         # every node on 1.13.x
just talos-health

# Control planes, one at a time. Talos refuses an upgrade that breaks quorum.
just talos-upgrade 192.168.1.211
just talos-upgrade 192.168.1.212
just talos-upgrade 192.168.1.213

# Workers. Start with one and prove Longhorn attaches there before the rest.
# worker1 runs the NetBird routing peer: public traffic drops while it reboots.
just talos-upgrade 192.168.1.223
just talos-upgrade 192.168.1.222
just talos-upgrade 192.168.1.221

# GPU worker last: new NVIDIA driver build, and its PDBs need drain=false.
just talos-upgrade 192.168.1.224 gpu false
```

Gate before every next node:

```bash
just talos-health                                   # clean
kubectl get nodes                                   # all Ready, none SchedulingDisabled
talosctl --talosconfig ~/.talos/config -n 192.168.1.211,192.168.1.212,192.168.1.213 etcd status
                                                    # 3 members, one leader, no errors
kubectl -n longhorn-system get volumes.longhorn.io \
  -o custom-columns=NAME:.metadata.name,ROBUSTNESS:.status.robustness | grep -E 'degraded|faulted'
                                                    # empty: replicas rebuilt
```

After worker4, check the GPU before calling it done:

```bash
talosctl --talosconfig ~/.talos/config -n 192.168.1.224 get extensions   # nvidia-* present
kubectl get node k8s-worker4 -o jsonpath='{.status.allocatable.nvidia\.com/gpu}'  # 5
```

Then Kubernetes, then Pulumi:

```bash
just talos-upgrade-k8s --dry-run
just talos-upgrade-k8s
just core talos preview
```

The preview should show updates only: `talos-secrets` (version contract), the
two image downloads, and the per-node config applies. A `replace` on
`talos-secrets` or a VM means stop. Then `just core talos up`.

Rollback before `just core talos up`: `talosctl rollback --nodes <ip>` boots the
previous A/B image with data intact, then revert the commit. After it, never
lower `talos.version` — roll nodes back with `talosctl rollback` and leave the
pin alone.

---

## Phase 2 — Cilium

**Risk:** High — CNI restart interrupts pod networking briefly.

```yaml
# config/versions.yaml
cilium: {repo: https://helm.cilium.io/, version: 1.X.Y}
```

**Rules:**
- One minor at a time (1.18 → 1.19 → 1.20), each on the latest patch of its line. That is the only path Cilium tests, for upgrade and for rollback.
- Verify compatibility with new Talos k8s version: https://docs.cilium.io/en/stable/network/kubernetes/compatibility/
- Read the version's upgrade notes for removed Helm values, then render the old and new chart with the values from `cilium.go` and diff the `cilium-config` ConfigMap (`helm template cilium cilium/cilium --version X -f values.yaml`).
- Gateway API CRDs go first, as their own apply: each Cilium minor needs a minimum Gateway API release, and the operator only enables the controller for CRDs present when it starts.
- Never add `wt0` to Cilium's `devices` list in `cilium.go` — it stays `["ens18", "eth0"]`. See [NetBird Peer](@/workloads/networking/netbird-peer/index.md).
- After upgrade check the `CiliumLoadBalancerIPPool` and `CiliumL2AnnouncementPolicy` CR specs in `cilium.go` for field renames

```bash
just core platform up
kubectl -n kube-system rollout status daemonset/cilium
kubectl get gateway -n kube-system homelab-gateway
```

### Gates

Run all of these from the laptop on the LAN before the first change, and again
after every apply. Do not start the next step until each one passes.

```bash
# 1. Every route Accepted by the Gateway. Any output is a failure.
kubectl get httproute,grpcroute -A -o json | jq -r '.items[]
  | select([.status.parents[]?.conditions[]? | select(.type=="Accepted" and .status=="True")] | length == 0)
  | "\(.kind) \(.metadata.namespace)/\(.metadata.name)"'
kubectl get gateway -n kube-system homelab-gateway   # PROGRAMMED True

# 2. Every LoadBalancer IP answers from the LAN, not just from inside the cluster.
kubectl get svc -A -o json | jq -r '.items[] | select(.spec.type=="LoadBalancer")
  | "\(.status.loadBalancer.ingress[0].ip) \(.spec.ports[0].port) \(.metadata.namespace)/\(.metadata.name)"' |
  while read -r ip port name; do
    nc -z -w 5 "$ip" "$port" && echo "ok   $name $ip:$port" || echo "FAIL $name $ip:$port"
  done
kubectl get lease -n kube-system | grep l2announce   # one lease per LoadBalancer Service

# 3. Host TCP on every node — the 1.19 regression killed this while ICMP still worked.
talosctl --talosconfig ~/.talos/config -n 192.168.1.211,192.168.1.212,192.168.1.213,192.168.1.221,192.168.1.222,192.168.1.223,192.168.1.224 version --short
just talos-health

# 4. No LoadBalancer IP is also a node IP. Any output: stop, see below.
comm -12 \
  <(kubectl get svc -A -o jsonpath='{range .items[*]}{.status.loadBalancer.ingress[*].ip}{"\n"}{end}' | tr ' ' '\n' | grep . | sort -u) \
  <(kubectl get nodes -o jsonpath='{range .items[*]}{.status.addresses[?(@.type=="InternalIP")].address}{"\n"}{end}' | sort -u)
```

Gate 4 exists because LB-IPAM hands out any address in the pool, node IPs
included, and the pool in `config/config.yml` spans worker IPs. From 1.19 the
datapath drops traffic to an LB-IPAM address on any port no Service uses
([cilium/cilium#44430](https://github.com/cilium/cilium/issues/44430)); 1.19.5
spares only the node that owns the IP, so a worker whose IP is also an LB IP
loses SSH-equivalent access (Talos API, kubelet) from everywhere else, and two
nodes answer ARP for it. If gate 4 prints anything, take the node IPs out of the
pool before touching Cilium.

### 1.18 → 1.20 runbook

Three applies, one commit each on the release branch — `git log --oneline --
config/versions.yaml core/platform/manifests/` lists them. Check out each in
turn, run `just core platform preview` and then `just core platform up`, and
pass every gate before the next.

1. **Gateway API CRDs v1.6.3, Cilium stays on 1.18.12.** The preview must show
   the CRDs updated, the `BackendLBPolicy` CRD deleted, the `safe-upgrades`
   admission policy created, and `cilium-ip-pool` as an in-place update to
   `cilium.io/v2` — never a replace, which would release every LoadBalancer IP.
   The experimental bundle keeps TLSRoute `v1alpha2`, which the 1.18 and 1.19
   operators refuse to start without
   ([cilium/cilium#45139](https://github.com/cilium/cilium/issues/45139)).
   Prove it before moving on:
   ```bash
   kubectl -n kube-system rollout restart deployment/cilium-operator
   kubectl -n kube-system rollout status deployment/cilium-operator
   kubectl -n kube-system logs deploy/cilium-operator | grep -i 'gateway' | grep -i -E 'error|failed'   # no output
   ```
2. **Cilium 1.19.10.** The minimum is 1.19.5, which carries the fix for
   [#44430](https://github.com/cilium/cilium/issues/44430) and
   [#46010](https://github.com/cilium/cilium/issues/46010). Watch gate 3 first:
   if a node's Talos API stops answering, roll back (below) before the agent
   reaches the rest of the nodes. 1.19 is not tested against Kubernetes 1.36,
   so do not stay on it longer than the gates take.
   ```bash
   kubectl -n kube-system rollout status daemonset/cilium --timeout=15m
   kubectl -n kube-system rollout status daemonset/cilium-envoy
   kubectl -n kube-system rollout status deployment/cilium-operator
   kubectl -n kube-system exec ds/cilium -c cilium-agent -- cilium-dbg status --brief
   ```
3. **Cilium 1.20.4.** Same commands and gates. An empty `l2announce` lease list
   with LB IPs still assigned is
   [#48799](https://github.com/cilium/cilium/issues/48799) (Services with
   `externalTrafficPolicy: Local` lose their announcement); the Gateway Service
   uses `Cluster` and is not affected.

### Rollback

Back one minor at a time, the reverse of the way up: 1.20 → 1.19.10 → 1.18.12.
Check out the previous commit (or revert the bump) and `just core platform up`;
Pulumi runs a `helm upgrade` to the older chart. If the API is reachable but
Pulumi is not an option, `helm history cilium -n kube-system` and
`helm rollback cilium <revision> -n kube-system`, then re-run Pulumi from the
matching commit so its state agrees.

Leave the Gateway API CRDs and the `cilium.io/v2` pool in place: 1.18 and 1.19
run on both. Rolling the CRDs below the 1.5 line is refused by the
`safe-upgrades.gateway.networking.k8s.io` admission policy, and would have to
start by deleting it.

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

The recipe only covers the Gateway. For any other LoadBalancer Service that
fails gate 2 the same way, delete its lease to force a re-election:
`kubectl delete lease -n kube-system cilium-l2announce-<namespace>-<service>`.

Pulumi reports the Cilium release as failed (`failed to become available within
allocated timeout`) if a single agent never goes ready, even when the DaemonSet
reached every node on the new image. Fix the node, then re-run
`just core platform up` — it is idempotent.

A node whose agent will not recover cannot be rebooted gracefully either: the
Longhorn RWX unmount hangs on a ClusterIP the broken datapath cannot reach.
Power-cycle it instead (see [If a node will not come back](#if-a-node-will-not-come-back)):

```bash
talosctl --talosconfig ~/.talos/config reboot --nodes <ip> --mode powercycle
```

If the Talos API itself is dead on that node, reset the VM from Proxmox. After
it returns, re-run the gates — a restarted agent is exactly when a stale L2
lease shows up.

---

## Phase 3 — Argo CD

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

Helm chart version and image tag must be compatible — check the OpenBao [tags](https://github.com/openbao/openbao/tags).

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

All image tags live in `core/cloud/bifrost/docker-compose.yml`:

| Image | Note |
|---|---|
| `traefik` | |
| `netbirdio/netbird-server` | |
| `netbirdio/reverse-proxy` | |
| `netbirdio/netbird` | agent |
| `netbirdio/dashboard` | **separate version line** — see below |
| `ghcr.io/goauthentik/server` | two services: server and worker |
| `postgres` | minor bumps only |

**NetBird rule:** `netbird-server`, `reverse-proxy` and the agent **must be on the
same version**. The dashboard is versioned independently (`v2.x`) and does not
follow them; take the dashboard release that was current when the server version shipped.
Also bump `netbird` in `config/versions.yaml` — the in-cluster peer reads its image from there.

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

`victoria-logs-single` is pinned in `workloads/cdk8s.yaml`:

```yaml
- helm:https://victoriametrics.github.io/helm-charts/victoria-logs-single@0.X.Y
```

`victoria-metrics-k8s-stack` is **not** in `cdk8s.yaml` — it is pinned at its call
site, on the chart Version in `workloads/observability/victoria_metrics.go`.

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

The full GPU Operator is deliberately not used. Two charts are pinned at their call
site in `workloads/hardware/nvidia_gpu_operator.go`:

- `nvidia-device-plugin` from `https://nvidia.github.io/k8s-device-plugin`
- `dcgm-exporter`

Verify the RTX 5070 Ti (sm_120, Blackwell) is still supported by the new release —
Blackwell support arrived in the 570.x driver series.

```bash
kubectl exec -n ollama deploy/ollama -- nvidia-smi
```

---

## Phase 10 — Workloads: Complex


### n8n

Chart and image are pinned in `config/versions.yaml` (`helm.n8n`, `images.n8n`).
Check latest: `helm show chart oci://8gears.container-registry.com/library/n8n`.
n8n runs its DB schema migrations on pod start; they are one-way, so an image
downgrade after a migrated start does not work — only a database restore does.

**Manual steps — back up the CNPG database before sync.** The CNPG cluster has
no scheduled backup, so take a dump before the change lands on the release
branch (CI then publishes the manifests branch and Argo CD syncs on its own):

```bash
primary=$(kubectl get pod -n n8n -l cnpg.io/cluster=n8n-pg,cnpg.io/instanceRole=primary -o name)
kubectl exec -n n8n "$primary" -c postgres -- \
  pg_dump -U postgres -d n8n -Fc > n8n-$(date +%F).dump
ls -lh n8n-$(date +%F).dump   # non-empty, keep it off-cluster
```

The first 2.x start can take several minutes of migrations (the startup probe
allows 10). Watch it, then verify:

```bash
kubectl logs -n n8n deploy/n8n -f          # migration progress, then "Editor is now accessible"
kubectl rollout status deployment n8n -n n8n
```

To roll back: scale n8n to 0, restore the dump into an empty `n8n` database
(`pg_restore -U postgres -d n8n --clean --if-exists`), and revert the pin.

**1.x → 2.x behaviour changes to check in workflows** (n8n's own list:
[2.0 breaking changes](https://docs.n8n.io/2-0-breaking-changes/)):

- Code node runs in a task runner (internal mode, child process in the n8n container); `$evaluateExpression()` no longer works there, and Code-node access to env vars is blocked (`N8N_BLOCK_ENV_ACCESS_IN_NODE=true`).
- Python Code node needs an external runner (`n8nio/runners` sidecar) — not deployed here.
- Execute Command and Local File Trigger nodes are disabled by default (`NODES_EXCLUDE`).
- Read/Write File nodes are confined to `~/.n8n-files` (`N8N_RESTRICT_FILE_ACCESS_TO`).
- Workflow activate/deactivate became Publish/Unpublish; the legacy Start node is removed.
- OAuth callbacks now require an authenticated session (`N8N_SKIP_AUTH_ON_OAUTH_CALLBACK=false`).
- In-memory binary data is gone; `N8N_DEFAULT_BINARY_DATA_MODE=filesystem` keeps it on the PVC.

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

Kubeflow is deployed via Kustomize from `kubeflow/community-distribution` (formerly `kubeflow/manifests`), not a Helm chart. Upgrade by updating the `ref` on each base in `workloads/ai/kubeflow/kustomization.yaml`. Check the [releases](https://github.com/kubeflow/community-distribution/releases) for the version compatible with the current k8s version, and read the "Version-Specific Upgrade Notes" in the upstream README — paths move between releases.

Build locally before pushing; CI runs the same command:

```bash
cd workloads && kustomize build ai/kubeflow/ > /dev/null
```

**Manual steps, 1.11.0 → 26.03.1** (dashboard moved to `kubeflow/dashboard` v2, notebooks to `notebooks-v1/`, Trainer 2.2.0):

Two Deployments changed their immutable `spec.selector`, so Argo CD cannot update them in place. Once the new manifests branch is published, delete them and let the sync recreate them:

```bash
kubectl delete deployment jobset-controller-manager -n kubeflow-system --ignore-not-found
kubectl delete deployment profiles-deployment -n kubeflow --ignore-not-found
```

The old `centraldashboard` and `admission-webhook` resources are no longer rendered, so Argo CD prunes them. The dashboard Service is now `dashboard`; the HTTPRoute in `workloads/ai/kubeflow.go` points at it.

> **Never delete the `profiles.kubeflow.org` CRD** — it cascades to every Profile and its namespace.

---

## cdk8s Import Regeneration

After changing any chart version in `cdk8s.yaml`, regenerate typed Go bindings:

```bash
cd workloads
cdk8s import        # regenerates workloads/imports/*
go mod tidy
just synth          # verify no compile errors before pushing
```

If Talos upgrades the embedded k8s version, also update the `k8s@X.Y.Z` import at the top of `cdk8s.yaml`.

---

## Post-Upgrade Verification

```bash
# Cluster health
talosctl --talosconfig ~/.talos/config health --nodes 192.168.1.210

# All Argo CD apps synced
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
| Cilium 1.18 → 1.19 | Datapath drops LB-IPAM traffic on unused ports — an LB IP equal to a node IP cuts that node off; use 1.19.5+ and pass gate 4 |
| Cilium 1.19 → 1.20 | Needs Gateway API CRDs v1.6.1+ applied first; keep the experimental channel while any TLSRoute `v1alpha2` reader remains |
| Grafana 10 → 11 | Angular plugins removed; datasource API changed |
| Longhorn minor | Upgrade one minor at a time only |
| NetBird any | All components (server/dashboard/proxy/agent/peer) must be on identical version |
| Authentik any | Migration race: `bootstrap.sh` runs `ak migrate` first; if health check times out SSH + force-migrate (see Phase 7) |
| Authentik 2026.5 → 2026.8 | Forwarded headers are honoured only from `AUTHENTIK_LISTEN__TRUSTED_PROXY_CIDRS` (the default covers `172.16.0.0/12`, so Traefik on `bifrost_net` qualifies). Upgrades refuse to skip a minor: the newest row of `authentik_version_history` must be the previous minor |
| Authentik any | After `netbird-agent` restart, verify `ip rule show \| grep 7120` — missing rules = restart agent again |
| Trivy any | `trivyVersion` const in `trivy.go` drives the CRD bundle; the chart comes from `cdk8s.yaml` |
| ComfyUI any | Needs a CUDA 12.8+ tag for sm_120; the `cu128` line is discontinued |
| n8n 1.x → 2.x | One-way DB migration — dump the CNPG database first; Code node moves to task runners, Execute Command node disabled, in-memory binary data removed |
| Kubeflow 1.11 → 26.03 | Repo renamed to `kubeflow/community-distribution`; dashboard, PodDefaults and profiles moved under `applications/dashboard/`, notebook apps under `applications/notebooks-v1/`; NetworkPolicies now ship in `common/kubeflow-namespace/base`; dashboard Service renamed `centraldashboard` → `dashboard` |
