---
title: Control plane HA
description: Three control planes share the Talos VIP 192.168.1.210 — failover, not load balancing, and what one lost node costs.
sidebar:
  order: 20
---

Three control planes run etcd and the API server; a Talos-managed VIP floats between them so every client uses one address. [Talos VIP docs](https://www.talos.dev/latest/talos-guides/network/vip/).

| | |
|---|---|
| Nodes | `k8s-controller1`–`3` · `192.168.1.211`–`213` |
| VIP | `192.168.1.210:6443` |
| Source | [`talos.go`](src:core/platform/talos.go) (`cpPatch`) |

## How it fits

One controller holds the VIP and answers ARP for it; if it dies, another takes over within seconds. Only the holder serves API traffic: this is failover, not load balancing. The built-in VIP was chosen because it needs no extra VM or keepalived, and the API server's load is small.

etcd needs a majority: with three members the cluster survives one loss. Losing two blocks all writes.

## During a failover

- New API connections succeed once another node announces the VIP; in-flight requests to the dead node fail.
- `kubectl` sees a few seconds of errors, then recovers.
- Running pods are unaffected; kubelets keep them running without the API server.
- Cilium reaches the API server through the VIP (`k8sServiceHost`), so it follows the failover.

Pulumi bootstraps etcd on the first control plane in `config.yml` (`k8s-controller1`); the others join it.

## Operate

```sh frame="terminal"
talosctl --talosconfig ~/.talos/config etcd members --nodes 192.168.1.211
talosctl --talosconfig ~/.talos/config get addresses --nodes 192.168.1.211,192.168.1.212,192.168.1.213   # VIP holder shows .210
kubectl cluster-info
```

Reboot control planes one at a time and wait for etcd to report three healthy members before the next.

## Troubleshooting

### API unreachable on the VIP but the nodes are up

Check which node holds `192.168.1.210` with the `get addresses` command above. If none does, check etcd membership on each controller: the VIP is only announced by a node whose etcd is healthy.
