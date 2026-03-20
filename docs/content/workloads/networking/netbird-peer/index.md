+++
title = "NetBird Peer"
description = "In-cluster WireGuard routing peer that advertises 192.168.1.0/24 into the NetBird mesh and forwards traffic to the Cilium Gateway."
weight = 10
+++

## What is the NetBird Peer?

The NetBird peer is an in-cluster WireGuard client that connects to the NetBird VPN mesh and acts as a **routing peer** for the cluster's LAN subnet (`192.168.1.0/24`). It runs on **worker1** (`192.168.1.221`) with `hostNetwork: true` and enables the Hetzner VPS (Bifrost) to reach cluster services via WireGuard, making public service routing through Traefik possible.

```
Bifrost Traefik
    │  proxy to 192.168.1.220:80
    ↓
netbird-agent (Bifrost · wt0: 100.109.47.211)
    │  WireGuard tunnel via relay
    ↓
netbird-peer-0 (worker1 · wt0: 100.109.244.71)
    │  kernel IP forward: wt0 → eth0
    │  CILIUM_POST_nat MASQUERADE: src → 192.168.1.221
    ↓
192.168.1.220 (Cilium Gateway · another node's eth0)
    │  Cilium BPF L7LB DNAT → Envoy :13507
    ↓
Grafana / other cluster pod
```

