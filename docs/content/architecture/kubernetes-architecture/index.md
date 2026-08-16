+++
title = "Kubernetes Architecture"
description = "Node layout, CNI, platform services, and workload placement on the Talos cluster."
weight = 20
+++

## Overview

Three Talos control-plane nodes share a virtual IP. Four worker nodes run all workloads. **Cilium** handles CNI, Gateway API ingress, and L2 LoadBalancer announcements. **ArgoCD** syncs all apps via GitOps.

---

## Node Inventory

| Name | Role | IP | CPU | RAM | Storage |
|------|------|----|-----|-----|---------|
| `k8s-controller1` | Control plane | 192.168.1.211 | 4 vCPU | 8 GiB | 50 GiB |
| `k8s-controller2` | Control plane | 192.168.1.212 | 4 vCPU | 8 GiB | 50 GiB |
| `k8s-controller3` | Control plane | 192.168.1.213 | 4 vCPU | 8 GiB | 50 GiB |
| `k8s-worker1` | Worker | 192.168.1.221 | 4 vCPU | 14 GiB | 200 GiB |
| `k8s-worker2` | Worker | 192.168.1.222 | 4 vCPU | 14 GiB | 200 GiB |
| `k8s-worker3` | Worker | 192.168.1.223 | 4 vCPU | 14 GiB | 200 GiB |
| `k8s-worker4` | Worker + GPU | 192.168.1.224 | 8 vCPU | 16 GiB | 250 GiB + RTX 5070 Ti |
| **Talos VIP** | Virtual IP | 192.168.1.210 | — | — | Floats across control-plane nodes |
| **Cilium L2 LB** | LoadBalancer pool | 192.168.1.220–230 | — | — | Assigned per LoadBalancer Service |

---

## Cluster Diagram

{% mermaid() %}
flowchart TB
    subgraph INTERNET["Internet"]
        USER["Browser / Client"]
        GH["GitHub\nrepo: madhank93/homelab"]
        CF["Cloudflare DNS\n*.madhan.app"]
    end

    subgraph BIFROST["Bifrost VPS · Hetzner · 178.156.199.250"]
        TRF["Traefik\nTLS termination + ForwardAuth"]
        AUTH["Authentik\nOIDC / SSO broker"]
        NBS["NetBird\nManagement + Signal + Relay"]
        NBA["netbird-agent\nWireGuard routing peer\n→ advertises 192.168.1.0/24"]
    end

    subgraph LAN["On-Prem LAN · 192.168.1.0/24"]
        subgraph CP["Control Plane · Talos"]
            VIP["Talos VIP\n192.168.1.210:6443"]
            CP1["controller1\n.211"]
            CP2["controller2\n.212"]
            CP3["controller3\n.213"]
            VIP --- CP1 & CP2 & CP3
        end

        subgraph PLT["Platform Layer"]
            CIL["Cilium\nCNI · kube-proxy replacement\nL2 LB 192.168.1.220–230\nGateway API"]
            CERT["cert-manager\nCloudflare DNS-01\nwildcard TLS"]
            ARGO["Argo CD\nApplicationSet → v0.1.7-manifests"]
        end

        subgraph SECRETS["Secrets Layer"]
            OB["OpenBao\nVault-compatible KV"]
            CSI["CSI Secrets Store\nfile mounts + k8s Secrets"]
        end

        subgraph STORAGE["Storage"]
            LONG["Longhorn\nReplicated block storage"]
            CNPG["CloudNativePG\nPostgreSQL operator"]
        end

        subgraph WORKERS["Workers · k8s-worker1–3"]
            OBS["Observability\nVictoriaMetrics · VictoriaLogs\nGrafana · OTel"]
            SEC["Security\nFalco · Kyverno · Trivy"]
            APPS["Applications\nn8n · Harbor · Headlamp\nOpenBao · NetBird peer"]
        end

        subgraph GPU["k8s-worker4 · RTX 5070 Ti"]
            AI["AI Workloads\nOllama · ComfyUI\nKubeflow"]
            NVIDIA["NVIDIA Device Plugin\nDCGM Exporter"]
        end
    end

    subgraph GITOPS["GitOps · GitHub"]
        SRC["release branch\nPulumi + CDK8s source"]
        MFST["v0.1.7-manifests branch\nSynthesized YAML"]
        GHA["GitHub Actions\ncdk8s synth + publish"]
    end

    USER -->|"HTTPS"| CF
    CF -->|"public services\nauth/netbird/grafana"| TRF
    CF -->|"LAN services via VPN\nharbor/headlamp/etc"| CIL
    TRF -->|"ForwardAuth"| AUTH
    TRF -->|"proxy via NetBird\nWireGuard tunnel"| CIL
    NBA <-->|"WireGuard mesh"| APPS
    NBS --- NBA
    GH --- SRC
    SRC -->|"push triggers"| GHA
    GHA -->|"publishes"| MFST
    MFST -->|"ArgoCD watches"| ARGO
    ARGO -->|"syncs"| WORKERS & GPU & SECRETS & STORAGE
    VIP --> CIL
    CIL --> WORKERS & GPU
    CERT -->|"ACME DNS-01"| CF
    OB --> CSI
    CSI -->|"volume mounts"| WORKERS & GPU & APPS
    LONG --> WORKERS & GPU
    CNPG --> WORKERS
{% end %}

