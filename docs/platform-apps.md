# Platform Apps Catalog

This document is the authoritative catalog of every application managed by CDK8s in `platform/cdk8s/cots/`.

---

## Table of Contents

- [All Apps at a Glance](#all-apps-at-a-glance)
- [Folder Structure](#folder-structure)
- [CDK8s Workflow](#cdk8s-workflow)
- [Apps by Folder](#apps-by-folder)
  - [ai/](#ai)
  - [automation/](#automation)
  - [compliance/](#compliance)
  - [management/](#management)
  - [monitoring/](#monitoring)
  - [registry/](#registry)
  - [security/](#security)
  - [storage/](#storage)

---

## All Apps at a Glance

| Folder | App | Namespace | External Endpoint | UI? | Purpose |
|--------|-----|-----------|-------------------|-----|---------|
| ai | nvidia-gpu-operator | nvidia-gpu-operator | — | No | GPU device plugin + NFD labels |
| ai | ollama | ollama | http://ollama.madhan.app | No (REST API) | LLM inference server |
| ai | comfyui | comfyui | http://comfyui.madhan.app | Yes | Stable Diffusion / Flux image generation |
| automation | n8n | n8n | http://n8n.madhan.app | Yes | Workflow automation |
| compliance | kyverno | kyverno | — | No | Admission controller + policy engine |
| compliance | trivy | trivy | — | No | Vulnerability scanning (CRD reports) |
| compliance | falco | falco | — | No | Runtime security (syscall monitoring) |
| management | rancher | cattle-system | http://rancher.madhan.app | Yes | Multi-cluster Kubernetes management |
| management | headlamp | headlamp | http://headlamp.madhan.app | Yes | Kubernetes dashboard |
| management | fleet | fleet | — | No | GitOps fleet management (Rancher Fleet) |
| monitoring | grafana | grafana | http://grafana.madhan.app | Yes | Dashboards and data visualization |
| monitoring | victoria-metrics | victoria-metrics | — | No | Time-series metrics storage |
| monitoring | victoria-logs | victoria-logs | — | No | Log storage (Loki-compatible API) |
| monitoring | alertmanager | alertmanager | — | No | Alert routing and deduplication |
| monitoring | otel-collector | opentelemetry | — | No | Metrics and log collection pipeline |
| registry | harbor | harbor | http://harbor.madhan.app | Yes | Container image registry |
| security | infisical | infisical | http://infisical.madhan.app | Yes | Secrets management platform |
| storage | longhorn | longhorn-system | — | No | Distributed block storage |

---

## Folder Structure

```
platform/cdk8s/cots/
├── ai/                     # GPU workloads and AI inference
│   ├── nvidia_gpu_operator.go
│   ├── ollama.go
│   └── comfyui.go
├── automation/             # Workflow and task automation
│   └── n8n.go
├── compliance/             # Policy enforcement and security auditing
│   ├── keyverno.go
│   ├── trivy.go
│   └── falco.go
├── management/             # Cluster management tooling
│   ├── headlamp.go
│   ├── rancher.go
│   └── fleet_device_manager.go
├── monitoring/             # Observability stack
│   ├── grafana.go
│   ├── victoria_metrics.go
│   ├── victoria_logs.go
│   ├── alert_manager.go
│   └── otel_collector.go
├── registry/               # Container image registry
│   └── harbor.go
├── security/               # Secrets and identity management
│   └── infisical.go
└── storage/                # Persistent storage
    └── longhorn.go
```

### Why This Organization

Each folder maps to a distinct operational concern:

- **ai/** — Workloads that require GPU hardware. All apps here use `runtimeClassName: nvidia` and request `nvidia.com/gpu: 1`.
- **automation/** — User-facing workflow tools. Currently n8n; future candidates include Temporal, Argo Workflows.
- **compliance/** — Non-interactive security agents. Kyverno enforces policy at admission time; Trivy scans images; Falco monitors syscalls at runtime. No UIs.
- **management/** — Operator-facing cluster UIs and fleet tooling. All have web interfaces or fleet-management APIs.
- **monitoring/** — The full observability stack. Separated from management because these are infrastructure concerns (metrics pipelines, storage, alerting), not operator UIs.
- **registry/** — Harbor is the only registry; isolated because image registries have distinct operational concerns (garbage collection, replication, proxy cache).
- **security/** — Infisical manages all runtime application secrets. Separated from compliance because it is an active runtime dependency for other apps, not a passive scanner.
- **storage/** — Longhorn is the cluster's default StorageClass provider. Isolated because all other apps depend on it; it must deploy first.

---

## CDK8s Workflow

### How platform/cdk8s/ Becomes Running Pods

```
1. Developer edits platform/cdk8s/cots/<folder>/<app>.go
        │
        ▼
2. Push to branch v0.1.5
        │
        ▼
3. GitHub Actions CI (.github/workflows/ci.yml)
   └── go build ./...         (compile check)
   └── go run main.go         (cdk8s synth)
       └── writes YAML to app/<appname>/*.yaml
        │
        ▼
4. CI commits generated YAML to v0.1.5-manifests branch
        │
        ▼
5. ArgoCD watches v0.1.5-manifests branch
   └── Detects change in app/<appname>/
   └── Applies YAML to cluster via kubectl
        │
        ▼
6. Kubernetes creates/updates resources
```

### Key Properties of This Workflow

- **No secrets in git** — CDK8s generates zero Secret resources. Bootstrap secrets are created by `just create-secrets` (SOPS-encrypted, run from laptop only).
- **Manifests are generated, not hand-written** — The `v0.1.5-manifests` branch contains machine-generated YAML. Never edit it by hand.
- **ArgoCD is the single source of truth** for the cluster state. Every app is an ArgoCD Application pointing to a subdirectory of `app/` in the manifests branch.
- **One CDK8s app per ArgoCD Application** — Each `main.go` entry creates output in a separate `app/<name>/` directory.

---

## Apps by Folder

### ai/

#### NVIDIA GPU Operator

| Property | Value |
|----------|-------|
| File | `platform/cdk8s/cots/ai/nvidia_gpu_operator.go` |
| Namespace | `nvidia-gpu-operator` |
| Helm chart | `gpu-operator` v25.10.1 (helm.ngc.nvidia.com/nvidia) |
| HTTPRoute | None |
| UI | No |

**Purpose**: Manages NVIDIA GPU access within the Kubernetes cluster. On Talos Linux, the GPU driver and container toolkit are provided by Talos system extensions (`nvidia-open-gpu-kernel-modules-production`, `nvidia-container-toolkit-production`). The operator's role is:
- Node Feature Discovery (NFD): labels GPU nodes with `nvidia.com/gpu.present=true`
- Device plugin: advertises `nvidia.com/gpu` as a schedulable resource
- DCGM exporter: exports GPU metrics to the monitoring stack

**Key configuration for Talos**:
- `driver.enabled: true` — the operator detects Talos extensions and sets `gpu.deploy.driver=pre-installed`, causing the driver DaemonSet DESIRED=0. The Talos extension provides the driver.
- `toolkit.enabled: false` — Talos extension provides the container toolkit.
- `DEVICE_LIST_STRATEGY=envvar` — avoids CDI hostPath issues on Talos (see `docs/nvidia-gpu-talos.md`).
- Includes a Talos Validation Bridge DaemonSet that writes marker files the GPU operator needs to unblock device plugin startup.

---

#### Ollama

| Property | Value |
|----------|-------|
| File | `platform/cdk8s/cots/ai/ollama.go` |
| Namespace | `ollama` |
| Helm chart | `ollama` v1.41.0 (otwld.github.io/ollama-helm) |
| HTTPRoute | `ollama.madhan.app` → `ollama:11434` |
| UI | No (REST API only) |
| Node | k8s-worker4 (GPU node) |

**Purpose**: Runs open-source LLMs (llama3.2, mistral, deepseek-r1, etc.) using the RTX 5070 Ti GPU. Exposes the OpenAI-compatible REST API at `ollama.madhan.app`.

**GPU configuration**: `runtimeClassName: nvidia`, `nvidia.com/gpu: 1`, `NVIDIA_VISIBLE_DEVICES=all`.

---

#### ComfyUI

| Property | Value |
|----------|-------|
| File | `platform/cdk8s/cots/ai/comfyui.go` |
| Namespace | `comfyui` |
| Image | `yanwk/comfyui-boot:latest-cu128` |
| HTTPRoute | `comfyui.madhan.app` → `comfyui:8188` |
| UI | Yes |
| Storage | 100Gi Longhorn PVC |

**Purpose**: Node-based image generation UI for Stable Diffusion and Flux models. Downloads models to persistent Longhorn storage.

**GPU configuration**: `runtimeClassName: nvidia`, `nvidia.com/gpu: 1`, strategy `Recreate` (only one GPU workload at a time).

---

### automation/

#### n8n

| Property | Value |
|----------|-------|
| File | `platform/cdk8s/cots/automation/n8n.go` |
| Namespace | `n8n` |
| HTTPRoute | `n8n.madhan.app` → n8n service |
| UI | Yes |

**Purpose**: Open-source workflow automation platform. Integrates with APIs, databases, and services for scheduled and event-driven automation.

---

### compliance/

#### Kyverno

| Property | Value |
|----------|-------|
| File | `platform/cdk8s/cots/compliance/keyverno.go` |
| Namespace | `kyverno` |
| HTTPRoute | None |
| UI | No |

**Purpose**: Kubernetes-native policy engine. Runs as a ValidatingWebhookConfiguration and MutatingWebhookConfiguration to enforce policies at admission time. Policies are defined as `ClusterPolicy` CRDs. Used for enforcing image signing, resource limits, and namespace conventions.

---

#### Trivy Operator

| Property | Value |
|----------|-------|
| File | `platform/cdk8s/cots/compliance/trivy.go` |
| Namespace | `trivy` |
| HTTPRoute | None |
| UI | No |

**Purpose**: Continuous vulnerability scanning. Watches all running workloads and generates `VulnerabilityReport` and `ConfigAuditReport` CRDs. Reports queryable via `kubectl get vulnerabilityreports -A`.

---

#### Falco

| Property | Value |
|----------|-------|
| File | `platform/cdk8s/cots/compliance/falco.go` |
| Namespace | `falco` |
| Helm chart | `falco` v4.8.0 (falcosecurity.github.io/charts) |
| HTTPRoute | None |
| UI | No |

**Purpose**: Runtime security. Monitors kernel syscalls on every node using eBPF (`modern_ebpf` driver — required on Talos because Talos locks down kernel module loading). Detects anomalies such as shell spawned in a container, unexpected network connections, and privilege escalation. Outputs JSON alerts to stdout, collected by OTel agent and forwarded to VictoriaLogs.

**Key Talos note**: Must use `driver.kind: modern_ebpf`. The `kmod` and `legacy_ebpf` drivers require kernel module loading which Talos does not permit.

---

### management/

#### Rancher

| Property | Value |
|----------|-------|
| File | `platform/cdk8s/cots/management/rancher.go` |
| Namespace | `cattle-system` |
| HTTPRoute | `rancher.madhan.app` → `rancher:80` |
| UI | Yes |

**Purpose**: Multi-cluster Kubernetes management UI. Provides fleet-level visibility into workloads, nodes, and events.

---

#### Headlamp

| Property | Value |
|----------|-------|
| File | `platform/cdk8s/cots/management/headlamp.go` |
| Namespace | `headlamp` |
| HTTPRoute | `headlamp.madhan.app` → `headlamp:80` |
| UI | Yes |

**Purpose**: Lightweight Kubernetes dashboard. A `headlamp-admin` ServiceAccount with `cluster-admin` binding is created by CDK8s; the token is stored in a Secret (`headlamp-admin-token`) and used to authenticate in the UI.

---

#### Fleet

| Property | Value |
|----------|-------|
| File | `platform/cdk8s/cots/management/fleet_device_manager.go` |
| Namespace | `fleet` |
| HTTPRoute | None |
| UI | No |

**Purpose**: Rancher Fleet GitOps agent. Manages multi-cluster GitOps deployments at scale.

---

### monitoring/

See `docs/observability.md` for the full observability stack architecture.

#### Grafana

| Property | Value |
|----------|-------|
| File | `platform/cdk8s/cots/monitoring/grafana.go` |
| Namespace | `grafana` |
| HTTPRoute | `grafana.madhan.app` → `grafana:80` |
| UI | Yes |

**Purpose**: Visualization and dashboarding. Datasources for VictoriaMetrics (Prometheus-compatible) and VictoriaLogs (Loki-compatible) are auto-provisioned via Helm values.

---

#### VictoriaMetrics

| Property | Value |
|----------|-------|
| File | `platform/cdk8s/cots/monitoring/victoria_metrics.go` |
| Namespace | `victoria-metrics` |
| HTTPRoute | None |
| UI | No |

**Purpose**: Horizontally scalable time-series metrics storage. Deploys as a cluster (vminsert, vmselect, vmstorage). Prometheus remote-write compatible for receiving metrics from OTel collector.

---

#### VictoriaLogs

| Property | Value |
|----------|-------|
| File | `platform/cdk8s/cots/monitoring/victoria_logs.go` |
| Namespace | `victoria-logs` |
| HTTPRoute | None |
| UI | No |

**Purpose**: Log storage with Loki-compatible query API. Receives logs from OTel collector via OTLP/HTTP. Queryable from Grafana using the Loki datasource plugin.

---

#### AlertManager

| Property | Value |
|----------|-------|
| File | `platform/cdk8s/cots/monitoring/alert_manager.go` |
| Namespace | `alertmanager` |
| HTTPRoute | None |
| UI | No |

**Purpose**: Alert routing, grouping, and deduplication. Deployed via `kube-prometheus-stack`.

---

#### OTel Collector

| Property | Value |
|----------|-------|
| File | `platform/cdk8s/cots/monitoring/otel_collector.go` |
| Namespace | `opentelemetry` |
| HTTPRoute | None |
| UI | No |

**Purpose**: Two-tier OpenTelemetry collection pipeline:
- **Agent (DaemonSet)** on every node: collects container logs, kubelet metrics, and host metrics.
- **Gateway (Deployment)**: collects cluster-level resource metrics and Kubernetes events.

Both export to VictoriaMetrics (metrics) and VictoriaLogs (logs).

---

### registry/

#### Harbor

| Property | Value |
|----------|-------|
| File | `platform/cdk8s/cots/registry/harbor.go` |
| Namespace | `harbor` |
| HTTPRoute | `harbor.madhan.app` → `harbor-core:80` |
| UI | Yes |

**Purpose**: Enterprise container image registry with vulnerability scanning, RBAC, and replication. Acts as a pull-through proxy cache for Docker Hub and other registries.

---

### security/

#### Infisical

| Property | Value |
|----------|-------|
| File | `platform/cdk8s/cots/security/infisical.go` |
| Namespace | `infisical` |
| HTTPRoute | `infisical.madhan.app` → Infisical service |
| UI | Yes |

**Purpose**: Central secrets management platform. All application runtime secrets are stored in Infisical and injected into pods via `InfisicalSecret` CRDs and the Infisical operator. Bootstrap secrets (Infisical DB password, Cloudflare API token) are handled separately via SOPS — see `docs/secrets-management.md`.

---

### storage/

#### Longhorn

| Property | Value |
|----------|-------|
| File | `platform/cdk8s/cots/storage/longhorn.go` |
| Namespace | `longhorn-system` |
| HTTPRoute | None |
| UI | No (internal Longhorn UI exists but not exposed via HTTPRoute) |

**Purpose**: Distributed block storage for Kubernetes. Provides the `longhorn` StorageClass used by all PVCs in the cluster. Runs on all 4 worker nodes (`k8s-worker1` through `k8s-worker4`).

**Key Talos configuration**:
- `createDefaultDiskLabeledNodes: false` — provisions default disk on all nodes (not just labeled ones).
- Pre-upgrade checker disabled (`preUpgradeChecker.jobEnabled: false`) for GitOps compatibility.
- Namespace has `pod-security.kubernetes.io/enforce: privileged` for the Longhorn DaemonSet.
