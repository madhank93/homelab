+++
title = "OpenBao"
description = "Runtime secrets store — Vault-compatible fork (MPL-2.0) with Kubernetes auth and CSI integration."
weight = 10
+++

## What is OpenBao?

[OpenBao](https://openbao.org/) is an open-source, community-maintained fork of HashiCorp Vault (MPL-2.0). It provides secrets management with a fully Vault-compatible API — all `vault` CLI knowledge and tooling applies directly using the `bao` CLI.

## Why OpenBao?

OpenBao's Kubernetes auth method allows pods to authenticate using their ServiceAccount token without any static credentials — the cluster itself is the identity provider. The Secrets Store CSI Driver integration mounts secrets directly into pods as files at startup, so no secret values appear in manifests, environment variables, or `kubectl get secret` output.

## How It's Used Here

OpenBao runs as a standalone (single-node) server in the `openbao` namespace. Pods authenticate using the Kubernetes auth method — they present their ServiceAccount JWT token, OpenBao verifies it with the Kubernetes tokenreviews API, and returns a short-lived token.

**Kubernetes auth roles:**

| App | SA | Namespace | Policy |
|-----|-----|-----------|--------|
| Grafana | `grafana` (default) | `grafana` | `grafana-policy` |
| Harbor | `secret-sync` | `harbor` | `harbor-policy` |
| n8n | `n8n` (default) | `n8n` | `n8n-policy` |
| Rancher | `secret-sync` | `cattle-system` | `rancher-policy` |
| NetBird | `netbird-peer` | `netbird` | `netbird-policy` |

**Secret paths (KV v2):**

| App | Path | Keys |
|-----|------|------|
| Grafana | `secret/data/grafana` | `ADMIN_PASSWORD`, `OAUTH_CLIENT_SECRET` |
| Harbor | `secret/data/harbor` | `HARBOR_ADMIN_PASSWORD` |
| n8n | `secret/data/n8n` | `ENCRYPTION_KEY` |
| Rancher | `secret/data/rancher` | `BOOTSTRAP_PASSWORD` |
| NetBird | `secret/data/netbird` | `NETBIRD_SETUP_KEY` |

Source: [`workloads/secrets/openbao.go`](https://github.com/madhank93/homelab/blob/v0.1.5/workloads/secrets/openbao.go)

## Secrets Patterns

### Pattern A — File-only (Grafana)

Secret is mounted as a file at `/mnt/secrets/ADMIN_PASSWORD`. The app reads it via an env var pointing to the file path:

```yaml
env:
  GF_SECURITY_ADMIN_PASSWORD__FILE: /mnt/secrets/ADMIN_PASSWORD
```

No k8s Secret is created. The secret value never appears in `kubectl get secret` output.

### Pattern B — secretObjects sync (Harbor, n8n, Rancher, NetBird)

The CSI volume mount triggers the SecretProviderClass `secretObjects` block, which creates a k8s Secret in the app's namespace. Required for Helm charts that only accept `existingSecret` references.

> The CSI volume mount is **required** to trigger the sync — if no pod mounts the volume, the k8s Secret is never created.

For Harbor and Rancher (whose Helm charts do not support `extraVolumes`), a dedicated `secret-sync` Deployment with a `pause` container mounts the CSI volume just to trigger the secretObjects sync.

## Configuration

| Setting | Value | Why |
|---------|-------|-----|
| Helm chart | `openbao` v0.25.6 | Pinned version |
| Storage | `10Gi` Longhorn | Persistent secrets storage |
| Storage backend | `file` | Simple, no Consul dependency |
| CSI provider | enabled | Bridges OpenBao → CSI driver |
| Injector | disabled | CSI-only approach |
| Metrics | `unauthenticated_metrics_access: true` | VMAgent scrapes `/v1/sys/metrics` |
| Retention | `prometheus_retention_time: 30s` | Prometheus metrics TTL |

## Auto-Unseal Sidecar

OpenBao starts in a sealed state after every pod restart. An `unseal` sidecar container (part of the Helm values `extraContainers`) polls every 15 seconds and unseals with the key from the `openbao-unseal-key` Secret:

```go
// workloads/secrets/openbao.go
"extraContainers": []map[string]any{
    {
        "name":  "unseal",
        "image": "openbao/openbao:2.5.1",
        "command": []string{"sh", "-c", `
while true; do
  STATUS=$(bao status -format=json 2>/dev/null || echo '{"sealed":true}')
  if echo "$STATUS" | grep -q '"sealed":true'; then
    bao operator unseal "$OPENBAO_UNSEAL_KEY" 2>/dev/null || true
  fi
  sleep 15
done`},
    },
}
```

> **Why `extraContainers` and not `extraInitContainers`?** Init containers must complete before the main container starts, but OpenBao must be running before it can accept an unseal request. A sidecar container runs alongside the main container and can poll until the server is ready.

The `openbao-unseal-key` Secret is created by `just create-secrets` from `secrets/bootstrap.sops.yaml`. It carries `argocd.argoproj.io/sync-options: Prune=false` so ArgoCD never deletes it.

## How It Connects

```
Pod starts
  → CSI volume mount → OpenBao CSI provider DaemonSet
  → OpenBao CSI provider: Kubernetes auth with pod SA token
  → OpenBao: verifies SA token via Kubernetes tokenreviews API
  → OpenBao: returns secret value(s)
  → CSI provider: writes secrets as files to pod filesystem
  → (Pattern B) CSI driver: creates/updates k8s Secret via secretObjects
```

## Screenshots

![OpenBao UI showing KV v2 secrets and Kubernetes auth configuration](/assets/screenshots/openbao/main-ui.png)

## Commands

All `bao` commands run inside the `openbao-0` pod. Export `ROOT_TOKEN` first:

```bash
ROOT_TOKEN=$(kubectl get secret openbao-unseal-key -n openbao \
  -o jsonpath='{.data.root-token}' | base64 -d)
```

### Read a secret

```bash
kubectl exec -n openbao openbao-0 -- env BAO_TOKEN=$ROOT_TOKEN \
  bao kv get secret/grafana
```

### Read a single key

```bash
kubectl exec -n openbao openbao-0 -- env BAO_TOKEN=$ROOT_TOKEN \
  bao kv get -field=ADMIN_PASSWORD secret/grafana
```

### Update / add a key (non-destructive patch)

```bash
kubectl exec -n openbao openbao-0 -- env BAO_TOKEN=$ROOT_TOKEN \
  bao kv patch secret/grafana ADMIN_PASSWORD=newvalue
```

### Replace all keys in a secret (full put)

```bash
kubectl exec -n openbao openbao-0 -- env BAO_TOKEN=$ROOT_TOKEN \
  bao kv put secret/grafana ADMIN_PASSWORD=value OAUTH_CLIENT_SECRET=value
```

### List all secrets in a path

```bash
kubectl exec -n openbao openbao-0 -- env BAO_TOKEN=$ROOT_TOKEN \
  bao kv list secret/
```

### Check OpenBao status

```bash
kubectl exec -n openbao openbao-0 -- bao status
```

### Check auth roles

```bash
kubectl exec -n openbao openbao-0 -- env BAO_TOKEN=$ROOT_TOKEN \
  bao read auth/kubernetes/role/grafana
```

---

## Troubleshooting

### Sealed After Restart

**Symptoms:** Pods that depend on OpenBao secrets fail to start; OpenBao pod shows `sealed=true`.

**Diagnosis:**

```bash
kubectl exec -n openbao openbao-0 -- bao status
```

**Fix:** The unseal sidecar should handle this automatically. If the sidecar is not working:

```bash
# Check sidecar logs
kubectl logs -n openbao openbao-0 -c unseal

# Manual unseal (get key from bootstrap secret)
UNSEAL_KEY=$(kubectl get secret openbao-unseal-key -n openbao -o jsonpath='{.data.unseal-key}' | base64 -d)
kubectl exec -n openbao openbao-0 -- bao operator unseal "$UNSEAL_KEY"
```

### 403 Permission Denied

**Symptoms:** CSI mount fails with `permission denied` or `403 Forbidden`.

**Diagnosis:**

```bash
# Check the OpenBao CSI provider logs on the affected node
kubectl logs -n openbao -l app.kubernetes.io/name=openbao-csi-provider

# Test the auth role manually
kubectl exec -n openbao openbao-0 -- bao read auth/kubernetes/role/<app-name>
```

**Fix:** The Kubernetes auth role may not include the pod's ServiceAccount. Run `just openbao-setup` to re-apply roles, or manually add the SA:

```bash
kubectl exec -n openbao openbao-0 -- env BAO_TOKEN=$ROOT_TOKEN \
  bao write auth/kubernetes/role/<app> \
    bound_service_account_names=<sa-name> \
    bound_service_account_namespaces=<namespace> \
    policies=<policy-name> \
    ttl=1h
```

### CSI Mount Failing

**Symptoms:** Pod stuck in `ContainerCreating` with `MountVolume.SetUp failed`.

**Diagnosis:**

```bash
kubectl describe pod <pod-name> -n <namespace>
# Look for: failed to get secretproviderclass

kubectl get secretproviderclass -n <namespace>
kubectl describe secretproviderclass <name> -n <namespace>
```

**Fix:** Ensure the `SecretProviderClass` exists in the same namespace as the pod. CDK8s should create it — check if ArgoCD has synced the namespace.

