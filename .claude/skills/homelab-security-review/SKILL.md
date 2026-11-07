---
name: homelab-security-review
description: "Homelab security review checklist for a change or the whole repo — rendered Secret payloads (scripts/manifest-secret-guard.sh, scripts/manifest-secret-allowlist.txt that may only shrink, TestNoSecretPayloads), the block-secrets.sh PreToolUse hook, gitleaks and trufflehog via .githooks and scripts/secret-scan.sh, OpenBao root-token handling in just openbao-root and scripts/openbao-setup.sh, internet exposure through publicServices in core/cloud/cloudflare.go and the Bifrost Traefik/ForwardAuth path, Authentik OIDC apps and their AccessGroup bindings, cluster-admin bindings (Headlamp), per-app OpenBao roles, kit.Namespace privileged namespaces, image pinning (TestImagesPinned), GitHub Actions SHA pinning, workflow permissions and the SBOM workflow. Use whenever a task asks to security-review, audit or harden the homelab, makes a service public, or adds a ServiceAccount, ClusterRoleBinding, OpenBao role, privileged namespace, workflow or just recipe that handles credentials. Do NOT use for writing the fix — use homelab-cdk8s or homelab-pulumi-go. Do NOT use for adding an SSO app — use homelab-sso. Do NOT use for Go code-level vulnerabilities — use golang-security. Do NOT use for a line-by-line diff review — use differential-review. Do NOT use for dependency risk — use supply-chain-risk-auditor."
---

# Homelab security review

The repo is public and `app/` is published to a public `<branch>-manifests` branch, so anything rendered is world-readable. Repo rules (no secret values, GitOps only, Pulumi from the laptop) are in [CLAUDE.md](../../../CLAUDE.md); the secrets model is in [concepts/secrets.mdx](../../../docs/src/content/docs/concepts/secrets.mdx); exposure in [cloud/exposure.mdx](../../../docs/src/content/docs/cloud/exposure.mdx); SSO in [cloud/authentik.md](../../../docs/src/content/docs/cloud/authentik.md).

Review code, not live values: never read `secrets/`, SOPS files, env files, kubeconfig or talosconfig, and never run a recipe that prints a credential.

## Scope the review

| Request | Do |
|---|---|
| Review a diff or branch | `git diff main...HEAD --stat`, then run only the areas whose files changed (table below) |
| Review the repo | Run every area in [references/checks.md](references/checks.md) in order |
| Make a service public | Exposure and Auth areas, before the user runs `just core hetzner up` |

| Changed path | Areas |
|---|---|
| `workloads/**` | Rendered Secrets, RBAC, Pod Security, Images |
| `scripts/manifest-secret-allowlist.txt`, `scripts/manifest-secret-guard.sh` | Rendered Secrets |
| `scripts/openbao-setup.sh`, `kit.SecretProviderClass` callers | OpenBao roles |
| `core/cloud/cloudflare.go`, `core/cloud/hetzner.go`, `core/cloud/bifrost/**` | Exposure |
| `core/cloud/authentik.go` | Auth |
| `justfile`, `.claude/hooks/**`, `.githooks/**`, `scripts/secret-scan.sh`, `.gitleaks.toml` | Secrets tooling, OpenBao root token |
| `.github/workflows/**` | CI, SBOM |

## Areas

Exact commands and failure signatures: [references/checks.md](references/checks.md).

| Area | Pass when |
|---|---|
| Rendered Secrets | `manifest-secret-guard.sh app` prints `OK`; the allowlist diff has no added lines; no `NewKubeSecret` outside `imports/` |
| Secrets tooling | `core.hooksPath` is `.githooks`; scanner images pinned by digest; every recipe or command that prints a credential is matched by `.claude/hooks/block-secrets.sh`; `.gitleaks.toml` allowlists only vendored example files |
| OpenBao root token | Minted only by `just openbao-token`, passed on stdin (never argv); revoked by the `EXIT` trap in `openbao-root` / `openbao-setup.sh`, or by `just openbao-revoke` when kept for the UI |
| OpenBao roles | One `app:sa` pair per app in `roles=`; the SA is the app's own, not `default` or shared across apps; policy is `read` on `secret/data/<app>` only; `TestSecretProviderClassRoles` passes |
| Exposure | Every `publicServices` entry is intended; `SkipAuth: true` only for an app that runs its own Authentik login; the Hetzner firewall opens only documented ports |
| Auth | Every `OIDCApp` that grants admin or data access sets `AccessGroup`; routes behind ForwardAuth are acceptable for any enrolled user |
| RBAC | Every `cluster-admin` binding is listed and justified; subjects are a named SA or the `oidc:grafana-admins` group |
| Pod Security | `kit.Namespace(..., true)` only for host-access workloads, each with a comment naming the host access |
| Images | `TestImagesPinned` passes; images in CR `{repository, tag}` fields checked by hand |
| CI | Every `uses:` pinned to a 40-char SHA with a version comment; top-level `permissions: contents: read`; each job-level write commented; no `pull_request_target` |
| SBOM | `.github/workflows/sbom.yml` runs on `release` and attaches the CycloneDX file |

## Report findings

One line per finding: `path:line — severity — problem — owner skill`.

| Severity | Meaning |
|---|---|
| Critical | Secret value in git, `app/` or the transcript; unauthenticated public admin UI |
| High | Public route any GitHub account can pass to admin access; cluster-admin to an unbound identity; unpinned Action |
| Medium | Over-broad OpenBao role or SA; privileged namespace without host access; hook gap for a credential-printing recipe; SSH open to the world |
| Low | Missing justification comment; tag-only pin where a digest is used elsewhere |

Cite the file and key name, never the value.

## Hand off

| Finding type | Skill |
|---|---|
| Fix in `workloads/` | homelab-cdk8s |
| Fix in `core/` | homelab-pulumi-go |
| New or changed SSO app | homelab-sso |
| Go code vulnerability (injection, crypto, unsafe input) | golang-security |
| Line-by-line diff review | differential-review |
| Dependency, chart or image provenance | supply-chain-risk-auditor |

## Verify

```bash
just check                               # lint, tests incl. policy tests, core build, synth, docs-check
./scripts/manifest-secret-guard.sh app   # after just synth
(cd workloads && go test -run 'TestNoSecretPayloads|TestImagesPinned|TestSecretProviderClassRoles|TestHTTPRoutesUseSharedGateway' .)
grep -nE 'uses: ' .github/workflows/*.yml | grep -vE '@[0-9a-f]{40}'   # must print nothing
just secret-scan-history                 # whole history, findings redacted; needs Docker
```
