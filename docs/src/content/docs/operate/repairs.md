---
title: Repairs
description: The just recipes and one-liners that fix known failure modes, and when to run each.
sidebar:
  order: 3
---

Each recipe below fixes a failure this cluster has actually hit. Run them from the repo root on a laptop with cluster access. Full recipe list: [justfile reference](/reference/justfile/).

## Recipes

| Recipe | When to run | What it does |
|---|---|---|
| `just gateway-repair-l2` | LAN times out on the Gateway IP while in-cluster requests to it still succeed, typically after a Cilium agent restart | Checks port 80 on the Gateway LoadBalancer IP; if unreachable, deletes its `cilium-l2announce-…` lease to force re-election, then waits up to a minute for recovery |
| `just longhorn-repair-iscsi` | No Longhorn volume on a node attaches after a Talos upgrade; volumes cycle `faulted` | In every instance-manager pod, enters the host `iscsid` namespaces, reads the parameter `iscsiadm` rejects, and strips it from `/var/lib/iscsi/nodes/`. Idempotent; `just talos-upgrade` runs it for you |
| `just comfyui on` / `off` | Before using ComfyUI; afterwards, to free VRAM for Ollama | Rewrites the `// comfyui-replicas` literal in [`comfyui.go`](src:workloads/ai/comfyui.go), runs `just synth`, commits and pushes. Git is the only place a scale sticks: Argo CD self-heal reverts `kubectl scale` |
| `just argocd-sync <app>` | Auto-sync is paused (`argocd_manual_sync: true`) or one app needs a sync now | Patches the Application with a sync operation (prune, `CreateNamespace`, `ServerSideApply`); no `argocd` CLI needed |
| `just talos-health` | Before and after any node change; must be clean before the next node | `talosctl health` against the first control plane via the VIP |
| `just talos-versions` | Checking an upgrade landed on every node | `talosctl version` on every node in `config/config.yml` |
| `just headlamp-token [duration]` | Logging in to Headlamp | Issues a short-lived cluster-admin token for `headlamp-admin` (default 24h) |
| `just openbao-token` | You need root on OpenBao | Generates a root token over the in-pod unix socket from the unseal key Secret. Refuses to print to a terminal: capture it with `export ROOT_TOKEN=$(just openbao-token)` |
| `just openbao-revoke [token]` | When done with a root token | Revokes the token passed, or `$ROOT_TOKEN` |
| `just openbao-get <path> [field]` | Reading one secret | Mints a throwaway root token, reads the KV path, revokes the token on exit |
| `just openbao-sa-get <sa> <ns> <role> <path> [field]` | Proving an app's role can read its secret | Logs in with the service account's token through Kubernetes auth, then reads the path |
| `just grafana-init` | Grafana fails to start because provisioning references org 2 | Creates Grafana's "Public" org and checks it got id 2. Idempotent |
| `just ping_scan` | Finding what is alive on the LAN | `nmap -sn` over the cluster subnet |

## Without a recipe

### Failed DaemonSet pods after DiskPressure

Pods evicted under DiskPressure stay `Failed`, and the DaemonSet does not replace them once the pressure clears, so Longhorn and the CSI driver stay down on that node.

```bash frame="terminal"
kubectl get pods -A --field-selector=status.phase=Failed
kubectl delete pods -A --field-selector=status.phase=Failed
```

### Stale L2 lease on a Service other than the Gateway

`gateway-repair-l2` covers only the Gateway. Delete the lease of any other LoadBalancer Service that answers in-cluster but not from the LAN:

```bash frame="terminal"
kubectl delete lease -n kube-system cilium-l2announce-<namespace>-<service>
```

### Node hangs on reboot

A node with a broken Cilium datapath cannot reboot gracefully: unmounting a Longhorn RWX volume blocks on a share-manager ClusterIP it can no longer reach. Skip the graceful path:

```bash frame="terminal"
talosctl --talosconfig ~/.talos/config reboot --nodes <ip> --mode powercycle
```

If the Talos API is dead too, reset the VM from Proxmox. Re-check the Gateway afterwards: a restarted agent is when a stale L2 lease appears.

### RWO volume deadlocked during a rollout

The new pod waits on a volume the old pod still holds. Force-delete the old pod, or scale its old ReplicaSet to zero:

```bash frame="terminal"
kubectl scale replicaset <old-rs> -n <namespace> --replicas=0
```
