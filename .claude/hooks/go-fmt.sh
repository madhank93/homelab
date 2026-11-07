#!/usr/bin/env bash
# PostToolUse on Edit/Write: format the touched .go file with the repo's
# golangci-lint formatters so `just lint` never fails on formatting.
set -euo pipefail

file="$(jq -r '.tool_input.file_path // empty')"
[[ "$file" == *.go && -f "$file" ]] || exit 0
case "$file" in */sdks/* | */imports/*) exit 0 ;; esac

root="$(git -C "$(dirname "$file")" rev-parse --show-toplevel)"
cd "$(dirname "$file")"
golangci-lint fmt --config "$root/.golangci.yml" "$file" >&2 || true
