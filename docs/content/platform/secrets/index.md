+++
title = "Secrets"
description = "SOPS/age for bootstrap secrets. OpenBao + Secrets Store CSI Driver for runtime secrets — zero credentials stored in git."
weight = 60
+++

## What is SOPS + OpenBao?

[SOPS](https://github.com/getsops/sops) (Secrets OPerationS) is a tool for encrypting secret files using age, PGP, or cloud KMS keys — allowing encrypted secrets to be safely committed to a public git repository. [OpenBao](https://openbao.org/) (a community fork of HashiCorp Vault) is a secrets management server that stores runtime credentials and serves them to pods via the Secrets Store CSI Driver.

## Why This Approach?

SOPS with age encryption keeps bootstrap secrets version-controlled and reproducible without any external service dependency. OpenBao handles runtime secrets at pod startup — no credentials ever appear in Kubernetes manifests, git history, or CDK8s-generated YAML.

## How It's Used Here

Bootstrap secrets (OpenBao unseal key, Cloudflare API token) are age-encrypted in `secrets/bootstrap.sops.yaml` and applied once with `just create-secrets`. All other app credentials live in OpenBao's KV store and are mounted into pods as files via the CSI driver at runtime, keeping CDK8s manifests completely secret-free.

## Architecture

Secrets management uses a two-tier approach:

1. **Bootstrap secrets** — encrypted with SOPS/age, committed safely to git, applied once with `just create-secrets`
2. **Runtime secrets** — stored in OpenBao (Vault fork, MPL-2.0), mounted into pods as files via the Secrets Store CSI Driver

```
secrets/bootstrap.sops.yaml  (age-encrypted, safe in git)
  └── just create-secrets
        ├── openbao/openbao-unseal-key    (Prune=false)
        └── cert-manager/cloudflare-api-token  (Prune=false)

OpenBao (ns: openbao, port 8200)
  ├── KV v2 at secret/
  │     ├── secret/data/grafana   ADMIN_PASSWORD
  │     ├── secret/data/harbor    HARBOR_ADMIN_PASSWORD
  │     ├── secret/data/n8n       ENCRYPTION_KEY
  │     ├── secret/data/rancher   BOOTSTRAP_PASSWORD
  │     └── secret/data/netbird   NETBIRD_SETUP_KEY
  └── Kubernetes Auth method
        └── per-app roles → bound to app ServiceAccount + namespace

Secrets Store CSI Driver (ns: kube-system)
  └── SecretProviderClass (per app ns)
        └── CSI volume in pod → mounts secrets as files
              └── secretObjects → syncs to k8s Secret (Pattern B only)
```

## Pattern A — File-only (no k8s Secret)

Used by: **Grafana**

Secret is mounted as a file at `/mnt/secrets/<KEY>`. The app reads it via an env var pointing to the file path (e.g. `GF_SECURITY_ADMIN_PASSWORD__FILE`).

No k8s Secret is created. The secret value never appears in `kubectl get secret` output.

## Pattern B — secretObjects sync (k8s Secret created)

Used by: **Harbor**, **n8n**, **Rancher**, **NetBird**

The CSI volume mount triggers the SecretProviderClass `secretObjects` block, which creates a k8s Secret in the app's namespace. This is required for Helm charts that only accept `existingSecret` references.

> The CSI volume mount is **required** to trigger the sync — if no pod mounts the volume, the k8s Secret is never created.

## Bootstrap Secrets

Only two Secrets are created by the bootstrap script and never managed by ArgoCD:

| Secret | Namespace | Keys | Purpose |
|--------|-----------|------|---------|
| `openbao-unseal-key` | `openbao` | `unseal-key` | Unseals OpenBao on pod startup via sidecar |
| `cloudflare-api-token` | `cert-manager` | `CLOUDFLARE_API_TOKEN` | DNS-01 ACME challenge for wildcard cert |

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

# Create/update bootstrap k8s Secrets
just create-secrets
```

## OpenBao Setup (one-time)

```bash
# 1. Deploy OpenBao + unseal it
just openbao-init        # initialises, saves root token to /tmp/openbao-init.json

# 2. Configure K8s auth, policies, roles, placeholder secrets
just openbao-setup       # runs scripts/openbao-setup.sh

# 3. Replace placeholder secrets with real values
ROOT_TOKEN=$(python3 -c "import json; print(json.load(open('/tmp/openbao-init.json'))['root_token'])")
kubectl exec -n openbao openbao-0 -- env BAO_TOKEN=$ROOT_TOKEN \
  bao kv put -mount=secret grafana  ADMIN_PASSWORD=<real>
kubectl exec -n openbao openbao-0 -- env BAO_TOKEN=$ROOT_TOKEN \
  bao kv put -mount=secret harbor   HARBOR_ADMIN_PASSWORD=<real>
kubectl exec -n openbao openbao-0 -- env BAO_TOKEN=$ROOT_TOKEN \
  bao kv put -mount=secret n8n      ENCRYPTION_KEY=<real>
kubectl exec -n openbao openbao-0 -- env BAO_TOKEN=$ROOT_TOKEN \
  bao kv put -mount=secret rancher  BOOTSTRAP_PASSWORD=<real>
kubectl exec -n openbao openbao-0 -- env BAO_TOKEN=$ROOT_TOKEN \
  bao kv put -mount=secret netbird  NETBIRD_SETUP_KEY=<real>
```

## Apps and Their Secret Paths

| App | OpenBao Path | Secret keys fetched | k8s Secret created | Pattern |
|-----|-------------|--------------------|--------------------|---------|
| Grafana | `secret/data/grafana` | `ADMIN_PASSWORD` | none | A (file) |
| Harbor | `secret/data/harbor` | `HARBOR_ADMIN_PASSWORD` | `harbor-admin` | B |
| n8n | `secret/data/n8n` | `ENCRYPTION_KEY` | `n8n-secrets` | B |
| Rancher | `secret/data/rancher` | `BOOTSTRAP_PASSWORD` | `rancher-bootstrap` | B |
| NetBird | `secret/data/netbird` | `NETBIRD_SETUP_KEY` | `netbird-setup-key` | B |

> n8n DB password is **not** in OpenBao — it is auto-managed by the CloudNativePG operator (`n8n-pg-app` Secret).

## CDK8s Generates Zero Secrets

The CI pipeline synthesizes CDK8s manifests to the `v0.1.5-manifests` branch. It requires **zero GitHub Actions secrets** — CDK8s never generates any `Secret` resources. All runtime secrets are pulled by the in-cluster CSI driver at mount time.
