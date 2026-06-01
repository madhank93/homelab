+++
title = "Exposing Headlamp via NetBird Reverse Proxy"
description = "Expose Headlamp to the internet through NetBird v0.66 expose with Authentik SSO group-based access control."
weight = 10
+++

## Overview

NetBird v0.66 introduced the `netbird expose` command — a built-in reverse proxy that publishes a local service to the internet behind your SSO provider. This guide walks through exposing Headlamp at `headlamp.proxy.madhan.app` with Authentik SSO authentication, requiring users to be in a specific Authentik group before they can reach the Headlamp UI.

### Architecture

```
User browser
  └─→ headlamp.proxy.madhan.app  (Cloudflare DNS → Hetzner VPS)
        └─→ Traefik on Bifrost   (TLS termination, *.proxy.madhan.app catch-all)
              └─→ NetBird Proxy Service  (port 443 on Bifrost)
                    └─→ WireGuard mesh tunnel
                          └─→ k8s-routing-peer (in cluster)
                                └─→ headlamp.headlamp.svc.cluster.local:80
```

### What's already in place

| Component | Status | Notes |
|-----------|--------|-------|
| `*.proxy.madhan.app` Cloudflare DNS | ✅ Done | Already points at Hetzner VPS |
| Traefik on Bifrost | ✅ Done | Handles TLS for `*.proxy.madhan.app` |
| `k8s-routing-peer` WireGuard peer | ✅ Done | Advertises `192.168.1.0/24` in the mesh |
| Headlamp service in-cluster | ✅ Done | `headlamp.headlamp.svc.cluster.local:80` |
| Authentik as IdP for NetBird | ✅ Done | OIDC via Dex connector |

### What you need to do

| Step | Where | Code change? |
|------|-------|-------------|
| 1. Create Authentik group | Authentik UI | No |
| 2. Register expose redirect URI in Authentik | `core/cloud/authentik.go` | **Yes** |
| 3. Assign your user to the group | Authentik UI | No |
| 4. Run `netbird expose` on the k8s-routing-peer | `workloads/networking/netbird_peer.go` | **Yes** |
| 5. Verify | CLI | No |

---

## Step 1 — Create an Authentik Group for Headlamp Access

In **Authentik UI → Directory → Groups → Create**:

| Field | Value |
|-------|-------|
| Name | `homelab-admins` |
| Notes | NetBird expose SSO gate for Headlamp |

Then **assign yourself** (and any other users): **Groups → homelab-admins → Users → Add User**.

---

## Step 2 — Register the Expose Redirect URI in Authentik (Code Change)

NetBird's SSO flow for exposed services redirects to `https://netbird.madhan.app/api/v1/sso/callback` after authentication. This must be added to the **NetBird OIDC app** in Authentik.

