+++
title = "Cilium"
description = "Cilium CNI as kube-proxy replacement, L2 LoadBalancer IPs, Gateway API, and Hubble."
weight = 50
+++

## Cilium CNI

[Cilium](https://cilium.io/) v1.16.6 is installed via Helm in the `kube-system` namespace by Pulumi (`core/platform/cilium.go`).

| Setting | Value | Purpose |
|---------|-------|---------|
| `ipam.mode` | `kubernetes` | IPAM managed by Kubernetes |
| `kubeProxyReplacement` | `true` | Full eBPF kube-proxy replacement |
| `l2Announcements.enabled` | `true` | L2 ARP announcements for LoadBalancer IPs |
| `gatewayAPI.enabled` | `true` | Kubernetes Gateway API support |
| `hubble.enabled` | `true` | Network observability |
| `hubble.relay.enabled` | `true` | Hubble relay for cross-node visibility |
| `hubble.ui.enabled` | `true` | Hubble UI dashboard |
| `k8sServiceHost` | `192.168.1.210` | VIP for API server connectivity |
| `k8sServicePort` | `6443` | API server port |

## Cilium as kube-proxy Replacement

Standard Kubernetes uses `kube-proxy` as a DaemonSet to manage `iptables` rules for Service-to-Pod routing. This homelab replaces kube-proxy entirely with Cilium's eBPF data plane.

**How it's disabled in Talos** (from [`core/platform/talos.go`](https://github.com/madhank93/homelab/blob/v0.1.5/core/platform/talos.go)):

```yaml
cluster:
  proxy:
    disabled: true
```

**How Cilium takes over** (from [`core/platform/cilium.go`](https://github.com/madhank93/homelab/blob/v0.1.5/core/platform/cilium.go)):

```go
"kubeProxyReplacement": pulumi.Bool(true),
"k8sServiceHost":       pulumi.String("192.168.1.210"),  // VIP
"k8sServicePort":       pulumi.Int(6443),
```

Cilium needs to know the API server address directly (not through kube-proxy) because it is replacing kube-proxy. This is why `k8sServiceHost` must be set to the VIP.

**What this means in practice:**

- No `kube-proxy` DaemonSet exists on any node
- All Service→Pod routing is handled in the Linux kernel via eBPF programs loaded by Cilium
- This eliminates iptables overhead and provides faster, more observable routing

**Verify:**

```bash
# No kube-proxy DaemonSet
kubectl get ds -n kube-system | grep kube-proxy  # should return nothing

# Cilium reports kube-proxy replacement active
cilium status | grep KubeProxyReplacement
# KubeProxyReplacement: True
```

## L2 Announcements and LoadBalancer IPs

Bare-metal clusters have no cloud provider to assign LoadBalancer IPs. Cilium's L2 announcement feature fills this role using ARP.

**IP Pool** (from [`core/platform/cilium.go`](https://github.com/madhank93/homelab/blob/v0.1.5/core/platform/cilium.go)):

```go
// IPs: 192.168.1.220 through 192.168.1.230 (11 addresses, each as /32)
for i := 220; i <= 230; i++ {
    cidr: fmt.Sprintf("192.168.1.%d/32", i)
}
```

A `CiliumLoadBalancerIPPool` named `address-pool` reserves these 11 IPs. Services with `type: LoadBalancer` are assigned IPs from this pool.

**L2 Announcement Policy:**

A `CiliumL2AnnouncementPolicy` named `l2-policy` restricts announcements to **worker nodes only** (no `node-role.kubernetes.io/control-plane` label):

```yaml
spec:
  nodeSelector:
    matchExpressions:
      - key: node-role.kubernetes.io/control-plane
        operator: DoesNotExist
```

**Traffic flow for a LoadBalancer Service:**

```
Client ARP: "Who has 192.168.1.220?"
  → Worker node responds (Cilium L2 announcement leader)
  → Client sends traffic to that worker's MAC
  → Cilium eBPF programs on the worker route to the correct pod
  → Pod responds directly back to client
```

**Lease-based leader election:** Each LoadBalancer IP has a Lease resource (`coordination.k8s.io/leases`) for leader election among worker nodes. The `cilium-leases-access` ClusterRole grants Cilium permission to manage these Leases.

**Primary gateway IP:** `192.168.1.220` is always the first IP assigned (to the homelab-gateway Service). The wildcard DNS record `*.madhan.app` resolves to this IP, so all cluster services are reachable at their hostnames.

## Gateway API

The [Gateway API](https://gateway-api.sigs.k8s.io/) is the successor to the Kubernetes Ingress spec. It separates infrastructure concerns (the Gateway, managed by the platform team) from application routing (HTTPRoutes, managed by app teams).

**Why Gateway API over Ingress:**

| Aspect | Ingress | Gateway API |
|--------|---------|-------------|
| Role separation | None | Gateway (infra) vs HTTPRoute (app) |
| Multi-protocol | HTTP/HTTPS only | HTTP, HTTPS, TCP, TLS, gRPC |
| Expressiveness | Limited header/path matching | Rich routing rules |
| Future | Legacy | Active development |

**The shared Gateway** (from [`core/platform/cilium.go`](https://github.com/madhank93/homelab/blob/v0.1.5/core/platform/cilium.go)):

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: homelab-gateway
  namespace: kube-system
spec:
  gatewayClassName: cilium
  listeners:
    - name: http
      protocol: HTTP
      port: 80
      allowedRoutes:
        namespaces:
          from: All       # Any namespace can attach an HTTPRoute
    - name: https
      protocol: HTTPS
      port: 443
      tls:
        mode: Terminate
        certificateRefs:
          - name: wildcard-madhan-app-tls   # cert-manager wildcard cert
            namespace: kube-system
      allowedRoutes:
        namespaces:
          from: All
```

`allowedRoutes.namespaces.from: All` is a deliberate single-tenant homelab choice. In a multi-tenant cluster you would restrict this to specific namespaces.

**Per-app HTTPRoutes:**

Each app creates its own `HTTPRoute` in its own namespace pointing back to the shared gateway. Example from `workloads/monitoring/grafana.go`:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: grafana
  namespace: grafana
spec:
  parentRefs:
    - name: homelab-gateway
      namespace: kube-system
  hostnames:
    - grafana.madhan.app
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: /
      backendRefs:
        - name: grafana
          port: 3000
```

Cilium runs an Envoy proxy per Gateway to handle HTTP routing and TLS termination.

## Traffic Flow

```
LAN Client
  → DNS: *.madhan.app → 192.168.1.220
  → ARP: worker node responds (Cilium L2 leader)
  → TCP: port 80 or 443 to worker
  → Cilium eBPF: routes to homelab-gateway (Envoy)
  → Envoy: matches HTTPRoute by hostname
  → Cilium eBPF: routes to app ClusterIP Service
  → Pod: handles request
```

## Hubble UI

Hubble UI is accessible at `http://hubble.madhan.app`. It provides network flow visibility across the cluster — which pods talk to which, which connections are dropped, and L7 protocol details.

```yaml
# HTTPRoute created by Pulumi in core/platform/cilium.go
hostnames: [hubble.madhan.app]
backendRefs: [{name: hubble-ui, port: 80, namespace: kube-system}]
```

## RBAC Patch

A `ClusterRole` and `ClusterRoleBinding` (`cilium-leases-access`) are created to allow Cilium to manage `coordination.k8s.io/leases` for L2 announcements. This is required because the Cilium Helm chart does not include these permissions by default.

## Troubleshooting

### Cilium CrashLoop (exit 137)

**Symptoms:** Cilium pods restart repeatedly with OOMKilled or exit 137.

**Diagnosis:**

```bash
kubectl get pods -n kube-system -l k8s-app=cilium
kubectl logs -n kube-system -l k8s-app=cilium --previous
```

**Fix:** Pod deletion alone does not fix BPF state corruption. Reboot the affected node:

```bash
talosctl --talosconfig ~/.talos/config reboot --nodes <node-IP>
```

**Why this happens:** The eBPF maps in the kernel can get corrupted if Cilium crashes mid-update. A node reboot clears all BPF state and lets Cilium start cleanly.

### L2 Announcement Not Working

```bash
# Check policy exists
kubectl get ciliuml2announcementpolicies

# Check IP pool
kubectl get ciliumloadbalancerippool

# Check leases
kubectl get leases -n kube-system | grep cilium

# Test ARP from a LAN machine
arping -I eth0 192.168.1.220
```
