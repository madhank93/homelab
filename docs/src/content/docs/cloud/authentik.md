---
title: Authentik
description: SSO on Bifrost — GitHub login, OIDC for Grafana and NetBird, and ForwardAuth for public cluster routes.
sidebar:
  order: 30
---

Authentik on Bifrost is the single sign-on for everything public: users log in with GitHub, Grafana and NetBird use it over OIDC, and Traefik gates other public routes with ForwardAuth. [Authentik docs](https://docs.goauthentik.io/).

| | |
|---|---|
| URL | `https://auth.madhan.app` (always public) |
| Runs on | [Bifrost](/cloud/bifrost/), started by `just core hetzner up` |
| Configured by | Pulumi stack `authentik`, [`authentik.go`](src:core/cloud/authentik.go), `just core authentik up` |
| Secrets | `AUTHENTIK_TOKEN`, `AUTHENTIK_GITHUB_SECRET`, `GRAFANA_OAUTH_CLIENT_SECRET`, `NETBIRD_CLIENT_SECRET` (SOPS) |

The container comes up on the Hetzner stack; the `authentik` stack then configures it through its API with `AUTHENTIK_TOKEN`, which is also the bootstrap token Authentik creates for `akadmin` on first boot.

## Configuration

### GitHub source

| Setting | Value | Why |
|---|---|---|
| Client ID | `Ov23liUPVh4nPuUJzGFp` | GitHub OAuth app |
| PKCE | S256 | |
| User matching | `identifier` | `email_link` loops when the admin's email matches a GitHub account that is not logged in |
| Login page | Bound to `default-authentication-identification` | Shows the **Login with GitHub** button |

### OIDC applications

| App | Client ID | Redirect URIs | Extra scopes |
|---|---|---|---|
| Grafana | `grafana-homelab` | `https://grafana.madhan.app/login/generic_oauth` | `groups` |
| NetBird | `aumenijDycfG1cQURqH9BNJpV3KVUCoMHGPUVUlT` | `https://netbird.madhan.app/oauth2/callback`, `http://localhost:53000` | — |

Both are confidential clients signed with the default self-signed key; access tokens last 1 hour, authorization codes 10 minutes. Every provider gets `openid`, `email`, `profile`, `offline_access` and a custom empty `api` scope that NetBird requests.

The `groups` scope returns the user's group names. Grafana's `org_mapping` makes members of `grafana-admins` Admin of the main org; everyone else is a Viewer.

The NetBird app is the upstream for NetBird's embedded Dex, not NetBird itself; the connector is added by hand in the NetBird UI ([setup](/cloud/netbird/)). Pulumi also creates service account `sa-netbird` (in `authentik Admins`) with a non-expiring token, exported as `NetbirdServiceToken`.

### ForwardAuth

The `Homelab ForwardAuth` proxy provider runs in `forward_domain` mode with cookie domain `.madhan.app`, so one login covers every subdomain. Traefik's `authentik-forwardauth` middleware calls `authentik-server:9000/outpost.goauthentik.io/auth/nginx`. Every route generated from `publicServices` uses it unless the entry sets `SkipAuth: true`; see [Exposure](/cloud/exposure/).

## Operate

```sh frame="terminal"
just core authentik up    # idempotent; re-run after drift or a failed run
cd core && pulumi stack select authentik && pulumi stack output NetbirdServiceToken --show-secrets
```

Grant Grafana admin: **Directory → Groups → grafana-admins → Add user**.

## Troubleshooting

### Grafana login works but the user has no admin rights

The user is not in `grafana-admins`. Add them, then log in again.

### ForwardAuth redirects in a loop

Check Authentik is healthy with `ssh root@178.156.199.250 'docker exec authentik-server ak healthcheck'`, then confirm the outpost in **Applications → Outposts**.
