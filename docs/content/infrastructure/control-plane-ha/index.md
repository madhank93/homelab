+++
title = "Control Plane HA"
description = "3-node control plane with Talos VIP — how HA works, what fails over, and what does not."
weight = 15
+++

## What Is Control Plane HA?

A highly available control plane means the Kubernetes API server remains reachable even when one control plane node goes down. This homelab runs 3 control plane nodes, which is the minimum count needed for etcd quorum-based HA.

## Node Configuration

| Node | IP | Role |
|------|----|------|
| k8s-controller1 | 192.168.1.211 | control plane |
| k8s-controller2 | 192.168.1.212 | control plane |
| k8s-controller3 | 192.168.1.213 | control plane |
| VIP | 192.168.1.210 | virtual (floats) |

Each control plane node runs the full Kubernetes control plane stack:

- `kube-apiserver` — serves the Kubernetes API
- `etcd` — distributed key-value store for cluster state
- `kube-controller-manager` — reconciliation loops
- `kube-scheduler` — places new pods onto nodes

## etcd Raft Consensus

etcd uses the [Raft consensus algorithm](https://raft.github.io/) to maintain consistency across nodes. To elect a leader and accept writes, etcd needs a **quorum** — more than half of the members.

| Total nodes | Quorum required | Can tolerate |
|-------------|-----------------|--------------|
| 1 | 1 | 0 failures |
| 3 | 2 | 1 failure |
| 5 | 3 | 2 failures |

With 3 control plane nodes: lose 1 → still have 2/3 → quorum maintained → cluster stays healthy. Lose 2 → only 1/3 → no quorum → writes block → cluster effectively frozen.

## VIP Load Balancing

Talos Linux has a built-in VIP feature that floats a virtual IP (`192.168.1.210`) between control plane nodes.

Configuration in the machine patch (see [`core/platform/talos.go`](https://github.com/madhank93/homelab/blob/v0.1.5/core/platform/talos.go)):

```yaml
machine:
  network:
    interfaces:
      - deviceSelector:
          physical: true
        dhcp: true
        vip:
          ip: 192.168.1.210
```

**How it works:**

1. At startup, all three control plane nodes compete via ARP leader election.
2. One node wins and announces `192.168.1.210` as its own via a gratuitous ARP.
3. All workers and external clients always connect to `https://192.168.1.210:6443`.
4. If the VIP holder crashes, another node detects the silence and takes over within 1–3 seconds.

This is **HA failover, not load balancing**. Only one node holds the VIP at any time. API server requests are not distributed across all three nodes — they all go to whichever node holds the VIP.

## Why This Approach?

| Option | Pros | Cons |
|--------|------|------|
| Talos built-in VIP | No extra components, works automatically | Not true load balancing |
| External HAProxy/keepalived | True round-robin | Extra VM or hardware, more complexity |
| Cloud load balancer | Managed, true LB | Not available on-premises |

For a homelab with a single Proxmox host and no redundant hardware, the Talos VIP is the right trade-off. The API server sees very little load; distribution matters far less than failover.

## During Failover

When the VIP holder crashes:

1. The remaining two control plane nodes detect no ARP reply (within ~1–3 seconds).
2. One of them wins the election and begins announcing the VIP.
3. New API server connections succeed immediately.
4. **In-flight requests** to the dead node's API server fail with connection reset/timeout.
5. kubectl commands see a 1–3 second error window, then recover.
6. **Running pods are unaffected** — the kubelet on worker nodes does not use the API server for normal pod operation. Pods keep running.
7. etcd quorum is maintained with 2 of 3 members — the cluster continues accepting writes.

## Bootstrap Sequence

Talos bootstraps the cluster against the first control plane node, then bootstraps etcd:

```go
// core/platform/talos.go — cluster bootstrap against 192.168.1.211 first
talos_cluster.NewBootstrap(ctx, "talos-bootstrap", &talos_cluster.BootstrapArgs{
    Node:            pulumi.String("192.168.1.211"),
    ClientConfiguration: ...
})
```

After `192.168.1.211` is bootstrapped, the other two nodes join and etcd syncs from the first member. The VIP becomes active once the first node is up and the cluster endpoint `https://192.168.1.210:6443` resolves correctly.

## Verifying HA

```bash
# Check all control plane nodes are members of etcd
talosctl --talosconfig ~/.talos/config etcd members --nodes 192.168.1.211

# Check API server is reachable via VIP
kubectl cluster-info

# Check which node holds the VIP (look for the extra IP on the interface)
talosctl --talosconfig ~/.talos/config get addresses --nodes 192.168.1.211,192.168.1.212,192.168.1.213

# Test failover: reboot one control plane node
talosctl --talosconfig ~/.talos/config reboot --nodes 192.168.1.211
# Wait 5s, then:
kubectl get nodes  # should still work
```
