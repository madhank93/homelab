#!/usr/bin/env bash
# Vendor third-party skills into .claude/skills/ at pinned commits:
#   samber/cc-skills-golang (MIT)    → golang-*
#   pulumi/agent-skills (Apache-2.0) → pulumi-*, provider-upgrade
#
#   sync-skills.sh
#
# Bump a SHA to update. Vendored skill dirs are replaced wholesale on every run, so
# never hand-edit them; repo-specific skills (homelab-*) are left untouched.
set -euo pipefail

dest="$(git rev-parse --show-toplevel)/.claude/skills"
notices="$dest/THIRD_PARTY_NOTICES.md"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

# vendor REPO SHA SKILLS_DIR SKILL...: copy skills and append the licence to the notices.
vendor() {
  local repo=$1 sha=$2 dir=$3 src="$tmp/${1//\//_}"
  shift 3
  mkdir -p "$src"
  curl -fsSL "https://codeload.github.com/$repo/tar.gz/$sha" | tar -xz -C "$src" --strip-components=1
  for s in "$@"; do
    [[ -f "$src/$dir/$s/SKILL.md" ]] || { echo "sync-skills: $s not found in $repo@$sha" >&2; exit 1; }
    rm -rf "${dest:?}/$s"
    cp -R "$src/$dir/$s" "$dest/$s"
    # A `paths:` glob hides the skill until a matching file is read; prompts here
    # often name Go code before opening it, so keep every skill listed.
    awk '/^---$/ { fm++ } fm == 1 && /^paths:/ { skip = 1; next } skip && /^  - / { next } { skip = 0; print }' \
      "$src/$dir/$s/SKILL.md" >"$dest/$s/SKILL.md"
  done
  {
    echo
    echo "## $repo"
    echo
    echo "From https://github.com/$repo at commit \`$sha\`:"
    echo
    printf -- '- `%s`\n' "$@"
    echo
    echo '```text'
    cat "$src/LICENSE"
    echo '```'
  } >>"$notices"
  echo "sync-skills: $# skills from $repo@${sha:0:12}"
}

mkdir -p "$dest"
{
  echo "# Third-party notices"
  echo
  echo "Skills vendored by \`scripts/sync-skills.sh\`. Do not edit them here; change"
  echo "the pinned SHA and re-run the script."
} >"$notices"

vendor samber/cc-skills-golang bb1d43d98a3f8ab639a0042f90d0ec524e39a477 skills \
  golang-code-style \
  golang-naming \
  golang-error-handling \
  golang-structs-interfaces \
  golang-context \
  golang-testing \
  golang-lint \
  golang-modernize \
  golang-documentation \
  golang-safety \
  golang-security \
  golang-troubleshooting \
  golang-dependency-management

vendor pulumi/agent-skills 82f222a529a499e8005a820000148cc5e28c934c pulumi/skills \
  pulumi-best-practices \
  pulumi-debug-failed-operation \
  provider-upgrade
