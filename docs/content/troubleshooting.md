+++
title = "Troubleshooting"
description = "Start from the symptom. Every entry links to the page with the full runbook."
weight = 65
+++

Twenty-odd pages carry a Troubleshooting section, which is only useful if you
already know which component is at fault. Start here instead — find the symptom,
follow the link.

## Something is unreachable

| Symptom | Likely cause | Where to look |
|---|---|---|
| `504 Gateway Timeout` on a public URL | Traefik cannot reach the cluster — NetBird tunnel or route | [NetBird Peer](@/workloads/networking/netbird-peer/index.md#troubleshooting) |
| Every LoadBalancer IP goes dark after a Cilium restart | Stale L2 announcement lease; the holder renews but stops answering ARP | [Upgrade Guide](@/upgrade-guide/index.md), then `just gateway-repair-l2` |
| A `*.madhan.app` host resolves but nothing answers | Service is LAN-only and you are off the LAN and off the VPN | [Service Access](@/infrastructure/access/index.md) |
| TLS certificate errors on a service | Wildcard cert or the Cilium secret mirror | [cert-manager](@/platform/cert-manager/index.md) |
| Argo CD UI loads but lists no applications | Usually a stale aggregated APIService breaking API discovery — check `kubectl api-resources` for errors | [Argo CD](@/platform/argocd/index.md) |

## A pod will not start

| Symptom | Likely cause | Where to look |
|---|---|---|
| GPU pod stays `Pending` with no clear reason | Missing the `dedicated=ai` toleration — worker4 is tainted | [GPU](@/hardware/gpu/index.md#the-dedicated-ai-taint) |
| GPU pod runs but CUDA is invisible | Missing `runtimeClassName: nvidia` | [GPU](@/hardware/gpu/index.md#gpu-workload-configuration) |
| `Multi-Attach error for volume` | RWO volume still held by the old pod during a rollout | [Harbor](@/workloads/registry/harbor/index.md#rwo-multi-attach-deadlock) |
| PVC stuck `Pending`, pod stuck `ContainerCreating` | Longhorn volume cannot attach | [Longhorn](@/workloads/storage/longhorn/index.md#volume-stuck-attaching) |
| No Longhorn volume on a node will attach after a Talos upgrade | New open-iscsi rejects a record its predecessor wrote | `just longhorn-repair-iscsi` — [Upgrade Guide](@/upgrade-guide/index.md) |
| Pod cannot find its k8s Secret | Nothing mounts the CSI volume, so the sync never fires | [Secrets](@/platform/secrets/index.md) |
| ComfyUI returns 503 | It ships scaled to zero | `just comfyui on` — [ComfyUI](@/workloads/ai/comfyui/index.md) |

## Something is missing or wrong

| Symptom | Likely cause | Where to look |
|---|---|---|
| `kubectl top` says `Metrics API not available` | Metrics Server | [Metrics Server](@/workloads/monitoring/metrics-server/index.md) |
| Metrics missing from Grafana | VMAgent is not scraping the target | [VictoriaMetrics](@/workloads/monitoring/victoria-metrics/index.md#troubleshooting) |
| Ollama cannot load a model | ComfyUI is holding VRAM, and the error surfaces on Ollama | [GPU](@/hardware/gpu/index.md#sharing-the-16-gb-of-vram) |
| Grafana OIDC login loops or errors | `root_url` / redirect URI mismatch | [Grafana](@/workloads/monitoring/grafana/index.md) |
| Falco UI never becomes ready | Redis wedged with `MISCONF` — TTL and volume size | [Falco](@/workloads/security/falco/index.md#falcosidekick-ui-retention) |
| App stuck `OutOfSync` in Argo CD | Sync wave, hook failure, or SSA ownership conflict | [Argo CD](@/platform/argocd/index.md#app-stuck-syncing) |
| A manual `kubectl` change reverted itself | `selfHeal=true` — this is intended | [Argo CD](@/platform/argocd/index.md) |

## A node is in trouble

| Symptom | Likely cause | Where to look |
|---|---|---|
| Node will not reboot after a Cilium problem | Unmount blocks on an NFS share-manager ClusterIP it can no longer reach | [Upgrade Guide](@/upgrade-guide/index.md) — powercycle |
| Talos upgrade hangs draining a node | Longhorn and single-instance CNPG PDBs allow zero disruptions | `just talos-upgrade <ip> <schematic> false` |
| Cluster health check fails between node upgrades | Stop and fix before continuing | `just talos-health` |

## First commands

```bash
just talos-health                       # cluster-level health
kubectl get applications -n argocd      # what Argo CD thinks is wrong
kubectl api-resources 2>&1 >/dev/null   # any output here means broken discovery
kubectl get events -A --sort-by=.lastTimestamp | tail -30
```