Source: [`workloads/networking/netbird_peer.go`](https://github.com/madhank93/homelab/blob/v0.1.5/workloads/networking/netbird_peer.go)

---

## Why a Kubernetes StatefulSet?

The in-cluster NetBird peer is deployed as a **StatefulSet** (not a Deployment) with a persistent PVC for `/var/lib/netbird/`:

- `/var/lib/netbird/` stores the WireGuard private key and peer registration state
- Without persistence, every pod restart generates a new private key → new peer registration in NetBird Management → accumulating duplicate peers in the UI
- StatefulSet + PVC ensures the same peer identity is reused across restarts

A simple Deployment would create a new peer registration on every restart (e.g., every ArgoCD sync that changes the pod spec), filling the NetBird Management UI with ghost peers.

---

## Configuration

| Setting | Value | Why |
|---------|-------|-----|
| Namespace | `netbird` | Privileged PSA (needs NET_ADMIN, SYS_MODULE) |
| Image | `netbirdio/netbird:0.66.2` | Pinned to match Bifrost server version |
| Kind | `StatefulSet` | Persistent identity across restarts |
| `hostNetwork: true` | true | WireGuard must manipulate host routing table |
| `dnsPolicy` | `ClusterFirstWithHostNet` | DNS works with hostNetwork |
| PVC | `100Mi` RWO | `/var/lib/netbird/` — private key + config |
| Capabilities | `NET_ADMIN`, `SYS_MODULE` | WireGuard kernel module management |
| `NB_MANAGEMENT_URL` | `https://netbird.madhan.app` | NetBird Management server on Bifrost |
| `NB_HOSTNAME` | `k8s-routing-peer` | Peer name in NetBird UI |
| `NB_SETUP_KEY` | From OpenBao (Pattern B) | Used only on first registration |

> **Critical:** PVC mount path is `/var/lib/netbird/` — **not** `/etc/netbird/`. NetBird v0.66 stores its private key at `/var/lib/netbird/` regardless of what `NB_CONFIG` points to. Mounting at the wrong path means the key is never persisted.

---

## Secrets (OpenBao)

Pattern B (secretObjects sync). `NETBIRD_SETUP_KEY` is fetched from OpenBao (`secret/data/netbird`) and synced into the `netbird-setup-key` k8s Secret.

The setup key is only used on **first registration**. Once the peer is registered, subsequent restarts reuse the private key from the PVC and ignore the setup key.

---

## MASQUERADE initContainer

The StatefulSet includes an `initContainer` that adds an iptables rule before the NetBird agent starts:

```bash
iptables -t nat -C POSTROUTING -s 100.109.0.0/16 -d 192.168.1.0/24 -j MASQUERADE 2>/dev/null \
  || iptables -t nat -A POSTROUTING -s 100.109.0.0/16 -d 192.168.1.0/24 -j MASQUERADE
```

The `-C` check prevents duplicate rules on pod restart. This rule ensures that traffic from Bifrost's WireGuard IP range (`100.109.x.x`) destined for the cluster LAN gets source-NAT'd to `192.168.1.221` (worker1's eth0 IP), allowing cluster nodes to send replies back via normal LAN routing.

In practice the actual NAT is performed by Cilium's `CILIUM_POST_nat` BPF chain, not the raw iptables rule. Both coexist without conflict.

---

## Cilium Constraint — wt0 Must NOT Be in Devices

Do **not** add `wt0` to Cilium's `devices` list in `core/platform/cilium.go`. WireGuard interfaces are `NOARP/POINTOPOINT` with no Ethernet header — Cilium's `cil_from_netdev` TC BPF silently drops all packets without monitor events, breaking all traffic from Bifrost.

`devices` must list only `eth0`. Traffic from `wt0` reaches Cilium BPF via another node's `eth0` after kernel IP forwarding and MASQUERADE on worker1. See [Network Flow — Why wt0 is NOT in Cilium Devices](/architecture/network-flow/#why-wt0-is-not-in-cilium-devices).

---

## Route Configuration

The peer registers in NetBird but does **not** automatically advertise routes. Routes must be configured in the NetBird Management UI:

**NetBird Management UI → Network → Routes → Add Route:**

| Field | Value |
|-------|-------|
| Network | `192.168.1.0/24` |
| Routing peer | `k8s-routing-peer` |
| Distribution Groups | `All` |

Without this route configuration, the Bifrost VPS cannot reach the cluster's LAN services even though the WireGuard tunnel is up.

---

## Troubleshooting

### Peer not connecting

```bash
# Check pod is running
kubectl get pods -n netbird -o wide

# Check NetBird agent status
kubectl exec -n netbird netbird-peer-0 -- netbird status
# Expect: Management: Connected, Peers count: 1/1 Connected

# Check NetBird agent logs for errors
kubectl logs -n netbird netbird-peer-0 | tail -30

# Verify management URL is reachable from the pod
kubectl exec -n netbird netbird-peer-0 -- \
  wget -qO- --timeout=5 https://netbird.madhan.app/api/v1/peers 2>&1 | head -5
```

### Route shows "Networks: -" on bifrost-agent

The peer is connected but not advertising the route. Either:
1. The Network Route hasn't been created in NetBird UI → Network Routes
2. The route references a stale/disconnected peer — delete old entries and re-assign to the current peer

```bash
# Verify from the peer side
kubectl exec -n netbird netbird-peer-0 -- netbird status
# Look for: Networks: 192.168.1.0/24

# Verify from Bifrost
ssh root@178.156.199.250 'docker exec netbird-agent netbird routes list'
# Expect: 192.168.1.0/24  Status: Selected
```

### 504 from public services but tunnel is up

The route is connected but traffic doesn't reach the cluster backend. Check the Cilium device configuration first:

```bash
# Confirm wt0 is NOT in Cilium's device list on worker1
CILIUM_POD=$(kubectl get pods -n kube-system -l app.kubernetes.io/name=cilium-agent \
  -o wide | grep k8s-worker1 | awk '{print $1}')
kubectl exec -n kube-system $CILIUM_POD -c cilium-agent -- cilium-dbg status | grep KubeProxy
# Must show only: [eth0  192.168.1.221 ...]
# If wt0 appears: update core/platform/cilium.go and run just core platform up

# Test connectivity to the Cilium LB from Bifrost
ssh root@178.156.199.250 \
  'curl -sv -H "Host: grafana.madhan.app" --connect-timeout 5 http://192.168.1.220/'
# Expect: HTTP/1.1 302 Found
```

### MASQUERADE rule not taking effect

The iptables MASQUERADE rule requires the pod to have `NET_ADMIN` capability. Check the initContainer ran:

```bash
kubectl describe pod -n netbird netbird-peer-0 | grep -A5 "Init Containers"
# setup-iptables should show: State: Terminated, Reason: Completed

# Check the rule is installed
kubectl exec -n netbird netbird-peer-0 -- \
  iptables -t nat -L POSTROUTING -n -v | grep -E "MASQUERADE|CILIUM"
# CILIUM_POST_nat chain should show high packet count
# MASQUERADE rule (backup) shows 0 packets — this is normal with Cilium BPF active
```
