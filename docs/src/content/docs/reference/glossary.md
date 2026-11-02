---
title: Glossary
description: Homelab-specific names and the jargon these docs use before defining it.
sidebar:
  order: 3
---

| Term | What it means here |
|---|---|
| **age** | Key format used by SOPS to encrypt `secrets/bootstrap.sops.yaml`. See **SOPS** |
| **Bifrost** | The Hetzner VPS at the public edge, running Traefik, Authentik, the NetBird server and Gatus under Docker Compose — see [Bifrost](/cloud/bifrost/) |
| **CDK8s** | Generates Kubernetes YAML from Go; every workload here is Go code — see [CDK8s](/platform/cdk8s/) |
| **ForwardAuth** | Traefik middleware that asks Authentik whether a request is authenticated before proxying it. Public services either use it or handle their own auth |
| **k8s-routing-peer** | NetBird peer name of the in-cluster `netbird-peer` pod. It advertises `192.168.1.0/24` into the mesh so Bifrost can reach cluster services |
| **manifests branch** | Branch CI force-pushes the CDK8s output to and Argo CD watches (`manifests_branch` in `config/config.yml`). Never edited by hand |
| **Pattern A / Pattern B** | The two ways a pod consumes an OpenBao secret: A mounts it as a file only; B also syncs a Kubernetes Secret — see [Secrets](/concepts/secrets/) |
| **schematic** | A Talos image built by factory.talos.dev with a set of system extensions. This cluster uses two: base and GPU |
| **share-manager** | The NFS server pod Longhorn starts for each RWX volume so several pods can mount it |
| **sm_120** | CUDA compute capability of the RTX 5070 Ti (Blackwell). Needs CUDA 12.8 or newer, which limits which images run |
| **SOPS** | Encrypts `secrets/bootstrap.sops.yaml` so it is safe to commit; `age` is the key format, `sops` the tool |
| **`src:` link** | Markdown link whose target starts with `src:`, which the docs build rewrites to the file on GitHub at the documented ref — see [Docs site](/reference/development/docs-site/) |
| **time-slicing** | Splitting one GPU into several schedulable `nvidia.com/gpu` units. Divides compute time, **not** VRAM — see [GPU](/hardware/gpu/) |
| **VMAgent / VMAlert / VMSingle** | VictoriaMetrics scraper, rule evaluator and single-node store. Cluster mode is not used |
| **wt0** | The WireGuard interface NetBird creates. NOARP/POINTOPOINT, so it must never be in Cilium's device list — see [Traffic](/concepts/traffic/) |
