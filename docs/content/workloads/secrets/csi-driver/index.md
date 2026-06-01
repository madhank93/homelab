+++
title = "Secrets Store CSI Driver"
description = "Mounts OpenBao secrets as files into pods without creating k8s Secrets."
weight = 20
+++

## What is the Secrets Store CSI Driver?

The [Secrets Store CSI Driver](https://secrets-store-csi-driver.sigs.k8s.io/) is a Kubernetes CSI (Container Storage Interface) plugin that mounts secrets from external secret stores directly into pods as files. It supports multiple backends via a provider model — this homelab uses the OpenBao (Vault-compatible) provider.

## Why the CSI Driver?

Kubernetes Secrets have a fundamental security limitation: they are stored as base64 in etcd and accessible to anyone with sufficient RBAC access. The CSI driver keeps secrets in OpenBao and only fetches them when a pod needs them.

| Approach | Secret lifecycle | Secret in etcd |
|----------|-----------------|----------------|
| k8s Secrets | Created manually, risk of git commit | Yes (base64) |
| envFrom Secret | Created manually | Yes (base64) |
| **CSI Driver** | Fetched at pod mount time, never stored | **No** |

## How It's Used Here

The CSI driver is deployed as a DaemonSet in `kube-system` (one pod per node). It intercepts CSI volume mounts and calls the OpenBao provider to fetch secrets at pod startup.

Source: [`workloads/secrets/csi_driver.go`](https://github.com/madhank93/homelab/blob/v0.1.5/workloads/secrets/csi_driver.go)

## Configuration

| Setting | Value | Why |
|---------|-------|-----|
| Helm chart | `secrets-store-csi-driver` v1.5.6 | Pinned version |
| Namespace | `kube-system` | Must be cluster-wide |
| `syncSecret.enabled` | `true` | Required for Pattern B (create k8s Secrets from secretObjects) |
| `enableSecretRotation` | `true` | Poll for updated secrets |
| `rotationPollInterval` | `2m` | Check every 2 minutes |
| Node tolerations | `operator: Exists` | Run on ALL nodes including control plane |

## How SecretProviderClass Works

Each app defines a `SecretProviderClass` that specifies:

1. Which provider to use (`openbao`)
2. The OpenBao address and role name
3. Which secrets to fetch (objectName, secretPath, secretKey)
4. Optional: `secretObjects` to sync fetched secrets into a k8s Secret

Example from `workloads/registry/harbor.go`:

```yaml
apiVersion: secrets-store.csi.x-k8s.io/v1
kind: SecretProviderClass
metadata:
  name: harbor-secrets
  namespace: harbor
spec:
  provider: openbao
  parameters:
    vaultAddress: http://openbao.openbao.svc.cluster.local:8200
    roleName: harbor
    objects: |
      - objectName: "HARBOR_ADMIN_PASSWORD"
        secretPath: "secret/data/harbor"
        secretKey: "HARBOR_ADMIN_PASSWORD"
  secretObjects:
    - secretName: harbor-admin
      type: Opaque
      data:
        - objectName: HARBOR_ADMIN_PASSWORD
          key: HARBOR_ADMIN_PASSWORD
```

When a pod mounts the CSI volume referencing `harbor-secrets`:
1. The CSI driver calls the OpenBao provider
2. The provider authenticates with the pod's SA token via Kubernetes auth
3. OpenBao returns the secret value
4. The driver writes the secret as a file into the pod at the specified mount path
5. The driver also creates/updates the `harbor-admin` k8s Secret (because `syncSecret.enabled=true`)

## Pattern A vs Pattern B

| Pattern | Secret as file | k8s Secret created | Used by |
|---------|---------------|-------------------|---------|
| A (file-only) | Yes | No | Grafana |
| B (secretObjects) | Yes | Yes | Harbor, n8n, Rancher, NetBird |

Pattern B is needed for Helm charts that only accept `existingSecret` references and cannot read secrets from file paths.

> **The CSI volume mount is required to trigger secretObjects sync.** If no pod mounts the volume, the k8s Secret is never created or updated.

## secret-sync Deployments

For Harbor and Rancher, whose Helm charts do not support `extraVolumes` on their component pods, a dedicated `secret-sync` Deployment runs a `pause` container whose only purpose is to mount the CSI volume and trigger secretObjects sync:

```go
// workloads/registry/harbor.go
{
    Name:  "pause",
    Image: "registry.k8s.io/pause:3.10",
    VolumeMounts: [{
        Name:      "openbao-secrets",
        MountPath: "/mnt/secrets",
        ReadOnly:  true,
    }],
}
```

The `pause` container uses essentially zero resources (a few KB of memory) — it just keeps the CSI volume mounted so the k8s Secret stays in sync.

## How It Connects

```
OpenBao (openbao namespace)
  ← OpenBao CSI Provider (DaemonSet, openbao namespace)
      ← Secrets Store CSI Driver (DaemonSet, kube-system)
          ← Pod with CSI volume mount (any namespace)
              → SecretProviderClass (same namespace as pod)
              → Files in pod at /mnt/secrets/
              → k8s Secret (Pattern B only, same namespace)
```

## Troubleshooting

### MountVolume.SetUp Failed

**Symptoms:** Pod stuck in `ContainerCreating`:

```
MountVolume.SetUp failed for volume "openbao-secrets":
  rpc error: code = Unknown desc = failed to get secretproviderclass
```

**Diagnosis:**

```bash
kubectl describe pod <pod-name> -n <namespace>
kubectl get secretproviderclass -n <namespace>
```

**Fix:** The `SecretProviderClass` must exist in the same namespace as the pod. Check ArgoCD sync status.

### Secret Not Updating After Change

**Symptoms:** App still using old secret value after updating OpenBao.

**Diagnosis:**

```bash
# Check rotation is running
kubectl logs -n kube-system -l app=secrets-store-csi-driver | grep rotation

# Check sync status
kubectl get secretproviderclasspodstatus -n <namespace>
```

**Fix:** With `enableSecretRotation: true` and `rotationPollInterval: 2m`, updated secrets are fetched within 2 minutes. The pod does NOT restart automatically — use Reloader (`reloader.stakater.com/auto: "true"`) on the app's Deployment to trigger a restart when the k8s Secret changes.

### Permission Denied from OpenBao

```bash
# Check OpenBao provider logs on the node
kubectl logs -n openbao -l app.kubernetes.io/name=openbao-csi-provider --all-containers

# Verify the Kubernetes auth role
kubectl exec -n openbao openbao-0 -- bao read auth/kubernetes/role/<app-name>
```
