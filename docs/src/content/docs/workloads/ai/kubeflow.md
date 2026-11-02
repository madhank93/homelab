---
title: Kubeflow
description: Notebooks, Pipelines, Katib and Trainer from the upstream manifests, without Istio, Knative, KServe or Dex.
sidebar:
  order: 30
  badge: { text: LAN, variant: note }
---

Kubeflow provides GPU notebooks, pipelines and hyperparameter tuning. It runs without Istio, so a Gateway API route and a small controller replace the mesh. [Upstream docs](https://www.kubeflow.org/docs/).

| | |
|---|---|
| Namespace | `kubeflow`, user namespace `kubeflow-user-example-com` |
| Source | [`kustomization.yaml`](src:workloads/ai/kubeflow/kustomization.yaml), [`kubeflow.go`](src:workloads/ai/kubeflow.go) |
| Exposure | LAN · `kubeflow.madhan.app` |
| Storage | Upstream PVCs for the bundled MySQL and SeaweedFS, on the default storage class |
| Auth | None — every route injects `kubeflow-userid: user@example.com` |

Why no mesh or serving layer: [Kubeflow without Istio](/reference/decisions/kubeflow/).

## How it fits

CI runs `kustomize build workloads/ai/kubeflow/` and commits the result to `app/kubeflow/` on the manifests branch, so Argo CD fetches no remote bases at sync time. The overlay pulls upstream components from `kubeflow/community-distribution` at the [pinned release](/reference/software-inventory/) and patches them:

| Patch | Why |
|---|---|
| Stub Istio CRDs ([`stub-istio-crds.yaml`](src:workloads/ai/kubeflow/stub-istio-crds.yaml)) | profiles-controller and others watch Istio kinds at startup and crash without the CRD |
| Delete every Istio resource | Pipeline and Katib overlays embed VirtualServices and AuthorizationPolicies even in platform-agnostic variants |
| Delete every upstream NetworkPolicy | They admit only Istio or same-namespace pods; cilium-envoy is hostNetwork, and the SeaweedFS policy blocks the pipelines profile controller |
| Rename the default `Profile` | A kustomize variable-substitution bug leaves it unnamed |

## Routing

The `kubeflow-dashboard` HTTPRoute maps each sub-app; sub-app rules precede the catch-all. Per-notebook routes come from the [notebook gateway controller](/workloads/ai/notebook-gateway-controller/).

| Path | Service | Rewrite prefix to `/` |
|---|---|---|
| `/jupyter/` | `jupyter-web-app-service` | Yes |
| `/volumes/` | `volumes-web-app-service` | Yes |
| `/tensorboards/` | `tensorboards-web-app-service` | Yes |
| `/pipeline/` | `ml-pipeline-ui` | Yes |
| `/katib/` | `katib-ui` | No — serves natively under `/katib/` |
| `/` | `dashboard` | No |

The Flask apps expect `/` (Istio's VirtualService used to rewrite it), hence the rewrite.

## GPU notebooks

A GPU notebook needs all four settings:

```yaml title="Notebook pod spec"
runtimeClassName: nvidia          # without it CUDA is unreachable even with a GPU request
nodeSelector:
  nvidia.com/gpu.present: "true"
tolerations:
- {key: dedicated, operator: Equal, value: ai, effect: NoSchedule}
containers:
- resources:
    limits:
      nvidia.com/gpu: "1"
```

The RTX 5070 Ti is Blackwell (`sm_120`). The `kubeflownotebookswg/jupyter-pytorch-cuda-full` images ship PyTorch without `sm_120` kernels, so install the nightly `cu128` wheel in the notebook and restart the kernel:

```bash title="Notebook terminal"
pip install --pre torch torchvision --index-url https://download.pytorch.org/whl/nightly/cu128 --force-reinstall
```

Check with `torch.randn(8, device="cuda")`; it must not raise.

## Operate

Add a user with a `Profile` resource for their email. The injected `kubeflow-userid` header is fixed, so real multi-user access needs an OIDC proxy in front.

## Troubleshooting

### Sub-app shows "not a valid page"

Its rule is missing from the route, or a Flask app lacks the rewrite: `kubectl get httproute kubeflow-dashboard -n kubeflow -o yaml`.

### CUDA errors in a GPU notebook

| Error | Cause | Fix |
|---|---|---|
| `Found no NVIDIA driver on your system` | Pod lacks `runtimeClassName: nvidia` | Add it to the notebook spec |
| `no kernel image is available for execution on the device` | PyTorch build lacks `sm_120` | Install the nightly `cu128` wheel above |

### Pipeline run stuck at `Init:0/1`

The `mlpipeline-minio-artifact` Secret is missing from the user namespace. The pipelines profile controller creates it through SeaweedFS's IAM port 8111; a NetworkPolicy that blocks that port stops it.

```bash frame="terminal"
kubectl get secret mlpipeline-minio-artifact -n kubeflow-user-example-com
kubectl get networkpolicy -n kubeflow          # expect none
kubectl logs deploy/kubeflow-pipelines-profile-controller -n kubeflow --tail=20
```

### PVC viewer stuck in `ContainerCreating`

Notebook volumes are RWO. While the notebook runs, the viewer pod on another node cannot attach the volume. Stop the notebook, then click **Connect** again.
