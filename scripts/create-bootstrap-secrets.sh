#!/usr/bin/env bash
# create-bootstrap-secrets.sh
#
# Creates the bootstrap Kubernetes Secrets that must exist BEFORE ArgoCD deploys
# the platform apps. These secrets are NOT managed by ArgoCD (Prune=false) so
# they persist across ArgoCD syncs.
#
# Usage (always via SOPS to inject env vars):
#   sops exec-env secrets/bootstrap.sops.yaml 'bash scripts/create-bootstrap-secrets.sh'
#
# Or via Justfile:
#   just create-secrets
#
# Secrets created:
#   - openbao/openbao-unseal-key      (unseal-key)
#   - cert-manager/cloudflare-api-token  (CLOUDFLARE_API_TOKEN)
#
# First-time OpenBao setup:
#   1. Run this script (creates namespace + placeholder-ready secret)
#   2. Push manifests → ArgoCD deploys OpenBao (will be sealed)
#   3. Run: just openbao-init   → initialises OpenBao, stores unseal key in SOPS
#   4. Update OPENBAO_UNSEAL_KEY in secrets/bootstrap.sops.yaml
#   5. Run this script again   → updates secret with real key
#   6. kubectl rollout restart statefulset/openbao -n openbao → sidecar unseals
#
# Idempotent: safe to run multiple times (uses kubectl apply).

set -euo pipefail

# ---------------------------------------------------------------------------
# Validate required environment variables
# ---------------------------------------------------------------------------
required_vars=(
  OPENBAO_UNSEAL_KEY
  CLOUDFLARE_API_TOKEN
)

for var in "${required_vars[@]}"; do
  if [[ -z "${!var:-}" ]]; then
    echo "❌  Required environment variable '$var' is not set."
    echo "    Run this script via SOPS:"
    echo "    sops exec-env secrets/bootstrap.sops.yaml 'bash scripts/create-bootstrap-secrets.sh'"
    exit 1
  fi
done

# ---------------------------------------------------------------------------
# Create namespaces (idempotent)
# ---------------------------------------------------------------------------
echo "→ Creating namespaces..."
kubectl create namespace openbao      --dry-run=client -o yaml | kubectl apply -f -
kubectl create namespace cert-manager --dry-run=client -o yaml | kubectl apply -f -

# ---------------------------------------------------------------------------
# openbao-unseal-key  (namespace: openbao)
# Used by the unseal sidecar container to automatically unseal OpenBao on
# pod restart. Populated from SOPS after first `bao operator init` run.
# ---------------------------------------------------------------------------
echo "→ Creating openbao/openbao-unseal-key..."
kubectl apply -f - <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: openbao-unseal-key
  namespace: openbao
  annotations:
    # Prevent ArgoCD from pruning this Secret — it is managed outside GitOps
    argocd.argoproj.io/sync-options: Prune=false
stringData:
  unseal-key: "${OPENBAO_UNSEAL_KEY}"
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
echo "Next steps (first-time setup):"
echo "  1. Push manifests → ArgoCD deploys OpenBao (will be sealed)"
echo "  2. Run: just openbao-init   → initialises OpenBao, outputs unseal key"
echo "  3. Add OPENBAO_UNSEAL_KEY to secrets/bootstrap.sops.yaml"
echo "  4. Run: just create-secrets → updates secret with real key"
echo "  5. kubectl rollout restart statefulset/openbao -n openbao"
echo "  6. Run: just openbao-setup  → configure K8s auth, policies, roles"
echo "  7. Write secrets to OpenBao: bao kv put secret/grafana ADMIN_PASSWORD=..."
