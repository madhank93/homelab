---
title: Notebook gateway controller
description: Small in-cluster controller that routes Kubeflow notebooks, TensorBoards and PVC viewers through the Cilium Gateway.
sidebar:
  order: 40
---

A Python controller that watches Kubeflow `Notebook`, `Tensorboard` and `PVCViewer` resources and creates a Gateway API `HTTPRoute` for each. Upstream controllers create Istio VirtualServices, which do nothing here, and Gateway API has no dynamic path routing of its own.

| | |
|---|---|
| Namespace | `kubeflow` |
| Source | [`notebook_gateway_controller.go`](src:workloads/ai/notebook_gateway_controller.go), image [`Dockerfile`](src:images/notebook-gateway-controller/Dockerfile) |
| Exposure | Routes it creates live on `kubeflow.madhan.app` |
| Storage | — |
| Auth | Routes inject `kubeflow-userid: user@example.com` |

## How it works

On start it syncs every existing resource, then watches all three kinds. ADDED or MODIFIED creates or patches the route; DELETED removes it. Routes live in `kubeflow` and point at Services in user namespaces, so it also creates a `ReferenceGrant` named `allow-kubeflow-gateway` in each user namespace.

| Resource | Route name | Path | Backend | URL rewrite |
|---|---|---|---|---|
| Notebook | `notebook-<ns>-<name>` | `/notebook/<ns>/<name>/` | `<name>:80` | No — Jupyter serves under its own prefix |
| Tensorboard | `tensorboard-<ns>-<name>` | `/tensorboard/<ns>/<name>/` | `<name>:80` | Prefix → `/` |
| PVCViewer | `pvcviewer-<ns>-<name>` | `/pvcviewers/<ns>/<name>/` | `pvcviewer-<name>:80` | Prefix → `/` |

## Configuration

| Setting | Value | Why |
|---|---|---|
| Image | `harbor.madhan.app/library/notebook-gateway-controller` | Python plus the Kubernetes client, built from `images/notebook-gateway-controller/` |
| Script | ConfigMap `notebook-gateway-controller-script`, mounted at `/app` | Edit the script in Go; Reloader restarts the pod on change |
| Security context | non-root `65534`, all capabilities dropped | |

## Operate

Rebuild the image only when its dependencies change; script edits ship through the ConfigMap.

```bash frame="terminal"
just build-push notebook-gateway-controller v1
```

## Troubleshooting

### Notebook URL returns 404

The route is missing. Check the routes and the controller log:

```bash frame="terminal"
kubectl get httproute -n kubeflow | grep -E 'notebook|tensorboard|pvcviewer'
kubectl logs -n kubeflow deploy/notebook-gateway-controller --tail=30
```

### Route exists but backend is refused

The `ReferenceGrant` is missing in the user namespace: `kubectl get referencegrant -n <user-namespace>`. The controller creates it with the first resource it sees there.
