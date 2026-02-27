+++
title = "Infrastructure"
description = "Pulumi stacks: Proxmox/Talos cluster, Cilium networking, Hetzner VPS, DNS/TLS, and secrets."
weight = 40
sort_by = "weight"
+++

Infrastructure is managed by **Pulumi** (Go), running from a developer laptop. It is never run in CI.

All Pulumi stacks live in `infra/pulumi/`. Secrets are injected at runtime via SOPS:

```bash
just pulumi talos up
just pulumi platform up
```

## Architecture

{% mermaid() %}
flowchart TD
    DEV[Developer Laptop\nPulumi + SOPS]

    subgraph Proxmox["Proxmox Host"]
        CP1[k8s-controller1\n192.168.1.211]
        CP2[k8s-controller2\n192.168.1.212]
        CP3[k8s-controller3\n192.168.1.213]
        W1[k8s-worker1\n192.168.1.221]
        W2[k8s-worker2\n192.168.1.222]
        W3[k8s-worker3\n192.168.1.223]
        W4[k8s-worker4 GPU\n192.168.1.224]
        VIP[VIP 192.168.1.210\nkube-apiserver]
    end

    subgraph Platform["Platform Layer"]
        CILIUM[Cilium CNI\nGateway API]
        ARGOCD[ArgoCD\nApplicationSet]
        LONGHORN[Longhorn\nStorage]
    end

    subgraph Hetzner["Hetzner Cloud"]
        VPS[Bifrost VPS\nNetBird + TURN]
    end

    DEV -->|pulumi up| Proxmox
    DEV -->|pulumi up| Hetzner
    CP1 & CP2 & CP3 --> VIP
    VIP --> CILIUM
    CILIUM --> ARGOCD
    ARGOCD --> LONGHORN
{% end %}

## Pulumi Stacks

| File | Responsibility |
|------|----------------|
| `talos.go` | Proxmox VMs + Talos cluster bootstrap |
| `cilium.go` | Cilium CNI + Gateway API + Hubble HTTPRoute |
| `argocd.go` | ArgoCD Helm + ApplicationSet |
| `hetzner_vps.go` | Hetzner VPS (Bifrost: NetBird + TURN) |
| `proxmox.go` | Proxmox provider + VM helpers |
