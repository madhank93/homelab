+++
title = "Harbor"
description = "Enterprise container image registry with vulnerability scanning and pull-through proxy."
weight = 10
+++

## What is Harbor?

[Harbor](https://goharbor.io/) is an open-source cloud-native container registry that provides role-based access control, vulnerability scanning (via Trivy), image signing, content trust, and pull-through proxy caching. It is CNCF Graduated status.

## Why Harbor?

Harbor is the most feature-complete self-hosted registry available. Alternatives like plain Docker Registry lack RBAC, vulnerability scanning, and a web UI. Harbor integrates directly with Trivy for on-push image scanning, making it useful alongside the cluster-wide Trivy Operator.

## How It's Used Here

Harbor stores private container images built for this homelab and provides pull-through proxy caches for Docker Hub, GHCR, and other public registries. Images are pushed to `harbor.madhan.app` and pulled into cluster deployments.

Source: [`workloads/registry/harbor.go`](https://github.com/madhank93/homelab/blob/v0.1.5/workloads/registry/harbor.go)

## Configuration

| Setting | Value | Why |
|---------|-------|-----|
| Namespace | `harbor` | Isolated namespace |
| HTTPRoute | `harbor.madhan.app` → `harbor:80` | Gateway API — points to nginx proxy, NOT harbor-core |
| Access | LAN / VPN only | Not in `publicServices` — DNS resolves to `192.168.1.220` |
| `externalURL` | `https://harbor.madhan.app` | Used in redirect URLs and push commands |
| `expose.type` | `clusterIP` | No NodePort/LoadBalancer — all via Gateway API |
| `expose.tls.enabled` | `false` | TLS terminated at Gateway |
| Registry PVC | `50Gi` RWX Longhorn | Multi-attach safe for rolling updates |
| Jobservice PVC | RWX Longhorn | Multi-attach safe for rolling updates |
| Database PVC | `10Gi` RWO Longhorn | Internal PostgreSQL (single writer) |
| `existingSecret` | `harbor-admin` | CSI-synced Secret from OpenBao |

## HTTPRoute Note

Harbor's HTTPRoute **must point to `harbor:80`** (the nginx proxy service), not `harbor-core:80` (the core API service directly). The nginx proxy handles routing between Harbor's multiple backend components (core, portal, registry, etc.). Pointing directly to `harbor-core` will cause the UI to fail to load assets and the registry protocol to break.

```go
// Correct:
"name": "harbor", "port": 80
// Wrong:
"name": "harbor-core", "port": 80
```

## Secrets (OpenBao)

Pattern B (secretObjects sync). The `HARBOR_ADMIN_PASSWORD` is fetched from OpenBao (`secret/data/harbor`) and synced into the `harbor-admin` k8s Secret by a dedicated `secret-sync` Deployment.

**Why a dedicated secret-sync Deployment?** Harbor's Helm chart does not support `extraVolumes` on its component pods. A `pause` container in a separate Deployment mounts the CSI volume, triggering the secretObjects sync. The Harbor chart then references `harbor-admin` via `existingSecret`.

```yaml
existingSecret: harbor-admin  # CSI-synced from OpenBao
```

## RWX PVCs

Harbor's registry and jobservice PVCs use `ReadWriteMany` via Longhorn's NFS share-manager. This eliminates the `Multi-Attach error` that occurs during rolling updates when both old and new pods need to mount the same volume simultaneously.

```go
"registry": {"accessMode": "ReadWriteMany"},
"jobservice": {"jobLog": {"accessMode": "ReadWriteMany"}},
```

## How It Connects

```
Docker push → harbor.madhan.app
  → Gateway API → harbor:80 (nginx)
  → harbor-core (API), harbor-portal (UI), harbor-registry (storage)
  → Registry PVC 50Gi RWX Longhorn
  → Trivy (vulnerability scan on push)

Secret sync:
  secret-sync pod → CSI volume → OpenBao → harbor-admin k8s Secret
  harbor-core → existingSecret: harbor-admin
```

## Screenshots

![Harbor UI showing projects, image repositories, and vulnerability scan results](/assets/screenshots/harbor/main-ui.png)

## Troubleshooting

### RWO Multi-Attach Deadlock (Legacy)

> This was the original issue before switching to RWX PVCs. Documented for reference.

**Symptoms:** Pod stuck with `Multi-Attach error for volume`.

**Fix:**

```bash
# Find the old ReplicaSet
kubectl get replicasets -n harbor

# Scale it down to 0
kubectl scale replicaset <old-rs-name> -n harbor --replicas=0

# Force-delete any stuck pod
kubectl delete pod -n harbor <stuck-pod> --grace-period=0 --force
```

### HTTPRoute Points to Wrong Service

**Symptoms:** Harbor UI loads a blank page, or registry push fails with 404.

**Fix:** Ensure the HTTPRoute backend is `harbor:80` (nginx proxy), not `harbor-core:80`.

```bash
kubectl get httproute harbor -n harbor -o yaml | grep -A5 backendRefs
# Should show: name: harbor, port: 80
```

### Admin Password Not Working

**Symptoms:** Can't log in as admin.

**Diagnosis:**

```bash
# Check the harbor-admin Secret is populated
kubectl get secret harbor-admin -n harbor -o yaml

# Check secret-sync pod is running and healthy
kubectl get pods -n harbor -l app=secret-sync
kubectl logs -n harbor -l app=secret-sync

# Verify OpenBao has the password
kubectl exec -n openbao openbao-0 -- bao kv get secret/harbor
```

### Configuring Pull-Through Proxy

1. Log in to `http://harbor.madhan.app` as admin
2. Go to **Administration → Registries → New Endpoint**
3. Add Docker Hub, GHCR, or other registries
4. Create a proxy project pointing to the registry endpoint
