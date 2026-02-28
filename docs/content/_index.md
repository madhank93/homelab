+++
template = "landing.html"
title = "Homelab"

[extra.hero]
title = "Homelab"
badge = "Talos · Proxmox · Pulumi · ArgoCD · CDK8s · NetBird · Authentik"
description = "A production-grade Kubernetes homelab on Talos Linux, provisioned entirely by Pulumi, with full GitOps via ArgoCD and CDK8s. One command deploys everything — from bare VMs to running apps."
cta_buttons = [
    { text = "Getting Started", url = "/getting-started", style = "primary" },
    { text = "View on GitHub", url = "https://github.com/madhank93/homelab", style = "secondary" },
]

[extra.features_section]
title = "What's Inside"
description = "Fully automated, GitOps-driven homelab — from bare metal to running apps"

[[extra.features_section.features]]
title = "Talos Linux"
desc = "Immutable, API-driven OS. No SSH, no shell, fully declarative. Provisioned by Pulumi on Proxmox with per-role machine configs."
icon = "fa-solid fa-shield-halved"

[[extra.features_section.features]]
title = "Pulumi IaC (Go)"
desc = "Infrastructure as typed Go code. Proxmox VMs, Talos cluster, Cilium, ArgoCD, Hetzner VPS, Cloudflare DNS — all in one codebase."
icon = "fa-solid fa-code"

[[extra.features_section.features]]
title = "ArgoCD GitOps"
desc = "CDK8s synthesizes manifests to a branch; ArgoCD ApplicationSet detects every new directory and syncs automatically. Zero manual kubectl."
icon = "fa-solid fa-arrows-rotate"

[[extra.features_section.features]]
title = "Bifrost Edge Layer"
desc = "Hetzner VPS running Traefik v3.3, NetBird v0.66, and Authentik. Fully bootstrapped by a single Pulumi command — no manual SSH steps."
icon = "fa-solid fa-globe"

[[extra.features_section.features]]
title = "Two-Tier Secrets"
desc = "Bootstrap secrets encrypted with SOPS/age, committed safely to git. Runtime secrets managed by Infisical operator and never written to manifests."
icon = "fa-solid fa-lock"

[[extra.features_section.features]]
title = "GPU Workloads"
desc = "NVIDIA RTX 5070 Ti with PCIe passthrough and time-slicing. Ollama for LLM inference, ComfyUI for Stable Diffusion / Flux image generation."
icon = "fa-solid fa-microchip"

[[extra.features_section.features]]
title = "Full Observability"
desc = "VictoriaMetrics + VictoriaLogs + Grafana + OpenTelemetry collector on every node. Falco eBPF syscall monitoring. Trivy vulnerability scanning."
icon = "fa-solid fa-chart-line"

[[extra.features_section.features]]
title = "Zero-Touch Bootstrap"
desc = "bootstrap.sh starts all services in order, waits for health checks, auto-provisions the NetBird IDP token via Authentik Django ORM, and substitutes secrets in config — all from one Pulumi deploy."
icon = "fa-solid fa-rocket"
+++
