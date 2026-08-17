+++
title = "Glossary"
description = "Terms these docs use before defining them."
weight = 66
+++

Homelab-specific names and the jargon that shows up without introduction.

| Term | What it means here |
|---|---|
| **Bifrost** | The Hetzner VPS that acts as the public edge. Named after the Norse rainbow bridge. Runs Traefik, Authentik and the NetBird server under Docker Compose — see [Hetzner Bifrost](@/infrastructure/hetzner-bifrost/index.md) |
| **wt0** | The WireGuard interface NetBird creates. It is NOARP/POINTOPOINT, which is why it must never appear in Cilium's device list — see [Network Flow](@/architecture/network-flow/index.md#why-wt0-is-not-in-cilium-devices) |
| **k8s-routing-peer** | The NetBird peer name of the in-cluster `netbird-peer` pod. It advertises `192.168.1.0/24` into the mesh so Bifrost can reach cluster services |
| **Pattern A / Pattern B** | The two ways a pod consumes an OpenBao secret. A mounts it as a file only; B additionally syncs a Kubernetes Secret. Defined in [Secrets](@/platform/secrets/index.md) |
| **ForwardAuth** | Traefik middleware that asks Authentik whether a request is authenticated before proxying it. Public services either use it or handle their own auth |
| **schematic** | A Talos image built by factory.talos.dev with a specific set of system extensions. This cluster uses two: a base one and a GPU one |
| **manifests branch** | `v0.1.7-manifests`. CI synthesizes CDK8s output and force-pushes it there; Argo CD watches it. Never edited by hand |
| **share-manager** | The NFS server pod Longhorn starts for each RWX volume, letting several pods attach at once |
| **VMAgent / VMAlert / VMSingle** | VictoriaMetrics components: the scraper, the rule evaluator, and the single-node store. This cluster does not run cluster mode |
| **time-slicing** | Splitting one physical GPU into several schedulable `nvidia.com/gpu` units. It divides compute time, **not** VRAM — see [GPU](@/hardware/gpu/index.md#sharing-the-16-gb-of-vram) |
| **sm_120** | The CUDA compute capability of the RTX 5070 Ti (Blackwell). It needs CUDA 12.8 or newer, which constrains which container images will run |
| **CDK8s** | Generates Kubernetes YAML from Go. Every workload here is Go code, not hand-written YAML — see [CDK8s](@/platform/cdk8s/index.md) |
| **SOPS / age** | The encryption used for `secrets/bootstrap.sops.yaml`, which is safe to commit. `age` is the key format; `sops` is the tool |
| **`src()`** | A Zola shortcode in these docs that links to a file in the repo at the right release ref — see [This Documentation Site](@/development/docs-site/index.md) |
