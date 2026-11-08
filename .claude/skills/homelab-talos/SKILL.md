---
name: homelab-talos
description: "Homelab Talos machine configuration and node operations — the talos Pulumi stack in core/platform/talos.go and proxmox.go (basePatch, cpPatch, workerPatch, gpuWorkerPatch, patchTalosConfig, ConfigurationApply, protected talos-secrets and VMs), which changes reboot a node, restart a service or do nothing, node inventory and IPs in config/config.yml, Talos and Kubernetes pins in config/versions.yaml, talosctl through the VIP and refreshing ~/.talos/config from the stack output, the worker4 GPU node (dedicated=ai taint, PCIe passthrough, eGPU enclosure over Thunderbolt), and the just talos-upgrade, talos-health, talos-versions, talos-upgrade-k8s and longhorn-repair-iscsi recipes with their gotchas (install disk, drain deadlock, one node at a time). Use whenever a task edits a Talos machine patch, adds or resizes a node, upgrades Talos, cannot reach a node with talosctl, or worker4 will not start. Do NOT use for Kubernetes component or app version bumps — use homelab-upgrade. Do NOT use for live outage triage — use homelab-incident. Do NOT use for general Pulumi Go code, resource options or reading previews — use homelab-pulumi-go."
---

# Homelab Talos

Seven Talos VMs on one Proxmox host, all defined by the `talos` stack. A bad patch here can reboot every node at once or replace the cluster PKI, so the job is to know the blast radius before `preview`.

## Hard rules

- Never print `core/talosconfig`, `core/kubeconfig`, `~/.talos/config` or the `talosconfig`/`kubeconfig`/`cp-config`/`worker-config` stack outputs. They hold cluster credentials.
- Never run `just core` or `pulumi`; the user runs `just core talos preview`, then `up`. Rules for that flow: [homelab-pulumi-go](../homelab-pulumi-go/SKILL.md).
- Never lower `talos.version` in `config/versions.yaml`; it would regenerate `talos-secrets` (new PKI). Roll back nodes with `talosctl rollback` instead.
- Never fix a node by hand (`talosctl apply-config`, `edit`, `patch`). Change the patch in `core/platform/talos.go` and let Pulumi apply it.
- Reboot nodes one at a time and run `just talos-health` before the next. Never pass `--preserve` to `talosctl upgrade`.
- Check upstream Talos and siderolabs GitHub issues before changing any pin in `config/versions.yaml`.

## Where things live

| Thing | Location |
|---|---|
| Patches for every node, control planes, workers, GPU worker | `basePatch`, `cpPatch`, `workerPatch`, `gpuWorkerPatch` in `core/platform/talos.go` |
| Hostname, static IP, default route, nameservers | `patchTalosConfig` / `patchMachineNetwork` (rewrites the DHCP interface) |
| VM shape, PCIe passthrough, `RebootAfterUpdate: false` | `NewProxmoxVM` in `core/platform/proxmox.go` |
| Node list, IPs, roles, sizes, `pcie_ids` | `cluster.nodes` in `config/config.yml` (`Node.HasGPU()` = has `pcie_ids`) |
| VIP, subnet, router, nameservers | `cluster.*` in `config/config.yml` |
| Talos version, base and GPU schematics, Kubernetes version | `talos`, `kubernetes` in `config/versions.yaml` |
| Patch-per-role table and VM settings | [Nodes](../../../docs/src/content/docs/hardware/nodes.mdx), [Talos](../../../docs/src/content/docs/platform/talos.md) |

Node configs are combined as `basePatch` + role patch (+ `gpuWorkerPatch` when `HasGPU()`), then `patchTalosConfig` runs per node. Each node gets its own `<node>-config-apply` resource; they have no ordering between them, so one `up` applies every node together.

## Change impact

Read the preview's `<node>-config-apply` updates against this table before the user runs `up`. `applyMode` is not set on `ConfigurationApply`, so Talos decides whether a change reboots.

