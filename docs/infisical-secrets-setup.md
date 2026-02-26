# Infisical Secrets Setup Guide

This guide covers everything needed to get secrets flowing from Infisical into the cluster.
It must be completed **after** Infisical is running and **before** apps that depend on secrets
will become healthy.

## Apps That Require Infisical Secrets

| App | Namespace | Secret Created | Keys Required |
|-----|-----------|---------------|---------------|
| Grafana | `grafana` | `grafana-admin` | `ADMIN_PASSWORD` |
| Harbor | `harbor` | `harbor-admin` | `HARBOR_ADMIN_PASSWORD` |
| n8n | `n8n` | `n8n-db` | `DB_PASSWORD` |
| Rancher | `cattle-system` | `rancher-bootstrap` | `BOOTSTRAP_PASSWORD` |

The Infisical operator reconciles these via `InfisicalSecret` CRDs in each namespace.
The managed secrets are then consumed by the respective Helm charts as `existingSecret`.

---

## Step 1 — Access the Infisical UI

```bash
# Port-forward if HTTPRoute is not yet resolving
kubectl port-forward svc/infisical-infisical-standalone-infisical -n infisical 8080:8080
# Then open http://localhost:8080
```

Or use the Gateway route: `http://infisical.madhan.app`

---

## Step 2 — Create Project and Environment

1. Log in to Infisical (register an admin account on first launch).
2. Create a new project named **`homelab-prod`**.
3. Confirm an environment named **`prod`** exists inside the project (created by default).

---

## Step 3 — Create Secrets

Navigate to **Secrets Dashboard → prod environment** and create the following folders and
secrets. All paths are relative to the project root.

### Path: `/grafana`

| Key | Value | Used by |
|-----|-------|---------|
| `ADMIN_PASSWORD` | `<your-secure-password>` | Grafana admin login (`admin` user) |

### Path: `/harbor`

| Key | Value | Used by |
|-----|-------|---------|
| `HARBOR_ADMIN_PASSWORD` | `<your-secure-password>` | Harbor admin login |

### Path: `/n8n`

| Key | Value | Used by |
|-----|-------|---------|
| `DB_PASSWORD` | `<your-secure-password>` | PostgreSQL password for `n8n` user |

