#!/usr/bin/env bash
# Check that each trigger prompt in .claude/skills/README.md loads the skill it names.
#
#   skill-trigger-check.sh [skill]
#
# Runs every prompt through `claude -p` with only the Skill tool (no MCP servers) and
# compares the first skill invoked. Pass a skill name to check only its rows.
# Each prompt is a real API call.
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
only="${1:-}"
pass=0
total=0

while IFS= read -r line; do
  [[ "$line" =~ ^\|\ \"(.*)\"\ \|\ \`([^\`]+)\`\ \|$ ]] || continue
  prompt="${BASH_REMATCH[1]}"
  want="${BASH_REMATCH[2]}"
  [[ -z "$only" || "$only" == "$want" ]] || continue

  fired="$(cd "$root" && claude -p "$prompt" --tools Skill --strict-mcp-config --output-format stream-json --verbose \
    --max-budget-usd 0.30 </dev/null 2>/dev/null |
    jq -r 'select(.type=="assistant") | .message.content[]? | select(.type=="tool_use" and .name=="Skill") | .input.skill' |
    head -n1 || true)"
  fired="${fired##*:}"
  fired="${fired%%@*}"

  total=$((total + 1))
  if [[ "$fired" == "$want" ]]; then
    pass=$((pass + 1))
    result=PASS
  else
    result=FAIL
  fi
  printf '%s  %-30s %-30s %s\n' "$result" "$want" "${fired:-none}" "$prompt"
done <"$root/.claude/skills/README.md"

echo "$pass/$total passed"
[[ "$pass" -eq "$total" ]]
