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
| Persistence | `10Gi` RWX Longhorn | Avoids rolling update deadlock |
| Admin user | `admin` | Set via env var |
| Admin password | File at `/mnt/secrets/ADMIN_PASSWORD` | Pattern A (CSI file mount) |
| OAuth secret | `grafana-oauth-secret` k8s Secret | Pattern B (CSI secretObjects sync) |

## Datasources

Configured statically in Helm values:

| Name | Type | URL | Notes |
|------|------|-----|-------|
| VictoriaMetrics | `prometheus` (default) | `http://victoria-metrics-victoria-metrics-cluster-vmselect.victoria-metrics.svc.cluster.local:8481/select/0/prometheus` | `timeInterval: 30s` |
| VictoriaLogs | `loki` | `http://victoria-logs-victoria-logs-single-server.victoria-logs.svc.cluster.local:9428/select` | Grafana appends `/loki/api/v1/...` automatically |

> **Important:** The VictoriaLogs URL ends at `/select` — do **not** append `/loki` or `/loki/api/v1`. Grafana's Loki plugin appends the API path automatically, so the correct URL is just `/select`. Using `/select/loki` results in double-path errors.

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

Grafana uses Authentik as an OIDC provider. GitHub is a social source in Authentik.

OIDC configuration in `grafana.go`:

```go
"auth.generic_oauth": map[string]any{
    "enabled":    true,
    "client_id":  "grafana-homelab",
    "auth_url":   "https://auth.madhan.app/application/o/grafana/authorize/",
    "token_url":  "https://auth.madhan.app/application/o/grafana/token/",
    "api_url":    "https://auth.madhan.app/application/o/userinfo/",
    "role_attribute_path": "contains(groups[*], 'grafana-admins') && 'Admin' || 'Viewer'",
},
```

Members of the `grafana-admins` group in Authentik get Grafana Admin role. Everyone else gets Viewer.

### One-Time SSO Setup

1. **Authentik UI → Directory → Federation → Create → GitHub source** (slug: `github`)
2. **Authentik UI → Applications → Providers → Create OAuth2/OIDC provider** — Name: `Grafana`, Redirect URI: `https://grafana.madhan.app/login/generic_oauth`, Scopes: `openid email profile`
3. **Authentik UI → Applications → Create** — bind to the Grafana provider
4. **Authentik UI → Directory → Groups → Create** — `grafana-admins`, add yourself
5. Store the Client Secret in OpenBao: `bao kv patch secret/grafana OAUTH_CLIENT_SECRET=<secret>`

### Admin Login (Bypass SSO)

The local `admin` account is always available at `/login`:

```bash
# Get admin password from OpenBao
kubectl exec -n openbao openbao-0 -- bao kv get -field=ADMIN_PASSWORD secret/grafana
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

### Pod Connection Timeout (kubelet nodeIP issue with NetBird)

**Symptoms:** Grafana pod shows connection timeouts to VictoriaMetrics or VictoriaLogs.

**Diagnosis:** Check if the NetBird peer is running on the same node as Grafana. NetBird adds a WireGuard interface, which can confuse kubelet's nodeIP detection.

**Fix:** The `kubelet.nodeIP.validSubnets: 192.168.1.0/24` setting in the Talos worker patch ensures kubelet reports the correct node IP. If nodes were provisioned before this patch was added, re-apply the patch via Pulumi.
