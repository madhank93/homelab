+++
title = "OpenBao"
description = "Runtime secrets store — Vault-compatible fork (MPL-2.0) with Kubernetes auth and CSI integration."
weight = 10
+++

## What is OpenBao?

[OpenBao](https://openbao.org/) is an open-source, community-maintained fork of HashiCorp Vault (MPL-2.0). It provides secrets management with a fully Vault-compatible API — all `vault` CLI knowledge and tooling applies directly using the `bao` CLI.

## Why OpenBao?

OpenBao's Kubernetes auth method allows pods to authenticate using their ServiceAccount token without any static credentials — the cluster itself is the identity provider. The Secrets Store CSI Driver integration mounts secrets directly into pods as files at startup, so no secret values appear in manifests, environment variables, or `kubectl get secret` output.

## First-Time Setup

Initial bootstrap (init, unseal, K8s auth configuration, writing app secrets) is covered in **[Deployment Guide — Phase 2](/getting-started/deployment/#phase-2-openbao-init-k8s-auth-write-app-secrets)**.

## How It's Used Here

OpenBao runs as a standalone (single-node) server in the `openbao` namespace. Pods authenticate using the Kubernetes auth method — they present their ServiceAccount JWT token, OpenBao verifies it with the Kubernetes tokenreviews API, and returns a short-lived token.

**Kubernetes auth roles:**

| App | SA | Namespace | Policy |
|-----|-----|-----------|--------|
| Grafana | `grafana` (default) | `grafana` | `grafana-policy` |
| Harbor | `secret-sync` | `harbor` | `harbor-policy` |
| n8n | `n8n` (default) | `n8n` | `n8n-policy` |
| NetBird | `default` | `netbird` | `netbird-policy` |
| Falco | `secret-sync` | `falco` | `falco-policy` |

**Secret paths (KV v2):**

| App | Path | Keys |
|-----|------|------|
| Grafana | `secret/data/grafana` | `ADMIN_PASSWORD`, `OAUTH_CLIENT_SECRET` |
| Harbor | `secret/data/harbor` | `HARBOR_ADMIN_PASSWORD` |
| n8n | `secret/data/n8n` | `ENCRYPTION_KEY` |
| NetBird | `secret/data/netbird` | `NETBIRD_SETUP_KEY` |
| Falco | `secret/data/falco` | `FALCOSIDEKICK_UI_USER` |

Source: {{ src(path="workloads/secrets/openbao.go") }}

## Secrets Patterns

Apps consume OpenBao secrets as mounted files, optionally synced to a k8s Secret.
Both patterns are explained once in [Secrets](@/platform/secrets/index.md) — the
short version is that the CSI volume mount is what triggers the sync, so a
SecretProviderClass with no pod mounting it produces nothing.

## Configuration

| Setting | Value | Why |
|---------|-------|-----|
| Helm chart | `openbao` | Version in the [Software Inventory](@/architecture/software-inventory.md) |
| Storage | `10Gi` Longhorn | Persistent secrets storage |
| Storage backend | `file` | Simple, no Consul dependency |
| CSI provider | enabled | Bridges OpenBao → CSI driver |
| Injector | disabled | CSI-only approach |
| Metrics | `unauthenticated_metrics_access: true` | VMAgent scrapes `/v1/sys/metrics` |
| Retention | `prometheus_retention_time: 30s` | Prometheus metrics TTL |
| Audit | `audit "file" "stdout"` | Every request logged to the pod's stdout; declared in config because API audit creation is off by default |
| Unix listener | `/tmp/openbao.sock`, generate-root enabled | Unauthenticated generate-root is off by default on every listener; only this in-pod socket allows it, for `just openbao-token` |

## Auto-Unseal Sidecar

OpenBao starts in a sealed state after every pod restart. An `unseal` sidecar container (part of the Helm values `extraContainers`) polls every 15 seconds and unseals with the key from the `openbao-unseal-key` Secret:

```go
// workloads/secrets/openbao.go
"extraContainers": []map[string]any{
    {
        "name":  "unseal",
        "image": "openbao/openbao:<version>",   // matches the server image
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

Only the sidecar gets the key; the server container has no reference to it.
The `openbao-unseal-key` Secret is created by `just create-secrets` from `secrets/bootstrap.sops.yaml`. It carries `argocd.argoproj.io/sync-options: Prune=false` so Argo CD never deletes it.

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

No root token is stored anywhere. `just openbao-token` mints a short-lived one
from the unseal key in the `openbao-unseal-key` Secret (the `generate-root`
ceremony over the in-pod socket). It refuses to print to a terminal: capture it,
pass it to `bao` on stdin so it never shows in a process list, and revoke it.

### Read a secret (recommended)

```bash
just openbao-get secret/grafana
just openbao-get secret/grafana ADMIN_PASSWORD
```

### Change a value

```bash
export ROOT_TOKEN=$(just openbao-token)
printf '%s\n%s' "$ROOT_TOKEN" '<new value>' | kubectl exec -i -n openbao openbao-0 -c openbao -- \
  sh -c 'read -r BAO_TOKEN; export BAO_TOKEN; exec bao kv patch -mount=secret grafana ADMIN_PASSWORD=-'
just openbao-revoke
```

### Other bao commands

```bash
export ROOT_TOKEN=$(just openbao-token)
printf '%s\n' "$ROOT_TOKEN" | kubectl exec -i -n openbao openbao-0 -c openbao -- \
  sh -c 'read -r BAO_TOKEN; export BAO_TOKEN; exec bao "$@"' bao kv list secret/
printf '%s\n' "$ROOT_TOKEN" | kubectl exec -i -n openbao openbao-0 -c openbao -- \
  sh -c 'read -r BAO_TOKEN; export BAO_TOKEN; exec bao "$@"' bao read auth/kubernetes/role/grafana
just openbao-revoke
```

### Check OpenBao status

```bash
kubectl exec -n openbao openbao-0 -- bao status
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
kubectl get secret openbao-unseal-key -n openbao -o jsonpath='{.data.unseal-key}' | base64 -d |
  kubectl exec -i -n openbao openbao-0 -c openbao -- sh -c 'read -r k; bao operator unseal "$k"'
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
printf '%s\n' "$ROOT_TOKEN" | kubectl exec -i -n openbao openbao-0 -c openbao -- \
  sh -c 'read -r BAO_TOKEN; export BAO_TOKEN; exec bao "$@"' bao \
  write auth/kubernetes/role/<app> \
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

**Fix:** Ensure the `SecretProviderClass` exists in the same namespace as the pod. CDK8s should create it — check if Argo CD has synced the namespace.

