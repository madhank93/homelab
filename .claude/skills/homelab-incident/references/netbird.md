# Public path incidents (NetBird, Bifrost)

The hop-by-hop path is in [Traffic](../../../../docs/src/content/docs/concepts/traffic.mdx). Runbooks: [NetBird peer](../../../../docs/src/content/docs/workloads/networking/netbird-peer.md#troubleshooting) and [NetBird](../../../../docs/src/content/docs/cloud/netbird.mdx#troubleshooting).

## Split the path

A 504 means Traefik on Bifrost is up and cannot reach `192.168.1.220`. Rule out the cluster first:

| Check | Result | Layer |
|---|---|---|
| LAN request to a `*.madhan.app` host | Times out | Cluster: Gateway L2 lease ([cilium.md](cilium.md)) |
| `kubectl exec -n netbird netbird-peer-0 -c netbird -- netbird status` | Not `Management: Connected` | Peer cannot reach the management server |
| Same, `Networks:` | Missing `192.168.1.0/24` | Route missing or on a stale peer in the NetBird UI |
| Peer connected, but `wt0` carries 0 bytes | Relay never connects | `NB_SKIP_SOCKET_MARK` was added to the peer env in `workloads/networking/netbird_peer.go`; remove it |
| Curl from Bifrost to `192.168.1.220` with a `Host` header | `302` | Path works; problem is in Traefik |

Commands for each row are on [NetBird peer](../../../../docs/src/content/docs/workloads/networking/netbird-peer.md#troubleshooting). SSH to Bifrost is read-only diagnosis; anything there changes through `just core hetzner up`, run by the user.

## Fixes by layer

- Stale route peer after a restart: in the NetBird UI, delete disconnected `k8s-routing-peer` duplicates and reassign the route (user step). Duplicates mean the PVC is not mounted at `/var/lib/netbird/`.
- Peer pod wedged: delete `netbird-peer-0`; the StatefulSet recreates it with the same key.
- `bifrost-agent` shows `Networks: -`: it must be in the `All` distribution group.

## Never

- `hostNetwork: true` on the peer. On the L2 lease holder's node a host-network peer cannot ARP for the IP its own node announces, and every public request times out.
- `wt0` in Cilium `devices`.
- `NB_SKIP_SOCKET_MARK=true`.
