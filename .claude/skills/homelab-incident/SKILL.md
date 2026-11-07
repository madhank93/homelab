---
name: homelab-incident
description: "Homelab incident triage for the live Talos cluster — symptom to first checks to likely cause to fix, ordered by blast radius: API or VIP 192.168.1.210 down, Cilium datapath or CrashLoop, Gateway LoadBalancer IP dark from a stale L2 lease (just gateway-repair-l2), new HTTPRoute 404 from a stale cilium-operator watch, Argo CD down or empty, Longhorn volumes not attaching (EngineImages, just longhorn-repair-iscsi), Failed DaemonSet pods after DiskPressure, OpenBao sealed, CSI secrets not mounting, NetBird public path 504, worker4 GPU or eGPU gone, flaky DNS. Read-only diagnosis first; fixes are GitOps, the justfile repair recipes or pod restarts. Use whenever something in the cluster is down, unreachable, stuck Pending or ContainerCreating, returning 404/503/504, or after a power outage, node reboot, Cilium restart or control-plane restart. Do NOT use for planned upgrades — use homelab-upgrade. Do NOT use for Talos machine config changes — use homelab-talos. Do NOT use for debugging Go code — use golang-troubleshooting. Do NOT use for a failed pulumi up — use pulumi-debug-failed-operation."
---

# Homelab incident triage

Diagnose read-only, find the layer, then apply the smallest fix the rules allow. The symptom index and runbooks already live in the docs: [troubleshooting](../../../docs/src/content/docs/operate/troubleshooting.md) and [repairs](../../../docs/src/content/docs/operate/repairs.md). This skill adds the triage order and the guardrails.

## Rules of engagement

- Diagnose with `kubectl get/describe/logs/events`, `talosctl get/health` and the `just` health recipes before changing anything.
- Allowed fixes, in order of preference: a code change pushed to the release branch (see "Deploy flow" in [CLAUDE.md](../../../CLAUDE.md)); a repair recipe below; deleting a pod, lease or `Failed` pod so a controller recreates it.
- Never `kubectl apply/patch/edit/scale` as a fix: Argo CD `selfHeal` reverts it. `just argocd-sync <app>` is the one sanctioned patch.
- Never read or print Secret or OpenBao data, and never run `openbao-*` or `sops-*` recipes or `just core`. Hand those to the user.
- Ask before anything irreversible or disruptive: node reboot or power cycle, force-deleting a pod holding an RWO volume, deleting a Deployment, detaching a Longhorn volume, touching the Proxmox host.

## First five commands

```bash
just talos-health                                   # nodes, etcd, kubelet
kubectl get nodes -o wide
kubectl get pods -A | grep -Ev 'Running|Completed'
kubectl get applications -n argocd
kubectl get events -A --sort-by=.lastTimestamp | tail -30
```

`kubectl api-resources 2>&1 >/dev/null` printing anything means broken API discovery.

## Triage table (highest blast radius first)

Work top-down: a lower row is often a symptom of a higher one.

