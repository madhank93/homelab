#!/usr/bin/env bash
# PreToolUse guard for Bash: secret material never enters the transcript.
# Backs up the permission deny rules, which only match command prefixes.
# Exit 2 blocks the call and shows stderr to Claude.
set -euo pipefail

cmd="$(jq -r '.tool_input.command // empty')"

deny() {
  echo "Blocked by .claude/hooks/block-secrets.sh: $1. Ask the user to run it." >&2
  exit 2
}

# Encrypted SOPS files, their decrypted output and runtime secret env files.
if grep -qE '\.sops\.ya?ml|secrets/bootstrap\.yaml|\.secrets\.env' <<<"$cmd"; then
  deny "command touches a SOPS or secrets file"
fi
if grep -qE '(^|[;&|([:space:]])sops[[:space:]]+(-d|--decrypt|decrypt|exec-|edit|set|unset)' <<<"$cmd"; then
  deny "sops runs only from the justfile, by the user"
fi
if grep -qE 'just[[:space:]]+(sops-view|sops-decrypt|openbao-get|openbao-sa-get|openbao-token|openbao-root)' <<<"$cmd"; then
  deny "recipe prints secret values"
fi
# Secret payloads from the cluster; names and `describe` (byte counts only) stay allowed.
if grep -qE 'kubectl[[:space:]].*get[[:space:]]+secrets?\b.*(-o|--output)' <<<"$cmd" ||
   grep -qE 'kubectl[[:space:]].*view-secret' <<<"$cmd"; then
  deny "command would print Kubernetes Secret data"
fi
if grep -qE 'bao[[:space:]]+(kv[[:space:]]+get|read)\b' <<<"$cmd"; then
  deny "command would print OpenBao secret data"
fi

exit 0
