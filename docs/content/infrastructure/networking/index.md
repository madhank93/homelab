+++
title = "Networking"
description = "Cilium CNI, Gateway API, L2 load balancer pool, and Hubble UI."
weight = 20
+++

## Cilium CNI

Cilium v1.16.6 is installed via Helm in the `kube-system` namespace. Key configuration:

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

## Load Balancer IP Pool

Cilium manages a `CiliumLoadBalancerIPPool` (`address-pool`) with the range `192.168.1.220–192.168.1.230`.

Services with `type: LoadBalancer` receive an IP from this pool via L2 ARP announcements. The `CiliumL2AnnouncementPolicy` (`l2-policy`) applies to all worker nodes (not control plane).

## Gateway API

Gateway API CRDs (experimental v1.2.1) are installed. A single shared `Gateway` resource (`homelab-gateway`) is created in `kube-system`:

```yaml
spec:
  gatewayClassName: cilium
  listeners:
    - name: http
      protocol: HTTP
      port: 80
      allowedRoutes:
        namespaces:
          from: All
    - name: https
      protocol: HTTPS
      port: 443
      tls:
        mode: Terminate
        certificateRefs:
          - name: wildcard-madhan-app-tls
            namespace: kube-system
```

> **Note:** The HTTPS listener requires the `wildcard-madhan-app-tls` Secret in `kube-system`. This Secret is provisioned by cert-manager once the `cloudflare-api-token` bootstrap secret exists. Until then, only HTTP routing is active.

## Traffic Flow

{% mermaid() %}
flowchart LR
    CLIENT[External Client]
    GW[homelab-gateway\nCiliumLoadBalancerIP\n192.168.1.220]
    HR_APP[HTTPRoute\napp.madhan.app]
    SVC[ClusterIP Service]
    POD[App Pod]

    CLIENT -->|HTTP port 80| GW
    GW --> HR_APP
    HR_APP --> SVC
    SVC --> POD
{% end %}

## HTTPRoutes

Each app creates an `HTTPRoute` in its own namespace pointing to the shared gateway:

```yaml
spec:
  parentRefs:
    - name: homelab-gateway
      namespace: kube-system
  hostnames:
    - app.madhan.app
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: /
      backendRefs:
        - name: app-service
          port: 80
```

## Hubble UI

Hubble UI is accessible at `http://hubble.madhan.app`. The HTTPRoute is created by Pulumi in `infra/pulumi/cilium.go`:

```yaml
hostnames: [hubble.madhan.app]
backendRefs: [{name: hubble-ui, port: 80}]
```

## RBAC Patch

A `ClusterRole` and `ClusterRoleBinding` (`cilium-leases-access`) are created to allow Cilium to manage `coordination.k8s.io/leases` for L2 announcements. This is required because the Cilium Helm chart does not include these permissions by default.
