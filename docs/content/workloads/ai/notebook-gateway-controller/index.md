+++
title = "Notebook Gateway Controller"
description = "Custom Python controller that creates HTTPRoutes and ReferenceGrants for Kubeflow Notebook, Tensorboard, and PVCViewer CRs."
weight = 40
+++

## What is the Notebook Gateway Controller?

The Notebook Gateway Controller is a small custom controller (Python, in-cluster) that watches Kubeflow `Notebook`, `Tensorboard`, and `PVCViewer` custom resources and creates the corresponding `HTTPRoute` and `ReferenceGrant` objects so each notebook URL is routable through the Cilium Gateway.

## Why Is This Needed?

Kubeflow's notebook controller is built for Istio — it creates Istio `VirtualService` resources per notebook, which are no-ops without Istio installed. The Gateway API has no built-in mechanism for dynamic path routing as notebooks come and go. This controller bridges the gap: it watches the Kubeflow CRs and manages `HTTPRoute` and `ReferenceGrant` objects in the `kubeflow` namespace so the Cilium Gateway can route traffic correctly.

**Without this controller:**

- Notebook URLs (`kubeflow.madhan.app/notebook/<ns>/<name>/`) return 404
- Tensorboard and PVCViewer UIs are unreachable through the Gateway

## How It's Used Here

The controller runs as a single-replica Deployment in the `kubeflow` namespace. The Python script is stored in a ConfigMap and mounted at `/app/controller.py`. On startup it installs the `kubernetes` Python client, then:

1. **Syncs all existing CRs** — creates HTTPRoutes for every Notebook, Tensorboard, and PVCViewer that already exists
2. **Watches for changes** — three background threads watch each CR type; on ADDED/MODIFIED it creates or patches the HTTPRoute; on DELETED it removes it

**Resource naming:**

| CR type | HTTPRoute name |
|---------|---------------|
| Notebook | `notebook-<namespace>-<name>` |
| Tensorboard | `tensorboard-<namespace>-<name>` |
| PVCViewer | `pvcviewer-<namespace>-<name>` |

**Cross-namespace routing:**

HTTPRoutes live in the `kubeflow` namespace but backend Services live in user namespaces (e.g. `kubeflow-user-example-com`). Gateway API requires a `ReferenceGrant` in the target namespace. The controller creates one `ReferenceGrant` named `allow-kubeflow-gateway` per user namespace the first time a CR is found there.

**URL rewrite:**

Tensorboard and PVCViewer use `URLRewrite` (strip the prefix to `/`) since those apps don't handle path prefixes. Notebooks do not use URL rewrite.

Source: [`workloads/ai/notebook_gateway_controller.go`](https://github.com/madhank93/homelab/blob/v0.1.5/workloads/ai/notebook_gateway_controller.go)

## Configuration

| Setting | Value |
|---------|-------|
| Namespace | `kubeflow` |
| Image | `python:3.11-slim` |
| Gateway | `homelab-gateway` in `kube-system` |
| Hostname | `kubeflow.madhan.app` |
| Replicas | 1 |

## Troubleshooting

### Notebook URL Returns 404

```bash
# Check HTTPRoutes exist
kubectl get httproute -n kubeflow | grep notebook

# Check controller logs
kubectl logs -n kubeflow -l app=notebook-gateway-controller
```

### ReferenceGrant Missing

```bash
kubectl get referencegrant -n <user-namespace>
```

If missing, the controller hasn't seen a CR in that namespace yet. Check that the Notebook CR exists:

```bash
kubectl get notebooks -A
```
