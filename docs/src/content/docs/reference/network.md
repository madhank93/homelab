---
title: Network
description: Every IP, CIDR, port and hostname in the homelab in one place, with the file that sets it.
sidebar:
  order: 5
---

Addresses come from [`config/config.yml`](src:config/config.yml) unless noted. Change them there, not in Go code. How traffic flows between them: [Traffic](/concepts/traffic/).

## LAN

| What | Address | Set in |
|---|---|---|
| Subnet | `192.168.1.0/24` | `cluster.subnet` |
| Router / default gateway | `192.168.1.254` | `cluster.router` |
| Nameservers | `1.1.1.1`, `192.168.1.254` | `cluster.nameservers` |
| Proxmox API | `https://192.168.1.235:8006/` | `proxmox.endpoint` |
| Kubernetes API VIP | `192.168.1.210:6443` | `cluster.vip`; Talos VIP on the control planes |

## Nodes

| Node | IP | Role |
|---|---|---|
| k8s-controller1 | `192.168.1.211` | control plane |
| k8s-controller2 | `192.168.1.212` | control plane |
| k8s-controller3 | `192.168.1.213` | control plane |
| k8s-worker1 | `192.168.1.221` | worker |
| k8s-worker2 | `192.168.1.222` | worker |
| k8s-worker3 | `192.168.1.223` | worker |
| k8s-worker4 | `192.168.1.224` | GPU worker (`dedicated=ai:NoSchedule`) |

Node ports that matter: Talos API `50000`, kubelet `10250`, KubePrism `127.0.0.1:7445`.

## LoadBalancer IPs

The Cilium LB-IPAM pool is `cluster.lb_pool` (`192.168.1.220`–`192.168.1.230`), announced over L2 by worker nodes. [`cilium.go`](src:core/platform/cilium.go) builds it as one `/32` per address and **skips node IPs**, so the usable pool is `.220` and `.225`–`.230`. Announcing a node's own IP takes that node off the LAN.

| IP | Service | Ports | Set in |
|---|---|---|---|
| `192.168.1.220` | Cilium Gateway `homelab-gateway` (`kube-system`) | 80, 443 | `cluster.gateway_ip` |
| `192.168.1.225` | Ollama | 11434 | [`ollama.go`](src:workloads/ai/ollama.go) |
| `192.168.1.230` | Counter-Strike 2 | 27015 UDP + TCP | [`cs2.go`](src:workloads/gaming/cs2.go) |

## Cluster CIDRs

Not overridden in the Talos config, so the Talos defaults apply: Services `10.96.0.0/12`, Pods `10.244.0.0/16`. Cilium runs with `ipam.mode: kubernetes`, taking each node's pod CIDR from Kubernetes.

## NetBird overlay

| What | Address |
|---|---|
| WireGuard interface | `wt0` on the Bifrost `netbird-agent` and on the `netbird-peer` pod (pod netns) |
| Overlay addresses | Assigned by NetBird management; peers here sit in `100.109.0.0/16`. Not set in this repo |
| Routed into the mesh | `192.168.1.0/24`, advertised by `k8s-routing-peer` |
| Policy routing on Bifrost | ip rule table `7120` |
| STUN | `3478/udp` on the VPS |

## Bifrost VPS

| What | Address | Set in |
|---|---|---|
| Public IP | `178.156.199.250` | `hetzner.vps_ip` |
| Firewall inbound | `22/tcp`, `80/tcp`, `443/tcp`, `3478/udp` | [`hetzner.go`](src:core/cloud/hetzner.go) |
| `bifrost_net` (Docker) | `172.30.0.0/24`, gateway `172.30.0.1` | [`docker-compose.yml`](src:core/cloud/bifrost/docker-compose.yml) |
| Traefik | `172.30.0.10` | same |
| Gatus | `172.30.0.12` | same |
| Authentik health (local only) | `127.0.0.1:9000` | same |

Containers resolve public hostnames to Traefik's `172.30.0.10` via `extra_hosts`, avoiding hairpin NAT.

## DNS

Cloudflare, all records `A`, DNS-only (not proxied), from [`cloudflare.go`](src:core/cloud/cloudflare.go).

| Record | Points at | Effect |
|---|---|---|
| `*.madhan.app` | `192.168.1.220` | Every unlisted subdomain resolves to the LAN Gateway: reachable on the LAN or over the VPN, nowhere else |
| `pve` | `192.168.1.235` | Proxmox host (`proxmox.host`, `proxmox.ip`); LAN only |
| `auth`, `netbird`, `proxy`, `uptime` | `178.156.199.250` | Always public |
| `*.proxy.madhan.app` | `178.156.199.250` | NetBird expose |
| Each entry in `publicServices` | `178.156.199.250` | Public via Traefik; currently `grafana` |

To make a service public, add it to `publicServices` and run `just core hetzner up` and `just core cloudflare up` — see [Exposure](/cloud/exposure/).

## Hostnames on the Gateway

HTTPRoutes on `homelab-gateway`, all `<name>.madhan.app`: `alertmanager`, `argocd`, `chat`, `comfyui`, `falco`, `grafana`, `harbor`, `headlamp`, `home`, `hubble`, `kubeflow`, `longhorn`, `n8n`, `ollama`, `openbao`, `victorialogs`, `vmselect`.
