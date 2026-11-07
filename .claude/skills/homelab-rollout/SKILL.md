---
name: homelab-rollout
description: "Homelab rollout procedure — how a change reaches the cluster: choosing the path (workloads/ through GitOps, a core/ Pulumi stack, or both), Pulumi stack order across talos, platform, hetzner, authentik and cloudflare, just core <stack> preview then up (run by the user), just check before push, pushing to the release branch rather than main, the cdk8s-seal-publish.yml workflow publishing <branch>-manifests, Argo CD following manifests_branch in config/config.yml, argocd_manual_sync and just argocd-sync <app>, apps needing special handling on sync (OnDelete OpenBao StatefulSet, immutable Deployment selectors, RWO volume deadlocks), post-deploy checks, and the release cut (docs ref, manifests_branch, release notes). Use whenever a task asks to deploy, ship, roll out, sync or release a change, or asks why a pushed change is not live. Do NOT use for version bumps — use homelab-upgrade. Do NOT use for writing the Go code — use homelab-cdk8s or homelab-pulumi-go. Do NOT use for a failed pulumi up — use pulumi-debug-failed-operation. Do NOT use for outage triage — use homelab-incident."
---

# Homelab rollout

Two delivery paths share one repo: Pulumi from the laptop for infrastructure, CI plus Argo CD for workloads. The pipeline is described in [concepts/gitops.mdx](../../../docs/src/content/docs/concepts/gitops.mdx); this skill is the procedure.

## Hard rules

- Never run `pulumi` or `just core`. They decrypt SOPS; ask the user to run them and paste the output.
- `just core <stack> up` passes `--yes`. Preview the same stack first, every time.
- Push to the release branch named in [CLAUDE.md](../../../CLAUDE.md), never `main`. `main` publishes `main-manifests`, which nothing syncs.
- GitOps only. No `kubectl apply/patch/edit` as a fix; Argo CD `selfHeal` reverts it. Deleting a pod or a Deployment so Argo CD recreates it from git is allowed.
- Commit or push only when the user asks.

## Pick the path

| Changed | Path | Apply |
|---|---|---|
| `workloads/**` | GitOps | push; CI publishes, Argo CD syncs |
| `config/config.yml` or `config/versions.yaml` read by workloads | GitOps | push (`config/**` triggers CI) |
| Same config read by `core/` | Pulumi | `just core <stack> preview`, then `up` |
| `core/**` | Pulumi | stack from the table below |
| `docs/**` | Docs CI | push; only release branches deploy the site |

A change in both trees: apply the Pulumi side first when workloads depend on it (CRDs, Gateway, Argo CD settings), otherwise push first. The "What to change where" table in [gitops.mdx](../../../docs/src/content/docs/concepts/gitops.mdx) maps common edits to stacks.

## Pulumi stack order

Stacks route by name in [core/main.go](../../../core/main.go).

| Stack | Needs first | Notes |
|---|---|---|
| `talos` | Proxmox | Writes `core/kubeconfig` and `core/talosconfig` |
| `platform` | `talos` (reads `core/kubeconfig`) | Cilium, Gateway, cert-manager, Argo CD, the ApplicationSet |
| `hetzner` | — | Bifrost VPS; regenerates Traefik routes from `publicServices` |
| `authentik` | Authentik running on Bifrost | OIDC apps |
| `cloudflare` | Bifrost IP in `config/config.yml` | DNS records from `publicServices` |

Public exposure change (`publicServices` in `core/cloud/cloudflare.go`): `cloudflare`, then `hetzner`.

Per stack:

1. `just lint core` and `cd core && go build ./...`.
2. User runs `just core <stack> preview`. Read it with [homelab-pulumi-go references/preview.md](../homelab-pulumi-go/references/preview.md); stop on any unexplained replace or delete.
3. User runs `just core <stack> up`.

## Workload path

