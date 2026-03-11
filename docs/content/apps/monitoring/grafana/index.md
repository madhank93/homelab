+++
title = "Grafana"
description = "Dashboards and data visualization for metrics and logs."
weight = 10
+++

## Overview

| Property | Value |
|----------|-------|
| CDK8s file | `workloads/monitoring/grafana.go` |
| Namespace | `grafana` |
| HTTPRoute | `grafana.madhan.app` |
| Access | Public — GitHub SSO via Authentik (Viewer for all, Admin for `grafana-admins` group) |

## Datasources

| Name | Type |
|------|------|
| VictoriaMetrics | Prometheus (default) |
| VictoriaLogs | Loki |

## Secrets (OpenBao)

Secrets live at `secret/data/grafana` in OpenBao:

| Key | Description |
|-----|-------------|
| `ADMIN_PASSWORD` | Grafana admin password — mounted as file via CSI |
| `OAUTH_CLIENT_SECRET` | Authentik OAuth2 client secret — synced to k8s Secret `grafana-oauth-secret` |

Write/update secrets:

```bash
# Set admin password (first time)
kubectl exec -n openbao openbao-0 -- bao kv put secret/grafana \
  ADMIN_PASSWORD=<password> \
  OAUTH_CLIENT_SECRET=<secret>

# Update a single key
kubectl exec -n openbao openbao-0 -- bao kv patch secret/grafana \
  OAUTH_CLIENT_SECRET=<new-secret>
```

## GitHub SSO via Authentik — Setup Steps

This is a one-time setup. Grafana uses Authentik as an OIDC provider; GitHub is a social source in Authentik.

### 1. Add GitHub social source in Authentik

Authentik UI → **Directory → Federation & Social login → Create → GitHub**
- Consumer key / secret: from your GitHub OAuth App
- Slug: `github`

### 2. Create OAuth2/OIDC provider in Authentik

Authentik UI → **Applications → Providers → Create → OAuth2/OpenID Provider**
- Name: `Grafana`
- Authentication flow: `default-authentication-flow`
- Client type: `Confidential`
- Redirect URI: `https://grafana.madhan.app/login/generic_oauth`
- Scopes: `openid`, `email`, `profile`
- Copy the **Client ID** and **Client Secret**

### 3. Create Application in Authentik

Authentik UI → **Applications → Applications → Create**
- Name: `Grafana`
- Slug: `grafana`
- Provider: bind to the provider created above

### 4. Create grafana-admins group (for Admin role)

Authentik UI → **Directory → Groups → Create**
- Name: `grafana-admins`
- Add your user to this group

### 5. Update grafana.go with the Client ID

In `workloads/monitoring/grafana.go`, replace:
```
"client_id": "REPLACE_WITH_AUTHENTIK_CLIENT_ID",
```
with the actual Client ID from step 2. Commit and let ArgoCD sync.

### 6. Store the Client Secret in OpenBao

```bash
kubectl exec -n openbao openbao-0 -- bao kv patch secret/grafana \
  OAUTH_CLIENT_SECRET=<client-secret-from-step-2>
```

### 7. Verify

- Visit `https://grafana.madhan.app` — should redirect to Authentik login
- Log in with GitHub — you get Viewer role
- Members of `grafana-admins` get Admin role automatically

## Admin Login (bypass SSO)

The local `admin` account is always available at `/login`:

```bash
# Get admin password
kubectl exec -n openbao openbao-0 -- bao kv get -field=ADMIN_PASSWORD secret/grafana
```

Username: `admin`, Password: value from OpenBao.

## Dashboards

Grafana auto-discovers dashboards from ConfigMaps with label `grafana_dashboard: "1"` across all namespaces. To add a dashboard, create a ConfigMap in any namespace with that label and the JSON content.
