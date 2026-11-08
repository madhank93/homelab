# Skills

Project skills Claude Code loads on demand. `golang-*` are vendored from
[samber/cc-skills-golang](https://github.com/samber/cc-skills-golang) (MIT) and
the Pulumi ones from [pulumi/agent-skills](https://github.com/pulumi/agent-skills)
(Apache-2.0); see `THIRD_PARTY_NOTICES.md`. Vendored skills are never edited
here; `homelab-*` are this repo's own.

| Skill | Source | Purpose |
|---|---|---|
| `homelab-cdk8s` | repo | `workloads/` charts, kit helpers, OpenBao/CSI secrets, Gateway routes, GPU scheduling, synth verification |
| `homelab-pulumi-go` | repo | `core/` stacks, SOPS-provided config, resource options, bridged SDKs, reading previews |
| `homelab-docs` | repo | `docs/` Starlight pages: placement, page shapes, inventory-only versions, `src:` links, verification |
| `homelab-sso` | repo | Authentik OIDC apps: client secret route, group binding, per-app wiring, login checks |
| `homelab-talos` | repo | Talos machine patches, change impact (reboot or not), talosctl, worker4 GPU node, node upgrades |
| `homelab-incident` | repo | Outage triage by blast radius, read-only first checks, sanctioned fixes, power-outage recovery |
| `homelab-rollout` | repo | Getting a change live: Pulumi stack order, release branch, CI, Argo CD sync, release cut |
| `homelab-upgrade` | repo | Version bumps: upstream research, where pins live, layer order, Renovate PRs |
| `homelab-security-review` | repo | Repo checklist: rendered Secrets, exposure, auth binding, RBAC, Pod Security, pinning, CI |
| `golang-code-style` | samber | Formatting, function shape, comments, readability |
| `golang-naming` | samber | Package, type, function, variable and constant names |
| `golang-error-handling` | samber | Wrapping, sentinel/custom errors, `errors.Is`/`As`, panics |
| `golang-structs-interfaces` | samber | Struct design, embedding, interface size and placement |
| `golang-context` | samber | `context.Context` propagation, cancellation, timeouts |
| `golang-testing` | samber | Table-driven tests, subtests, fixtures, coverage |
| `golang-lint` | samber | golangci-lint configuration and fixing findings |
| `golang-modernize` | samber | Replacing old idioms with current Go features |
| `golang-documentation` | samber | Doc comments, Example tests, READMEs |
| `golang-safety` | samber | Nil panics, aliasing, truncation and other accidental bugs |
| `golang-security` | samber | Injection, crypto, path traversal, SSRF, secrets handling |
| `golang-troubleshooting` | samber | Root-causing crashes, races, deadlocks and odd behaviour |
| `golang-dependency-management` | samber | `go.mod`, upgrades, `govulncheck` |
| `pulumi-best-practices` | pulumi | Outputs, components, aliases, secrets and config (TypeScript/Python examples) |
| `pulumi-debug-failed-operation` | pulumi | Reading a failed `up`/`preview` from Pulumi Cloud and fixing it |
| `provider-upgrade` | pulumi | Bumping a provider and reconciling the resulting diff |
| `differential-review` | Trail of Bits plugin | Security review of a diff |
| `supply-chain-risk-auditor` | Trail of Bits plugin | Advisory and maintenance risk of Go and npm dependencies |

Vendored skills mention libraries this repo does not use (testify, samber/oops,
zap) or Pulumi in TypeScript/Python; ignore those parts rather than editing the
files. `homelab-pulumi-go` holds the Go specifics. Trail of Bits skills are
CC-BY-SA, so they are enabled as plugins in `.claude/settings.json` instead of
being copied into this repo.

## Trigger check

Sample prompts and the skill that should load. Re-run these after editing a
`description`.

| Prompt | Expected skill |
|---|---|
| "Deploy Uptime Kuma to the cluster behind the gateway" | `homelab-cdk8s` |
| "Mount n8n's encryption key from OpenBao into the pod" | `homelab-cdk8s` |
| "Export the Bifrost VPS IPv6 address as a stack output" | `homelab-pulumi-go` |
| "Add a Cloudflare DNS record for a new LAN host" | `homelab-pulumi-go` |
| "Document the new Uptime Kuma workload" | `homelab-docs` |
| "Write the release notes page for the next release" | `homelab-docs` |
| "Add Authentik login to Open WebUI" | `homelab-sso` |
| "Will this Talos patch reboot every node?" | `homelab-talos` |
| "Every madhan.app URL times out from the LAN" | `homelab-incident` |
| "Get this change live on the cluster" | `homelab-rollout` |
| "Upgrade Longhorn to the next minor" | `homelab-upgrade` |
| "Security-review exposing n8n publicly" | `homelab-security-review` |
| "Clean up the formatting and comments in kit.go" | `golang-code-style` |
| "Is `cfg.K` a good name for the config instance?" | `golang-naming` |
| "Wrap the errors returned by DeployHetznerVPS with context" | `golang-error-handling` |
| "Should AuthentikContext be passed by value or pointer?" | `golang-structs-interfaces` |
| "Add a timeout to the HTTP call in getDefaultIdentificationStage" | `golang-context` |
| "Write a table-driven test for serverCfg in cs2.go" | `golang-testing` |
| "Should we enable the godot linter in .golangci.yml?" | `golang-lint` |
| "Replace the manual key loop with maps/slices helpers" | `golang-modernize` |
| "Write the doc comment for the kit package" | `golang-documentation` |
| "Could this map access in serverCfg panic on nil?" | `golang-safety` |
| "Is the exec.Command in hetzner.go injectable?" | `golang-security` |
| "synth hangs with no output, find out why" | `golang-troubleshooting` |
| "go mod tidy pulls in two versions of x/net in workloads; resolve it" | `golang-dependency-management` |
| "`just core platform up` failed, find out why" | `pulumi-debug-failed-operation` |
| "After the Proxmox provider upgrade, preview replaces every VM; reconcile it" | `provider-upgrade` |
| "Do a differential review of the last three commits for security regressions" | `differential-review` |

## Re-sync

```bash
scripts/sync-skills.sh
```

Bump a SHA in the script to move to a newer upstream commit; edit the skill
list after it to add or drop a vendored skill. The script replaces each listed
directory and rewrites `THIRD_PARTY_NOTICES.md`.
