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
| UI | Yes (`http://infisical.madhan.app`) |
| Bootstrap Secret | `infisical/infisical-secrets` (created by `just create-secrets`) |

## Purpose

Infisical is the central secrets management platform. All application runtime secrets are stored in Infisical projects and injected into pods via `InfisicalSecret` CRDs managed by the Infisical operator.

## Architecture

```
infisical-secrets (k8s Secret, created by bootstrap)
    │
    └── Infisical Helm chart (backend + frontend + PostgreSQL)
            │
            └── Infisical Operator (secrets-operator chart)
                    │
                    └── InfisicalSecret CRs (one per app, kubernetesAuth)
                            │
                            └── Kubernetes Secrets (grafana-admin, harbor-admin, etc.)
                                    │
                                    └── App pods consume secrets via envFrom or volume
```

## Authentication — Kubernetes Auth

All `InfisicalSecret` CRs use **Kubernetes Auth** — the operator authenticates with Infisical using its own ServiceAccount JWT. No credentials are stored in Kubernetes Secrets.

Machine Identity ID: `47aef6c1-bdeb-40fa-be46-63bbcfe6a4df`
ServiceAccount: `infisical-opera-controller-manager` (namespace: `infisical`)

The operator SA is bound to the `infisical-token-reviewer` ClusterRole (grants `tokenreviews:create`) so Infisical can verify the JWT via the K8s token review API.

## Bootstrap Setup

```bash
# 1. Create bootstrap secrets (from SOPS-encrypted file)
just create-secrets

# 2. After ArgoCD deploys Infisical, open http://infisical.madhan.app
#    - Complete org/project setup
#    - Access Control → Machine Identities → create identity
#    - Enable Kubernetes Auth (host: https://192.168.1.210:6443)
#    - Note the identity ID

# 3. Add secrets for each app under their paths:
#    /grafana, /harbor, /n8n, /rancher, /netbird, /infisical
```

## PostgreSQL

Uses `docker.io/library/postgres:17` (official image). The Bitnami image was removed from Docker Hub.
Password sourced from `infisical-secrets` k8s Secret via `secretKeyRef`.

## InfisicalSecret Pattern

All `InfisicalSecret` resources carry `argocd.argoproj.io/sync-options: ServerSideApply=false` because the Infisical CRD schema omits `projectSlug` from `serviceToken.secretsScope`, which breaks ArgoCD's SSA schema validation.

## Operator Bug Workarounds (v0.10.25)

The `secrets-operator` v0.10.25 chart hardcodes `subjects[0].namespace: default` in three bindings. Three override resources correct the namespace to `infisical`:

- `infisical-opera-leader-election-rolebinding` (RoleBinding)
- `infisical-opera-manager-rolebinding` (ClusterRoleBinding)
- `infisical-opera-metrics-auth-rolebinding` (ClusterRoleBinding)
