# Skills

Project skills Claude Code loads on demand. `golang-*` are vendored from
[samber/cc-skills-golang](https://github.com/samber/cc-skills-golang) (MIT, see
`THIRD_PARTY_NOTICES.md`) and are never edited here; `homelab-*` are this repo's
own.

| Skill | Source | Purpose |
|---|---|---|
| `homelab-cdk8s` | repo | `workloads/` charts, kit helpers, OpenBao/CSI secrets, Gateway routes, GPU scheduling, synth verification |
| `homelab-pulumi-go` | repo | `core/` stacks, SOPS-provided config, resource options, bridged SDKs, reading previews |
| `golang-code-style` | samber | Formatting, function shape, comments, readability |
| `golang-naming` | samber | Package, type, function, variable and constant names |
| `golang-error-handling` | samber | Wrapping, sentinel/custom errors, `errors.Is`/`As`, panics |
| `golang-structs-interfaces` | samber | Struct design, embedding, interface size and placement |
| `golang-context` | samber | `context.Context` propagation, cancellation, timeouts |
| `golang-testing` | samber | Table-driven tests, subtests, fixtures, coverage |
| `golang-lint` | samber | golangci-lint configuration and fixing findings |
| `golang-modernize` | samber | Replacing old idioms with current Go features |

Vendored skills mention libraries this repo does not use (testify, samber/oops,
zap); ignore those parts rather than editing the files.

## Trigger check

Sample prompts and the skill that should load. Re-run these after editing a
`description`.

| Prompt | Expected skill |
|---|---|
| "Deploy Uptime Kuma to the cluster behind the gateway" | `homelab-cdk8s` |
| "Harbor needs a new OIDC secret from OpenBao" | `homelab-cdk8s` |
| "Pin the Proxmox VM disk so a new Talos image doesn't replace it" | `homelab-pulumi-go` |
| "Add an Authentik OIDC application for Open WebUI" | `homelab-pulumi-go` |
| "Clean up the formatting and comments in kit.go" | `golang-code-style` |
| "Is `cfg.K` a good name for the config instance?" | `golang-naming` |
| "Wrap the errors returned by DeployHetznerVPS with context" | `golang-error-handling` |
| "Should AuthentikContext be passed by value or pointer?" | `golang-structs-interfaces` |
| "Add a timeout to the HTTP call in getDefaultIdentificationStage" | `golang-context` |
| "Write a table-driven test for serverCfg in cs2.go" | `golang-testing` |
| "golangci-lint reports gosec G304 in hetzner.go" | `golang-lint` |
| "Replace the manual key loop with maps/slices helpers" | `golang-modernize` |

## Re-sync

```bash
scripts/sync-skills.sh
```

Bump `SHA` in the script to move to a newer upstream commit; edit `SKILLS` to
add or drop a vendored skill. The script replaces each listed directory and
rewrites `THIRD_PARTY_NOTICES.md`.
