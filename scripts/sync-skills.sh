#!/usr/bin/env bash
# Vendor Go skills from samber/cc-skills-golang (MIT) into .claude/skills/.
#
#   sync-skills.sh
#
# Bump SHA to update. Vendored skill dirs are replaced wholesale on every run, so
# never hand-edit them; repo-specific skills (homelab-*) are left untouched.
set -euo pipefail

REPO=samber/cc-skills-golang
SHA=bb1d43d98a3f8ab639a0042f90d0ec524e39a477
SKILLS=(
  golang-code-style
  golang-naming
  golang-error-handling
  golang-structs-interfaces
  golang-context
  golang-testing
  golang-lint
  golang-modernize
)

dest="$(git rev-parse --show-toplevel)/.claude/skills"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

curl -fsSL "https://codeload.github.com/$REPO/tar.gz/$SHA" | tar -xz -C "$tmp" --strip-components=1

mkdir -p "$dest"
for s in "${SKILLS[@]}"; do
  [[ -f "$tmp/skills/$s/SKILL.md" ]] || { echo "sync-skills: $s not found at $SHA" >&2; exit 1; }
  rm -rf "${dest:?}/$s"
  cp -R "$tmp/skills/$s" "$dest/$s"
done

{
  echo "# Third-party notices"
  echo
  echo "The following skills are vendored by \`scripts/sync-skills.sh\` from"
  echo "https://github.com/$REPO at commit \`$SHA\`:"
  echo
  printf -- '- `%s`\n' "${SKILLS[@]}"
  echo
  echo "Do not edit them here; change the pinned SHA and re-run the script."
  echo
  echo '```text'
  cat "$tmp/LICENSE"
  echo '```'
} >"$dest/THIRD_PARTY_NOTICES.md"

echo "sync-skills: ${#SKILLS[@]} skills from $REPO@${SHA:0:12}"
