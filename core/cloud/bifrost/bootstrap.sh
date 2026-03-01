#!/usr/bin/env bash
# ==============================================================================
# bifrost/bootstrap.sh
#
# Starts bifrost services in dependency order, gates each step on health checks.
# Processes netbird/config.yaml by substituting secret placeholders before
# starting netbird-server.
#
# NetBird v0.66 combined server ALWAYS runs embedded Dex OIDC.
# auth.issuer is the NetBird server's own /oauth2 URL.
# After first deploy: configure Authentik connector via NetBird Settings UI.
#
# Invoked via: pulumi remote.Command → bash /etc/bifrost/bootstrap.sh
# Runs as:     root on the Hetzner VPS
# Safe to re-run — docker compose up -d is idempotent.
# ==============================================================================
set -euo pipefail

readonly BIFROST=/etc/bifrost
readonly COMPOSE="docker compose -f $BIFROST/docker-compose.yml"

# ── Logging ───────────────────────────────────────────────────────────────────

log() {
  printf '[bifrost] %s  %s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)" "$*"
}

log_step() {
  printf '\n[bifrost] ══════════════════════════════════════════════\n'
  printf '[bifrost]   %s\n' "$*"
  printf '[bifrost] ══════════════════════════════════════════════\n'
}

die() {
  log "FATAL: $*"
  exit 1
}

# ── Helpers ───────────────────────────────────────────────────────────────────

# wait_healthy <container> [timeout_secs]
# Polls until the container is running + healthy (or running with no healthcheck).
wait_healthy() {
  local name="$1"
  local max="${2:-120}"
  local interval=5
  local elapsed=0

  log "  waiting for $name  (timeout=${max}s, poll every ${interval}s)"

  while true; do
    local state health
    state=$(docker inspect \
              --format='{{.State.Status}}' \
              "$name" 2>/dev/null || echo "missing")
    health=$(docker inspect \
              --format='{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}' \
              "$name" 2>/dev/null || echo "missing")

    log "  [$name  t=${elapsed}s]  state=$state  health=$health"

    if [[ "$state" == "running" ]] && [[ "$health" == "healthy" || "$health" == "none" ]]; then
      log "  $name ready ✓"
      return 0
    fi

    if (( elapsed >= max )); then
      log "  ── last 30 log lines for $name ──"
      docker logs --tail=30 "$name" 2>&1 || true
      die "$name did not become ready within ${max}s"
    fi

    sleep "$interval"
    elapsed=$(( elapsed + interval ))
  done
}

# has_secret <varname>
# Returns 0 if varname=<non-empty-value> exists in .secrets.env.
has_secret() {
  grep -q "^${1}=[^[:space:]]" "$BIFROST/.secrets.env" 2>/dev/null
}

# secret_len <varname>
# Prints character count of the value for logging (never the value itself).
secret_len() {
  grep "^${1}=" "$BIFROST/.secrets.env" 2>/dev/null \
    | cut -d= -f2- | tr -d '\n' | wc -c | tr -d ' '
}

# read_secret <varname>
# Prints the value of a secret from .secrets.env (used internally — never logged).
read_secret() {
  grep "^${1}=" "$BIFROST/.secrets.env" 2>/dev/null | cut -d= -f2-
}

# ── Pre-flight ────────────────────────────────────────────────────────────────

preflight() {
  log_step "Pre-flight"

  # On a fresh VPS docker is installed by cloud-init — wait for it to finish.
  if command -v cloud-init &>/dev/null; then
    log "  waiting for cloud-init to finish ..."
    cloud-init status --wait 2>&1 \
      && log "  cloud-init: complete" \
      || log "  cloud-init: non-zero exit (may be normal on pre-provisioned servers)"
  else
    log "  cloud-init not found — assuming pre-provisioned server"
  fi

  docker compose version > /dev/null 2>&1 \
    || die "docker compose not available; check /var/log/cloud-init-output.log"
  log "  docker compose: OK"

  # Validate required secrets
  log "  checking required secrets in .secrets.env ..."
  local missing=0
  for var in CF_DNS_API_TOKEN NB_DATA_STORE_KEY NB_RELAY_SECRET AUTHENTIK_BOOTSTRAP_TOKEN NB_OWNER_PASSWORD; do
    if has_secret "$var"; then
      log "  ✓  $var  ($(secret_len "$var") chars)"
    else
      log "  ✗  $var  MISSING"
      missing=$(( missing + 1 ))
    fi
  done

  if [ "$missing" -gt 0 ]; then
    die "$missing required secret(s) absent from .secrets.env — run: just core hetzner up"
  fi

  # Warn about optional secrets
  for var in NB_PROXY_TOKEN NB_BIFROST_SETUP_KEY; do
    if has_secret "$var"; then
      log "  ✓  $var  ($(secret_len "$var") chars)"
    else
      log "  ⚠  $var  not set  (will be skipped below)"
    fi
  done

  log "  pre-flight OK"
}