**File to change:** [`core/cloud/authentik.go`](file:///Volumes/work/git-repos/homelab/core/cloud/authentik.go)

Find the `createOIDCApp` call for the NetBird app (around line 265) and add the expose callback redirect:

```go
// BEFORE
Redirects: []string{
    "https://netbird.madhan.app/oauth2/callback",
    "http://localhost:53000",
},

// AFTER
Redirects: []string{
    "https://netbird.madhan.app/oauth2/callback",      // Dex embedded IdP callback
    "http://localhost:53000",                           // CLI device-auth callback
    "https://netbird.madhan.app/api/v1/sso/callback",  // expose SSO callback (v0.66+)
},
```

Then apply:

```bash
just core authentik up
```

---

## Step 3 — Modify the netbird-peer Deployment to Run `expose` (Code Change)

The existing `k8s-routing-peer` deployment runs `netbird up --advertise-routes=192.168.1.0/24`. You need to run `netbird expose` as a **second container (sidecar)** in the same pod — it uses the same WireGuard interface already established by the primary container.

**File to change:** [`workloads/networking/netbird_peer.go`](file:///Volumes/work/git-repos/homelab/workloads/networking/netbird_peer.go)

Add a sidecar container to the existing Deployment's container list:

```go
// Add after the existing netbird container in Containers slice:
{
    Name:  jsii.String("netbird-expose-headlamp"),
    Image: jsii.String("netbirdio/netbird:0.66"),
    Command: &[]*string{
        jsii.String("netbird"),
        jsii.String("expose"),
        // The in-cluster service address to proxy to
        jsii.String("headlamp.headlamp.svc.cluster.local:80"),
        // Public hostname (matches *.proxy.madhan.app wildcard in Cloudflare)
        jsii.String("--domain=headlamp.proxy.madhan.app"),
        // SSO: only members of this Authentik group get through
        jsii.String("--with-user-groups=homelab-admins"),
        jsii.String("--hostname=k8s-routing-peer"),
    },
    Env: &[]*k8s.EnvVar{
        {
            Name: jsii.String("NB_SETUP_KEY"),
            ValueFrom: &k8s.EnvVarSource{
                SecretKeyRef: &k8s.SecretKeySelector{
                    Name: jsii.String("netbird-setup-key"),
                    Key:  jsii.String("NETBIRD_SETUP_KEY"),
                },
            },
        },
        {
            Name:  jsii.String("NB_MANAGEMENT_URL"),
            Value: jsii.String("https://netbird.madhan.app"),
        },
    },
    SecurityContext: &k8s.SecurityContext{
        Capabilities: &k8s.Capabilities{
            Add: &[]*string{
                jsii.String("NET_ADMIN"),
                jsii.String("SYS_MODULE"),
            },
        },
    },
},
```

> **Note:** Pin to `netbirdio/netbird:0.66` rather than `latest` to match the server version. The `expose` command shares the WireGuard tunnel from the primary container — it doesn't need to run `netbird up` again, just `expose` on top.

Synthesize and push:

```bash
just synth
git add workloads/networking/netbird_peer.go app/netbird/
git commit -m "feat: add netbird expose sidecar for headlamp"
git push origin v0.1.5-manifests
```

ArgoCD will roll the `netbird-peer` Deployment with the new sidecar within 3 minutes.

---

## Step 4 — Verify

### 4a. Check the expose sidecar is running

```bash
kubectl get pods -n netbird
kubectl logs -n netbird -l app=netbird-peer -c netbird-expose-headlamp --tail=30
# Look for: "service exposed at: headlamp.proxy.madhan.app"
```

### 4b. Check the NetBird dashboard

**NetBird UI → Network → Exposed Services** — you should see:

| Service | Domain | Auth |
|---------|--------|------|
| headlamp | headlamp.proxy.madhan.app | SSO (homelab-admins) |

### 4c. Browser test

Open **`https://headlamp.proxy.madhan.app`** in a private window:

1. Browser hits Cloudflare → Hetzner VPS Traefik → NetBird Proxy service
2. NetBird proxy detects unauthenticated request → redirects to `https://netbird.madhan.app/api/v1/sso/...`
3. NetBird SSO flow redirects to Authentik (via embedded Dex)
4. You log in with GitHub → Authentik verifies you're in `homelab-admins`
5. Redirected back → Headlamp UI loads

### 4d. Test unauthorized access

Log in with an account that is **not** in `homelab-admins` — you should get a `403 Forbidden` from the NetBird proxy layer.

---

## DNS Note — No Cloudflare Change Needed

The `*.proxy.madhan.app` wildcard A record already exists in `cloudflare.go`:

```go
// Already in core/cloud/cloudflare.go
if err := newRecord("*.proxy.madhan.app", "wildcard-proxy-madhan-app", hetznerIP,
    "NetBird expose wildcard"); err != nil {
    return err
}
```

`headlamp.proxy.madhan.app` is automatically covered — no additional DNS record needed.

---

## Adding More Services Later

The same pattern applies to any other internal service. For each new service, add another sidecar container:

```go
{
    Name:  jsii.String("netbird-expose-grafana"),
    Image: jsii.String("netbirdio/netbird:0.66"),
    Command: &[]*string{
        jsii.String("netbird"),
        jsii.String("expose"),
        jsii.String("grafana.grafana.svc.cluster.local:80"),
        jsii.String("--domain=grafana.proxy.madhan.app"),
        jsii.String("--with-user-groups=homelab-admins"),
        jsii.String("--hostname=k8s-routing-peer"),
    },
    // ... same Env + SecurityContext as above
},
```

> If you use `grafana.proxy.madhan.app` for the NetBird expose subdomain, you can keep the **existing** `grafana.madhan.app` DNS record pointing at the Hetzner VPS via Traefik ForwardAuth (Authentik) — they are separate, independent exposure paths.

---

## Troubleshooting

| Symptom | Likely cause | Fix |
|---------|-------------|-----|
| `headlamp.proxy.madhan.app` returns `502` | Sidecar not running or expose registration failed | Check sidecar logs |
| SSO redirect loop | Authentik redirect URI not registered | Apply `just core authentik up` and check `netbird.madhan.app/api/v1/sso/callback` is in the allowed redirects |
| `403` after successful login | User not in `homelab-admins` group | Add user to group in Authentik UI |
| Sidecar CrashLoopBackOff | Version mismatch between `netbird` client and server | Use the same version tag: `netbirdio/netbird:0.66` |
| `expose` command not found | Client image too old | Pin image to `netbirdio/netbird:0.66` or later |