---

## Talos Configuration

Talos Linux is provisioned by Pulumi (`core/platform/talos.go`). Each role gets a machine config with role-specific patches:

| Patch | Controller | Worker | Worker4 (GPU) |
|-------|-----------|--------|---------------|
| `cpPatch` | ✓ | — | — |
| `workerPatch` | — | ✓ | — |
| `gpuWorkerPatch` | — | — | ✓ |

These are inline Go strings in {{ src(path="core/platform/talos.go") }}, not
separate YAML files.

**Talos image schematics** (from factory.talos.dev):

| Schematic | Extensions | Used by |
|-----------|-----------|---------|
| Base | `iscsi-tools`, `util-linux-tools`, `qemu-guest-agent` | All nodes |
| GPU | Base + `nvidia-container-toolkit`, `nvidia-open-gpu-kernel-modules` | `k8s-worker4` |

The cluster endpoint is `https://192.168.1.210:6443` (Talos VIP).

---

## Cilium + Gateway API

Cilium handles both CNI and north-south ingress via the Gateway API:

| Feature | Config |
|---------|--------|
| CNI mode | kube-proxy replacement |
| L2 announcements | `192.168.1.220–230` pool (LAN) |
| Gateway class | `cilium` |
| HTTPRoute for Hubble UI | `hubble.madhan.app → hubble-ui:80` |
| ForwardAuth | Via Traefik on Bifrost (not in-cluster) |

The Gateway API `GatewayClass` is provisioned by `core/platform/cilium.go`. App HTTPRoutes are defined in CDK8s (`workloads/**/*.go`).

---

## Workload Placement

| Package | Components | Node affinity |
|---------|-----------|--------------|
| `storage/` | Longhorn | DaemonSet — all workers |
| `secrets/` | OpenBao + CSI Driver | Any worker |
| `observability/` | VictoriaMetrics, VictoriaLogs, OTel | Deployment + DaemonSet |
| `monitoring/` | Grafana | Any worker |
| `security/` | Falco (eBPF), Kyverno, Trivy | DaemonSet + CronJob |
| `hardware/` | NVIDIA device plugin + DCGM | DaemonSet, NodeFeatureDiscovery |
| `networking/` | NetBird peer | `hostNetwork: true`, any worker |
| `registry/` | Harbor | Deployments + RWO PVCs |
| `automation/` | n8n + PostgreSQL | Any worker |
| `ai/` | Ollama, ComfyUI, Kubeflow | **`k8s-worker4` only** (GPU) |
| `management/` | Headlamp | Any worker |
| `support/` | Stakater Reloader | Any worker |

---

## Service Access

| Service URL | DNS resolves to | Access |
|------------|----------------|--------|
| `grafana.madhan.app` | `178.156.199.250` (public) | Via Bifrost + ForwardAuth |
| `auth.madhan.app` | `178.156.199.250` (public) | Authentik on Bifrost |
| `netbird.madhan.app` | `178.156.199.250` (public) | NetBird on Bifrost |
| `harbor.madhan.app` | `192.168.1.220` (LAN) | LAN or VPN only |
| `headlamp.madhan.app` | `192.168.1.220` (LAN) | LAN or VPN only |
| `hubble.madhan.app` | `192.168.1.220` (LAN) | LAN or VPN only |

See [Network Flow](@/architecture/network-flow/index.md) for the complete traffic path
breakdown, and the [Software Inventory](@/architecture/software-inventory.md) for the
version of everything named above.
