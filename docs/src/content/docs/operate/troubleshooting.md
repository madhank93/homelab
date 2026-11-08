---
title: Troubleshooting
description: Start from the symptom. Every entry names the likely cause and the page with the full runbook.
sidebar:
  order: 2
---

Find the symptom, then follow the link. Component pages carry the detailed fixes.

## Something is unreachable

| Symptom | Likely cause | Where to look |
|---|---|---|
| `504 Gateway Timeout` on a public URL | Traefik cannot reach the cluster: NetBird tunnel or route | [NetBird peer](/workloads/networking/netbird-peer/) |
| Every LoadBalancer IP goes dark after a Cilium restart | Stale L2 announcement lease; the holder renews but stops answering ARP | `just gateway-repair-l2` — [Repairs](/operate/repairs/) |
| A `*.madhan.app` host resolves but nothing answers | Service is LAN-only and you are off the LAN and off the VPN | [Exposure](/cloud/exposure/) |
| TLS certificate errors on a service | Wildcard cert, or Cilium's copy of it in `cilium-secrets` | [cert-manager](/platform/cert-manager/) |
| Argo CD UI loads but lists no applications | Stale aggregated APIService breaks API discovery; `kubectl api-resources` shows errors | [Argo CD](/platform/argocd/) |

## A pod will not start

| Symptom | Likely cause | Where to look |
|---|---|---|
| GPU pod stays `Pending` with no clear reason | Missing the `dedicated=ai:NoSchedule` toleration; worker4 is tainted | [GPU](/hardware/gpu/) |
| GPU pod runs but CUDA is invisible | Missing `runtimeClassName: nvidia` | [GPU](/hardware/gpu/) |
| `Multi-Attach error for volume` | RWO volume still held by the old pod during a rollout | [Harbor](/workloads/registry/harbor/) |
| PVC stuck `Pending`, pod stuck `ContainerCreating` | Longhorn volume cannot attach | [Longhorn](/workloads/storage/longhorn/) |
| No Longhorn volume on a node attaches after a Talos upgrade | New open-iscsi rejects a record its predecessor wrote | `just longhorn-repair-iscsi` — [Repairs](/operate/repairs/) |
| Longhorn or CSI still down after DiskPressure cleared | DaemonSet pods evicted under pressure stay `Failed` and are never replaced | Delete the `Failed` pods — [Repairs](/operate/repairs/) |
| Pod cannot find its Kubernetes Secret | Nothing mounts the CSI volume, so the sync never fires | [Secrets](/concepts/secrets/) |
| ComfyUI returns 503 | It ships scaled to zero | `just comfyui on` — [ComfyUI](/workloads/ai/comfyui/) |

## Something is missing or wrong

| Symptom | Likely cause | Where to look |
|---|---|---|
| `kubectl top` says `Metrics API not available` | Metrics Server | [Metrics Server](/workloads/monitoring/metrics-server/) |
| Metrics missing from Grafana | VMAgent is not scraping the target | [VictoriaMetrics](/workloads/observability/victoria-metrics/) |
| Ollama cannot load a model | ComfyUI holds VRAM; the error surfaces on Ollama | [GPU](/hardware/gpu/), `just comfyui off` |
| Grafana OIDC login loops or errors | `root_url` / redirect URI mismatch | [Grafana](/workloads/monitoring/grafana/) |
| Falco UI never becomes ready | Redis wedged with `MISCONF`: TTL and volume size | [Falco](/workloads/security/falco/) |
| App stuck `OutOfSync` in Argo CD | Sync wave, hook failure, or SSA ownership conflict | [Argo CD](/platform/argocd/) |
| A manual `kubectl` change reverted itself | `selfHeal: true` — intended; change it in git | [Argo CD](/platform/argocd/) |

## A node is in trouble

| Symptom | Likely cause | Where to look |
|---|---|---|
| Node will not reboot after a Cilium problem | Unmount blocks on an NFS share-manager ClusterIP it can no longer reach | Power-cycle — [Repairs](/operate/repairs/) |
| Talos upgrade hangs draining a node | Longhorn and single-instance CNPG PDBs allow zero disruptions | `just talos-upgrade <ip> <schematic> false` — [Talos upgrade](/operate/upgrades/talos/) |
| Cluster health check fails between node upgrades | Stop and fix before continuing | `just talos-health` |

## First commands

```bash frame="terminal"
just talos-health                       # cluster-level health
kubectl get applications -n argocd      # what Argo CD thinks is wrong
kubectl api-resources 2>&1 >/dev/null   # any output means broken discovery
kubectl get events -A --sort-by=.lastTimestamp | tail -30
```