| Change | Effect on running nodes |
|---|---|
| Image or schematic (`DownloadImage`, VM `disks[0].fileId`) | Nothing; only new VMs. Upgrade with `just talos-upgrade` |
| `talos.version` bump | Updates `talos-secrets` and the image downloads; nodes stay on the old OS until `just talos-upgrade` |
| `kubernetes.version` bump | New kubelet/control-plane images on every node at once; run `just talos-upgrade-k8s` first |
| `machine.install` | Nothing until the next upgrade |
| VM cores, memory, balloon, `hostpci` | Nothing until the VM is cold-started (`RebootAfterUpdate: false`); passthrough changes need a cold boot |
| Kernel modules, anything Talos cannot apply live | Node reboot, on all matching nodes at once; stop and plan a per-node rollout with the user |
| Node labels, taints, kubelet config, apiServer `extraArgs`, network | Applied live or by restarting the affected service |
| `cluster.network.cni`, `cluster.proxy` | Do not change; Cilium in the `platform` stack depends on both |
| Any `replace` on `talos-secrets` or a VM | Stop. Both are `Protect(true)`; the preview means a PKI or disk loss |

Adding a node: append to `cluster.nodes` with an unused IP in `cluster.subnet`, then preview. Changing an IP: edit `config/config.yml` and preview.

## talosctl

| Task | Command |
|---|---|
| Endpoint | Always the VIP `192.168.1.210`; the recipes use `TALOSCTL` from the `justfile` |
| Cluster health (gate between nodes) | `just talos-health` |
| OS version on every node | `just talos-versions` |
| Extensions on worker4 | `talosctl --talosconfig ~/.talos/config -n 192.168.1.224 get extensions` |
| Stale config / certificate rejected | The user runs `cd core && pulumi stack output --stack talos talosconfig --show-secrets > ~/.talos/config`; redirect only, never echo it |

Use read-only `talosctl get`/`version`/`health`. Reboots go through `just talos-upgrade`, or the power-cycle in [Repairs](../../../docs/src/content/docs/operate/repairs.md) when a node will not reboot gracefully.

## Upgrades

Follow the runbook in [Talos and Kubernetes upgrades](../../../docs/src/content/docs/operate/upgrades/talos.mdx); do not restate it. The decisions that go wrong:

| Situation | Do |
|---|---|
| Order | Control planes `.211`, `.212`, `.213`, then workers, then `192.168.1.224 gpu` last; gate with the runbook's checks between each |
| Which schematic | `gpu` only for the node with `pcie_ids`; the recipe refuses a mismatch |
| Node hosts `n8n-pg-1` (`kubectl get pod -n n8n n8n-pg-1 -o wide`), or is worker4 | Pass drain `false`; the Longhorn and CNPG PDBs deadlock the drain |
| Drain already timed out | Node is cordoned with the new image unbooted: re-run with `false`, then `kubectl uncordon <node>` |
| Undrained reboot sits in "rebooting" | Wait about ten minutes for the Longhorn CSI unmount before anything else |
| Volumes faulted after reboot | `just longhorn-repair-iscsi` (the upgrade recipe already runs it) |
| Kubernetes version | Separate step: `just talos-upgrade-k8s --dry-run`, then without the flag, before `just core talos up` |

Longhorn manager must not tolerate the control-plane taint, or engine images stall and no volume attaches. That lives in `workloads/storage/longhorn.go`; use homelab-cdk8s to change it.

## worker4 (GPU node)

| Fact | Where |
|---|---|
| Taint `dedicated=ai:NoSchedule` | `registerWithTaints` in `gpuWorkerPatch` for first registration; `TaintGPUNodes` (`core/platform/nodes.go`, platform stack) for the live node. Never `machine.nodeTaints`: NodeRestriction rejects it |
| NVIDIA kernel modules | `gpuWorkerPatch` |
| PCIe `0000:09:00.0`, `Pcie: true`, `Xvga: false` | `pcie_ids` in `config/config.yml`, `NewProxmoxVM` |
| Pod requirements (toleration, `runtimeClassName: nvidia`) | [GPU](../../../docs/src/content/docs/hardware/gpu.md); workload side is homelab-cdk8s |

The GPU sits in an external USB4 enclosure behind a Thunderbolt add-in card. After a power loss the enclosure can stay off, and Proxmox fails to start the VM with `no PCI device found for '0000:09:00.0'`. Ask the user to check enclosure power and cable, then run `boltctl list` and `lspci -nn | grep 10de` on the Proxmox host (no SSH access from here). Thunderbolt authorization is not the cause.

## Verify

1. `just lint core` and `cd core && go build ./...` pass.
2. `just check` passes when `config/` or docs changed.
3. The user runs `just core talos preview`; it shows only the expected `<node>-config-apply` updates, no `replace` on `talos-secrets` or any VM, and no reboot-class change across all nodes at once.
4. After `up` or an upgrade: `just talos-health` is clean, `just talos-versions` matches `config/versions.yaml`, `kubectl get nodes` shows all Ready and none SchedulingDisabled.
