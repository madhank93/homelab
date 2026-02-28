#!/usr/bin/env bash
# ==============================================================================
# bifrost/bootstrap.sh
#
# Starts bifrost services in dependency order, gates each step on health checks,
# and auto-provisions NB_IDP_MGMT_TOKEN by creating an Authentik API token via
# `docker exec authentik-server ak shell` (Django ORM — bypasses the unreliable
# AUTHENTIK_BOOTSTRAP_TOKEN env-var mechanism in Authentik 2025.10+).
# Logs every action with timestamps. Never prints secret values.
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
  for var in CF_DNS_API_TOKEN NB_DATA_STORE_KEY NB_RELAY_SECRET AUTHENTIK_BOOTSTRAP_TOKEN; do
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
  for var in NB_IDP_MGMT_TOKEN NB_PROXY_TOKEN NB_BIFROST_SETUP_KEY; do
    if has_secret "$var"; then
      log "  ✓  $var  ($(secret_len "$var") chars)"
    else
      log "  ⚠  $var  not set  (will be provisioned or skipped below)"
    fi
  done

  log "  pre-flight OK"
}

# ── Provision NB_IDP_MGMT_TOKEN ──────────────────────────────────────────────
# Uses `docker exec authentik-server ak shell` to create or retrieve the akadmin
# API token via the Authentik Django ORM.
#
# Why not AUTHENTIK_BOOTSTRAP_TOKEN env var?
#   In Authentik ≥ 2025.10 the env var is present in the container but the
#   bootstrap job does NOT write the token to the database on a fresh deploy.
#   Directly calling ak shell is reliable and idempotent.
#
# Token identifier: 'netbird-mgmt-token' — stable across re-runs.
# Token key:        taken from AUTHENTIK_BOOTSTRAP_TOKEN in .secrets.env so
#                   it matches the SOPS-managed value and survives rotations.
# Output protocol:  prints "TOKEN:<key>" to stdout on success; errors go to
#                   stderr so they appear in logs without exposing token values.
# Retries up to 12 times (2 min) to tolerate slow Django startup.

