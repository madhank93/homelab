#!/usr/bin/env bash
# create-bootstrap-secrets.sh
#
# Creates the bootstrap Kubernetes Secrets that must exist BEFORE ArgoCD deploys
# the platform apps. These secrets are NOT managed by ArgoCD (Prune=false) so
# they persist across ArgoCD syncs.
#
# Usage (always via SOPS to inject env vars):
#   sops exec-env infra/secrets/bootstrap.env.sops -- bash infra/scripts/create-bootstrap-secrets.sh
#
# Or via Justfile:
#   just create-secrets
#
# Secrets created:
#   - infisical/infisical-secrets     (DB_PASSWORD, AUTH_SECRET, ENCRYPTION_KEY,
#                                      DB_CONNECTION_URI, REDIS_PASSWORD)
#   - cert-manager/cloudflare-api-token  (CLOUDFLARE_API_TOKEN)
#
# Idempotent: safe to run multiple times (uses kubectl apply).

set -euo pipefail

# ---------------------------------------------------------------------------
# Validate required environment variables
# ---------------------------------------------------------------------------
required_vars=(
  INFISICAL_DB_PASSWORD
  INFISICAL_ENCRYPTION_KEY
  INFISICAL_AUTH_SECRET
  REDIS_PASSWORD
  CLOUDFLARE_API_TOKEN
)

for var in "${required_vars[@]}"; do
  if [[ -z "${!var:-}" ]]; then
    echo "❌  Required environment variable '$var' is not set."
    echo "    Run this script via SOPS:"
    echo "    sops exec-env infra/secrets/bootstrap.env.sops -- bash infra/scripts/create-bootstrap-secrets.sh"
    exit 1
  fi
done

# ---------------------------------------------------------------------------
# Build derived values
# ---------------------------------------------------------------------------
DB_CONNECTION_URI="postgresql://infisical:${INFISICAL_DB_PASSWORD}@postgresql:5432/infisical"

# ---------------------------------------------------------------------------
# Create namespaces (idempotent)
# ---------------------------------------------------------------------------
echo "→ Creating namespaces..."
kubectl create namespace infisical  --dry-run=client -o yaml | kubectl apply -f -
kubectl create namespace cert-manager --dry-run=client -o yaml | kubectl apply -f -

# ---------------------------------------------------------------------------
# infisical-secrets  (namespace: infisical)
# ---------------------------------------------------------------------------
echo "→ Creating infisical/infisical-secrets..."
kubectl apply -f - <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: infisical-secrets
  namespace: infisical
  annotations:
    # Prevent ArgoCD from pruning this Secret — it is managed outside GitOps
    argocd.argoproj.io/sync-options: Prune=false
stringData:
  DB_PASSWORD: "${INFISICAL_DB_PASSWORD}"
  AUTH_SECRET: "${INFISICAL_AUTH_SECRET}"
  ENCRYPTION_KEY: "${INFISICAL_ENCRYPTION_KEY}"
  DB_CONNECTION_URI: "${DB_CONNECTION_URI}"
  REDIS_PASSWORD: "${REDIS_PASSWORD}"
EOF

# ---------------------------------------------------------------------------
# cloudflare-api-token  (namespace: cert-manager)
# ---------------------------------------------------------------------------
echo "→ Creating cert-manager/cloudflare-api-token..."
kubectl apply -f - <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: cloudflare-api-token
  namespace: cert-manager
  annotations:
    # Prevent ArgoCD from pruning this Secret — it is managed outside GitOps
    argocd.argoproj.io/sync-options: Prune=false
stringData:
  CLOUDFLARE_API_TOKEN: "${CLOUDFLARE_API_TOKEN}"
EOF

# ---------------------------------------------------------------------------
echo ""
echo "✅  Bootstrap secrets created successfully."
echo ""
echo "Next steps:"
echo "  1. Run Pulumi to provision the cluster:  just pulumi talos up"
echo "  2. Wait for ArgoCD to deploy Infisical (~5 min after cluster ready)"
echo "  3. Log in to Infisical and store all secrets there for ongoing management"
echo "  4. Configure Infisical's InfisicalSecret CRs to sync secrets to other apps"
