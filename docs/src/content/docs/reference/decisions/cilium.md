---
title: Cilium for networking
description: Why one Cilium install replaces kube-proxy, MetalLB and an Ingress controller.
---
**Decision:** Cilium as CNI with kube-proxy replacement, L2 announcements for LoadBalancer IPs, and its Gateway API implementation for all HTTP routing.

**Context:** Bare metal has no cloud load balancer, Talos ships no CNI here, and every app needs hostname routing behind one IP the wildcard DNS record can point at.

| Option | For | Against |
|---|---|---|
| Cilium (CNI + L2 + Gateway API) | One component; eBPF datapath; Hubble flow visibility; Gateway API role split | Talos-specific tuning (`hostLegacyRouting`, socket LB off); restarts can leave stale L2 leases |
| Flannel/Calico + MetalLB + ingress-nginx | Familiar, widely documented | Three components to version; Ingress has no Gateway/route split |
| Ingress API instead of Gateway API | Simpler, every chart supports it | HTTP only, limited matching, no infra/app separation |

**Consequences:**
- Gateway API CRDs are vendored and applied before Cilium; Cilium upgrades go one minor at a time.
- `wt0` can never be a Cilium device, so NetBird traffic reaches pods by kernel forwarding to another node.
- An L2 lease can go stale after an agent restart; `just gateway-repair-l2` re-elects it.