1. `just check` passes (lint, tests, core build, synth, docs-check).
2. Commit and push to the release branch.
3. [cdk8s-seal-publish.yml](../../../.github/workflows/cdk8s-seal-publish.yml) runs lint, tests, synth, the Kubeflow kustomize build and `scripts/manifest-secret-guard.sh`, then force-publishes `app/` to `<branch>-manifests`. Check the run: `gh run list --workflow cdk8s-seal-publish.yml --branch <branch> --limit 3`.
4. A push touching only `core/` or `docs/` does not trigger it; rerun with `gh workflow run cdk8s-seal-publish.yml --ref <branch>` if the manifests branch must refresh.
5. Argo CD syncs the branch named by `manifests_branch` in [config/config.yml](../../../config/config.yml). ApplicationSet `cots-applications` ([core/platform/argocd.go](../../../core/platform/argocd.go)) makes one Application per top-level directory.

## Manual sync

| `argocd_manual_sync` | Behaviour |
|---|---|
| `false` | `automated: {prune, selfHeal}`; a published change syncs on its own |
| `true` | Nothing syncs until `just argocd-sync <app>` |

Toggling it changes the ApplicationSet, so it needs `just core platform up`. Use `true` for a release that must roll out app by app; set it back to `false` after.

## Apps that need a hand on sync

| App or symptom | Action | Source |
|---|---|---|
| OpenBao config change | StatefulSet is `OnDelete`: `kubectl delete pod -n openbao openbao-0` after the sync | [openbao.md](../../../docs/src/content/docs/workloads/secrets/openbao.md) |
| Sync fails with immutable `spec.selector` | Delete the Deployment once; Argo CD recreates it | [argocd.md](../../../docs/src/content/docs/platform/argocd.md) |
| Kubeflow selector change | Delete the Deployments named in the runbook | [upgrades/workloads.mdx](../../../docs/src/content/docs/operate/upgrades/workloads.mdx) |
| New pod stuck on an RWO volume (Harbor) | Scale the old ReplicaSet to zero | [repairs.md](../../../docs/src/content/docs/operate/repairs.md) |
| StatefulSet permanently OutOfSync | Add the field to the global `apps_StatefulSet` rule in `argocd.go` | [argocd.md](../../../docs/src/content/docs/platform/argocd.md) |
| Grafana org change | `just grafana-init` before the change syncs | [repairs.md](../../../docs/src/content/docs/operate/repairs.md) |
| Anything else stuck | Load `homelab-incident` | [troubleshooting.md](../../../docs/src/content/docs/operate/troubleshooting.md) |

## Release cut

1. Create branch `vX.Y.Z` and push it. The `create` trigger publishes `vX.Y.Z-manifests` even without a workload change.
2. Confirm the branch exists: `git ls-remote origin vX.Y.Z-manifests`.
3. Move together, in one commit: `ref` in [docs/astro.config.mjs](../../../docs/astro.config.mjs), `manifests_branch` in `config/config.yml`, the branch named in `CLAUDE.md`, and any changed inventory row. See [docs-site.md](../../../docs/src/content/docs/reference/development/docs-site.md).
4. Release notes: full notes in `docs/releases/vX.Y.Z-RELEASE-NOTES.md`, condensed page in `docs/src/content/docs/releases/` with its card in `releases/index.mdx`. Load `homelab-docs`.
5. User runs `just core platform preview`, then `up`, to repoint the ApplicationSet. Never before step 2: a missing branch drops every Application's source at once.

## Verify

```bash
just check
gh run list --workflow cdk8s-seal-publish.yml --branch <branch> --limit 1   # completed, success
kubectl get applications -n argocd                    # Synced + Healthy
kubectl get application <app> -n argocd -w            # one app after just argocd-sync
kubectl get pods -n <app>                             # Running, no restarts climbing
kubectl get httproute -n <app>                        # Accepted on the Gateway
curl -sI https://<host>.madhan.app                    # route answers
just talos-health                                     # after any talos or platform up
```

A new HTTPRoute with no status after a control-plane restart is a known Cilium operator issue; hand it to `homelab-incident`.
