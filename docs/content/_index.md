+++
template = "landing.html"
title = "Homelab"

[extra.hero]
title = "Homelab Documentation"
badge = "Talos · Proxmox · ArgoCD · CDK8s"
description = "A production-grade Kubernetes homelab running on Talos Linux, provisioned by Pulumi on Proxmox, with full GitOps via ArgoCD and CDK8s."
cta_buttons = [
    { text = "Hardware", url = "/hardware", style = "primary" },
    { text = "View on GitHub", url = "https://github.com/madhank93/homelab", style = "secondary" },
]

[extra.features_section]
title = "What's Inside"
description = "A fully automated, GitOps-driven homelab"

[[extra.features_section.features]]
title = "Talos Linux"
desc = "Immutable, API-driven OS for Kubernetes. No SSH. No shell. Fully declarative."
icon = "fa-solid fa-shield-halved"

[[extra.features_section.features]]
title = "Pulumi IaC"
desc = "Infrastructure as Go code. Proxmox VMs, Talos cluster, Cilium, ArgoCD — all in code."
icon = "fa-solid fa-code"

[[extra.features_section.features]]
title = "ArgoCD GitOps"
desc = "CDK8s synthesizes manifests to a branch; ArgoCD syncs the cluster automatically."
icon = "fa-solid fa-arrows-rotate"

[[extra.features_section.features]]
title = "SOPS + Infisical"
desc = "Bootstrap secrets encrypted with age/SOPS. Runtime secrets via Infisical operator."
icon = "fa-solid fa-lock"

[[extra.features_section.features]]
title = "GPU Workloads"
desc = "NVIDIA RTX 5070 Ti with time-slicing. Ollama LLM inference + ComfyUI image generation."
icon = "fa-solid fa-microchip"

[[extra.features_section.features]]
title = "Full Observability"
desc = "VictoriaMetrics + VictoriaLogs + Grafana + OTel collector on every node."
icon = "fa-solid fa-chart-line"
+++
