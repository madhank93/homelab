---
name: homelab-sso
description: "Homelab SSO through Authentik — adding or changing an OIDC application with createOIDCApp in core/cloud/authentik.go (OIDCApp fields, client ID, slug, redirect URIs, GrantTypes, the groups scope, AccessGroup policy binding to grafana-admins), the grafana_admins list in config/config.yml, where each client secret travels (SOPS <APP>_OAUTH_CLIENT_SECRET to Pulumi, or through the just openbao-setup seed into OpenBao and the CSI driver), the per-app integrations in use (Grafana generic_oauth, Homepage OIDC env, Argo CD oidc.config with a Pulumi-made Secret, Harbor CONFIG_OVERWRITE_JSON, Headlamp plus kube-apiserver oidc-* flags in the Talos control-plane patch, the Proxmox authentik OpenID realm), the Traefik ForwardAuth provider, and checking the login handoff end to end (invalid_request, redirect_uri mismatch, user has no admin role). Use whenever a task adds SSO to an app, changes who may log in, rotates an OAuth client secret or debugs an Authentik login. Do NOT use for general Pulumi resource options or preview reading — use homelab-pulumi-go. Do NOT use for mounting the secret into a pod or kit helpers — use homelab-cdk8s. Do NOT use for reviewing exposure or auth posture — use homelab-security-review."
---

# Homelab SSO

Authentik on Bifrost is the only identity provider; users enroll with GitHub. Every OIDC app is declared in `core/cloud/authentik.go` (stack `authentik`); the app side lives wherever the app is deployed. The component page [authentik.md](../../../docs/src/content/docs/cloud/authentik.md) holds the current app table; keep it in step.

## Hard rules

- Never read or print a client secret. A new one goes into the SOPS bootstrap file; ask the user to add it.
- The user runs `just core <stack> preview`, then `up`, and `just openbao-setup`. Never run them.
- Bind every new admin-facing app to `grafana-admins` with `AccessGroup`: any GitHub account can enroll in Authentik, so an unbound app admits strangers.
- Keep each app's local admin login working as the break-glass path.
- Never render the client secret into a manifest; it arrives through OpenBao and CSI, or kubelet `$(VAR)` expansion.

## Add an OIDC app

1. **Pick names.** `Slug` sets the issuer `https://<auth host>/application/o/<slug>/`. `ClientID` names the Pulumi resources `provider-<ClientID>`, `app-<ClientID>` and `binding-<ClientID>`, so changing it replaces them. Convention: client ID `<app>-homelab`, secret `<APP>_OAUTH_CLIENT_SECRET`.
2. **Secret.** Ask the user to add `<APP>_OAUTH_CLIENT_SECRET` (a random string) to the SOPS file; Authentik stores whatever Pulumi sends.
3. **Authentik side.** Admin UIs that map `groups` to a role join the `{name, slug, clientID, secret, base, callback}` loop in `DeployAuthentik`; anything else gets its own `createOIDCApp(ac, OIDCApp{...})` call.

| Field | Set it to |
|---|---|
| `ClientType` | `"confidential"` (every app here is) |
| `Redirects` | Exact callback URL; a `*` switches that entry to regex matching |
| `GrantTypes` | Empty for `authorization_code` + `refresh_token`; to add one, list all (NetBird adds `urn:ietf:params:oauth:grant-type:device_code`) |
| `ExtraPropertyMappings` | `groupsMapping.ID().ToStringOutput()` when the app maps groups to roles |
| `AccessGroup` | `grafanaAdminsGroup.ID().ToStringOutput()`; creates the policy binding |

4. **Secret to the app.** Pick the route in the next section.
5. **App side.** Set issuer, client ID and scopes; copy the closest pattern in [references/apps.md](references/apps.md).
6. **Docs.** Add the app to the OIDC table and Secrets row in `authentik.md`, and an Auth row on the app's own page (homelab-docs).
7. Run Verify.

## Where the client secret goes

| App runs | Route | Example |
|---|---|---|
| In-cluster, synced by Argo CD | SOPS → `seed <app> <KEY> "${<APP>_OAUTH_CLIENT_SECRET:?...}"` in `scripts/openbao-setup.sh` → `secret/<app>` → `kit.SecretProviderClass` + `kit.Sync` | Grafana, Homepage, Harbor, Headlamp |
| Argo CD, which deploys OpenBao | SOPS → `corev1.NewSecret` in Pulumi, value wrapped in `pulumi.ToSecret` | `argocd-oidc` in `core/platform/argocd.go` |
| Outside the cluster, Pulumi-managed | SOPS → resource input wrapped in `pulumi.ToSecret` | Proxmox realm `ClientKey` |

For the OpenBao route, also add `<app>:<serviceaccount>` to `roles=` in `openbao-setup.sh` when the app has no role yet. `seed` never overwrites an existing value, so rotating a secret means patching OpenBao as well as SOPS. Mount, sync and rotation rules: [homelab-cdk8s secrets](../homelab-cdk8s/references/secrets.md).

## Who gets in, and as what

- Login: membership of the `AccessGroup` (policy binding, order 0).
- Admin: the app maps the `groups` claim value `grafana-admins` to its own admin role.
- Add a person: append their email to `grafana_admins` in `config/config.yml` after their first GitHub login (`lookupGrafanaAdmins` warns and skips unknown emails), then `just core authentik up`.

## Rollout order

1. `just core authentik preview`, then `up`: the provider exists before any app redirects to it.
2. `just openbao-setup` when a `seed` line or role was added.
3. The stack holding the app side: `platform` (Argo CD), `talos` (Proxmox realm, kube-apiserver flags), or a push to the release branch for workloads ([CLAUDE.md](../../../CLAUDE.md) deploy flow).
4. Restart the app pod if it mounted the SecretProviderClass before the key existed.

## Symptoms

| Symptom | Cause | Fix |
|---|---|---|
| `invalid_request` on every login | Provider lists no grant types | Confirm `GrantTypes` reached the provider; re-run `just core authentik up` |
| `redirect_uri` mismatch | App sends another scheme, host, port or path | Match `Redirects` exactly; Grafana needs `server.root_url` for https; Proxmox sends its origin with `:8006` |
| Authentik denies access | User not in `grafana-admins` | Add to `grafana_admins`, re-run the authentik stack |
| Logged in, no admin rights | `groups` scope missing on the provider or in the app's request | Add `groupsMapping`; request `groups` in the app |
| Grafana "User sync failed" | Authentik user recreated with a new `sub` | Keep `oauth_allow_insecure_email_lookup` in `grafana.go` |
| Headlamp signs in, API returns 401/403 | kube-apiserver `oidc-*` flags missing, or no RBAC for `oidc:<group>` | Check the control-plane patch in `talos.go` and the `oidc-grafana-admins` ClusterRoleBinding |

## Verify

1. `just lint core` and `cd core && go build ./...`; for workload changes `just synth` and check `app/<dir>/` has the env and SecretProviderClass but no `Secret`.
2. `just check` before pushing.
3. Ask the user for `just core authentik preview`: expect creates of `provider-<ClientID>`, `app-<ClientID>` and, when bound, `binding-<ClientID>`. A replace of an existing provider is a stop.
4. Issuer answers: `curl -s https://auth.madhan.app/application/o/<slug>/.well-known/openid-configuration | jq .issuer`.
5. Log in through the app in a private window as a `grafana-admins` member (admin role) and, if possible, as a non-member (refused by Authentik).
6. Confirm the local admin login still works.
