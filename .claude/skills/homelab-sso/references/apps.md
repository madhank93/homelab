# Per-app integration patterns

Every app below uses the issuer `https://<auth host>/application/o/<slug>/` built from `config.C.Host("auth")`, never a hard-coded domain.

| App | Slug / client ID | App-side code | Secret reaches the app as | Admin mapping |
|---|---|---|---|---|
| Grafana | `grafana` / `grafana-homelab` | `auth.generic_oauth` in `workloads/monitoring/grafana.go` | `kit.Sync` key `GF_AUTH_GENERIC_OAUTH_CLIENT_SECRET` into `grafana-oauth-secret`, loaded by `envFromSecret` | `org_mapping` + `role_attribute_path` on `groups` |
| Homepage | `homepage` / `homepage-homelab` | `HOMEPAGE_OIDC_*` env in `workloads/management/homepage.go` | `secretEnv("HOMEPAGE_OIDC_CLIENT_SECRET", "OAUTH_CLIENT_SECRET")` from `homepage-secrets` | None; `AccessGroup` is the whole gate |
| Argo CD | `argocd` / `argocd-homelab` | `configs.cm.oidc.config` and `configs.rbac` in `core/platform/argocd.go` | `$argocd-oidc:clientSecret`, from the Pulumi Secret `argocd-oidc` | `policy.csv`: `g, grafana-admins, role:admin`; `policy.default` empty |
| Harbor | `harbor` / `harbor-homelab` | `CONFIG_OVERWRITE_JSON` on `core.extraEnvVars` in `workloads/registry/harbor.go` | `OIDC_CLIENT_SECRET` env from `harbor-admin`, referenced as `$(OIDC_CLIENT_SECRET)` inside the JSON | `oidc_admin_group: grafana-admins` |
| Headlamp | `kubernetes` / `kubernetes` | `config.oidc.externalSecret` + `OIDC_*` env in `workloads/management/headlamp.go` | `kit.Sync` into `headlamp-oidc` | ClusterRoleBinding `oidc-grafana-admins` → `cluster-admin` |
| kube-apiserver | (same `kubernetes` client) | `cluster.apiServer.extraArgs` `oidc-*` in the control-plane patch, `core/platform/talos.go` | None; validates ID tokens only | Groups arrive as `oidc:<group>` |
| Proxmox | `proxmox` / `proxmox-homelab` | `realm.NewOpenid` `authentik` in `NewProxmoxSSO`, `core/platform/proxmox.go` | `ClientKey` from SOPS via Pulumi | Synced group `grafana-admins-authentik` holds `Administrator` on `/` |
| NetBird | `netbird` / `aumenijDycfG1cQURqH9BNJpV3KVUCoMHGPUVUlT` | Embedded Dex connector, added by hand in the NetBird UI ([netbird.mdx](../../../../docs/src/content/docs/cloud/netbird.mdx)) | Entered in the NetBird UI | None |

## Pattern notes

- **Grafana** uses the shared endpoints (`/application/o/authorize/`, `token/`, `userinfo/`) rather than the slug issuer, and needs `server.root_url` so its redirect URI is https behind TLS termination. Form login is disabled; break-glass is in the comment beside `disable_login_form`.
- **Argo CD** cannot read from OpenBao because it deploys OpenBao. The Secret carries the label `app.kubernetes.io/part-of: argocd`, which `argocd-cm` `$secret:key` references require. Lives in stack `platform`.
- **Harbor** stores auth settings in its database; `CONFIG_OVERWRITE_JSON` makes `harbor-core` re-apply them on every start. Build the JSON with `json.Marshal` and leave the secret as a `$(VAR)` reference. Local admin stays at `/account/sign-in`.
- **Headlamp** passes the user's ID token straight to the API server, so the client ID must equal the apiserver's `oidc-client-id` (`kubernetes`). Both the Talos patch (stack `talos`) and the workload change are needed. Fallback: `just headlamp-token`.
- **Proxmox** redirects to its own origin with `:8006` and no path, hence the empty callback in the loop. Users are auto-created from `preferred_username`; groups sync with the `-authentik` suffix. Lives in stack `talos`. `root@pam` is the fallback.
- **ForwardAuth** (`createHomelabForwardAuth`) is not per app: Traefik gates every `publicServices` entry in `core/cloud/cloudflare.go` unless it sets `SkipAuth: true`. Set `SkipAuth` for an app with its own OIDC login to avoid a second prompt (Grafana does).

App pages: [argocd.md](../../../../docs/src/content/docs/platform/argocd.md), [harbor.mdx](../../../../docs/src/content/docs/workloads/registry/harbor.mdx), [headlamp.md](../../../../docs/src/content/docs/workloads/management/headlamp.md), [homepage.md](../../../../docs/src/content/docs/workloads/management/homepage.md).