| # | Symptom | First checks | Likely cause | Fix |
|---|---|---|---|---|
| 1 | `kubectl` times out on `192.168.1.210` | `talosctl --talosconfig ~/.talos/config get addresses --nodes 192.168.1.211,192.168.1.212,192.168.1.213`; `etcd members` | No controller holds the VIP; etcd lost quorum | [Control plane HA](../../../docs/src/content/docs/platform/control-plane-ha.md); stale talosconfig: [Talos](../../../docs/src/content/docs/platform/talos.md#troubleshooting) |
| 2 | Node `NotReady`, Cilium agent `0/1` or CrashLoop exit 137 | `kubectl get pods -n kube-system -l app.kubernetes.io/name=cilium-agent -o wide` | Corrupt BPF state; datapath wedged | Reboot the node (confirm first); hangs → [references/cilium.md](references/cilium.md) |
| 3 | Every `*.madhan.app` times out on LAN, works in-cluster | Lease holder of `cilium-l2announce-kube-system-cilium-gateway-homelab-gateway` | Stale L2 lease after an agent restart | `just gateway-repair-l2`; other Services: [references/cilium.md](references/cilium.md) |
| 4 | New HTTPRoute has no `status.parents`, host 404s, old routes fine | Gateway `attachedRoutes` vs `kubectl get httproute -A` count | Stale operator watch after apiserver restarts | Delete the operator leader pod: [references/cilium.md](references/cilium.md) |
| 5 | Argo CD UI down, or lists no applications | `kubectl get pods -n argocd -o wide`, Cilium on those nodes, `kubectl api-resources` | Cilium on its node (row 2); stale APIService | [Argo CD](../../../docs/src/content/docs/platform/argocd.md#troubleshooting) |
| 6 | PVCs `Pending`, pods `ContainerCreating`, volumes `attaching`/`faulted` | `kubectl get engineimages -n longhorn-system`, `kubectl get volumes.longhorn.io -n longhorn-system` | EngineImage not on all nodes; bad iSCSI records after upgrade | [references/storage.md](references/storage.md) |
| 7 | Longhorn `ManagerPodDown` or `driver name secrets-store.csi.k8s.io not found` after DiskPressure | `kubectl get pods -A --field-selector=status.phase=Failed` | Rejected DaemonSet pods stay `Failed` | Delete the `Failed` pods: [references/storage.md](references/storage.md) |
| 8 | Apps with secrets hang in `ContainerCreating` | `kubectl exec -n openbao openbao-0 -c openbao -- bao status`; `kubectl logs -n openbao openbao-0 -c unseal` | OpenBao sealed, unseal sidecar broken | [OpenBao](../../../docs/src/content/docs/workloads/secrets/openbao.md#sealed-after-a-restart); hand-unseal is the user's step |
| 9 | Pod cannot find its Secret, `failed to get secretproviderclass`, 403 on mount | `kubectl get secretproviderclass,secretproviderclasspodstatus -n <ns>`; CSI provider logs | No pod mounts the CSI volume; SPC missing; role/SA mismatch | [CSI driver](../../../docs/src/content/docs/workloads/secrets/csi-driver.md#troubleshooting), [Secrets](../../../docs/src/content/docs/concepts/secrets.mdx) |
| 10 | Public URLs return 504, LAN works | `kubectl exec -n netbird netbird-peer-0 -c netbird -- netbird status` | Tunnel down, stale route peer, wt0 0 bytes | [references/netbird.md](references/netbird.md) |
| 11 | worker4 missing, GPU pods `Pending`, CUDA invisible, Ollama cannot load | `kubectl get node k8s-worker4`; `kubectl describe node k8s-worker4 \| grep nvidia.com/gpu` | eGPU enclosure off after power loss; ComfyUI holds VRAM; pod spec | [references/gpu.md](references/gpu.md) |
| 12 | Intermittent name resolution failures on one node | `kubectl get pods -n kube-system -l k8s-app=kube-dns -o wide`; `talosctl --talosconfig ~/.talos/config get resolvers --nodes <ip>` | Node upstream resolver, or CoreDNS on a broken-datapath node | Restart CoreDNS pods; node resolvers come from [Nodes](../../../docs/src/content/docs/hardware/nodes.mdx#addressing) |

Certificate errors, Grafana, Falco, Harbor `Multi-Attach` and the rest: [troubleshooting](../../../docs/src/content/docs/operate/troubleshooting.md).

## After a power outage

Recover in this order, checking each before the next:

1. Proxmox host up, all seven VMs running; worker4 not starting → [references/gpu.md](references/gpu.md).
2. `just talos-health` clean; VIP answers (row 1).
3. Every Cilium agent ready (row 2), then the Gateway IP from the LAN (row 3).
4. OpenBao unsealed (row 8), Longhorn volumes attached (row 6), no `Failed` pods (row 7).
5. Argo CD apps `Synced/Healthy`, then the public path (row 10).

## Repair recipes

| Recipe | Fixes |
|---|---|
| `just gateway-repair-l2` | Gateway LB IP dark on the LAN; checks port 80 first, no-op when reachable |
| `just longhorn-repair-iscsi` | Unparseable iSCSI node records after a Talos upgrade; idempotent |
| `just comfyui off` | VRAM held by ComfyUI; commits and pushes |
| `just argocd-sync <app>` | Sync one app now, or while `argocd_manual_sync` is true |
| `just talos-health` / `just talos-versions` | Cluster health; versions per node |

Full list: [justfile reference](../../../docs/src/content/docs/reference/justfile.md).

## Done means

```bash
just talos-health                                        # clean
kubectl get nodes                                        # all Ready
kubectl get pods -A | grep -Ev 'Running|Completed'       # nothing unexpected
kubectl get applications -n argocd                       # Synced / Healthy
just gateway-repair-l2                                   # "reachable — nothing to do"
```

Then confirm the original symptom is gone from where it was seen (LAN, public URL, or the app). If the fix was a code change, it is pushed to the release branch and `just check` passed. If the incident revealed a new failure mode, add a row to the troubleshooting page (load `homelab-docs`).
