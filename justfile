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

ping_scan:
    nmap -sn 192.168.1.0/24