provision_nb_idp_mgmt_token() {
  local max_attempts=12
  local attempt=0

  log "  provisioning NB_IDP_MGMT_TOKEN via authentik-server ak shell ..."

  while (( attempt < max_attempts )); do
    attempt=$(( attempt + 1 ))
    log "  [ak shell  attempt=${attempt}/${max_attempts}]"

    local out
    out=$(docker exec authentik-server ak shell -c "
from authentik.core.models import Token, TokenIntents, User
import os, sys

key = os.environ.get('AUTHENTIK_BOOTSTRAP_TOKEN', '')
if not key:
    print('ERROR: AUTHENTIK_BOOTSTRAP_TOKEN not set in container environment', file=sys.stderr)
    sys.exit(1)

try:
    user = User.objects.filter(username='akadmin').first()
    if not user:
        print('ERROR: akadmin user not found in DB', file=sys.stderr)
        sys.exit(1)

    t, created = Token.objects.get_or_create(
        identifier='netbird-mgmt-token',
        defaults={
            'user': user,
            'intent': TokenIntents.INTENT_API,
            'key': key,
            'expiring': False,
            'description': 'NetBird IDP management token (auto-provisioned by bootstrap.sh)',
        }
    )
    if not created and t.key != key:
        # Secret was rotated in SOPS — update the stored key to match.
        t.key = key
        t.save(update_fields=['key'])
        print('updated existing token key', file=sys.stderr)
    elif created:
        print('created new token', file=sys.stderr)
    else:
        print('token already exists with matching key', file=sys.stderr)

    # Only this line goes to stdout — bash greps for the TOKEN: prefix.
    print(f'TOKEN:{t.key}')
    sys.exit(0)
except Exception as exc:
    print(f'ERROR: {exc}', file=sys.stderr)
    sys.exit(1)
" 2>&1)

    local token
    token=$(printf '%s\n' "$out" | grep '^TOKEN:' | cut -d: -f2-)

    if [ -n "$token" ]; then
      log "  token provisioned ($(printf '%s' "$token" | wc -c | tr -d ' ') chars)"
      echo "NB_IDP_MGMT_TOKEN=${token}" >> "$BIFROST/.secrets.env"
      log "  NB_IDP_MGMT_TOKEN written to .secrets.env ✓"
      return 0
    fi

    # Log non-secret diagnostic lines (exclude any line matching TOKEN:)
    log "  ak shell did not return a token — diagnostic output:"
    printf '%s\n' "$out" | grep -v '^TOKEN:' | head -5 | while IFS= read -r line; do
      log "    $line"
    done
    sleep 10
  done

  die "Failed to provision NB_IDP_MGMT_TOKEN after ${max_attempts} attempts — check authentik-server logs"
}

# ── Process netbird config template ──────────────────────────────────────────
# NetBird v0.66 does NOT expand ${VAR} in its config.yaml — the YAML is read
# verbatim. This function substitutes the three secret placeholders in-place
# using sed before netbird-server is started.
#
# Placeholders in netbird/config.yaml:
#   ${NB_RELAY_SECRET}    authSecret
#   ${NB_DATA_STORE_KEY}  store.encryptionKey
#   ${NB_IDP_MGMT_TOKEN}  idp.authentik.managementToken
#
# Safe: base64 values (relay/store) only use A-Za-z0-9+/= which never conflict
# with the sed | delimiter. The IDP token is alphanumeric only.
# Idempotent: if placeholders are already gone (re-run, CopyToRemote unchanged)
# the sed commands match nothing and the file is unchanged.

process_netbird_config() {
  local cfg="$BIFROST/netbird/config.yaml"
  local relay store idp

  relay=$(read_secret NB_RELAY_SECRET)
  store=$(read_secret NB_DATA_STORE_KEY)
  idp=$(read_secret NB_IDP_MGMT_TOKEN)

  log "  substituting secrets in netbird/config.yaml ..."
  local tmp
  tmp=$(mktemp)
  sed \
    -e 's|${NB_RELAY_SECRET}|'"${relay}"'|g' \
    -e 's|${NB_DATA_STORE_KEY}|'"${store}"'|g' \
    -e 's|${NB_IDP_MGMT_TOKEN}|'"${idp}"'|g' \
    "$cfg" > "$tmp" && mv "$tmp" "$cfg"
  log "  netbird/config.yaml processed ✓"
}

# ── Main startup sequence ─────────────────────────────────────────────────────

preflight

log_step "1/6  traefik — TLS termination + routing"
$COMPOSE up -d traefik
wait_healthy traefik 60

log_step "2/6  authentik-postgres — database"
$COMPOSE up -d authentik-postgres
wait_healthy authentik-postgres 120

log_step "3/6  authentik-server + authentik-worker — SSO"
$COMPOSE up -d authentik-server authentik-worker
log "  containers started — waiting for Authentik to be healthy ..."
wait_healthy authentik-server 300

# ── Provision NB_IDP_MGMT_TOKEN (between steps 3 and 4) ──────────────────────

if has_secret "NB_IDP_MGMT_TOKEN"; then
  log "  NB_IDP_MGMT_TOKEN already in .secrets.env ($(secret_len NB_IDP_MGMT_TOKEN) chars) — skipping provisioning"
else
  provision_nb_idp_mgmt_token
fi

process_netbird_config

log_step "4/6  netbird-server — management + signal + relay"
$COMPOSE up -d netbird-server
wait_healthy netbird-server 120

log_step "5/6  netbird-dashboard — UI"
$COMPOSE up -d netbird-dashboard
wait_healthy netbird-dashboard 60

log_step "6/6  netbird-proxy — TCP expose feature"
if has_secret "NB_PROXY_TOKEN"; then
  log "  NB_PROXY_TOKEN present ($(secret_len NB_PROXY_TOKEN) chars)"
  $COMPOSE up -d --force-recreate netbird-proxy
  wait_healthy netbird-proxy 60
else
  log "  NB_PROXY_TOKEN not set — netbird-proxy skipped"
  log ""
  log "  ┌─ One-time setup needed ──────────────────────────────────────┐"
  log "  │  1. Open https://netbird.madhan.app  →  log in via Authentik  │"
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
