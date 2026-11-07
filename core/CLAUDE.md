# core/ — Pulumi (Go)

- One binary, five stacks; `main.go` switches on `ctx.Stack()`. Add a stack there and in the justfile comment.
- Config and secrets reach Pulumi as env vars via `sops exec-env` (loaded into `cfg.K`); never add a secret to `Pulumi.<stack>.yaml` or code.
- `core/sdks/` is generated and gitignored (parameterized/Terraform-bridged providers pinned in `Pulumi.yaml`); never edit it.
- A refactor must keep resource names, inputs, options and parents identical, or preview shows replaces against live infra. Ask the user for `just core <stack> preview` after any change here.
- Never create resources inside `ApplyT`; never change a resource's logical name without an `Aliases` option.

Procedures: `homelab-pulumi-go` skill.
