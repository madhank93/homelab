+++
title = "Infisical"
description = "Central secrets management platform for all runtime app secrets."
weight = 30
+++

## Overview

| Property | Value |
|----------|-------|
| CDK8s file | `workloads/secrets/infisical.go` |
| Namespace | `infisical` |
| HTTPRoute | `infisical.madhan.app` → Infisical service |
| UI | Yes |
| Bootstrap Secret | `infisical/infisical-secrets` (created by `just create-secrets`) |

## Purpose

Infisical is the central secrets management platform. All application runtime secrets are stored in Infisical projects and injected into pods via `InfisicalSecret` CRDs + the Infisical operator.

## Architecture

```
infisical-secrets (k8s Secret, created by bootstrap)
    │
    └── Infisical Helm chart (backend + frontend + operator + PostgreSQL)
            │
            └── InfisicalSecret CRs (one per app)
                    │
                    └── Kubernetes Secrets (grafana-admin, harbor-admin, etc.)
                            │
                            └── App pods consume secrets via envFrom or volume
```

## Backend

Infisical backend requires:
- `INFISICAL_DB_PASSWORD` — PostgreSQL password
- `INFISICAL_ENCRYPTION_KEY` — at-rest encryption key
- `INFISICAL_AUTH_SECRET` — JWT signing secret

These are provided via the `infisical-secrets` k8s Secret.

## PostgreSQL

The PostgreSQL backend uses `docker.io/library/postgres:17`. The bitnami PostgreSQL image was removed from Docker Hub; only the official image is used.

## Bootstrap Setup

```bash
# 1. Create bootstrap secrets (from SOPS-encrypted file)
just create-secrets

# 2. After ArgoCD deploys Infisical:
# - Create an org and project at http://infisical.madhan.app
# - Add secrets for each app under their paths (/grafana, /harbor, etc.)
# - Create a Service Token with read access

# 3. Create the service token Secret
kubectl create secret generic infisical-service-token \
  --from-literal=infisicalToken=<token> \
  -n infisical
```

See `docs/infisical-secrets-setup.md` for the complete setup guide.

## InfisicalSecret Pattern

All apps that use Infisical carry `ServerSideApply=false` on their InfisicalSecret resources because the Infisical CRD schema omits `projectSlug`, which breaks ArgoCD's SSA validation.
