+++
title = "NetBird Peer"
description = "In-cluster WireGuard peer that routes LAN traffic (192.168.1.0/24) through the NetBird VPN mesh."
weight = 10
+++

## What is the NetBird Peer?

The NetBird peer is an in-cluster WireGuard client that connects to the NetBird VPN mesh and acts as a **routing peer** for the cluster's LAN subnet (`192.168.1.0/24`). It enables the Hetzner VPS (Bifrost) to reach cluster services via WireGuard, making public service routing possible.

## Why a Kubernetes StatefulSet?

The in-cluster NetBird peer is deployed as a **StatefulSet** (not a Deployment) with a persistent PVC for `/etc/netbird/`:

- `/etc/netbird/` stores the WireGuard private key and peer registration state
- Without persistence, every pod restart generates a new private key → new peer registration in NetBird Management → accumulating duplicate peers in the UI
- StatefulSet + PVC ensures the same peer identity is reused across restarts

A simple Deployment would create a new peer registration on every restart (e.g., every ArgoCD sync that changes the pod spec), filling the NetBird Management UI with ghost peers.

Source: [`workloads/networking/netbird_peer.go`](https://github.com/madhank93/homelab/blob/v0.1.5/workloads/networking/netbird_peer.go)

## Configuration

| Setting | Value | Why |
|---------|-------|-----|
| Namespace | `netbird` | Privileged PSA (needs NET_ADMIN, SYS_MODULE) |
| Image | `netbirdio/netbird:0.66.2` | Pinned to match Bifrost server version |
| Kind | `StatefulSet` | Persistent identity across restarts |
| `hostNetwork: true` | true | WireGuard must manipulate host routing table |
| `dnsPolicy` | `ClusterFirstWithHostNet` | DNS works with hostNetwork |
| PVC | `100Mi` RWO | `/etc/netbird/` — private key + config |
| Capabilities | `NET_ADMIN`, `SYS_MODULE` | WireGuard kernel module management |
| `NB_MANAGEMENT_URL` | `https://netbird.madhan.app` | NetBird Management server on Bifrost |
| `NB_HOSTNAME` | `k8s-routing-peer` | Peer name in NetBird UI |
| `NB_SETUP_KEY` | From OpenBao (Pattern B) | Used only on first registration |

## Secrets (OpenBao)

Pattern B (secretObjects sync). `NETBIRD_SETUP_KEY` is fetched from OpenBao (`secret/data/netbird`) and synced into the `netbird-setup-key` k8s Secret.

The setup key is only used on **first registration**. Once the peer is registered, subsequent restarts reuse the private key from the PVC and ignore the setup key.

## Route Configuration

The peer registers in NetBird but does **not** automatically advertise routes. Routes must be configured in the NetBird Management UI:

**NetBird Management UI → Network → Routes → Add Route:**

| Field | Value |
|-------|-------|
| Network | `192.168.1.0/24` |
| Routing peer | `k8s-routing-peer` |

Without this route configuration, the Bifrost VPS cannot reach the cluster's LAN services even though the WireGuard tunnel is up.

## How It Connects

```
NetBird Management (netbird.madhan.app on Bifrost)
  ↕ WireGuard tunnel
NetBird peer pod in cluster (k8s-routing-peer)
  → hostNetwork: true → accesses 192.168.1.0/24 subnet
  ← Bifrost VPS sends traffic for 192.168.1.0/24
  → Routes to cluster nodes and services

Bifrost Traefik → NetBird WireGuard → k8s-routing-peer
  → 192.168.1.220 (homelab-gateway) → pod
```

## Troubleshooting

### Peer Not Connecting

```bash
# Check pod is running
kubectl get pods -n netbird

# Check NetBird agent logs
kubectl logs -n netbird netbird-peer-0

# Verify management URL is reachable
kubectl exec -n netbird netbird-peer-0 -- \
  curl -s https://netbird.madhan.app/api/v1/peers
```

### Duplicate Peers in NetBird UI

**Why:** This happens when the old PVC was deleted (or if the pod was previously a Deployment). Each pod restart without the persisted private key creates a new peer identity.

**Fix:** In the NetBird Management UI, manually delete the stale/ghost peers. Then verify the StatefulSet's PVC persists across restarts:

```bash
kubectl get pvc -n netbird
# Should show: netbird-config-netbird-peer-0 (Bound)
```

### Route Not Working

If Bifrost can reach the NetBird peer but cannot route to 192.168.1.x addresses:

1. Check the route is configured in NetBird Management UI (Network → Routes)
2. Check the peer shows as "Connected" in the NetBird UI
3. Verify `hostNetwork: true` is set on the pod
