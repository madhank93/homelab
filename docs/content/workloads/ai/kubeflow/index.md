+++
title = "Kubeflow"
description = "ML platform on Kubernetes: Notebooks, Pipelines, Katib, Training Operator — without Istio or KServe."
weight = 30
+++

## What is Kubeflow?

[Kubeflow](https://www.kubeflow.org/) is an open-source ML platform for Kubernetes. It provides:

- **Notebooks** — JupyterLab environments with GPU access, lifecycle management, and per-user namespaces
- **Pipelines** — DAG-based ML workflow orchestration (KFP v2)
- **Katib** — Automated hyperparameter tuning and neural architecture search
- **Training Operator** — Distributed training jobs (PyTorchJob, TFJob, MPIJob)
- **Tensorboard** — Training metric visualisation
- **Volumes** — PVC management UI

## Why No Istio / Knative / KServe?

Kubeflow's default installation bundles Istio (service mesh), Knative (serverless), Dex (OIDC), and KServe (model serving). This cluster does not use any of them:

| Component | Status | Reason |
|-----------|--------|--------|
| Istio | Excluded | Cilium already provides network policy and L7 observability |
| Knative | Excluded | No serverless workloads; adds unnecessary complexity |
| KServe | Excluded | Not needed for current workloads |
| Dex / oauth2-proxy | Excluded | Identity handled by Authentik at the Bifrost VPS edge |

Without Istio, several upstream Kubeflow controllers still watch Istio CRDs at startup (they call it unconditionally). Stub CRDs are installed to prevent crashes:

```
workloads/ai/kubeflow/stub-istio-crds.yaml
  ├── AuthorizationPolicy  (security.istio.io/v1beta1)   — profiles-controller
  ├── VirtualService       (networking.istio.io/v1alpha3) — notebook-controller
  └── DestinationRule      (networking.istio.io/v1alpha3) — metacontroller / KFP profile controller
```

These stubs contain only the minimum schema needed for the watch to register. No actual Istio resources exist — the controllers watch empty lists.

## How It's Used Here

Kubeflow v1.11.0 is deployed via a kustomize overlay pointing at the upstream manifests repository. ArgoCD manages it from the `v0.1.5` branch.

```
workloads/ai/kubeflow/kustomization.yaml
  ├── remote bases → github.com/kubeflow/manifests v1.11.0 (no Istio overlay)
  ├── stub-istio-crds.yaml
  └── patches:
      ├── Profile name fix (kustomize 5.8.0 variable substitution bug)
      ├── Delete centraldashboard NetworkPolicy (blocks Cilium-envoy hostNetwork traffic)
      ├── Delete seaweedfs NetworkPolicy (blocks profile controller IAM port 8111)
      └── Delete all Istio resources (DestinationRule, VirtualService, AuthorizationPolicy)
```

Source: [`workloads/ai/kubeflow/kustomization.yaml`](https://github.com/madhank93/homelab/blob/v0.1.5/workloads/ai/kubeflow/kustomization.yaml)

Components installed:

| Component | Path |
|-----------|------|
| cert-manager Kubeflow issuer | `common/cert-manager/kubeflow-issuer/base` |
| Kubeflow namespace + roles | `common/kubeflow-namespace/base` + `kubeflow-roles/base` |
| Kubeflow Pipelines | `applications/pipeline/upstream/env/cert-manager/platform-agnostic-multi-user` |
| Katib | `applications/katib/upstream/installs/katib-with-kubeflow` |
| Central Dashboard | `applications/centraldashboard/upstream/base` |
| Admission Webhook (PodDefaults) | `applications/admission-webhook/upstream/overlays/cert-manager` |
| Jupyter Web App | `applications/jupyter/jupyter-web-app/upstream/base` |
| Notebook Controller | `applications/jupyter/notebook-controller/upstream/overlays/kubeflow` |
| Profiles + KFAM | `applications/profiles/pss` |
| PVC Viewer Controller | `applications/pvcviewer-controller/upstream/base` |
| Volumes Web App | `applications/volumes-web-app/upstream/base` |
| Tensorboard Controller | `applications/tensorboard/tensorboard-controller/upstream/overlays/kubeflow` |
| Tensorboards Web App | `applications/tensorboard/tensorboards-web-app/upstream/base` |
| Training Operator (Trainer v2) | `applications/trainer/overlays` |
| Spark Operator | `applications/spark/spark-operator/overlays/kubeflow` |
| User namespace | `common/user-namespace/base` |

## Access

All Kubeflow sub-apps are exposed at `https://kubeflow.madhan.app` via the Cilium Gateway. A static `kubeflow-dashboard` HTTPRoute routes path prefixes to each service. No authentication — the gateway injects `kubeflow-userid: user@example.com` header on all requests.

| Path | Service | URL rewrite |
|------|---------|-------------|
| `/` | `centraldashboard:80` | None (catch-all) |
| `/jupyter/` | `jupyter-web-app-service:80` | `ReplacePrefixMatch: /` |
| `/volumes/` | `volumes-web-app-service:80` | `ReplacePrefixMatch: /` |
| `/tensorboards/` | `tensorboards-web-app-service:80` | `ReplacePrefixMatch: /` |
| `/pipeline/` | `ml-pipeline-ui:80` | `ReplacePrefixMatch: /` |
| `/katib/` | `katib-ui:80` | **None** — katib-ui is a Go binary that serves at its path natively |

> **Why URL rewrite?** Flask-based web apps (jupyter, volumes, tensorboards, pipeline UI) serve their HTML at `/` and generate relative links. The gateway strips the path prefix so the Flask app receives `/` and everything works. The `katib-ui` Go binary is aware of its path prefix and serves correctly without rewriting.

Source: [`app/kubeflow/HTTPRoute.kubeflow-dashboard.k8s.yaml`](https://github.com/madhank93/homelab/blob/v0.1.5/app/kubeflow/HTTPRoute.kubeflow-dashboard.k8s.yaml)

## Notebook and Tensorboard Routing

Notebook and Tensorboard CRs each get their own HTTPRoute dynamically. The upstream controllers create Istio VirtualServices per CR — without Istio those are no-ops. A custom controller watches the CRs and creates Gateway API HTTPRoutes instead.

Source: [`workloads/ai/notebook_gateway_controller.go`](https://github.com/madhank93/homelab/blob/v0.1.5/workloads/ai/notebook_gateway_controller.go)

The controller runs as a Python deployment in the `kubeflow` namespace and watches three CRD types:

| CR type | URL pattern | URL rewrite | Notes |
|---------|-------------|-------------|-------|
| `notebooks.kubeflow.org/v1` | `/notebook/<ns>/<name>/` | **None** | notebook-controller sets `--NotebookApp.base_url` — Jupyter serves at its own prefix |
| `tensorboards.tensorboard.kubeflow.org/v1alpha1` | `/tensorboard/<ns>/<name>/` | `ReplacePrefixMatch: /` | TensorBoard starts without `--path_prefix`; serves at `/` |
| `pvcviewers.kubeflow.org/v1alpha1` | `/pvcviewers/<ns>/<name>/` | `ReplacePrefixMatch: /` | filebrowser uses `FB_BASEURL` for link generation but serves at `/` |

For each CR in a user namespace, the controller also creates a `ReferenceGrant` (once per namespace) allowing the `kubeflow` namespace HTTPRoute to reference services in the user namespace.

### How the Controller Works

```
notebook-gateway-controller (Python, kubeflow ns)
  sync_all() on startup → patch all existing CRs
  watch_notebooks()     → main thread
  watch_tensorboards()  → daemon thread
  watch_pvcviewers()    → daemon thread

On ADDED/MODIFIED event:
  1. apply_refgrant(user_ns)  → ensure ReferenceGrant exists
  2. apply_httproute(...)      → create/patch HTTPRoute in kubeflow ns

On DELETED event:
  delete_httproute(...)        → remove HTTPRoute from kubeflow ns
```

## GPU Notebooks

### Required Pod Spec

GPU notebooks on this cluster require three specific settings:

```yaml
spec:
  template:
    spec:
      runtimeClassName: nvidia          # REQUIRED — without this the nvidia-container-runtime
                                        # hook never fires and CUDA is inaccessible even with
                                        # nvidia.com/gpu: 1 in the resource limits
      nodeSelector:
        nvidia.com/gpu.present: "true"  # Routes to k8s-worker4 (the GPU node)
                                        # NOTE: worker4 has no dedicated=ai label or taint;
                                        # this GFD label is the correct selector
      containers:
      - resources:
          requests:
            nvidia.com/gpu: "1"         # Requests one time-sliced virtual GPU
          limits:
            nvidia.com/gpu: "1"
```

> **Why `runtimeClassName: nvidia`?** On Talos, NVIDIA libraries live inside a squashfs at `/usr/local/glibc/usr/lib/` — not at standard paths. The `nvidia-container-runtime` hook (configured by the `nvidia-container-toolkit-production` Talos extension) knows the Talos-specific paths and injects them. Without `runtimeClassName: nvidia`, the default `runc` runtime runs and no injection happens. `torch.cuda.is_available()` returns `True` (the device is visible via `NVIDIA_VISIBLE_DEVICES`), but any CUDA kernel call fails with `no kernel image is available for execution on the device`.

### PyTorch and sm_120 (Blackwell)

The RTX 5070 Ti is **NVIDIA Blackwell architecture, compute capability sm_120**. Standard PyTorch wheels do not include sm_120 kernels until PyTorch 2.7.0 (May 2025). The `kubeflownotebookswg/jupyter-pytorch-cuda-full:v1.10.0-rc.1` image ships PyTorch compiled for sm_50–sm_90 only.

**Symptom:** `CUDA available: True` and `GPU: NVIDIA GeForce RTX 5070 Ti` but any tensor operation on the GPU raises:
```
RuntimeError: CUDA error: no kernel image is available for execution on the device
```

**Fix — run once per notebook session, then restart the kernel:**
```python
import subprocess, sys
subprocess.run(
    [sys.executable, "-m", "pip", "install", "--pre",
     "torch", "torchvision",
     "--index-url", "https://download.pytorch.org/whl/nightly/cu128",
     "--force-reinstall", "--quiet"],
    check=True
)
print("Restart kernel now.")
```

**Verify after restart:**
```python
import torch
print(torch.__version__)        # e.g. 2.12.0.dev20260316+cu128
print(torch.cuda.is_available())   # True
print(torch.cuda.get_device_name(0))  # NVIDIA GeForce RTX 5070 Ti
x = torch.randn(64, 784, device="cuda")  # must not raise
print("sm_120 works!")
```

### Recommended Notebook Image

| Image | CUDA | PyTorch | sm_120 | Notes |
|-------|------|---------|--------|-------|
| `kubeflownotebookswg/jupyter-pytorch-cuda-full:v1.10.0-rc.1` | 12.x | 2.5.x | No — install nightly | Best Kubeflow-native image; handles `NB_PREFIX`, JupyterLab startup |
| `kubeflownotebookswg/jupyter-pytorch-cuda-full:v1.9.0` | 12.x | 2.3.x | No | Older; works the same way |
| `nvcr.io/nvidia/pytorch:25.xx-py3` | 12.8 | 2.6+ | Yes | Does **not** handle `NB_PREFIX` — requires custom startup wrapper to work as a Kubeflow notebook |

### GPU Time-Slicing

The RTX 5070 Ti is configured with 5 time-sliced virtual GPUs. VRAM (16 GB) is **not isolated** between processes — all virtual GPU holders share the same physical VRAM pool.

| Workload | Virtual GPUs | Typical VRAM |
|----------|-------------|--------------|
| Ollama (7B model) | 1 | ~4 GB |
| ComfyUI (SDXL) | 1 | ~6 GB |
| Kubeflow notebook | 1 | variable |
| Training job | 1 | variable |
| Katib trial | 1 | variable |

Running Flux.1 (~12 GB) + Ollama simultaneously will OOM. The user is responsible for managing concurrent VRAM usage.

## Pipelines (KFP v2)

Pipelines use SeaweedFS (bundled S3-compatible object store) and MySQL (bundled). Artifact credentials are provisioned per user namespace by the `kubeflow-pipelines-profile-controller` via metacontroller.

### mlpipeline-minio-artifact Secret

This secret is automatically created in every user namespace (e.g. `kubeflow-user-example-com`) by the profile controller. It contains S3 credentials for writing pipeline artifacts to SeaweedFS. Pipeline pods mount it as a volume — if it is missing, all pipeline runs fail with:

```
MountVolume.SetUp failed: secret "mlpipeline-minio-artifact" not found
```

If the secret is missing, check:

```bash
# Is the profile controller responding?
kubectl logs deployment/kubeflow-pipelines-profile-controller -n kubeflow --tail=20

# Is metacontroller retrying?
kubectl logs metacontroller-0 -n kubeflow --tail=20 | grep error
```

**Known cause in this cluster:** The upstream SeaweedFS NetworkPolicy (installed by the pipeline manifests) only allows port 8333 (S3 API) to SeaweedFS, not port 8111 (IAM/STS API). The profile controller uses `AWS_ENDPOINT_URL=http://seaweedfs.kubeflow:8111` which routes all boto3 calls (including S3 bucket creation) through port 8111. Without access to 8111, the profile controller times out and never creates the secret. The fix is applied via a kustomize patch that deletes the seaweedfs NetworkPolicy.

## PVC Viewer

The PVC viewer (filebrowser) is launched by the `pvcviewer-controller` when a `PVCViewer` CR exists. The controller is triggered from the Volumes Web App when a user clicks **Connect** on a volume.

### RWO Limitation

Longhorn RWO (ReadWriteOnce) volumes can only be attached to one node at a time. If the PVC is in use by a running notebook (attached to its node), the PVC viewer pod will be scheduled on a **different** node and get stuck in `ContainerCreating` indefinitely:

```
AttachVolume: waiting for volumes to attach/detach  (PVC already attached to another node)
```

**Workaround:** Stop the notebook first (releases the PVC from its node), then click Connect in the Volumes page. The PVC viewer pod can then attach and mount the PVC.

**Future fix:** Use an RWX (ReadWriteMany) storage class for notebook PVCs. Longhorn supports RWX via NFS. RWX PVCs can be mounted on multiple nodes simultaneously, allowing the PVC viewer and notebook to coexist.

## Default User

The default user is `user@example.com`. The user namespace is `kubeflow-user-example-com`.

Authentication is not configured — the gateway injects `kubeflow-userid: user@example.com` on every request via a `RequestHeaderModifier` filter. The central dashboard and all sub-apps receive this header and treat the user as authenticated.

To create additional users: add a `Profile` CR pointing to the new user's email, and update the `kubeflow-userid` header value (or deploy a proper OIDC proxy).

## How It Connects

```
Browser → kubeflow.madhan.app
  → homelab-gateway (Cilium, kube-system)
  → HTTPRoute: kubeflow-dashboard (static, kubeflow ns)
      ├── /jupyter/      → jupyter-web-app-service (URLRewrite /)
      ├── /volumes/      → volumes-web-app-service (URLRewrite /)
      ├── /tensorboards/ → tensorboards-web-app-service (URLRewrite /)
      ├── /katib/        → katib-ui (no rewrite)
      ├── /pipeline/     → ml-pipeline-ui (URLRewrite /)
      └── /              → centraldashboard
  → HTTPRoute: notebook-<ns>-<name> (dynamic, kubeflow ns)
      → /notebook/<ns>/<name>/ → <name>.<ns>:80 (no rewrite)
  → HTTPRoute: tensorboard-<ns>-<name> (dynamic, kubeflow ns)
      → /tensorboard/<ns>/<name>/ → <name>.<ns>:80 (URLRewrite /)
  → HTTPRoute: pvcviewer-<ns>-<name> (dynamic, kubeflow ns)
      → /pvcviewers/<ns>/<name>/ → pvcviewer-<name>.<ns>:80 (URLRewrite /)

notebook-gateway-controller (Python, kubeflow ns)
  → watches Notebooks, Tensorboards, PVCViewers
  → creates/deletes HTTPRoutes + ReferenceGrants per CR

kubeflow-pipelines-profile-controller
  → called by metacontroller on Namespace events
  → creates S3 bucket in SeaweedFS (port 8111 IAM API)
  → creates mlpipeline-minio-artifact Secret in user ns
  → creates kfp-launcher, artifact-repositories, metadata-grpc-configmap ConfigMaps

Notebook pod (GPU) on k8s-worker4:
  → runtimeClassName: nvidia
  → nvidia-container-runtime injects libnvidia-ml.so.1
  → RTX 5070 Ti (sm_120, CUDA 12.8)
  → PyTorch nightly cu128 required (stable wheels don't include sm_120 kernels)
```

## Troubleshooting

### Sub-app page shows "not a valid page"

The path prefix rules in the HTTPRoute are missing or wrong.

```bash
kubectl get httproute kubeflow-dashboard -n kubeflow -o yaml
```

Each sub-app must have an explicit `PathPrefix` match in the `kubeflow-dashboard` HTTPRoute. Flask apps need `URLRewrite ReplacePrefixMatch: /`; Go binaries (katib-ui) do not.

### Notebook URL 404

The `notebook-gateway-controller` creates the HTTPRoute. Check it is running and healthy:

```bash
kubectl get pods -n kubeflow | grep notebook-gateway
kubectl logs deployment/notebook-gateway-controller -n kubeflow --tail=30
```

If the controller is crashlooping, check logs for Python errors — the ConfigMap version should bump on each deploy (Reloader restarts the pod on ConfigMap change).

### Tensorboard URL 404

TensorBoard serves at `/` — it is started without `--path_prefix`. The HTTPRoute must have `URLRewrite ReplacePrefixMatch: /` to strip the prefix.

```bash
# Confirm URLRewrite is present
kubectl get httproute tensorboard-<ns>-<name> -n kubeflow \
  -o jsonpath='{.spec.rules[0].filters}' | python3 -m json.tool
```

### CUDA Not Working in GPU Notebook

**Symptom 1:** `Found no NVIDIA driver on your system`

Pod is running without `runtimeClassName: nvidia`. The nvidia-container-runtime hook never fires.

```bash
kubectl get pod <notebook>-0 -n kubeflow-user-example-com \
  -o jsonpath='{.spec.runtimeClassName}'
# Must return: nvidia
```

**Symptom 2:** `no kernel image is available for execution on the device`

PyTorch was compiled without sm_120 (Blackwell) kernels. Install nightly:

```python
import subprocess, sys
subprocess.run([sys.executable, "-m", "pip", "install", "--pre", "torch", "torchvision",
    "--index-url", "https://download.pytorch.org/whl/nightly/cu128",
    "--force-reinstall", "--quiet"], check=True)
# Then: Kernel → Restart Kernel
```

### Pipeline Run Stuck at Init:0/1

The init container mounts `mlpipeline-minio-artifact` secret. If the secret does not exist, the pod never starts.

```bash
kubectl get secret mlpipeline-minio-artifact -n kubeflow-user-example-com
```

If missing, check the profile controller logs for SeaweedFS connection errors. The seaweedfs NetworkPolicy should be absent (deleted by the kustomize patch):

```bash
kubectl get networkpolicy seaweedfs -n kubeflow
# Should return: Error from server (NotFound)
```

### Controller CrashLoop After ConfigMap Update

The Stakater Reloader annotation (`reloader.stakater.com/auto: "true"`) on the controller Deployment causes a rolling restart when the ConfigMap changes. If the restart loops, check the Python script for syntax errors in the controller logs.
