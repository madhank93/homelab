+++
title = "Authentik"
description = "OIDC identity provider running on the Bifrost VPS — GitHub SSO, ForwardAuth, and OIDC for Grafana and NetBird."
weight = 35
+++

## Overview

[Authentik](https://goauthentik.io/) is the identity provider for the homelab. It runs on the Bifrost Hetzner VPS (not inside the cluster) and is managed by Pulumi (`core/cloud/authentik.go`, stack: `authentik`).

Authentik handles three things:

1. **GitHub OAuth login** — users authenticate with their GitHub account
2. **OIDC for cluster apps** — Grafana and NetBird use Authentik as their OIDC provider
3. **ForwardAuth for public services** — Traefik's `authentik-forwardauth` middleware gates public cluster services behind a session cookie

---

## How It's Deployed

Authentik runs as two containers (`authentik-server` and `authentik-worker`) in `docker-compose.yml` on the Bifrost VPS, managed automatically by `just core hetzner up`.

Managed separately by Pulumi's `authentik` stack — configures OIDC apps, GitHub source, scopes, and ForwardAuth outpost against the live Authentik API:

```bash
just core authentik up
```

See [Hetzner Bifrost](/infrastructure/hetzner-bifrost) for the VPS setup and container details.

---

## Configuration

| Setting | Value |
|---------|-------|
| URL | `https://auth.madhan.app` |
| Stack | `authentik` |
| Source file | `core/cloud/authentik.go` |
| GitHub OAuth ClientID | `Ov23liUPVh4nPuUJzGFp` |
| Grafana OIDC ClientID | `grafana-homelab` |
| NetBird OIDC ClientID | `aumenijDycfG1cQURqH9BNJpV3KVUCoMHGPUVUlT` |

---

## GitHub OAuth Source

Authentik's `default-authentication-identification` stage is configured to show a **Login with GitHub** button. Users log in with their GitHub account, which Authentik links by email (`email_link` mode).

```
User → auth.madhan.app → "Login with GitHub" → GitHub OAuth → Authentik (identity confirmed)
```

The GitHub OAuth app uses PKCE (S256) for additional security.

---

## OIDC Applications

### Grafana

Grafana uses Authentik as a **confidential** OIDC provider:

| Setting | Value |
|---------|-------|
| ClientID | `grafana-homelab` |
| ClientSecret | `GRAFANA_OAUTH_CLIENT_SECRET` (SOPS) |
| Redirect URI | `https://grafana.madhan.app/login/generic_oauth` |
| Scopes | `openid email profile groups offline_access` |
| Token validity | 1 hour access, 10 min authorization code |

The `groups` scope is a custom property mapping that returns the list of Authentik group names the user belongs to. Grafana uses this via `role_attribute_path` to assign the **Admin** role to members of the `grafana-admins` group.

```python
# groups scope expression (in Authentik)
return [group.name for group in request.user.ak_groups.all()]
```

```bash
# Add a user to grafana-admins via Authentik UI → Directory → Groups → grafana-admins
```

### NetBird

NetBird uses Authentik through its **embedded Dex OIDC connector**. Authentik acts as the upstream provider; Dex federates to it.

| Setting | Value |
|---------|-------|
| ClientID | `aumenijDycfG1cQURqH9BNJpV3KVUCoMHGPUVUlT` |
| ClientSecret | `NETBIRD_CLIENT_SECRET` (SOPS) |
| Redirect URIs | `https://netbird.madhan.app/oauth2/callback`, `http://localhost:53000` |
| Scopes | `openid email profile api offline_access` |

The `api` scope is a custom empty mapping required by NetBird's documentation.

Configure in NetBird: **Settings → Identity Providers → Add → Authentik**:
- Issuer: `https://auth.madhan.app/application/o/netbird/`
- Client ID: `aumenijDycfG1cQURqH9BNJpV3KVUCoMHGPUVUlT`
- Client Secret: value of `NETBIRD_CLIENT_SECRET` from SOPS

---

## ForwardAuth (Public Services)

Traefik on the Bifrost VPS uses Authentik's **embedded outpost** to protect public cluster services behind a session cookie.

**How it works:**

```
Browser → grafana.madhan.app
  → Traefik → authentik-forwardauth middleware
    → Authentik: is there a valid session cookie?
      No → redirect to auth.madhan.app → GitHub login
      Yes → forward to k8s-gateway (http://192.168.1.220)
```

The ForwardAuth provider uses **forward_domain** mode with cookie domain `.madhan.app`, so a single login covers all `*.madhan.app` subdomains. The embedded outpost is managed entirely in Pulumi — no manual outpost configuration needed.

Services with `SkipAuth: true` in `cloudflare.go/publicServices` (e.g. Grafana, which handles its own auth) bypass the ForwardAuth middleware.

---

## NetBird Service Account

Authentik creates a `sa-netbird` service account in the `authentik Admins` group with a non-expiring API token. NetBird uses this token to sync users from Authentik.

```bash
# Get the service account token (Pulumi output):
cd core && pulumi stack select authentik && pulumi stack output NetbirdServiceToken --show-secrets
```

---

## Secrets

All secrets are stored in `secrets/bootstrap.sops.yaml` and injected via `sops exec-env` during `just core authentik up`:

| Variable | Purpose |
|----------|---------|
| `AUTHENTIK_TOKEN` | Bootstrap admin API token for Pulumi Authentik provider |
| `AUTHENTIK_GITHUB_SECRET` | GitHub OAuth app client secret |
| `GRAFANA_OAUTH_CLIENT_SECRET` | Grafana OIDC client secret |
| `NETBIRD_CLIENT_SECRET` | NetBird Dex connector client secret |

---

## Troubleshooting

### Grafana shows "Login with GitHub" but users get no roles

The user is not in the `grafana-admins` group. In Authentik:
**Directory → Groups → grafana-admins → Add user**

### ForwardAuth loops or redirects forever

```bash
# Check the embedded outpost is healthy
docker exec -n bifrost-net authentik-server ak healthcheck

# Verify outpost is registered in Authentik UI → Applications → Outposts
```

### Re-running the Authentik stack

If Authentik resources drift or the stack errors, re-run:

```bash
just core authentik up
```

The stack is idempotent — it imports existing resources where possible.
