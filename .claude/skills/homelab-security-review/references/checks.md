# Security checks

Run from the repo root. Each command is followed by `# fail:` — what a failing result looks like.

## Rendered Secrets

```bash
just synth && ./scripts/manifest-secret-guard.sh app
# fail: "::error file=...::Secret <ns>/<name> has data/stringData", or
#       "<entry> is allowlisted but renders no Secret payload" (remove the stale entry)
git diff main...HEAD -- scripts/manifest-secret-allowlist.txt | grep '^+[^+]'
# fail: any output — the allowlist may only shrink
(cd workloads && go test -run TestNoSecretPayloads .)
# fail: "Secret ... has data/stringData and would be published"
grep -rn 'NewKubeSecret' workloads --include='*.go' | grep -v imports/
# fail: any output — source the value from OpenBao (homelab-cdk8s)
```

## Secrets tooling

```bash
git config core.hooksPath
# fail: not .githooks — fix with `just hooks`
grep -nE '^(GITLEAKS|TRUFFLEHOG)=' scripts/secret-scan.sh
# fail: an image without @sha256:
grep -n -A2 '^secret-scan-history' justfile
# fail: tag-only images while secret-scan.sh pins digests
cat .gitleaks.toml
# fail: an allowlist path outside vendored skill examples, or a regexes allowlist
grep -nE 'kubectl create token|-field=token|-field=' justfile
# fail: a recipe that prints a token or secret value whose name is missing from
#       the just[[:space:]]+(...) list in .claude/hooks/block-secrets.sh
jq '.permissions.deny' .claude/settings.json
# fail: kubeconfig, talosconfig, SOPS or env read patterns missing
git check-ignore -v scripts/kubeconfig scripts/talosconfig
# fail: no match — the file is committable
```

## OpenBao root token

```bash
grep -n -A7 '^openbao-root' justfile
grep -n 'trap\|ROOT_TOKEN=' scripts/openbao-setup.sh
# fail: token in argv or an exported env var, or no `token revoke -self` in an EXIT trap
grep -n 'openbao-token\|bao_root' justfile scripts/*.sh
# fail: a new caller that does not revoke, or echoes the token
```

## OpenBao roles

```bash
grep -n '^roles=' scripts/openbao-setup.sh
# fail: SA `default` (every pod in the namespace can read the app's secrets),
#       or one SA name shared by unrelated apps
grep -n -A3 'policy write' scripts/openbao-setup.sh
grep -n 'capabilities' scripts/openbao-setup.sh
# fail: a path other than secret/data/<app>, a wildcard, or a capability beyond read
(cd workloads && go test -run TestSecretProviderClassRoles .)
# fail: "roleName ... is not a role", wrong namespace, or wrong ServiceAccount
```

## Exposure

```bash
grep -n -A6 'var publicServices' core/cloud/cloudflare.go
# fail: an admin UI listed, or SkipAuth: true on an app without its own Authentik login
grep -n 'range \[\]string{' core/cloud/cloudflare.go
# fail: an always-public name not in the table of cloud/exposure.mdx
grep -n -B2 -A4 'SkipAuth' core/cloud/hetzner.go
# fail: authentik-forwardauth no longer added for entries without SkipAuth
grep -n -A8 'hcloud.FirewallRuleArgs' core/cloud/hetzner.go
# fail: a port beyond 22, 80, 443, 3478/udp; SSH open to 0.0.0.0/0 is a finding to record
grep -rn 'LoadBalancer' workloads --include='*.go' | grep -v imports/
# fail: an LB IP not in the reachability table of cloud/exposure.mdx
```

Reference: [cloud/exposure.mdx](../../../../docs/src/content/docs/cloud/exposure.mdx).

## Auth

```bash
grep -n 'AccessGroup:\|Slug:' core/cloud/authentik.go
# fail: an app granting admin or data access without AccessGroup — any GitHub account can enroll
grep -n -A30 'func createHomelabForwardAuth' core/cloud/authentik.go | grep -n 'NewPolicyBinding'
# fail: no output means every enrolled user passes ForwardAuth, so a public service
#       without SkipAuth is open to any GitHub account
grep -n 'grafana_admins' config/config.yml
# fail: an unexpected member
```

## RBAC and service accounts

```bash
grep -rn '"cluster-admin"' workloads core --include='*.go' | grep -v -e imports/ -e sdks/
# fail: a binding to anything other than headlamp-admin or oidc:grafana-admins without a stated reason
grep -n -A2 '^headlamp-token' justfile
# fail: a longer default duration, or the token stored in a Secret
grep -rn '"verbs"' workloads --include='*.go' | grep -v imports/
# fail: "*" verbs or resources, or cluster-wide writes where a namespaced Role fits
grep -rn 'rbacv1.NewClusterRoleBinding' core --include='*.go' | grep -v sdks/
# fail: a new binding without a narrow ClusterRole
```

## Pod Security

```bash
grep -rn -B1 'kit.Namespace(.*true)' workloads --include='*.go'
# fail: a workload with no hostPath, hostNetwork, hostPID, kernel module or device need,
#       or a new entry without a comment naming the host access
grep -rnE 'Privileged|hostNetwork|hostPID' workloads --include='*.go' | grep -v imports/
# fail: host access in a namespace not marked privileged, or host access it does not need
```

## Images

```bash
(cd workloads && go test -run 'TestImagesPinned|TestPinned' .)
# fail: "image ... has no tag/digest or uses :latest"
grep -n 'latest' config/versions.yaml workloads/cdk8s.yaml
# fail: any output — TestImagesPinned does not see CR {repository, tag} fields
```

## CI

```bash
grep -nE 'uses: ' .github/workflows/*.yml | grep -vE '@[0-9a-f]{40}'
# fail: any output
grep -n -A3 'permissions:' .github/workflows/*.yml
# fail: top level not contents: read; a job-level write without a comment;
#       id-token or packages write without need
grep -nE 'pull_request_target|workflow_run' .github/workflows/*.yml
# fail: any output
grep -n 'manifest-secret-guard\|go test\|actions-gh-pages' .github/workflows/cdk8s-seal-publish.yml
# fail: guard or tests missing, or after the publish step
```

## SBOM

```bash
grep -n -A3 '^on:' .github/workflows/sbom.yml
grep -n 'format:' .github/workflows/sbom.yml
# fail: no release trigger, or a format other than cyclonedx-json
```
