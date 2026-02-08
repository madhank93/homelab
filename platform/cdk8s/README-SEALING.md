# Secure CDK8s CI Pipeline with Sealed Secrets

## Overview

This pipeline ensures that plaintext Kubernetes Secrets are **never committed to Git**. The flow is:

```
CDK8s Synth → Detect Secrets → Seal with kubeseal → Publish SealedSecrets
```

### Directory Structure

```
homelab/
├── app/                   # ❌ NEVER COMMIT - CDK8s output with plaintext Secrets
│   ├── cert-manager/
│   ├── grafana/
│   ├── harbor/
│   └── ...
└── platform/cdk8s/
    ├── sealed/            # ❌ NEVER COMMIT - Intermediate SealedSecrets
    ├── publish/           # ❌ NEVER COMMIT - Final manifests for GitOps
    ├── scripts/
    │   └── seal-secrets.sh    # Bash script to seal secrets
    └── .gitignore         # Ensures sealed/, publish/ are ignored
```

**Note:** The `app/` directory is at the repository root and is already protected by the root `.gitignore` file (`**/app`).

**GitOps Repo** (separate repository):
```
homelab-gitops/
└── manifests/         # ✅ SAFE TO COMMIT - Only SealedSecrets
```

## Local Development

### Prerequisites

```bash
# Install kubeseal
KUBESEAL_VERSION=0.27.1
curl -LO "https://github.com/bitnami-labs/sealed-secrets/releases/download/v${KUBESEAL_VERSION}/kubeseal-${KUBESEAL_VERSION}-linux-amd64.tar.gz"
tar -xzf kubeseal-${KUBESEAL_VERSION}-linux-amd64.tar.gz
sudo mv kubeseal /usr/local/bin/

# Install yq
sudo wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64
sudo chmod +x /usr/local/bin/yq
```

### Usage

```bash
# 1. Synthesize manifests (from platform/cdk8s directory)
go run main.go
# Output will be in ../../app/

# 2. Seal secrets
chmod +x scripts/seal-secrets.sh
./scripts/seal-secrets.sh ../../app ./sealed

# 3. Copy non-Secret manifests + SealedSecrets to publish/
mkdir -p publish
find ../../app -type f -name "*.yaml" -exec sh -c '
  if ! yq eval "select(.kind == \"Secret\")" "$1" > /dev/null 2>&1; then
    rel_path="${1#../../app/}"
    mkdir -p "publish/$(dirname "$rel_path")"
    cp "$1" "publish/$rel_path"
  fi
' _ {} \;
cp sealed/* publish/
```

## CI/CD Pipeline

The GitHub Actions workflow (`.github/workflows/cdk8s-seal-publish.yml`) automatically:

1. Synthesizes CDK8s manifests (`go run main.go`)
2. Seals all Secret resources using the Bash script
3. Publishes only SealedSecrets to the GitOps repository
4. Cleans up all temporary files

### Required GitHub Secrets

**Only 1 secret needed:**

- `INFISICAL_DB_PASSWORD`: Infisical PostgreSQL database password (bootstrap only)

**All other application secrets are managed in Infisical.** See [SECRETS-GUIDE.md](./SECRETS-GUIDE.md) for complete configuration.

### Workflow Triggers

- Push to `main` branch with changes in `platform/cdk8s/**`
- Manual workflow dispatch

## Security Guarantees

✅ **Plaintext secrets never touch disk in CI** (piped via stdin to kubeseal)  
✅ **`app/` (CDK8s output), `sealed/`, `publish/` are gitignored**  
✅ **Only SealedSecrets are pushed to GitOps repo**  
✅ **Temporary files are cleaned up even on failure** (`trap` in Bash, `if: always()` in GHA)  
✅ **Controller name/namespace are explicit** (no defaults)

## Sealed Secrets Controller

Ensure your cluster has the Sealed Secrets controller running:

```bash
kubectl get pods -n kube-system -l app.kubernetes.io/name=sealed-secrets
```

Controller configuration used:
- **Name**: `sealed-secrets`
- **Namespace**: `kube-system`

## Troubleshooting

### kubeseal fails with "cannot fetch certificate"

```bash
# Verify controller is running
kubectl get pods -n kube-system -l app.kubernetes.io/name=sealed-secrets

# Manually fetch certificate
kubeseal --fetch-cert \
  --controller-name=sealed-secrets \
  --controller-namespace=kube-system
```

### Secret not being sealed

Check if the Secret has the correct structure:

```bash
# View Secret in app/
yq eval 'select(.kind == "Secret")' ../../app/grafana/your-manifest.yaml

# Manually seal for testing
kubectl create secret generic test-secret \
  --from-literal=key=value \
  --dry-run=client -o yaml | \
  kubeseal \
    --controller-name=sealed-secrets \
    --controller-namespace=kube-system \
    --format=yaml
```

### GitOps repo not updating

Check GitHub Actions logs and ensure:
- `GITOPS_PAT` secret is set
- Token has write permissions
- GitOps repo exists and is accessible

## References

- [Sealed Secrets Documentation](https://github.com/bitnami-labs/sealed-secrets)
- [CDK8s Documentation](https://cdk8s.io/)
- [GitHub Actions Workflows](https://docs.github.com/en/actions)
