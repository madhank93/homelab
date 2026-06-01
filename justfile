# justfile

# Pulumi — infra stacks (talos | platform | hetzner | authentik | cloudflare)
[working-directory: 'core']
core stack action:
    SOPS_AGE_KEY_FILE="$HOME/.config/sops/age/keys.txt" \
      sops exec-env ../secrets/bootstrap.sops.yaml \
      'pulumi stack select {{stack}} && pulumi {{action}} --yes'

# CDK8s synthesis → writes to app/
[working-directory: 'workloads']
synth:
    go run .

# Build and push a custom image to Harbor
# Usage: just build-push <image-name> <tag>
# Example: just build-push notebook-gateway-controller v1
build-push image tag:
    docker buildx build \
      --platform linux/amd64 \
      --push \
      -t harbor.madhan.app/library/{{image}}:{{tag}} \
      images/{{image}}

# Bootstrap secrets (creates k8s Secrets from sops-encrypted values)
create-secrets:
    SOPS_AGE_KEY_FILE="$HOME/.config/sops/age/keys.txt" \
      sops exec-env secrets/bootstrap.sops.yaml 'bash scripts/create-bootstrap-secrets.sh'

# One-time OpenBao initialisation (run after first ArgoCD deploy of OpenBao)
openbao-init:
    kubectl exec -n openbao openbao-0 -- bao operator init \
      -key-shares=1 -key-threshold=1 -format=json > /tmp/openbao-init.json
    @echo "✅  Init output saved to /tmp/openbao-init.json"
    @echo "    Copy the unseal_keys_b64[0] value into secrets/bootstrap.sops.yaml as OPENBAO_UNSEAL_KEY"
    @echo "    Then run: just create-secrets && kubectl rollout restart statefulset/openbao -n openbao"

# One-time OpenBao K8s auth + policy + role setup (run after openbao-init)
openbao-setup:
    SOPS_AGE_KEY_FILE="$HOME/.config/sops/age/keys.txt" \
      sops exec-env secrets/bootstrap.sops.yaml 'bash scripts/openbao-setup.sh'

# Generate a temporary OpenBao root token from the stored unseal key.
# The token is printed to stdout — export it for subsequent bao commands.
# Revoke it when done: just openbao-revoke <token>
# OpenBao 2.x: client must generate OTP first, then pass it to -init.
openbao-token:
    #!/usr/bin/env bash
    set -euo pipefail
    UNSEAL_KEY=$(kubectl get secret openbao-unseal-key -n openbao \
      -o jsonpath='{.data.unseal-key}' | base64 -d)
    kubectl exec -n openbao openbao-0 -c openbao -- \
      bao operator generate-root -cancel -format=json 2>/dev/null || true
    OTP=$(kubectl exec -n openbao openbao-0 -c openbao -- \
      bao operator generate-root -generate-otp)
    INIT=$(kubectl exec -n openbao openbao-0 -c openbao -- \
      bao operator generate-root -init -otp="$OTP" -format=json)
    NONCE=$(echo "$INIT" | python3 -c "import sys,json; print(json.load(sys.stdin)['nonce'])")
    RESULT=$(kubectl exec -n openbao openbao-0 -c openbao -- \
      bao operator generate-root -nonce="$NONCE" -format=json "$UNSEAL_KEY")
    ENCODED=$(echo "$RESULT" | python3 -c "import sys,json; print(json.load(sys.stdin)['encoded_token'])")
    kubectl exec -n openbao openbao-0 -c openbao -- \
      bao operator generate-root -decode="$ENCODED" -otp="$OTP"