# ── Process netbird config template ──────────────────────────────────────────
# NetBird v0.66 does NOT expand ${VAR} in its config.yaml — the YAML is read
# verbatim. This function substitutes secret placeholders before starting
# netbird-server.
#
# Placeholders in netbird/config.yaml:
#   ${NB_RELAY_SECRET}    authSecret         (base64 — safe with sed)
#   ${NB_DATA_STORE_KEY}  store.encryptionKey (base64 — safe with sed)
#   ${NB_OWNER_HASH}      auth.owner.password (bcrypt hash — contains $ and /,
#                                              must use Python for substitution)
#
# Idempotent: if placeholders are already gone (re-run), the commands match
# nothing and the file is unchanged.

process_netbird_config() {
  local cfg="$BIFROST/netbird/config.yaml"
  local relay store owner_pass owner_hash

  relay=$(read_secret NB_RELAY_SECRET)
  store=$(read_secret NB_DATA_STORE_KEY)
  owner_pass=$(read_secret NB_OWNER_PASSWORD)

  # Generate bcrypt hash from owner password.
  # Bcrypt hash contains $, /, and other chars that break sed — use Python.
  log "  generating bcrypt hash for owner password ..."
  python3 -c "import bcrypt" 2>/dev/null \
    || { log "  installing python3-bcrypt ..."; apt-get install -y python3-bcrypt > /dev/null 2>&1; }

  owner_hash=$(_OWNER_PASS="$owner_pass" python3 - <<'PYEOF'
import bcrypt, os
p = os.environ['_OWNER_PASS'].encode()
print(bcrypt.hashpw(p, bcrypt.gensalt(10)).decode())
PYEOF
)
  log "  bcrypt hash generated ($(printf '%s' "$owner_hash" | wc -c | tr -d ' ') chars)"

  log "  substituting secrets in netbird/config.yaml ..."

  # Step 1: substitute base64 values via sed (safe — no special delimiter chars)
  local tmp
  tmp=$(mktemp)
  sed \
    -e 's|${NB_RELAY_SECRET}|'"${relay}"'|g' \
    -e 's|${NB_DATA_STORE_KEY}|'"${store}"'|g' \
    "$cfg" > "$tmp" && mv "$tmp" "$cfg"

  # Step 2: substitute bcrypt hash via Python (safe — handles $2b$10$... correctly)
  _OWNER_HASH="$owner_hash" python3 - "$cfg" <<'PYEOF'
import sys, os
with open(sys.argv[1]) as f:
    d = f.read()
d = d.replace('${NB_OWNER_HASH}', os.environ['_OWNER_HASH'])
with open(sys.argv[1], 'w') as f:
    f.write(d)
PYEOF

  log "  netbird/config.yaml processed ✓"
}

# ── Main startup sequence ─────────────────────────────────────────────────────

preflight

log_step "1/5  traefik — TLS termination + routing"
$COMPOSE up -d traefik
wait_healthy traefik 60

log_step "2/5  authentik-postgres — database"
$COMPOSE up -d authentik-postgres
wait_healthy authentik-postgres 120

log_step "3/5  authentik-server + authentik-worker — SSO"
$COMPOSE up -d authentik-server authentik-worker
log "  containers started — waiting for Authentik to be healthy ..."
wait_healthy authentik-server 300

# Process netbird config before starting netbird-server
process_netbird_config

log_step "4/5  netbird-server + netbird-dashboard — NetBird"
$COMPOSE up -d netbird-server netbird-dashboard
wait_healthy netbird-server 120
wait_healthy netbird-dashboard 60

log ""
log "  ┌─ First-time setup (if not already done) ────────────────────────┐"
log "  │  1. Open https://netbird.madhan.app                              │"
log "  │  2. Log in with local admin: admin@madhan.app + NB_OWNER_PASSWORD│"
log "  │  3. Settings → Identity Providers → Add → Authentik              │"
log "  │     Client ID:     aumenijDycfG1cQURqH9BNJpV3KVUCoMHGPUVUlT     │"
log "  │     Client Secret: (from sops: NETBIRD_CLIENT_SECRET)            │"
log "  │     Issuer:        https://auth.madhan.app/application/o/netbird/│"
log "  └─────────────────────────────────────────────────────────────────┘"

log_step "5/5  netbird-proxy — TCP expose feature"
if has_secret "NB_PROXY_TOKEN"; then
  log "  NB_PROXY_TOKEN present ($(secret_len NB_PROXY_TOKEN) chars)"
  $COMPOSE up -d --force-recreate netbird-proxy
  wait_healthy netbird-proxy 60
else
  log "  NB_PROXY_TOKEN not set — netbird-proxy skipped"
  log ""
  log "  ┌─ One-time setup needed ──────────────────────────────────────┐"
  log "  │  1. Open https://netbird.madhan.app  →  log in               │"
  log "  │  2. Settings → Access Tokens → Create Personal Access Token   │"
  log "  │  3. sops edit secrets/bootstrap.sops.yaml                     │"
  log "  │     Add:  NB_PROXY_TOKEN=<token>                              │"
  log "  │  4. just core hetzner up   (this script re-runs automatically) │"
  log "  └──────────────────────────────────────────────────────────────┘"
fi

log_step "Container status"
docker ps --format 'table {{.Names}}\t{{.Status}}' \
  | grep -E '(NAMES|traefik|netbird|authentik)' \
  || docker ps --format 'table {{.Names}}\t{{.Status}}'

log "Bootstrap complete."
