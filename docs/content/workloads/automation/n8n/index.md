+++
title = "n8n"
description = "Open-source workflow automation platform with PostgreSQL backend and OpenBao secrets."
weight = 10
+++

## What is n8n?

[n8n](https://n8n.io/) is an open-source workflow automation platform. It provides a visual node-based editor for building integrations between 400+ services and supports scheduled, webhook, and event-driven workflows. Unlike hosted solutions (Zapier, Make), n8n is fully self-hosted with no per-execution pricing.

## Why n8n?

n8n is the leading self-hosted automation platform with native Kubernetes support (the 8gears Helm chart). Its open-source nature means there are no rate limits or execution caps, making it suitable for high-frequency workflows in a homelab.

## How It's Used Here

n8n is deployed using the [8gears Helm chart](https://github.com/8gears/n8n-helm-chart) (OCI chart `oci://8gears.container-registry.com/library/n8n`, v2.0.1). It uses CloudNativePG for PostgreSQL instead of the embedded Bitnami PostgreSQL subchart.

Source: [`workloads/automation/n8n.go`](https://github.com/madhank93/homelab/blob/v0.1.5/workloads/automation/n8n.go)

## Configuration

| Setting | Value | Why |
|---------|-------|-----|
| Namespace | `n8n` | Isolated namespace |
| HTTPRoute | `n8n.madhan.app` → `n8n:80` | Gateway API |
| n8n image tag | `1.78.0` | Pinned — never use `latest` |
| Chart | 8gears OCI v2.0.1 | Native Kubernetes features (`extraVolumes`, etc.) |
| Persistence | `10Gi` RWX Longhorn | Multi-attach safe |
| Database | CloudNativePG PostgreSQL | CNPG manages credential lifecycle |
| DB host | `n8n-pg-rw` | CNPG read-write service |
| `postgresql.enabled` | `false` | Bitnami subchart disabled |

## Database

n8n uses a [CloudNativePG](https://cloudnative-pg.io/) `Cluster` CR for PostgreSQL:

```yaml
instances: 1
storage:
  size: 10Gi
  storageClass: longhorn
bootstrap:
  initdb:
    database: n8n
    owner: n8n
```

CNPG auto-creates:
- Secret `n8n-pg-app` with `username` and `password` for the `n8n` app user
- Service `n8n-pg-rw` for the read-write endpoint (primary)

The DB password is **not in OpenBao** — CNPG owns the credential lifecycle and auto-rotates it. n8n reads it via:

```go
"DB_POSTGRESDB_PASSWORD": {
    "valueFrom": {"secretKeyRef": {"name": "n8n-pg-app", "key": "password"}},
}
```

## Secrets (OpenBao)

Pattern B (secretObjects sync). `ENCRYPTION_KEY` is fetched from OpenBao (`secret/data/n8n`) and synced into the `n8n-secrets` k8s Secret.

```go
// workloads/automation/n8n.go
"N8N_ENCRYPTION_KEY": {
    "valueFrom": {"secretKeyRef": {"name": "n8n-secrets", "key": "N8N_ENCRYPTION_KEY"}},
}
```

> **Warning: Never rotate `N8N_ENCRYPTION_KEY` after workflows are created.** n8n uses this key to encrypt stored credentials (API tokens, passwords). If the key changes, existing credentials become unreadable and must be re-entered manually. Store the key permanently in OpenBao and treat it as immutable.

## How It Connects

```
Browser → n8n.madhan.app
  → homelab-gateway → n8n:80 → n8n pod:5678
  → PostgreSQL via n8n-pg-rw:5432 (CNPG)
  → OpenBao CSI → n8n-secrets k8s Secret → ENCRYPTION_KEY env var
  → External APIs (via workflow nodes)
```

## Screenshots

![n8n workflow editor showing node connections and execution history](/assets/screenshots/n8n/workflow-editor.png)

## Troubleshooting

### Credentials Cannot Be Decrypted

**Symptoms:** n8n shows `Credentials could not be decrypted. Please update or delete them`.

**Why:** The `N8N_ENCRYPTION_KEY` in the running pod does not match the key used when credentials were first stored.

**Fix:** There is no automated fix. Options:
1. Restore the original encryption key in OpenBao (`bao kv patch secret/n8n ENCRYPTION_KEY=<original>`)
2. Delete all affected credentials in the n8n UI and re-enter them

### Database Connection Failed

```bash
# Check CNPG cluster status
kubectl get cluster n8n-pg -n n8n

# Check CNPG pod is running
kubectl get pods -n n8n -l cnpg.io/cluster=n8n-pg

# Check DB secret exists
kubectl get secret n8n-pg-app -n n8n
```

### Deployment Selector Immutable

If migrating from the community n8n chart to the 8gears chart, the Deployment's `spec.selector` changes. Kubernetes does not allow modifying selectors after creation.

**Fix:** Delete the old Deployment before ArgoCD syncs:

```bash
kubectl delete deployment n8n -n n8n
# Then trigger ArgoCD sync
argocd app sync n8n
```
