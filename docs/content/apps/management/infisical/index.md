+++
title = "Infisical"
description = "Central secrets management platform. Operator authenticates via Kubernetes Auth — zero stored credentials."
weight = 30
+++

## Overview

| Property | Value |
|----------|-------|
| CDK8s file | `workloads/secrets/infisical.go` |
| Namespace | `infisical` |
| HTTPRoute | `infisical.madhan.app` |
| Operator version | `secrets-operator` v0.10.25 |
| App chart | `infisical-standalone` v1.7.2 |
| Auth method | **Kubernetes Auth (Option C)** — zero stored credentials |
| Bootstrap Secret | `infisical/infisical-secrets` (from `just create-secrets`) |

## Purpose

Infisical is the central runtime secrets platform. All app secrets (Grafana, Harbor, n8n, Rancher, NetBird) are stored in Infisical projects and injected into Kubernetes pods via `InfisicalSecret` CRs managed by the Infisical operator.

CDK8s generates zero `Secret` manifests — secrets never touch git or the manifests branch.

## Architecture

```
Bootstrap (SOPS/age)
  └── infisical-secrets (k8s Secret)
        └── Infisical Server (PostgreSQL backend, Redis)
              │
              └── Infisical Operator (secrets-operator)
                    │  ← authenticates via Kubernetes JWT (Option C)
                    └── InfisicalSecret CR → infisical-synced-secrets
                          └── App pods (envFrom / volumeMount)
```

## Kubernetes Auth (Option C)

The operator authenticates to Infisical using its own ServiceAccount JWT — **no service tokens, no stored credentials anywhere**.

### How it works

```
Operator reconcile loop
  → mounts its own SA token (/var/run/secrets/...)
  → POST /api/v1/auth/kubernetes-auth/login { jwt: <SA token> }
  → Infisical calls k8s tokenreviews API to verify the JWT is legitimate
  → Infisical returns a short-lived access token
  → Operator fetches secrets using that token (auto-rotates every 60s)
```

### What CDK8s provisions

| Resource | Name | Purpose |
|----------|------|---------|
| `ClusterRole` | `infisical-token-reviewer` | Grants `create` on `tokenreviews` |
| `ClusterRoleBinding` | `infisical-token-reviewer` | Binds role to operator SA |
| `InfisicalSecret` | `infisical-bootstrap-secret` | Drives secret sync with kubernetesAuth |

### One-time setup (must do after first deploy)

After ArgoCD sync applies the RBAC resources, register the cluster in the Infisical UI:

**1. Get the token reviewer JWT:**
```bash
kubectl create token infisical-operator-controller-manager \
  -n infisical --duration=8760h
```

**2. Infisical UI → Access Control → Machine Identities → Create → "k8s-homelab":**

| Field | Value |
|-------|-------|
| Method | Kubernetes Auth |
| Kubernetes Host | `https://192.168.1.210:6443` |
| Token Reviewer JWT | *(paste from step 1)* |
| Allowed SA Names | `infisical-operator-controller-manager` |
| Allowed Namespaces | `infisical` |

**3. Copy the `identityId`** and replace the placeholder in `workloads/secrets/infisical.go`:

```go
"identityId": "REPLACE_WITH_IDENTITY_ID",
//              ↑ paste the UUID from the Infisical UI here
```

Then re-synthesize and push:
```bash
just synth
git add workloads/secrets/infisical.go app/infisical/
git commit -m "feat: set Infisical kubernetesAuth identityId"
```

### Verification

```bash
# CR created
kubectl get infisicalsecret -n infisical

# Operator authenticated successfully (look for "Successfully authenticated" log)
kubectl logs -n infisical -l app.kubernetes.io/name=secrets-operator --tail=50

# Synced secret created by operator
kubectl get secret infisical-synced-secrets -n infisical

# CR status conditions
kubectl describe infisicalsecret infisical-bootstrap-secret -n infisical
```

## Backend Components

### Infisical Server

Deployed via `infisical-standalone` v1.7.2. Requires three secrets from `infisical-secrets`:

| Key | Purpose |
|-----|---------|
| `DB_PASSWORD` | PostgreSQL password |
| `ENCRYPTION_KEY` | At-rest secret encryption (AES-256) |
| `AUTH_SECRET` | JWT signing secret |

### PostgreSQL

StatefulSet using `docker.io/library/postgres:17` on Longhorn PVC (10 Gi). The Bitnami PostgreSQL image is **not used** — it was removed from Docker Hub.

### Redis

In-cluster Redis (from the `infisical-standalone` chart). Password is the chart default; it's only reachable inside the `infisical` namespace.

## Apps That Use Infisical

| App | Infisical Path | k8s Secret synced |
|-----|---------------|-------------------|
| Grafana | `/grafana` | `grafana-admin` (`ADMIN_PASSWORD`) |
| Harbor | `/harbor` | `harbor-admin` (`HARBOR_ADMIN_PASSWORD`) |
| n8n | `/n8n` | `n8n-db` (`DB_PASSWORD`, `N8N_ENCRYPTION_KEY`) |
| Rancher | `/rancher` | `rancher-bootstrap` (`BOOTSTRAP_PASSWORD`) |
| NetBird | `/netbird` | `netbird-setup-key` (`NETBIRD_SETUP_KEY`) |

## ArgoCD Note

All `InfisicalSecret` resources use:

```yaml
annotations:
  argocd.argoproj.io/sync-options: ServerSideApply=false
```

The Infisical CRD schema omits `projectSlug` from `serviceToken.secretsScope`, which breaks ArgoCD's SSA diff engine. Disabling SSA for these resources is the correct workaround.
