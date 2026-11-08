# homelab

Talos Kubernetes on Proxmox, built entirely from this repo. Public GitHub repo.

## Map

| Path | What | Entry |
|---|---|---|
| `core/` | Pulumi (Go): stacks `talos`, `platform`, `hetzner`, `authentik`, `cloudflare` | `core/main.go` routes by stack name |
| `workloads/` | CDK8s (Go): every in-cluster app | `workloads/main.go` `apps` table → `app/<dir>` |
| `config/` | Shared values for core, workloads and the justfile (`yq`) | `config/config.yml`, `config/versions.yaml` |
| `secrets/` | SOPS-encrypted bootstrap values | never read it (see Rules) |
| `scripts/` | Helpers run through `just`, never directly | `scripts/README.md` |
| `docs/` | Astro Starlight site | `docs/src/content/docs/` |
| `app/` | Synth output, gitignored; CI publishes it to `<branch>-manifests` | — |

`core/CLAUDE.md` and `workloads/CLAUDE.md` hold module-specific rules.

## Rules

- **No secret values in git, the transcript or the manifests.** Never read or decrypt `secrets/`, `*.env`, kubeconfig or talosconfig, and never print Secret or OpenBao data. Runtime secrets live in OpenBao. New code renders no `Secret`; the existing chart-rendered ones are in `scripts/manifest-secret-allowlist.txt`, which may only shrink. `.claude/hooks/block-secrets.sh` enforces this.
- **GitOps only.** No manual `kubectl apply/patch/edit` or host changes as a fix; change the code. Restarting a pod to pick up synced config is fine.
- **Pulumi runs from the laptop, never CI**, and always `just core <stack> preview` before `up`. `just core` decrypts SOPS, so the user runs it.
- **Check upstream GitHub issues before any version bump.** Versions are pinned in `config/versions.yaml`, `workloads/cdk8s.yaml` and the files listed in `docs/src/content/docs/reference/software-inventory.md`.
- **Docs state no versions** outside the software inventory; `just docs-check` fails otherwise.
- Comments state durable intent in 1–3 lines; no history of how the code got here.

## Done means

```bash
just check   # lint (golangci-lint, .golangci.yml) + tests + core build + synth + docs-check
```

- Lint config is shared by all modules; generated `core/sdks/` and `workloads/imports/` are excluded.
- `just fmt` applies gofumpt + goimports; the PostToolUse hook already does this per edited file.
- Fix lint findings properly. `//nolint:<linter> // <reason>` only for a real false positive, on the line.
- CI (`.github/workflows/cdk8s-seal-publish.yml`) gates publishing on lint + tests for `config/` and `workloads/`. `core/` is checked only locally because its SDKs are generated on the laptop.

## Deploy flow

1. Commit and push to the release branch (`v0.1.8` today), not `main`.
2. CI synthesizes `workloads/` and publishes `v0.1.8-manifests`.
3. Argo CD syncs from `manifests_branch` in `config/config.yml`.
4. Infra changes: `just core <stack> preview`, then `up` (user runs both).

## Skills

Before writing, reviewing or answering a question about Go code, load the `golang-*` skill that matches the topic (error handling, naming, safety, testing, …), even for a one-line answer. They are vendored from samber/cc-skills-golang by `scripts/sync-skills.sh`; don't edit them by hand. Repo conventions: `homelab-cdk8s`, `homelab-pulumi-go`. Index and trigger table: `.claude/skills/README.md`.

Agents (`.claude/agents/`): `cluster-inspector` for read-only live-cluster diagnosis, `change-reviewer` for a diff review before commit. `just skill-check` verifies the README trigger table.
