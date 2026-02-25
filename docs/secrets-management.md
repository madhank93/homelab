# Secrets Management

## Architecture Overview

```
┌──────────────────────────────────────────────────────────────────────┐
│  infra/secrets/bootstrap.env.sops  ◄─ encrypted with age, in git    │
└──────────────────┬───────────────────────────────────────────────────┘
                   │  sops exec-env (decrypts at runtime, never on disk)
         ┌─────────▼──────────┐          ┌────────────────────┐
         │  just pulumi up    │          │  just create-secrets│
         │  (Pulumi sees vars │          │  (kubectl apply)    │
         │   as env vars)     │          └─────────┬──────────┘
         └─────────┬──────────┘                    │
                   │                     ┌──────────▼──────────────────────┐
                   │                     │ k8s Secrets (Prune=false)       │
                   │                     │  infisical/infisical-secrets     │
                   │                     │  cert-manager/cloudflare-api-tok │
                   ▼                     └──────────┬──────────────────────┘
         Talos cluster + ArgoCD                     │  ArgoCD syncs platform apps
                   │                                │
                   └──────────────────────────────► ▼
                                          Infisical deployed by ArgoCD
                                          reads from infisical-secrets
                                                     │
                                          InfisicalSecret CRs sync
                                          runtime secrets to all apps
```

**Single source of truth:**
- Bootstrap time: SOPS-encrypted file in git → k8s Secrets via bootstrap script
- Runtime: Infisical manages all app secrets via InfisicalSecret CRs
- No Sealed Secrets controller needed
- CI pipeline requires zero GitHub Actions secrets

---

## First-Time Setup

### 1. Install prerequisites

```bash
# age (encryption tool)
brew install age

# sops (secrets manager)
brew install sops
```

### 2. Generate an age key pair

```bash
mkdir -p ~/.config/sops/age
age-keygen -o ~/.config/sops/age/keys.txt
# Output looks like:
# Public key: age1abc123...
```

The file `~/.config/sops/age/keys.txt` holds both keys. **Never commit it.**

### 3. Register your public key in `.sops.yaml`

Edit `.sops.yaml` at the repo root and replace the placeholder:

```yaml
creation_rules:
  - path_regex: infra/secrets/.*\.sops$
    age: age1abc123...   # ← your actual public key here
```

### 4. Gather required secrets before encrypting

You need the real values for every key in `bootstrap.env` before you can
encrypt the file. Collect them now:

#### Cloudflare API Token (`CLOUDFLARE_API_TOKEN`)

Go to **Cloudflare Dashboard → My Profile → API Tokens → Create Token**.
Choose **"Create Custom Token"** (not the Global API Key).

| Field | Value |
|-------|-------|
| Token name | `cert-manager-homelab` |
| Permissions | **Zone → Zone → Read** |
| Permissions | **Zone → DNS → Edit** |
| Zone Resources | **Include → Specific zone → `madhan.app`** |
| TTL | *(optional)* set an annual expiry |

Why both permissions:
- `Zone:Zone:Read` — cert-manager must resolve the domain name to its Cloudflare
  zone ID before it can write any DNS records.
- `Zone:DNS:Edit` — cert-manager creates and deletes the `_acme-challenge`
  TXT record to prove ownership to Let's Encrypt.

Scope to `madhan.app` only — a leaked token cannot touch any other domain.

Copy the generated token string; it is shown **only once**.

#### Other secrets

| Key | How to generate |
|-----|----------------|
| `PROXMOX_PASSWORD` | Your Proxmox root (or API user) password |
| `HCLOUD_TOKEN` | Hetzner Cloud → Project → Security → API Tokens → Generate |
| `INFISICAL_DB_PASSWORD` | `openssl rand -hex 16` |
| `INFISICAL_ENCRYPTION_KEY` | `openssl rand -hex 16` |
| `INFISICAL_AUTH_SECRET` | `openssl rand -base64 32` |
| `REDIS_PASSWORD` | `openssl rand -hex 24` (48 hex chars) |

### 5. Create and encrypt the bootstrap secrets file

The template contains only `KEY=changeme` lines — no comments. SOPS encrypts
comments too, so keeping them out produces a readable encrypted file where keys
remain visible.

```bash
# Copy the template
cp infra/secrets/bootstrap.env.sops.example infra/secrets/bootstrap.env

# Fill in every value collected in step 4 (no comments — keys only)
$EDITOR infra/secrets/bootstrap.env

# Encrypt (requires age key from step 2 and .sops.yaml from step 3)
sops --encrypt infra/secrets/bootstrap.env > infra/secrets/bootstrap.env.sops

# Delete the plaintext file — gitignored, but remove it anyway
rm infra/secrets/bootstrap.env
```

The encrypted file will look like:
```
PROXMOX_PASSWORD=ENC[AES256_GCM,data:...,type:str]
HCLOUD_TOKEN=ENC[AES256_GCM,data:...,type:str]
...
sops_version=3.x.x
```

Keys are visible, values are ciphertext. Commit `infra/secrets/bootstrap.env.sops` — safe to push to a public repo.

---

## Day-to-Day Workflow

### Run Pulumi (secrets injected automatically)

```bash
just pulumi talos preview
just pulumi talos up
```

SOPS decrypts `bootstrap.env.sops` into memory and injects vars as environment
variables for the duration of the `pulumi` command. Nothing is written to disk.

### Bootstrap a fresh cluster

