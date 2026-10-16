#!/usr/bin/env bash
# Fail if a rendered manifest under <dir> carries a Secret with a payload.
#
#   manifest-secret-guard.sh <dir>
#
# The synthesized output is published to a public branch, so any Secret with
# non-empty data/stringData is a leak. Files listed (path relative to <dir>) in
# manifest-secret-allowlist.txt are tolerated; an entry that no longer renders a
# Secret is also an error, so the allowlist can only shrink. Requires yq v4.
set -euo pipefail

dir=${1:?usage: $0 <manifest-dir>}
dir=${dir%/}
allowlist="$(dirname "$0")/manifest-secret-allowlist.txt"

allowed=$(sed -e 's/#.*//' -e 's/[[:space:]]*$//' -e '/^$/d' "$allowlist")

# One line per leaking Secret: "<namespace>/<name>".
leaks() {
  yq e -N 'select(.kind == "Secret" and
      (((.data // {}) | length) > 0 or ((.stringData // {}) | length) > 0))
    | (.metadata.namespace // "") + "/" + .metadata.name' "$1"
}

failed=0
found=""
while IFS= read -r -d '' file; do
  rel=${file#"$dir"/}
  # Cheap superset pre-filter; yq is the real check.
  grep -q 'Secret' "$file" || continue
  secrets=$(leaks "$file")
  [[ -z $secrets ]] && continue
  found+="$rel"$'\n'
  if grep -qxF -- "$rel" <<<"$allowed"; then
    echo "allowlisted: $rel ($(tr '\n' ' ' <<<"$secrets"))"
    continue
  fi
  while IFS= read -r s; do
    echo "::error file=$rel::Secret $s has data/stringData; it would be published. Source it from OpenBao instead."
  done <<<"$secrets"
  failed=1
done < <(find "$dir" -type f \( -name '*.yaml' -o -name '*.yml' \) -print0 | sort -z)

while IFS= read -r entry; do
  [[ -z $entry ]] && continue
  if ! grep -qxF -- "$entry" <<<"$found"; then
    echo "::error::$entry is allowlisted but renders no Secret payload; remove it from $allowlist."
    failed=1
  fi
done <<<"$allowed"

if ((failed)); then
  echo "manifest-secret-guard: FAILED" >&2
  exit 1
fi
echo "manifest-secret-guard: OK"
