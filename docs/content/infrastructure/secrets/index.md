+++
title = "Secrets"
description = "SOPS + age for bootstrap secrets, Infisical for runtime secrets."
weight = 50
+++

## Architecture

Secrets management uses a two-tier approach:

1. **Bootstrap secrets** — encrypted with SOPS/age, stored in git, created once by `just create-secrets`
2. **Runtime secrets** — stored in Infisical, synced to pods via `InfisicalSecret` CRDs

{% mermaid() %}
flowchart TD
    SOPS["secrets/bootstrap.sops.yaml<br/>Encrypted with age — safe in git"]
    LAPTOP["just create-secrets<br/>Decrypts via sops exec-env"]
    PULUMI["just pulumi talos up<br/>Decrypts via sops exec-env"]

    subgraph Bootstrap ["Bootstrap Secrets (created once)"]
        IS["infisical/infisical-secrets<br/>DB_PASSWORD, AUTH_SECRET, ENCRYPTION_KEY<br/>Prune=false"]
        CF["cert-manager/cloudflare-api-token<br/>CLOUDFLARE_API_TOKEN<br/>Prune=false"]
    end

    subgraph Runtime ["Runtime Secrets (Infisical)"]
        INFISICAL["Infisical Platform<br/>infisical.madhan.app"]
        ISR["InfisicalSecret CRs<br/>per-app namespace"]
        APPSEC["App Secrets<br/>grafana-admin, harbor-admin<br/>n8n-db, rancher-bootstrap"]
    end

    SOPS --> LAPTOP
    SOPS --> PULUMI
    LAPTOP --> Bootstrap
    Bootstrap --> INFISICAL
    INFISICAL --> ISR
    ISR --> APPSEC
{% end %}

## Bootstrap Secrets

Only two Secrets are created by the bootstrap script:

| Secret | Namespace | Keys | Created by |
|--------|-----------|------|------------|
| `infisical-secrets` | `infisical` | DB_PASSWORD, AUTH_SECRET, ENCRYPTION_KEY, DB_CONNECTION_URI, REDIS_PASSWORD | `just create-secrets` |
| `cloudflare-api-token` | `cert-manager` | CLOUDFLARE_API_TOKEN | `just create-secrets` |

Both secrets carry `argocd.argoproj.io/sync-options: Prune=false` — ArgoCD will never delete them even though they do not exist in the manifests branch.

## SOPS + age Setup

### First-Time Setup

```bash
# 1. Install tools
brew install age sops

# 2. Generate age key pair
mkdir -p ~/.config/sops/age
age-keygen -o ~/.config/sops/age/keys.txt
# Public key: age1abc123...

# 3. Add to shell profile (REQUIRED — sops 3.12+ does not auto-discover)
export SOPS_AGE_KEY_FILE="$HOME/.config/sops/age/keys.txt"

# 4. Register public key in .sops.yaml at repo root
# creation_rules:
#   - path_regex: secrets/.*\.sops$
#     age: age1abc123...

# 5. Encrypt the bootstrap secrets file
cp secrets/bootstrap.sops.yaml.example secrets/bootstrap.yaml
$EDITOR secrets/bootstrap.yaml  # fill in real values
sops --encrypt secrets/bootstrap.yaml > secrets/bootstrap.sops.yaml
rm secrets/bootstrap.yaml
```

### Day-to-Day

```bash
# Edit encrypted file (sops opens $EDITOR, re-encrypts on save)
sops secrets/bootstrap.sops.yaml

# Run Pulumi (secrets injected as env vars, never written to disk)
just pulumi talos up

# Create/update bootstrap k8s Secrets
just create-secrets
```

### SOPS exec-env Syntax

```bash
# Correct (command as a single quoted string)
sops exec-env file.sops.yaml 'bash script.sh'

# Wrong (-- not supported by sops)
sops exec-env file.sops.yaml -- bash script.sh
```

## Runtime Secrets (Infisical)

All application runtime secrets are stored in Infisical projects and synced to Kubernetes Secrets via `InfisicalSecret` CRDs.

### Apps That Use Infisical

| App | Infisical Path | k8s Secret | Keys |
|-----|---------------|------------|------|
| Grafana | `/grafana` | `grafana-admin` | `ADMIN_PASSWORD` |
| Harbor | `/harbor` | `harbor-admin` | `HARBOR_ADMIN_PASSWORD` |
| n8n | `/n8n` | `n8n-db` | `DB_PASSWORD` |
| Rancher | `/rancher` | `rancher-bootstrap` | `BOOTSTRAP_PASSWORD` |

### InfisicalSecret CRD Pattern

```yaml
apiVersion: secrets.infisical.com/v1alpha1
kind: InfisicalSecret
metadata:
  name: grafana-admin
  namespace: grafana
  annotations:
    argocd.argoproj.io/sync-options: ServerSideApply=false
spec:
  hostAPI: https://infisical.madhan.app/api
  resyncInterval: 60
  authentication:
    serviceToken:
      serviceTokenSecretReference:
        secretName: infisical-service-token
        secretNamespace: infisical
      secretsScope:
        envSlug: prod
        secretsPath: /grafana
  managedSecretReference:
    secretName: grafana-admin
    secretNamespace: grafana
```

> `argocd.argoproj.io/sync-options: ServerSideApply=false` is required on all `InfisicalSecret` resources because the Infisical CRD schema omits `projectSlug` from `serviceToken.secretsScope`, which breaks ArgoCD's server-side apply validation.

## CDK8s Generates Zero Secrets

The CI pipeline synthesizes CDK8s manifests and publishes them to the `*-manifests` branch. It requires **zero GitHub Actions secrets** because CDK8s never generates any `Secret` resources.
