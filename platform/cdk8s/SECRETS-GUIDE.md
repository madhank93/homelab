# Infisical Secrets Configuration Guide

## Quick Reference

**Total Secrets Needed: 6**

1. **GitHub Secret (1):** `INFISICAL_DB_PASSWORD`
2. **Infisical UI Secrets (4):** App passwords
3. **Kubernetes Secret (1):** `infisical-service-token`

---

## 1. GitHub Secret (Bootstrap)

### INFISICAL_DB_PASSWORD

**Purpose:** Infisical PostgreSQL database password  
**Where:** GitHub Repository → Settings → Secrets → Actions

```bash
# Generate password
openssl rand -base64 32

# Add to GitHub as: INFISICAL_DB_PASSWORD
```

---

## 2. Infisical UI Secrets (After Deployment)

### Access Infisical

```
URL: https://infisical.madhan.app
```

### Initial Setup

1. Create admin account
2. Create project: `homelab-prod`
3. Environment: `prod`

### Secrets to Add

Generate passwords:
```bash
openssl rand -base64 32  # Run 4 times
```

| Path | Key | Value | Used By |
|------|-----|-------|---------|
| `/grafana` | `ADMIN_PASSWORD` | `<generated>` | Grafana admin login |
| `/harbor` | `ADMIN_PASSWORD` | `<generated>` | Harbor admin login |
| `/rancher` | `BOOTSTRAP_PASSWORD` | `<generated>` | Rancher bootstrap |
| `/n8n` | `DB_PASSWORD` | `<generated>` | n8n PostgreSQL |

---

## 3. Service Token (Kubernetes Secret)

### Generate in Infisical UI

1. Settings → Service Tokens → Create
2. Configure:
   - Name: `kubernetes-operator`
   - Project: `homelab-prod`
   - Environment: `prod`
   - Permissions: Read-only
3. Copy token (starts with `st.`)

### Create in Kubernetes

```bash
kubectl create secret generic infisical-service-token \
  --from-literal=token=st.xxx.yyy.zzz \
  -n infisical
```

---

## Deployment Steps

```bash
# 1. Add GitHub Secret
# Add INFISICAL_DB_PASSWORD to GitHub

# 2. Deploy
git push

# 3. Configure Infisical UI
# Access https://infisical.madhan.app
# Add 4 app secrets

# 4. Create service token secret
kubectl create secret generic infisical-service-token \
  --from-literal=token=st.YOUR_TOKEN \
  -n infisical

# 5. Verify
kubectl get infisicalsecret -A
kubectl get secret grafana-admin -n monitoring
```

---

## Verification

```bash
# Check operator
kubectl get pods -n infisical-operator-system

# Check secrets synced
kubectl get secret grafana-admin -n monitoring -o yaml
kubectl get secret harbor-admin -n harbor -o yaml
kubectl get secret rancher-bootstrap -n rancher -o yaml
kubectl get secret n8n-db -n n8n -o yaml
```

---

## Troubleshooting

**Operator not syncing?**
```bash
kubectl logs -n infisical-operator-system deployment/secrets-operator-controller-manager
# Common: service token invalid or missing
```

**App won't start?**
```bash
kubectl describe infisicalsecret <name> -n <namespace>
# Check status and events
```