> **Note — n8n encryption key**: n8n also requires `N8N_ENCRYPTION_KEY` but this is
> managed separately by the Helm chart (auto-generated into `n8n-encryption-key-secret-v2`).
> It is NOT sourced from Infisical. If you see a "Mismatching encryption keys" crash,
> see the [N8n Encryption Key section](#n8n-encryption-key-mismatch) below.

### Path: `/rancher`

| Key | Value | Used by |
|-----|-------|---------|
| `BOOTSTRAP_PASSWORD` | `<your-secure-password>` | Rancher initial bootstrap |

---

## Step 4 — Create a Service Token

1. In Infisical, go to **Project Settings → Service Tokens**.
2. Create a token with:
   - **Name**: `homelab-prod-token`
   - **Scopes** (all Read):
     - `/grafana`
     - `/harbor`
     - `/n8n`
     - `/rancher`
   - **Expiration**: Never (or set a rotation schedule)
3. **Copy the token immediately** — it is not shown again.

---

## Step 5 — Create the k8s Service Token Secret

This is the bridge between the Infisical operator and your project. Create it once:

```bash
kubectl create secret generic infisical-service-token \
  --from-literal=infisicalToken="<PASTE_YOUR_TOKEN_HERE>" \
  -n infisical \
  --dry-run=client -o yaml | kubectl apply -f -
```

Once created, the Infisical operator will start reconciling all `InfisicalSecret` CRDs
across namespaces automatically (within ~60 seconds per the `resyncInterval`).

---

## Step 6 — Verify

```bash
# Check InfisicalSecret status in each namespace
kubectl get infisicalsecret -A

# Confirm the managed secrets were created
kubectl get secret grafana-admin -n grafana
kubectl get secret harbor-admin -n harbor
kubectl get secret n8n-db -n n8n
kubectl get secret rancher-bootstrap -n cattle-system

# Check operator logs if something is missing
kubectl logs -l app.kubernetes.io/name=infisical-operator -n infisical --tail=50
```

---

## Known Issues

### ArgoCD Sync Error: `projectSlug field not declared in schema`

**Symptom**: ArgoCD shows sync error on grafana/harbor/n8n/rancher apps:
```
failed to create typed patch object: .spec.authentication.serviceToken.secretsScope.projectSlug:
field not declared in schema
```

**Root cause**: The Infisical CRD schema for `authentication.serviceToken.secretsScope` does
not include `projectSlug`. ArgoCD's `ServerSideApply=true` validates against the schema and
rejects the field.

**Fix (already in code)**: All `InfisicalSecret` resources in CDK8s have the annotation:
```yaml
argocd.argoproj.io/sync-options: ServerSideApply=false
```
This makes ArgoCD use client-side apply for these resources only, bypassing schema validation.
The annotation is generated into each app's manifests automatically by CDK8s.

If you still see the error after the next CI run and ArgoCD sync, manually patch the live
resources:
```bash
for ns in grafana harbor n8n cattle-system; do
  name=$(kubectl get infisicalsecret -n $ns -o jsonpath='{.items[0].metadata.name}')
  kubectl annotate infisicalsecret $name -n $ns \
    argocd.argoproj.io/sync-options=ServerSideApply=false --overwrite
done
```

### N8n Encryption Key Mismatch

**Symptom**: n8n pod crashes with:
```
Mismatching encryption keys. The encryption key in the settings file /home/node/.n8n/config
does not match the N8N_ENCRYPTION_KEY env var.
```

**Root cause**: n8n stores its encryption key in the PVC at `/home/node/.n8n/config`. The
Helm chart auto-generates `n8n-encryption-key-secret-v2` with the `N8N_ENCRYPTION_KEY` env
var. If the pod was recreated and the secret was regenerated, the values diverge.

**Fix options**:

Option A — Fresh start (loses workflow data):
```bash
kubectl delete pvc -n n8n -l app.kubernetes.io/name=n8n
# ArgoCD or helm will recreate with matching keys on next deploy
```

Option B — Recover existing key (preserves workflow data):
```bash
# Run a temporary pod to read the config file from the PVC
kubectl run -it --rm n8n-recovery --image=busybox --restart=Never \
  --overrides='{"spec":{"volumes":[{"name":"data","persistentVolumeClaim":{"claimName":"n8n-main"}}],"containers":[{"name":"n8n-recovery","image":"busybox","command":["cat","/data/.n8n/config"],"volumeMounts":[{"name":"data","mountPath":"/data"}]}]}}' -n n8n
# Extract encryptionKey from the JSON output
# Then patch the secret:
kubectl patch secret n8n-encryption-key-secret-v2 -n n8n \
  --type=merge -p '{"data":{"N8N_ENCRYPTION_KEY":"<base64-encoded-key>"}}'
```

### Harbor RWO PVC Multi-Attach

**Symptom**: `harbor-jobservice` or `harbor-registry` pod stuck in `ContainerCreating` with:
```
Multi-Attach error: Volume is already used by pod(s) harbor-jobservice-<old-pod>
```

**Root cause**: Harbor Helm chart hardcodes `RollingUpdate` strategy. When ArgoCD syncs a
change, the rolling update deadlocks because old pods hold RWO PVCs on different nodes.

**Fix** (run after each occurrence):
```bash
kubectl delete pod -n harbor -l "app=harbor,component=jobservice" --grace-period=0 --force
kubectl delete pod -n harbor -l "app=harbor,component=registry" --grace-period=0 --force
```
The old pods release the PVCs and new pods attach successfully within ~30 seconds.
