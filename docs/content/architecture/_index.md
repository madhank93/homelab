+++
title = "Architecture"
description = "System architecture: network topology, Kubernetes layout, GitOps pipeline, and secrets management."
weight = 15
sort_by = "weight"
+++

The homelab is designed around four principles: **everything as code**, **no manual cluster changes**, **secrets never in git plaintext**, and **zero-touch automation** for repetitive operational tasks.

---

## Four-Layer Stack

```
┌─────────────────────────────────────────────────────────────────┐
│  APPS                                                           │
│  ComfyUI · Ollama · Grafana · Harbor · n8n · Falco · Trivy …  │
│  Managed by: ArgoCD (GitOps) + CDK8s (manifest synthesis)      │
├─────────────────────────────────────────────────────────────────┤
│  PLATFORM                                                       │
│  Talos Linux K8s · Cilium CNI · Gateway API · cert-manager     │
│  Managed by: Pulumi (Go)                                        │
├─────────────────────────────────────────────────────────────────┤
│  INFRASTRUCTURE                                                 │
│  Proxmox VMs (7 nodes) · Hetzner VPS (Bifrost edge)            │
│  Managed by: Pulumi (Go)                                        │
├─────────────────────────────────────────────────────────────────┤
│  HARDWARE                                                       │
│  Proxmox host · NVIDIA RTX 5070 Ti                             │
│  Managed by: Manual                                             │
└─────────────────────────────────────────────────────────────────┘
```

---

## Architecture Sections

| Section | What it covers |
|---------|---------------|
| [Network Flow](/architecture/network-flow) | How traffic reaches services — public internet, LAN, and VPN paths |
| [Kubernetes Architecture](/architecture/kubernetes-architecture) | Node layout, CNI, platform services, workload placement |
| [GitOps Flow](/architecture/gitops-flow) | Pulumi infra path vs CDK8s workload path; CI pipeline |
| [Secrets Flow](/architecture/secrets-flow) | SOPS bootstrap secrets + Infisical runtime secrets + Bifrost auto-provisioning |