# Read an OpenBao secret. Optionally pass a field name to get a single value.
# Usage:
#   just openbao-get secret/grafana
#   just openbao-get secret/grafana ADMIN_PASSWORD
openbao-get path field='':
    #!/usr/bin/env bash
    set -euo pipefail
    UNSEAL_KEY=$(kubectl get secret openbao-unseal-key -n openbao \
      -o jsonpath='{.data.unseal-key}' | base64 -d)
    kubectl exec -n openbao openbao-0 -c openbao -- \
      bao operator generate-root -cancel -format=json 2>/dev/null || true
    OTP=$(kubectl exec -n openbao openbao-0 -c openbao -- \
      bao operator generate-root -generate-otp)
    INIT=$(kubectl exec -n openbao openbao-0 -c openbao -- \
      bao operator generate-root -init -otp="$OTP" -format=json)
    NONCE=$(echo "$INIT" | python3 -c "import sys,json; print(json.load(sys.stdin)['nonce'])")
    RESULT=$(kubectl exec -n openbao openbao-0 -c openbao -- \
      bao operator generate-root -nonce="$NONCE" -format=json "$UNSEAL_KEY")
    ENCODED=$(echo "$RESULT" | python3 -c "import sys,json; print(json.load(sys.stdin)['encoded_token'])")
    ROOT_TOKEN=$(kubectl exec -n openbao openbao-0 -c openbao -- \
      bao operator generate-root -decode="$ENCODED" -otp="$OTP")
    if [ -n "{{field}}" ]; then
      kubectl exec -n openbao openbao-0 -c openbao -- \
        env VAULT_TOKEN="$ROOT_TOKEN" bao kv get -field="{{field}}" "{{path}}"
    else
      kubectl exec -n openbao openbao-0 -c openbao -- \
        env VAULT_TOKEN="$ROOT_TOKEN" bao kv get "{{path}}"
    fi
    kubectl exec -n openbao openbao-0 -c openbao -- \
      env VAULT_TOKEN="$ROOT_TOKEN" bao token revoke "$ROOT_TOKEN" 2>/dev/null || true

# Revoke a previously generated root token
openbao-revoke token:
    #!/usr/bin/env bash
    set -euo pipefail
    kubectl exec -n openbao openbao-0 -c openbao -- \
      env VAULT_TOKEN="{{token}}" bao token revoke "{{token}}"

# Get a short-lived OpenBao token via Kubernetes auth using a service account.
# The service account must be bound to an OpenBao role (see scripts/openbao-setup.sh).
# Usage:
#   just openbao-sa-token <serviceaccount> <namespace> <role>
# Examples:
#   just openbao-sa-token secret-sync harbor harbor     → harbor-policy token
#   just openbao-sa-token grafana     grafana grafana   → grafana-policy token
openbao-sa-token sa ns role:
    #!/usr/bin/env bash
    set -euo pipefail
    SA_TOKEN=$(kubectl create token {{sa}} -n {{ns}} --duration=10m)
    kubectl exec -n openbao openbao-0 -c openbao -- \
      bao write -field=token auth/kubernetes/login role={{role}} jwt="$SA_TOKEN"

# Read a secret from OpenBao using a Kubernetes service account.
# Usage:
#   just openbao-sa-get <serviceaccount> <namespace> <role> <secret-path> [field]
# Examples:
#   just openbao-sa-get secret-sync harbor harbor secret/harbor
#   just openbao-sa-get secret-sync harbor harbor secret/harbor HARBOR_ADMIN_PASSWORD
openbao-sa-get sa ns role path field='':
    #!/usr/bin/env bash
    set -euo pipefail
    SA_TOKEN=$(kubectl create token {{sa}} -n {{ns}} --duration=10m)
    VAULT_TOKEN=$(kubectl exec -n openbao openbao-0 -c openbao -- \
      bao write -field=token auth/kubernetes/login role={{role}} jwt="$SA_TOKEN")
    if [ -n "{{field}}" ]; then
      kubectl exec -n openbao openbao-0 -c openbao -- \
        env VAULT_TOKEN="$VAULT_TOKEN" bao kv get -field={{field}} {{path}}
    else
      kubectl exec -n openbao openbao-0 -c openbao -- \
        env VAULT_TOKEN="$VAULT_TOKEN" bao kv get {{path}}
    fi

# Decrypt and view a SOPS-encrypted file (opens in $EDITOR by default)
# Usage:
#   just sops-view secrets/bootstrap.sops.yaml
# To print to stdout instead:
#   just sops-decrypt secrets/bootstrap.sops.yaml
sops-view file:
    SOPS_AGE_KEY_FILE="$HOME/.config/sops/age/keys.txt" sops {{file}}

sops-decrypt file:
    SOPS_AGE_KEY_FILE="$HOME/.config/sops/age/keys.txt" sops -d {{file}}

ping_scan:
    nmap -sn 192.168.1.0/24
