+++
title = "Secrets"
description = "SOPS/age for bootstrap secrets. Infisical + Kubernetes Auth for runtime secrets — zero credentials stored on-cluster."
weight = 50
+++

## Architecture

Secrets management uses a two-tier approach:

1. **Bootstrap secrets** — encrypted with SOPS/age, committed safely to git, applied once with `just create-secrets`
2. **Runtime secrets** — stored in Infisical, synced to pods via `InfisicalSecret` CRs using **Kubernetes Auth** (operator JWT, no stored tokens)

{% mermaid() %}
flowchart TD
    SOPS["secrets/bootstrap.sops.yaml<br/>Encrypted with age — safe in git"]
    LAPTOP["just create-secrets<br/>Decrypts via sops exec-env"]
    PULUMI["just pulumi talos up<br/>Decrypts via sops exec-env"]

    subgraph Bootstrap ["Bootstrap Secrets (created once)"]
        IS["infisical/infisical-secrets<br/>DB_PASSWORD, AUTH_SECRET, ENCRYPTION_KEY<br/>Prune=false"]
        CF["cert-manager/cloudflare-api-token<br/>CLOUDFLARE_API_TOKEN<br/>Prune=false"]
    end

    subgraph K8sAuth ["Kubernetes Auth (Option C)"]
        SA["infisical-operator-controller-manager<br/>ServiceAccount JWT"]
        CR["ClusterRole: infisical-token-reviewer<br/>tokenreviews create"]
        CRB["ClusterRoleBinding"]
    end

    subgraph Runtime ["Runtime Secrets (Infisical — kubernetesAuth)"]
        INFISICAL["Infisical Platform<br/>infisical.madhan.app"]
        ISR["InfisicalSecret CR<br/>infisical-bootstrap-secret"]
        APPSEC["Synced k8s Secrets<br/>grafana-admin, harbor-admin<br/>n8n-db, rancher-bootstrap"]
    end

    SOPS --> LAPTOP
    SOPS --> PULUMI
    LAPTOP --> Bootstrap
    Bootstrap --> INFISICAL
    SA --> CR
    CR --> CRB
    CRB --> SA
    SA -- "JWT login" --> INFISICAL
    INFISICAL -- "tokenreviews verify" --> SA
    INFISICAL --> ISR
    ISR --> APPSEC
{% end %}

## Bootstrap Secrets

Only two Secrets are created by the bootstrap script and never managed by ArgoCD:

| Secret | Namespace | Keys | Created by |
|--------|-----------|------|------------|
| `infisical-secrets` | `infisical` | DB_PASSWORD, AUTH_SECRET, ENCRYPTION_KEY, DB_CONNECTION_URI, REDIS_PASSWORD | `just create-secrets` |
| `cloudflare-api-token` | `cert-manager` | CLOUDFLARE_API_TOKEN | `just create-secrets` |

Both carry `argocd.argoproj.io/sync-options: Prune=false` so ArgoCD never deletes them.

## SOPS + age Setup

### First-Time Setup

```bash
# 1. Install tools
brew install age sops

# 2. Generate age key pair (back up this file!)
mkdir -p ~/.config/sops/age
age-keygen -o ~/.config/sops/age/keys.txt
# Output: Public key: age1abc123...

# 3. Add to shell profile (REQUIRED — sops 3.12+ does not auto-discover)
echo 'export SOPS_AGE_KEY_FILE="$HOME/.config/sops/age/keys.txt"' >> ~/.zshrc
source ~/.zshrc

# 4. Register public key in .sops.yaml at repo root
# creation_rules:
#   - path_regex: secrets/.*\.sops$
#     age: age1abc123...

# 5. Populate the bootstrap secrets file
sops secrets/bootstrap.sops.yaml   # opens $EDITOR, re-encrypts on save
```

### Day-to-Day Commands

```bash
# Edit encrypted file in-place
sops secrets/bootstrap.sops.yaml

# Run Pulumi (secrets injected as env vars, never written to disk)
just pulumi talos up

# Create/update bootstrap k8s Secrets
just create-secrets
```

### SOPS exec-env Syntax

```bash
# Correct — command as a single quoted string
sops exec-env file.sops.yaml 'bash script.sh'

# Wrong — double-dash not supported by sops
sops exec-env file.sops.yaml -- bash script.sh
```

## Runtime Secrets — Kubernetes Auth (Option C)

All application runtime secrets are stored in Infisical and synced to Kubernetes via `InfisicalSecret` CRs. The operator authenticates to Infisical using its own **ServiceAccount JWT** — no service tokens or passwords are stored anywhere on-cluster.

### Auth Flow

```
Operator reconcile loop
  → mounts its own SA token (/var/run/secrets/...)
  → POST /api/v1/auth/kubernetes-auth/login { jwt: <SA token> }
  → Infisical verifies JWT via k8s tokenreviews API
  → Infisical returns a short-lived access token
  → Operator fetches/syncs secrets using that token (resyncInterval: 60s)
```

### RBAC Resources (CDK8s-managed)

| Resource | Name | What it does |
|----------|------|-------------|
| `ClusterRole` | `infisical-token-reviewer` | Grants `create` on `authentication.k8s.io/tokenreviews` |
| `ClusterRoleBinding` | `infisical-token-reviewer` | Binds that role to the operator's ServiceAccount |
| `InfisicalSecret` | `infisical-bootstrap-secret` | Drives sync with `kubernetesAuth` spec |

### InfisicalSecret CR Pattern

```yaml
apiVersion: secrets.infisical.com/v1alpha1
kind: InfisicalSecret
metadata:
  name: infisical-bootstrap-secret
  namespace: infisical
  annotations:
    argocd.argoproj.io/sync-options: ServerSideApply=false
spec:
  hostAPI: https://infisical.madhan.app/api
  resyncInterval: 60
  authentication:
    kubernetesAuth:
      identityId: "<machine-identity-id>"   # from Infisical UI
      serviceAccountRef:
        name: infisical-operator-controller-manager
        namespace: infisical
  managedSecretReference:
    secretName: infisical-synced-secrets
    secretNamespace: infisical
    creationPolicy: "Orphan"
```

> `ServerSideApply=false` is required on all `InfisicalSecret` resources because the Infisical CRD schema omits `projectSlug`, which breaks ArgoCD's SSA diff engine.

### One-Time Infisical UI Setup

See the [Infisical app page](/apps/management/infisical) for the full step-by-step setup to register the cluster and obtain the `identityId`.

### Apps That Use Infisical

| App | Path | k8s Secret | Keys |
|-----|------|------------|------|
| Grafana | `/grafana` | `grafana-admin` | `ADMIN_PASSWORD` |
| Harbor | `/harbor` | `harbor-admin` | `HARBOR_ADMIN_PASSWORD` |
| n8n | `/n8n` | `n8n-db` | `DB_PASSWORD`, `N8N_ENCRYPTION_KEY` |
| Rancher | `/rancher` | `rancher-bootstrap` | `BOOTSTRAP_PASSWORD` |
| NetBird | `/netbird` | `netbird-setup-key` | `NETBIRD_SETUP_KEY` |

## CDK8s Generates Zero Secrets

The CI pipeline synthesizes CDK8s manifests to the `v0.1.5-manifests` branch. It requires **zero GitHub Actions secrets** — CDK8s never generates any `Secret` resources. All runtime secrets are pulled by the in-cluster operator at reconcile time.
