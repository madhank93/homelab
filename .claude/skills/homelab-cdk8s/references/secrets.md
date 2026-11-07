# OpenBao secrets through the Secrets Store CSI driver

Runtime secrets live in OpenBao KV v2 at `secret/<app>`. Pods read them through the Secrets Store CSI driver (`workloads/secrets/csi_driver.go`, provider `openbao`). The rendered manifests contain only references, never values.

## Wiring a new secret

| Step | Where | What |
|---|---|---|
| 1 | `scripts/openbao-setup.sh` `roles=` | Add `<app>:<serviceaccount>`. The role binds that service account in namespace `<app>` to a read-only policy on `secret/data/<app>` |
| 2 | `scripts/openbao-setup.sh` `seed` lines | Seed each key: `$(rand)` for generated values, `${VAR:?...}` for values that come from SOPS via `sops exec-env` |
| 3 | Chart | `kit.SecretProviderClass(chart, "<app>-secrets", "<app>", []string{"KEY"}, "<k8s-secret>", kit.Sync{...})`; object names must match the seeded keys |
| 4 | Chart | Mount the CSI volume in the consuming pod (`kit.SecretsVolume`/`SecretsMount`, the `*Values` forms, or `kit.SecretSync`) |
| 5 | Pod template | Add `reloader.stakater.com/auto: "true"` so a rotated value rolls the pod |
| 6 | User | `just openbao-setup` (needs cluster access and SOPS; ask the user to run it) |

The role name, the KV path and the namespace are all `<app>`; keep them equal or the Kubernetes auth login is denied.

## Choosing the delivery

| Pattern | How the app reads it | Use when | Example |
|---|---|---|---|
| File-only | `/mnt/secrets/<KEY>` | App supports reading the value from a file | `monitoring/grafana.go` `ADMIN_PASSWORD` via `$__file{}` |
| Synced Secret | env `valueFrom.secretKeyRef`, `envFrom`, or chart `existingSecret` | App only takes env vars or an existing Secret name | `automation/n8n.go`, `registry/harbor.go` |
| Secret-sync pod | Same as synced, but the mount lives in a pause pod | Chart cannot add volumes to its pods | `kit.SecretSync` in `registry/harbor.go`, `security/falco.go` |
| Operator-owned | Operator writes the Secret | CNPG, cert-manager and similar generate the credential | `n8n-pg-app` in `automation/n8n.go` |

Prefer file-only: the value never becomes a Kubernetes Secret object at all. Check the image actually supports it — the distroless Grafana image ignores `GF_*__FILE`, which is why it uses `$__file{}` instead.

## Rules

- The driver creates the mirrored Secret only while some pod mounts the CSI volume; no mount, no Secret, and every `secretKeyRef` to it fails.
- `kit.Sync{Object, Key}` renames on the way: `Object` is the OpenBao key, `Key` is the key in the Kubernetes Secret (n8n maps `ENCRYPTION_KEY` to `N8N_ENCRYPTION_KEY`).
- Rotation is enabled with a 2m poll (`enableSecretRotation`, `rotationPollInterval` in `csi_driver.go`); Reloader then restarts annotated pods when the Secret changes.
- After adding an object to an existing SecretProviderClass, restart the pod that mounts it so the volume is mounted again and the mirrored Secret gains the key. Do not wait on the rotation poll for structural changes.
- Embed a secret in a config string through kubelet `$(VAR)` expansion, never by templating the value into Go: Harbor's `CONFIG_OVERWRITE_JSON` carries `$(OIDC_CLIENT_SECRET)`.
- Keep `SecretProviderClass` a raw `cdk8s.NewApiObject` (no typed import exists); add fields in `kit.SecretProviderClass`, not per chart.
- Never generate a Secret in Go, never set a chart's inline password value, and blank chart defaults that would render one (Harbor sets `harborAdminPassword: ""` with `existingSecretAdminPassword`).
- `scripts/manifest-secret-allowlist.txt` lists existing chart-rendered Secrets being migrated; never add to it.