```bash
# 1. Provision infrastructure and create k8s secrets in one command:
just bootstrap

# Or run the steps separately:
just create-secrets        # creates infisical-secrets + cloudflare-api-token
just pulumi talos up       # provisions VMs, Talos, ArgoCD
```

### Update a secret

```bash
# Decrypt in-place for editing
sops infra/secrets/bootstrap.env.sops

# sops opens $EDITOR with the decrypted content, re-encrypts on save
```

---

## Secret Inventory

| Secret | Namespace | Keys | Managed by |
|--------|-----------|------|------------|
| `infisical-secrets` | `infisical` | DB_PASSWORD, AUTH_SECRET, ENCRYPTION_KEY, DB_CONNECTION_URI, REDIS_PASSWORD | bootstrap script |
| `cloudflare-api-token` | `cert-manager` | CLOUDFLARE_API_TOKEN | bootstrap script |

All other application secrets are managed by Infisical at runtime via
`InfisicalSecret` custom resources.

---

## Edge Cases & Operational Notes

### ArgoCD will not prune bootstrap secrets

Both secrets have the annotation:
```yaml
argocd.argoproj.io/sync-options: Prune=false
```
ArgoCD will never delete them during a sync, even though they are not present
in the manifests branch.

### Safe to re-run the bootstrap script

`create-bootstrap-secrets.sh` uses `kubectl apply` throughout. Running it
again updates the secrets in place without downtime.

### Fresh cluster ordering

On a fresh cluster, namespaces do not yet exist. The bootstrap script creates
them before creating the secrets:
```
infisical → created
cert-manager → created
Secrets → created
```
ArgoCD (deployed by Pulumi) will later sync and manage the rest of the
resources in those namespaces without conflicting with the secrets.

### Rotating a secret

```bash
# 1. Edit the encrypted file
sops infra/secrets/bootstrap.env.sops

# 2. Re-run the bootstrap script to update the k8s Secret
just create-secrets

# 3. Restart the affected pod to pick up the new value
kubectl rollout restart deployment/... -n infisical
```

### Migrating an existing infisical-secrets (from Sealed Secrets)

If the cluster has an existing `infisical-secrets` Secret that was created by
the Sealed Secrets controller:

1. Note: `kubectl apply` will overwrite it with the plain Secret from the
   bootstrap script — this is correct and desired.
2. The Sealed Secrets controller is removed when ArgoCD prunes the
   `sealed-secrets` application (no manifests in the branch after this commit).
3. Any `SealedSecret` CRs left on the cluster become harmless orphans (no
   controller to reconcile them). Delete them manually if desired:
   ```bash
   kubectl delete sealedsecret --all -A
   ```

### Adding a team member

1. Get their age public key (`age1...`)
2. Add it to `.sops.yaml` under the same rule (comma-separated or as a list)
3. Run `sops updatekeys infra/secrets/bootstrap.env.sops` to re-encrypt with
   the new key set
4. Commit the updated `.sops.yaml` and `bootstrap.env.sops`

### GitHub Actions (CI)

The CI pipeline (`cdk8s-seal-publish.yml`) synthesizes CDK8s manifests and
publishes them to the `*-manifests` branch. It requires **zero GitHub Actions
secrets** because:
- CDK8s no longer generates any Secret resources
- All secrets are created by the bootstrap script on the operator's laptop
- The pipeline just runs `go run main.go` (no sensitive input)

If you ever need Pulumi to run in CI, add `SOPS_AGE_KEY` as a repository
secret (the full contents of `~/.config/sops/age/keys.txt`) and add
`sops exec-env` wrapping to the workflow step.

### Redis password

The Infisical Helm chart's Redis subchart reads the password from
`infisical-secrets` via:
```yaml
auth:
  existingSecret: infisical-secrets
  existingSecretPasswordKey: REDIS_PASSWORD
```
This means the password is never embedded in the CDK8s-generated manifests.

### Cloudflare API Token and cert-manager

The token is created during **First-Time Setup → Step 4** above, before the
bootstrap secrets file is encrypted. It is stored in the
`cert-manager/cloudflare-api-token` Secret (created by `just create-secrets`).

The bootstrap script creates it directly rather than via an InfisicalSecret CR
to avoid a circular dependency: Infisical needs cert-manager to be healthy;
cert-manager needs the Cloudflare token to become healthy.

#### Verifying the certificate issuance

After `just create-secrets` and `just pulumi talos up`, check that cert-manager
picks up the token and the ACME DNS-01 challenge progresses:

```bash
# Watch the Certificate status
kubectl describe certificate wildcard-madhan-app -n kube-system

# Watch the CertificateRequest and Order objects
kubectl get certificaterequests,orders -n kube-system

# Check cert-manager logs for Cloudflare API errors
kubectl logs -n cert-manager deployment/cert-manager | grep -i cloudflare
```

A successful issuance ends with:
```
Status: True  Type: Ready
Message: Certificate is up to date and has not expired
```

Once `wildcard-madhan-app-tls` exists in `kube-system`, re-enable the HTTPS
listener in `infra/pulumi/cilium.go` and run `just pulumi talos up` again.

---

## Long-Term Vision

Once Infisical is healthy and all secrets are stored there:

1. Store all new app secrets in Infisical projects
2. Create `InfisicalSecret` CRs in each app namespace to sync secrets
3. The bootstrap secrets (`infisical-secrets`, `cloudflare-api-token`) remain
   as the minimal set of "chicken-and-egg" secrets that must exist before
   Infisical starts

The bootstrap script becomes a one-time operation for fresh cluster provisioning.
