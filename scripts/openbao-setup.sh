#!/usr/bin/env bash
# openbao-setup.sh
#
# One-time OpenBao configuration script. Run AFTER:
#   1. OpenBao is deployed and unsealed (just openbao-init + just create-secrets)
#   2. OPENBAO_UNSEAL_KEY is in SOPS and the openbao-unseal-key Secret is updated
#
# Usage (via Justfile):
#   just openbao-setup
#
# What this script does:
#   1. Enables Kubernetes auth method
#   2. Configures K8s auth with the cluster's API server + CA
#   3. Enables KV v2 secrets engine at secret/
#   4. Creates per-app policies (read-only access to their secret path)
#   5. Creates K8s auth roles (bind app ServiceAccount → policy)
#   6. Writes placeholder secrets at each path (to be replaced with real values)

set -euo pipefail

# Root token is written to /tmp/openbao-init.json by `just openbao-init`
if [[ ! -f /tmp/openbao-init.json ]]; then
  echo "❌  /tmp/openbao-init.json not found. Run: just openbao-init first."
  exit 1
fi
ROOT_TOKEN=$(python3 -c "import json; print(json.load(open('/tmp/openbao-init.json'))['root_token'])")

BAO="kubectl exec -i -n openbao openbao-0 -- env BAO_TOKEN=${ROOT_TOKEN} bao"

echo "→ Configuring OpenBao K8s auth method..."
$BAO auth enable kubernetes 2>/dev/null || echo "  (kubernetes auth already enabled)"

# Configure K8s auth — use the in-cluster service account token reviewer
$BAO write auth/kubernetes/config \
  kubernetes_host="https://kubernetes.default.svc.cluster.local" \
  disable_local_ca_jwt=false

echo "→ Enabling KV v2 secrets engine at secret/..."
$BAO secrets enable -path=secret kv-v2 2>/dev/null || echo "  (secret/ engine already enabled)"

# ---------------------------------------------------------------------------
# Policies — each app gets read-only access to its own secret path
# ---------------------------------------------------------------------------
echo "→ Creating policies..."

$BAO policy write grafana-policy - <<'POLICY'
path "secret/data/grafana" {
  capabilities = ["read"]
}
POLICY

$BAO policy write harbor-policy - <<'POLICY'
path "secret/data/harbor" {
  capabilities = ["read"]
}
POLICY

$BAO policy write n8n-policy - <<'POLICY'
path "secret/data/n8n" {
  capabilities = ["read"]
}
POLICY

$BAO policy write rancher-policy - <<'POLICY'
path "secret/data/rancher" {
  capabilities = ["read"]
}
POLICY

$BAO policy write netbird-policy - <<'POLICY'
path "secret/data/netbird" {
  capabilities = ["read"]
}
POLICY

# ---------------------------------------------------------------------------
# K8s auth roles — bind ServiceAccount + namespace → policy
# ---------------------------------------------------------------------------
echo "→ Creating K8s auth roles..."

# Grafana: SA created by the Grafana Helm chart (named "grafana" by default)
$BAO write auth/kubernetes/role/grafana \
  bound_service_account_names=grafana \
  bound_service_account_namespaces=grafana \
  policies=grafana-policy \
  ttl=1h

# Harbor: secret-sync SA (dedicated pod; Harbor chart doesn't support extraVolumes)
$BAO write auth/kubernetes/role/harbor \
  bound_service_account_names=secret-sync \
  bound_service_account_namespaces=harbor \
  policies=harbor-policy \
  ttl=1h

# N8n: SA created by the N8n Helm chart (release name → SA name "n8n")
$BAO write auth/kubernetes/role/n8n \
  bound_service_account_names=n8n \
  bound_service_account_namespaces=n8n \
  policies=n8n-policy \
  ttl=1h

# Rancher: secret-sync SA (dedicated pod; Rancher chart doesn't support extraVolumes)
$BAO write auth/kubernetes/role/rancher \
  bound_service_account_names=secret-sync \
  bound_service_account_namespaces=cattle-system \
  policies=rancher-policy \
  ttl=1h

# NetBird peer: uses the default SA in the netbird namespace
$BAO write auth/kubernetes/role/netbird \
  bound_service_account_names=default \
  bound_service_account_namespaces=netbird \
  policies=netbird-policy \
  ttl=1h

# ---------------------------------------------------------------------------
# Placeholder secrets — replace with real values before apps are deployed
# ---------------------------------------------------------------------------
echo "→ Writing placeholder secrets (REPLACE THESE with real values)..."

$BAO kv put -mount=secret grafana  ADMIN_PASSWORD="CHANGEME"
$BAO kv put -mount=secret harbor   HARBOR_ADMIN_PASSWORD="CHANGEME"
$BAO kv put -mount=secret n8n      DB_PASSWORD="CHANGEME"
$BAO kv put -mount=secret rancher  BOOTSTRAP_PASSWORD="CHANGEME"
$BAO kv put -mount=secret netbird  NETBIRD_SETUP_KEY="CHANGEME"

echo ""
echo "✅  OpenBao setup complete."
echo ""
echo "⚠️   Replace placeholder secrets with real values:"
echo "    ROOT_TOKEN=\$(python3 -c \"import json; print(json.load(open('/tmp/openbao-init.json'))['root_token'])\")"
echo "    kubectl exec -n openbao openbao-0 -- env BAO_TOKEN=$ROOT_TOKEN bao kv put -mount=secret grafana  ADMIN_PASSWORD=<real>"
echo "    kubectl exec -n openbao openbao-0 -- env BAO_TOKEN=$ROOT_TOKEN bao kv put -mount=secret harbor   HARBOR_ADMIN_PASSWORD=<real>"
echo "    kubectl exec -n openbao openbao-0 -- env BAO_TOKEN=$ROOT_TOKEN bao kv put -mount=secret n8n      DB_PASSWORD=<real>"
echo "    kubectl exec -n openbao openbao-0 -- env BAO_TOKEN=$ROOT_TOKEN bao kv put -mount=secret rancher  BOOTSTRAP_PASSWORD=<real>"
echo "    kubectl exec -n openbao openbao-0 -- env BAO_TOKEN=$ROOT_TOKEN bao kv put -mount=secret netbird  NETBIRD_SETUP_KEY=<real>"
echo ""
echo "After writing secrets, trigger ArgoCD sync to start app pods."
