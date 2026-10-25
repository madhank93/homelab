#!/usr/bin/env bash
# Secret scan run by the git hooks in .githooks/ (enable: `just hooks`).
#
#   secret-scan.sh staged            staged changes (pre-commit)
#   secret-scan.sh push              commits about to be pushed (pre-push; refs on stdin)
#
# Both scanners run from pinned images, so the only local requirement is Docker.
# A scanner that cannot run blocks the commit/push: an unscanned change is treated
# as a leak. Override a false positive with an inline `gitleaks:allow` comment,
# or skip once with --no-verify.
set -euo pipefail

GITLEAKS=ghcr.io/gitleaks/gitleaks:v8.30.1@sha256:c00b6bd0aeb3071cbcb79009cb16a60dd9e0a7c60e2be9ab65d25e6bc8abbb7f
TRUFFLEHOG=trufflesecurity/trufflehog:3.99.2@sha256:47a84bc18a0d04a165498bbbd3bacfd84176661cbc91f8e0f85467f6a771e99a

repo=$(git rev-parse --show-toplevel)
# In a worktree, .git is a file pointing into the main repo's git dir; mount that
# at its own path too, or the scanners see no repository.
common=$(git rev-parse --path-format=absolute --git-common-dir)

if ! docker info >/dev/null 2>&1; then
  echo "secret-scan: Docker is not running; start it or skip once with --no-verify." >&2
  exit 1
fi

run() { docker run --rm -v "$repo:/repo" -v "$common:$common" -w /repo "$@"; }

gitleaks() { run "$GITLEAKS" git --no-banner --redact "$@" /repo; }

# --results=verified,unknown: report live secrets and ones the verifier could not
# check, skip the ones it proved dead. Verification calls each provider's API.
trufflehog() { run "$TRUFFLEHOG" git file:///repo --no-update --fail --results=verified,unknown "$@"; }

case "${1:-}" in
staged)
  gitleaks --pre-commit --staged
  # trufflehog's git mode only reads commits, so give it the staged blobs as files.
  stage=$(mktemp -d)
  trap 'rm -rf "$stage"' EXIT
  git diff --cached --name-only --diff-filter=ACMR -z | xargs -0 -r git checkout-index --prefix="$stage/" --
  if [[ -n $(ls -A "$stage") ]]; then
    docker run --rm -v "$stage:/staged:ro" "$TRUFFLEHOG" filesystem /staged \
      --no-update --fail --results=verified,unknown
  fi
  ;;
push)
  zero=0000000000000000000000000000000000000000
  while read -r _ local_sha _ remote_sha; do
    [[ $local_sha == "$zero" ]] && continue # branch deletion
    if [[ $remote_sha == "$zero" ]]; then
      # New branch: scan only commits no remote has yet.
      oldest=$(git rev-list --reverse "$local_sha" --not --remotes | head -1)
      [[ -z $oldest ]] && continue
      range="$local_sha --not --remotes"
      base=$(git rev-parse --verify --quiet "$oldest^" || true)
    else
      range="$remote_sha..$local_sha"
      base=$remote_sha
    fi
    gitleaks --log-opts="$range"
    trufflehog --branch "$local_sha" ${base:+--since-commit "$base"}
  done
  ;;
*)
  echo "usage: $0 staged|push" >&2
  exit 2
  ;;
esac
