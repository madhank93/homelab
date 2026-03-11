+++
title = "n8n"
description = "Open-source workflow automation platform."
weight = 10
+++

## Overview

| Property | Value |
|----------|-------|
| CDK8s file | `workloads/automation/n8n.go` |
| Namespace | `n8n` |
| Helm chart | [8gears n8n](https://github.com/8gears/n8n-helm-chart) `oci://8gears.container-registry.com/library/n8n` v2.0.1 |
| HTTPRoute | `n8n.madhan.app` → `n8n:80` |
| UI | Yes |
| Database | PostgreSQL via CloudNativePG (`n8n-pg` Cluster) |

## Purpose

n8n is an open-source workflow automation platform. It integrates with 400+ services via nodes and supports scheduled, webhook, and event-driven workflows.

## Architecture

```
OpenBao (secret/data/n8n)
  └── ENCRYPTION_KEY
        └── CSI Driver → SecretProviderClass → n8n-secrets Secret
              └── n8n pod: N8N_ENCRYPTION_KEY env var

CloudNativePG Operator (cnpg-system)
  └── Cluster CR: n8n-pg (namespace: n8n)
        ├── Pod: n8n-pg-1 (PostgreSQL 17)
        ├── Service: n8n-pg-rw:5432 (read-write)
        └── Secret: n8n-pg-app (username/password — auto-managed by CNPG)
              └── n8n pod: DB_POSTGRESDB_PASSWORD env var
```

## Database

n8n uses a **CloudNativePG** single-instance PostgreSQL cluster (`n8n-pg`) deployed in the `n8n` namespace.

CNPG auto-creates:
- `n8n-pg-app` — k8s Secret with `username`, `password`, `uri`, etc.
- `n8n-pg-rw` — read-write ClusterIP Service (primary)
- `n8n-pg-ro` — read-only Service
- `n8n-pg-r` — round-robin Service

The DB password is **fully managed by CNPG** — it auto-generates a stable password that persists across operator restarts. No manual password management needed.

## Encryption Key

n8n encrypts stored credentials with `N8N_ENCRYPTION_KEY`. This key is stored in OpenBao at `secret/data/n8n` under the key `ENCRYPTION_KEY` and synced into the cluster via the Secrets Store CSI Driver.

The key **must remain stable** — changing it will break decryption of any stored credentials.

> If you see `Error: Credentials could not be decrypted`, check that the value at `secret/data/n8n` `ENCRYPTION_KEY` in OpenBao matches what was used when credentials were first stored.

## Secret Flow

```
OpenBao path: secret/data/n8n
  ENCRYPTION_KEY → SecretProviderClass (n8n-secrets) → k8s Secret n8n-secrets
                    └── N8N_ENCRYPTION_KEY key

n8n Deployment:
  extraEnv.N8N_ENCRYPTION_KEY  ← secretKeyRef: n8n-secrets / N8N_ENCRYPTION_KEY
  extraEnv.DB_POSTGRESDB_PASSWORD ← secretKeyRef: n8n-pg-app / password
  main.config.db.postgresdb.host = n8n-pg-rw
```

## Accessing the UI

Navigate to `http://n8n.madhan.app`. On first access (fresh database), create an owner account.

## Known Issue: RWO Deadlock (not applicable)

Unlike Harbor, n8n's 8gears chart uses `Recreate` deployment strategy by default, so RWO PVC multi-attach issues do not occur.
