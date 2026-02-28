+++
title = "Kubernetes Architecture"
description = "Node layout, CNI, platform services, and workload placement on the Talos cluster."
weight = 20
+++

## Overview

Three Talos control-plane nodes sit behind a KubeVIP virtual IP. Four worker nodes run all workloads. **Cilium** handles CNI, Gateway API ingress, and L2 LoadBalancer announcements. **ArgoCD** syncs all apps via GitOps.

---

## Node Inventory

| Name | Role | IP | CPU | RAM | Storage |
|------|------|----|-----|-----|---------|
| `k8s-controller1` | Control plane | 192.168.1.211 | 4 vCPU | 6 GiB | 30 GiB |
| `k8s-controller2` | Control plane | 192.168.1.212 | 4 vCPU | 6 GiB | 30 GiB |
| `k8s-controller3` | Control plane | 192.168.1.213 | 4 vCPU | 6 GiB | 30 GiB |
| `k8s-worker1` | Worker | 192.168.1.221 | 4 vCPU | 6 GiB | 125 GiB |
| `k8s-worker2` | Worker | 192.168.1.222 | 4 vCPU | 6 GiB | 125 GiB |
| `k8s-worker3` | Worker | 192.168.1.223 | 4 vCPU | 6 GiB | 125 GiB |
| `k8s-worker4` | Worker + GPU | 192.168.1.224 | 4 vCPU | 6 GiB | 125 GiB + RTX 5070 Ti |
| **KubeVIP** | Virtual IP | 192.168.1.210 | — | — | Floats across control-plane nodes |
| **Cilium L2 LB** | LoadBalancer pool | 192.168.1.220–230 | — | — | Assigned per LoadBalancer Service |

---

## Cluster Diagram

{% mermaid() %}
flowchart TB
    subgraph CP["Control Plane · 3 nodes"]
        VIP["KubeVIP<br/>192.168.1.210:6443"]
        CP1["k8s-controller1<br/>.211"]
        CP2["k8s-controller2<br/>.212"]
        CP3["k8s-controller3<br/>.213"]
        VIP --- CP1 & CP2 & CP3
    end

    subgraph WRK["Workers"]
        W1["k8s-worker1<br/>.221"]
        W2["k8s-worker2<br/>.222"]
        W3["k8s-worker3<br/>.223"]
        W4["k8s-worker4<br/>.224<br/>RTX 5070 Ti"]
    end

    subgraph PLT["Platform Services"]
        CIL["Cilium L2 LB<br/>192.168.1.220"]
        ARGO["ArgoCD<br/>ApplicationSet"]
        LONG["Longhorn<br/>distributed storage"]
        INF["Infisical<br/>operator + secrets"]
        CERT["cert-manager<br/>wildcard TLS"]
    end

    subgraph BIFROST["Bifrost VPS · Hetzner"]
        WG["NetBird routing peer<br/>WireGuard mesh"]
    end

    VIP --> CIL
    CIL --> W1 & W2 & W3 & W4
    ARGO -->|"syncs workloads<br/>from manifests branch"| W1 & W2 & W3 & W4
    LONG --> W1 & W2 & W3 & W4
    INF -->|"InfisicalSecret CRs"| W1 & W2 & W3 & W4
    CERT -->|"cert-manager ACME"| CIL
    WG <-->|"WireGuard<br/>192.168.1.0/24"| CIL
{% end %}

---

## Talos Configuration

Talos Linux is provisioned by Pulumi (`core/platform/talos.go`). Each role gets a machine config with role-specific patches:

| Patch | Controller | Worker | Worker4 (GPU) |
|-------|-----------|--------|---------------|
| `controlplane.patch.yaml` | ✓ | — | — |
| `worker.patch.yaml` | — | ✓ | ✓ |
| `nvidia.patch.yaml` | — | — | ✓ |

**Talos image schematics** (from factory.talos.dev):

| Schematic | Extensions | Used by |
|-----------|-----------|---------|
| Base | `iscsi-tools`, `qemu-guest-agent` | All nodes |
| GPU | Base + `nvidia-container-toolkit` | `k8s-worker4` |

The cluster endpoint is `https://192.168.1.210:6443` (KubeVIP).

---

## Cilium + Gateway API

Cilium handles both CNI and north-south ingress via the Gateway API:

| Feature | Config |
|---------|--------|
| CNI mode | kube-proxy replacement |
| L2 announcements | `192.168.1.220–230` pool (LAN) |
| Gateway class | `cilium` |
| HTTPRoute for Hubble UI | `hubble.madhan.app → hubble-relay:80` |
| ForwardAuth | Via Traefik on Bifrost (not in-cluster) |

The Gateway API `GatewayClass` is provisioned by `core/platform/cilium.go`. App HTTPRoutes are defined in CDK8s (`workloads/**/*.go`).

---

## Workload Placement

| Package | Components | Node affinity |
|---------|-----------|--------------|
| `storage/` | Longhorn | DaemonSet — all workers |
| `secrets/` | Infisical + PostgreSQL | Any worker |
| `observability/` | VictoriaMetrics, VictoriaLogs, OTel | Deployment + DaemonSet |
| `monitoring/` | Grafana | Any worker |
| `security/` | Falco (eBPF), Kyverno, Trivy | DaemonSet + CronJob |
| `hardware/` | NVIDIA GPU Operator | DaemonSet, NodeFeatureDiscovery |
| `networking/` | NetBird peer | `hostNetwork: true`, any worker |
| `registry/` | Harbor | Deployments + RWO PVCs |
| `automation/` | n8n + PostgreSQL | Any worker |
| `ai/` | Ollama, ComfyUI | **`k8s-worker4` only** (GPU) |
| `management/` | Headlamp, Rancher, Fleet | Any worker |
| `support/` | Stakater Reloader | Any worker |

---

## Service Access

| Service URL | DNS resolves to | Access |
|------------|----------------|--------|
| `grafana.madhan.app` | `178.156.199.250` (public) | Via Bifrost + ForwardAuth |
| `harbor.madhan.app` | `178.156.199.250` (public) | Via Bifrost + ForwardAuth |
| `auth.madhan.app` | `178.156.199.250` (public) | Authentik on Bifrost |
| `netbird.madhan.app` | `178.156.199.250` (public) | NetBird on Bifrost |
| `headlamp.madhan.app` | `192.168.1.220` (LAN) | LAN or VPN only |
| `infisical.madhan.app` | `192.168.1.220` (LAN) | LAN or VPN only |
| `hubble.madhan.app` | `192.168.1.220` (LAN) | LAN or VPN only |

See [Network Flow](/architecture/network-flow) for the complete traffic path breakdown.
