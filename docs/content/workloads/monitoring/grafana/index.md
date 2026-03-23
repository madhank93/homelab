+++
title = "Grafana"
description = "Dashboards and data visualization for metrics and logs."
weight = 10
+++

## What is Grafana?

[Grafana](https://grafana.com/) is an open-source observability platform for visualizing metrics, logs, and traces. It connects to multiple datasources and provides a rich dashboard editor with alerting capabilities.

## Why Grafana?

Grafana is the de-facto standard for Kubernetes dashboards. It has native support for VictoriaMetrics (Prometheus-compatible) and VictoriaLogs (Loki-compatible), and its sidecar model for dashboard provisioning fits perfectly with the GitOps workflow.

## How It's Used Here

Grafana is the primary observability UI for the cluster. It visualizes:
- Node and pod metrics from VictoriaMetrics
- Container logs from VictoriaLogs
- GPU metrics from DCGM Exporter
- ArgoCD application health
- Longhorn volume status
- Falco security events

Source: [`workloads/monitoring/grafana.go`](https://github.com/madhank93/homelab/blob/v0.1.5/workloads/monitoring/grafana.go)

## Configuration

| Setting | Value | Why |
|---------|-------|-----|
| Namespace | `grafana` | Isolated namespace |
| HTTPRoute | `grafana.madhan.app` → port 3000 | Gateway API |
| Access | Public (Cloudflare A record) | GitHub SSO via Authentik |
| `root_url` | `https://grafana.madhan.app` | Forces `https://` in OAuth redirect_uri (see below) |
| Persistence | `10Gi` RWX Longhorn | Avoids rolling update deadlock |
| Admin user | `admin` | Set via env var |
| Admin password | File at `/mnt/secrets/ADMIN_PASSWORD` | Pattern A (CSI file mount) |
| OAuth secret | `grafana-oauth-secret` k8s Secret | Pattern B (CSI secretObjects sync) |

## Datasources

Configured statically in Helm values:

| Name | Type | URL | Notes |
|------|------|-----|-------|
| VictoriaMetrics | `prometheus` (default) | `http://vmsingle-vm-stack.victoria-metrics.svc.cluster.local:8428` | `timeInterval: 30s` |
| VictoriaLogs | `loki` | `http://victoria-logs-victoria-logs-single-server.victoria-logs.svc.cluster.local:9428/select` | `logsVolumeEnabled: false` (VictoriaLogs doesn't implement Loki index/volume API) |

> **Important:** The VictoriaLogs URL ends at `/select` — Grafana's Loki plugin appends `/loki/api/v1/...` automatically. The `logsVolumeEnabled` flag is disabled to suppress the unsupported `/index/volume` endpoint error in Drilldown → Logs.

## Dashboard Provisioning

Grafana's sidecar watches ConfigMaps across all namespaces with the label `grafana_dashboard: "1"`:

```yaml
sidecar:
  dashboards:
    enabled: true
    searchNamespace: ALL
    label: grafana_dashboard
    labelValue: "1"
    folderAnnotation: grafana_folder
    provider:
      foldersFromFilesStructure: true
```

To add a dashboard, create a ConfigMap in any namespace with:

```yaml
metadata:
  labels:
    grafana_dashboard: "1"
  annotations:
    grafana_folder: "My Folder"
data:
  my-dashboard.json: |
    { ... Grafana dashboard JSON ... }
```

## Secrets (OpenBao)

Secrets at `secret/data/grafana` in OpenBao:

| Key | Pattern | How read by Grafana |
|-----|---------|---------------------|
| `ADMIN_PASSWORD` | A (file) | `GF_SECURITY_ADMIN_PASSWORD__FILE=/mnt/secrets/ADMIN_PASSWORD` |
| `OAUTH_CLIENT_SECRET` | B (secretObjects) | Synced to `grafana-oauth-secret` Secret → `envFromSecret` |

Write/update secrets:

```bash
# Set all secrets (first time)
kubectl exec -n openbao openbao-0 -- bao kv put secret/grafana \
  ADMIN_PASSWORD=<password> \
  OAUTH_CLIENT_SECRET=<secret>

# Update a single key
kubectl exec -n openbao openbao-0 -- bao kv patch secret/grafana \
  OAUTH_CLIENT_SECRET=<new-secret>
```

## GitHub SSO via Authentik

Grafana uses Authentik as an OIDC provider. GitHub is a social login source in Authentik — users log in with their GitHub account.

### Login Flow (Public User)

1. Browse to **`https://grafana.madhan.app`** — Grafana redirects to the Authentik login page
2. Click **"Login with GitHub via Authentik"**
3. GitHub OAuth consent screen appears — authorise the app
4. GitHub redirects back to Authentik, which creates an Authentik user (using the `default-source-enrollment` flow)
5. Authentik issues an OIDC token and redirects back to Grafana
6. Grafana reads the `groups` claim and assigns role: **Admin** if in `grafana-admins`, otherwise **Viewer**

> First-time users get the **Viewer** role automatically. To promote to Admin, add the user to the `grafana-admins` group in Authentik → Directory → Groups.

### OIDC Configuration

Set in `workloads/monitoring/grafana.go`:

```go
"auth.generic_oauth": map[string]any{
    "enabled":              true,
    "name":                 "GitHub via Authentik",
    "client_id":            "grafana-homelab",
    "scopes":               "openid email profile",
    "auth_url":             "https://auth.madhan.app/application/o/authorize/",
    "token_url":            "https://auth.madhan.app/application/o/token/",
    "api_url":              "https://auth.madhan.app/application/o/userinfo/",
    "use_pkce":             true,
    "allow_sign_up":        true,
    "role_attribute_path":  "contains(groups[*], 'grafana-admins') && 'Admin' || 'Viewer'",
},
// root_url forces https:// in redirect_uri — TLS terminates at Bifrost/Traefik,
// so without this Grafana would construct an http:// callback that mismatches
// the https:// redirect URI registered in Authentik.
"server": map[string]any{
    "root_url": "https://grafana.madhan.app",
},
```

### One-Time SSO Setup (Managed by Pulumi)

The Authentik configuration is fully automated via `just core authentik up` (`core/cloud/authentik.go`). Manual steps are only needed to:

1. **Grant Admin role** — Authentik UI → Directory → Groups → `grafana-admins` → Add user
2. **Store Client Secret in OpenBao** — run once after creating the Grafana OIDC provider:
   ```bash
   just openbao-get secret/grafana   # verify current value
   bao kv patch secret/grafana OAUTH_CLIENT_SECRET=<secret>
   ```

### Admin Login (Bypass SSO)

The local `admin` account is always available at `/login`:

```bash
# Get admin password
just openbao-get secret/grafana ADMIN_PASSWORD
```

## How It Connects

```
Browser → grafana.madhan.app
  → Cloudflare → Traefik on Bifrost (public only)
  → homelab-gateway → HTTPRoute → Grafana pod
  → Grafana → Authentik OIDC (login)
  → Grafana → VictoriaMetrics (metrics queries)
  → Grafana → VictoriaLogs (log queries)
  → OpenBao CSI → /mnt/secrets/ADMIN_PASSWORD (admin password file)
  → grafana-oauth-secret k8s Secret (OAuth client secret)
```

## Screenshots

![Grafana main dashboard showing VictoriaMetrics datasource and node resource panels](/assets/screenshots/grafana/main-dashboard.png)

## Troubleshooting

### RWX PVC Migration

If Grafana was previously using an RWO PVC and rolling updates cause Multi-Attach errors, delete the old PVC and recreate with `ReadWriteMany`. Grafana must be temporarily scaled down during migration.

### OAuth Client Secret Not Loading

**Symptoms:** Grafana shows "Invalid client_secret" at OAuth login.

**Diagnosis:**

```bash
# Check the k8s Secret is populated
kubectl get secret grafana-oauth-secret -n grafana -o yaml

# Check CSI volume is mounted
kubectl describe pod -n grafana -l app.kubernetes.io/name=grafana | grep -A5 "openbao-secrets"
```

**Fix:** If the `grafana-oauth-secret` k8s Secret is missing, the CSI volume mount may have failed. Check OpenBao is unsealed and the `grafana` role exists.

### GitHub SSO Works on LAN but Fails Externally

**Symptoms:** Clicking "GitHub via Authentik" on LAN completes successfully, but from the internet Authentik returns a `redirect_uri mismatch` or `invalid_grant` error.

**Root cause:** Without `root_url` set, Grafana detects its scheme from the incoming request. TLS terminates at Bifrost/Traefik, which proxies to the cluster as plain HTTP. Grafana sees an HTTP request and constructs the OAuth callback as `http://grafana.madhan.app/login/generic_oauth`. Authentik has `https://...` registered — the mismatch causes OAuth to fail. On LAN, the request is plain HTTP end-to-end, so `http://` is correct and it works.

**Fix:** `root_url = https://grafana.madhan.app` is set in `[server]` in `grafana.ini` (already applied). This forces Grafana to always construct the correct `https://` callback URL regardless of the incoming request scheme.

### Pod Connection Timeout (kubelet nodeIP issue with NetBird)

**Symptoms:** Grafana pod shows connection timeouts to VictoriaMetrics or VictoriaLogs.

**Diagnosis:** Check if the NetBird peer is running on the same node as Grafana. NetBird adds a WireGuard interface, which can confuse kubelet's nodeIP detection.

**Fix:** The `kubelet.nodeIP.validSubnets: 192.168.1.0/24` setting in the Talos worker patch ensures kubelet reports the correct node IP. If nodes were provisioned before this patch was added, re-apply the patch via Pulumi.
