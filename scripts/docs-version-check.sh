#!/usr/bin/env bash
#
# Keeps the docs honest about versions. Two rules:
#
#   1. Every version in docs/content/architecture/software-inventory.md must
#      still appear in the file that page says pins it.
#   2. No other page under docs/content/ may state a version at all — they
#      describe behaviour and link to the inventory.
#
# Rule 2 is what makes rule 1 hold: without it, numbers reappear on component
# pages and drift silently the next time a chart moves.
#
# Run from anywhere: just docs-check

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

inventory="docs/content/architecture/software-inventory.md"
failures=0

fail() {
  printf '  \033[31mFAIL\033[0m %s\n' "$1"
  failures=$((failures + 1))
}

# --- Rule 1: inventory rows vs the files they name ---------------------------
#
# A row looks like:
#   | Cilium | 1.18.12 | {{ src(path="core/platform/cilium.go") }} |
# "same" in the last column means "the file from the row above".

echo "Checking inventory rows against source files..."

last_path=""
while IFS= read -r line; do
  # Table body rows only: three or more pipe-separated cells, not the header
  # rule and not the header itself.
  [[ "$line" =~ ^\|[[:space:]] ]] || continue
  [[ "$line" =~ ^\|[[:space:]]*-+ ]] && continue

  component=$(awk -F'|' '{print $2}' <<<"$line" | xargs)
  version=$(awk -F'|' '{print $3}' <<<"$line" | xargs)

  [[ "$component" == "Software" ]] && continue
  [[ -z "$version" ]] && continue

  # The path is wherever src() appears in the row; "same" reuses the row above.
  if [[ "$line" =~ src\(path=\"([^\"]+)\" ]]; then
    last_path="${BASH_REMATCH[1]}"
  elif [[ "$line" != *same* ]]; then
    continue  # prose cell with no source link
  fi

  path="$last_path"
  [[ -z "$path" ]] && continue

  if [[ ! -f "$path" ]]; then
    fail "$component: $path does not exist"
    continue
  fi

  # Accept the version with or without a leading v — docs write v1.21.1 where
  # some charts pin 1.21.1 and vice versa.
  bare="${version#v}"
  if ! grep -qF -- "$bare" "$path"; then
    fail "$component: version '$version' not found in $path"
  fi
done <"$inventory"

# --- Rule 2: no versions outside the inventory -------------------------------
#
# Legitimate exceptions, matched as substrings of the offending line.

allow=(
  'gateway-api-v1.2.1'      # the vendored CRD manifest is named after its version
  '-manifests'              # the Argo CD source branch, checked separately below
  'HTTP/1.1'                # curl output in troubleshooting examples
  'HTTP/2'
  '192.168.'                # LAN addresses
  '178.156.199.250'         # Bifrost VPS
  '100.109.'                # NetBird overlay
  '10.96.'                  # service CIDR
  '172.30.0.'               # bifrost_net
  '1.1.1.1'                 # upstream DNS
  '0.0.0.0'
  # Versions of things this repo does not pin: notebook images a user picks at
  # spawn time, sample command output, and third-party bugs worth naming.
  'kubeflownotebookswg/'
  'torch.__version__'
  'kustomize 5.8.0'
)

# The upgrade guide legitimately names version boundaries a change happened at
# ("Grafana 10 to 11 moved the datasource schema"). Those are historical facts,
# not claims about what is deployed, so they cannot drift.
echo "Checking that no other page states a version..."

while IFS= read -r hit; do
  file="${hit%%:*}"
  rest="${hit#*:}"
  lineno="${rest%%:*}"
  text="${rest#*:}"

  skip=false
  for a in "${allow[@]}"; do
    [[ "$text" == *"$a"* ]] && skip=true && break
  done
  # Escape hatch for a version that is a historical fact rather than a claim
  # about what is deployed — "expose was added in v0.66", "Talos v1.10+ made
  # /etc read-only". Those describe a boundary that already happened, so they
  # cannot drift. Mark the line with a trailing comment.
  [[ "$text" == *"docs-check: historical"* ]] && skip=true
  $skip && continue

  fail "$file:$lineno states a version — move it to the inventory and link there"
  printf '       %s\n' "$(echo "$text" | cut -c1-100)"
done < <(
  # Three-component versions bare or v-prefixed, plus two-component ones only
  # when v-prefixed — "v0.66" is a version, "12.8" is usually a CUDA line.
  grep -rnE '\bv?[0-9]+\.[0-9]+\.[0-9]+\b|\bv[0-9]+\.[0-9]+\b' docs/content \
    --include='*.md' \
    --exclude-dir=upgrade-guide \
    | grep -v "^$inventory:" \
    || true
)

# --- Rule 3: the manifests branch docs name is the one Argo CD watches -------
#
# Allowlisted above so rule 2 ignores it, but it still rots on a release cut,
# so check it against the ApplicationSet directly.

echo "Checking the manifests branch reference..."

argocd_branch=$(grep -oE '"(revision|targetRevision)": "[^"]*-manifests"' core/platform/argocd.go \
  | head -1 | grep -oE '[^" ]*-manifests')

if [[ -z "$argocd_branch" ]]; then
  fail "could not find a *-manifests branch in core/platform/argocd.go"
else
  while IFS= read -r hit; do
    file="${hit%%:*}"
    rest="${hit#*:}"
    lineno="${rest%%:*}"
    fail "$file:$lineno names a manifests branch other than $argocd_branch"
  done < <(
    grep -rnE '\bv[0-9]+\.[0-9]+\.[0-9]+-manifests\b' docs/content --include='*.md' \
      | grep -v "$argocd_branch" || true
  )
fi

echo
if [[ "$failures" -gt 0 ]]; then
  echo "$failures problem(s). The inventory is the single source of truth:"
  echo "  docs/content/architecture/software-inventory.md"
  exit 1
fi
echo "Docs versions agree with the code."
