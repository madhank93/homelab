#!/usr/bin/env bash
# openbao-setup.sh — configure OpenBao K8s auth, per-app policies/roles and app secrets.
#
# Run via `just openbao-setup` once OpenBao is initialised and unsealed. Safe to
# re-run: auth/policies/roles are rewritten, but existing KV values are never
# overwritten — missing keys are seeded, present keys are left alone.
#
# Authenticates with a root token generated from the unseal key (the
# openbao-unseal-key Secret) and revoked on exit. No token is printed or put in argv.

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ROOT_TOKEN=""

# The token travels as the first stdin line so it never appears in a process list;
# the rest of stdin (policy body, secret value) passes through to bao.
bao_stdin() {
  { printf '%s\n' "$ROOT_TOKEN"; cat; } |
    kubectl exec -i -n openbao openbao-0 -c openbao -- \
      sh -c 'read -r BAO_TOKEN; export BAO_TOKEN; exec bao "$@"' bao "$@"
}
bao() { bao_stdin "$@" </dev/null; }

# fd 3 = original stderr: an EXIT trap keeps the failing command's redirections.
exec 3>&2
revoke() {
  [[ -n "$ROOT_TOKEN" ]] || return 0
  if bao token revoke -self >/dev/null; then
    echo "→ Setup root token revoked." >&3
  else
    echo "⚠️  Setup root token NOT revoked; find it with: bao list auth/token/accessors" >&3
  fi
}
trap revoke EXIT

ROOT_TOKEN="$(just -f "$repo_root/justfile" openbao-token)"

rand() { openssl rand -base64 32 | tr -d '/+=\n'; }

# Prints the secret's JSON, or nothing if the path does not exist. Other errors abort.
read_secret() {
  local app="$1" err=""
  if bao kv get -mount=secret -format=json "$app" 2>/dev/null; then return 0; fi
  err="$(bao kv get -mount=secret "$app" 2>&1 >/dev/null || true)"
  if [[ "$err" == *"No value found"* ]]; then return 0; fi
  echo "❌  reading secret/$app: $err" >&2
  exit 1
}

# Writes key=value at secret/<app> only if the key is absent. Value goes via stdin.
seed() {
  local app="$1" key="$2" value="$3" json=""
  json="$(read_secret "$app")"
  if [[ -z "$json" ]]; then
    printf '%s' "$value" | bao_stdin kv put -cas=0 -mount=secret "$app" "$key=-" >/dev/null
    echo "  secret/$app $key: created"
  elif jq -e --arg k "$key" '.data.data | has($k)' <<<"$json" >/dev/null; then
    echo "  secret/$app $key: exists, kept"
  else
    printf '%s' "$value" | bao_stdin kv patch -mount=secret "$app" "$key=-" >/dev/null
    echo "  secret/$app $key: added"
  fi
}

echo "→ Configuring Kubernetes auth..."
bao auth enable kubernetes >/dev/null 2>&1 || echo "  (kubernetes auth already enabled)"
bao write auth/kubernetes/config \
  kubernetes_host="https://kubernetes.default.svc.cluster.local" \
  disable_local_ca_jwt=false >/dev/null

echo "→ Enabling KV v2 at secret/..."
bao secrets enable -path=secret kv-v2 >/dev/null 2>&1 || echo "  (secret/ already enabled)"

# app:ServiceAccount that mounts the app's SecretProviderClass (namespace = app).
roles="grafana:grafana harbor:secret-sync n8n:n8n netbird:default falco:secret-sync"

echo "→ Writing read-only policies and K8s auth roles..."
for role in $roles; do
  app="${role%%:*}"
  sa="${role#*:}"
  printf 'path "secret/data/%s" {\n  capabilities = ["read"]\n}\n' "$app" |
    bao_stdin policy write "$app-policy" - >/dev/null
  bao write "auth/kubernetes/role/$app" \
    bound_service_account_names="$sa" \
    bound_service_account_namespaces="$app" \
    policies="$app-policy" \
    ttl=1h >/dev/null
  echo "  $app (sa $sa)"
done

# Keys must match the objects list of each app's kit.SecretProviderClass.
# The NetBird key comes from its UI, so it starts as a random placeholder that
# lets the CSI mount succeed; replace it with `kv patch` later.
echo "→ Seeding missing secrets (existing values are never changed)..."
seed grafana ADMIN_PASSWORD "$(rand)"
seed grafana OAUTH_CLIENT_SECRET "${GRAFANA_OAUTH_CLIENT_SECRET:?missing from secrets/bootstrap.sops.yaml}"
seed harbor HARBOR_ADMIN_PASSWORD "$(rand)"
# n8n's encryption key must never change, or saved credentials stop decrypting.
# Older deploy docs wrote it as N8N_ENCRYPTION_KEY; carry that value over if present.
legacy_n8n_key="$(jq -r '.data.data.N8N_ENCRYPTION_KEY // empty' <<<"$(read_secret n8n)")"
seed n8n ENCRYPTION_KEY "${legacy_n8n_key:-$(rand)}"
seed netbird NETBIRD_SETUP_KEY "$(rand)"
seed falco FALCOSIDEKICK_UI_USER "admin:$(rand)"

cat <<'EOF'

✅  OpenBao setup complete.

If netbird NETBIRD_SETUP_KEY was just created, replace it with the
k8s-routing-peer setup key from the NetBird UI.
Read values with: just openbao-get secret/<app> [KEY]
EOF
