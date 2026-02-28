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

ping_scan:
    nmap -sn 192.168.1.0/24
