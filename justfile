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

TALOSCTL := 'talosctl --talosconfig "$HOME/.talos/config" -e 192.168.1.210'
NODES    := '192.168.1.211,192.168.1.212,192.168.1.213,192.168.1.221,192.168.1.222,192.168.1.223,192.168.1.224'

# Upgrade one running Talos node.
# Upgrade workers first, then GPU workers, then controllers; verify cluster
# health between nodes.
#
# Never pass --preserve: it forces the deprecated MachineService.Upgrade path,
# whose installer runs in metal mode and rejects configs without a
# machine.install section. Talos 1.13 upgrades via LifecycleService, which
# keeps user data and drains the node on its own.
#
# Pass drain=false for a node whose pods cannot be evicted: Longhorn gives each
# instance-manager a PDB allowing zero disruptions while a volume is attached
# there, and a single-instance CNPG cluster can never release its only pod. The
# drain then burns its timeout and talosctl exits before rebooting, leaving the
# node cordoned and un-upgraded.
#   just talos-upgrade 192.168.1.223
#   just talos-upgrade 192.168.1.224 gpu
#   just talos-upgrade 192.168.1.224 gpu false
talos-upgrade node schematic='base' drain='true':
    #!/usr/bin/env bash
    set -euo pipefail

    # Derive the version and schematic IDs from Pulumi configuration.
    SRC=core/platform/talos.go
    VERSION=$(sed -nE 's/.*talosVersion[[:space:]]*=[[:space:]]*"(v[0-9.]+)".*/\1/p' "$SRC")
    IDS=$(sed -nE 's|.*factory\.talos\.dev/image/([0-9a-f]{64}).*|\1|p' "$SRC")

    case "{{schematic}}" in
      base) ID=$(echo "$IDS" | sed -n 1p) ;;
      gpu)  ID=$(echo "$IDS" | sed -n 2p) ;;
      *) echo "schematic must be 'base' or 'gpu'" >&2; exit 1 ;;
    esac

    [ -n "$VERSION" ] && [ -n "$ID" ] || {
      echo "could not parse $SRC" >&2
      exit 1
    }

    IMAGE="factory.talos.dev/installer/$ID:$VERSION"
    echo "upgrading {{node}} to $VERSION"
    echo "  $IMAGE"
    read -rp "reboot this node? [y/N] " ok
    case "$ok" in y|Y|yes|YES) ;; *) echo "aborted"; exit 1 ;; esac

    {{TALOSCTL}} upgrade --nodes {{node}} --image "$IMAGE" --drain={{drain}}
    {{TALOSCTL}} -n 192.168.1.211 health --wait-timeout=10m
    just longhorn-repair-iscsi

# Drop iSCSI node records the current open-iscsi refuses to parse.
#
# An upgrade can ship an open-iscsi that rejects a parameter its predecessor
# wrote into /var/lib/iscsi/nodes/<iqn>/<portal>/default. `iscsiadm -m node`
# reads every record, so one unparseable file fails the whole call and no
# Longhorn volume on that node can attach: the engine's frontend never starts,
# Longhorn marks the volume faulted and retries forever. The records sit on
# persistent /var, so rebooting does not clear them.
#
# Reads the rejected parameter out of iscsiadm's own error, so it keeps working
# when a later release renames a different one.
longhorn-repair-iscsi:
    #!/usr/bin/env bash
    set -euo pipefail

    for pod in $(kubectl get pods -n longhorn-system \
        -l longhorn.io/component=instance-manager -o name); do
      pod=${pod#pod/}
      node=$(kubectl get pod -n longhorn-system "$pod" \
        -o jsonpath='{.metadata.labels.longhorn\.io/node}')

      # iscsiadm lives on the host, not in this container, and the records it
      # reads are the host's. Enter the iscsid namespaces to validate; repair
      # through the /host bind mount, where the same files are writable.
      pid=$(kubectl exec -n longhorn-system "$pod" -- bash -c \
        'for p in /host/proc/[0-9]*; do grep -qa iscsid "$p/comm" 2>/dev/null && { echo "${p#/host/proc/}"; break; }; done')
      [ -n "$pid" ] || { echo "$node: iscsid not running, skipped"; continue; }

      for _ in $(seq 1 10); do
        err=$(kubectl exec -n longhorn-system "$pod" -- nsenter \
          --mount=/host/proc/"$pid"/ns/mnt --net=/host/proc/"$pid"/ns/net \
          iscsiadm -m node -o show 2>&1 >/dev/null) && break
        param=$(printf '%s' "$err" \
          | sed -nE 's/.*Unknown parameter name ([A-Za-z0-9_.]+).*/\1/p' | head -1)
        [ -n "$param" ] || break
        echo "$node: stripping $param"
        kubectl exec -n longhorn-system "$pod" -- bash -c \
          "grep -rl '$param' /host/var/lib/iscsi/nodes/ | while read -r f; do sed -i '/$param/d' \"\$f\"; done"
      done
    done

# Re-elect the Gateway's L2 announcement when its LoadBalancer IP goes dark.
#
# A Cilium agent restart can leave the lease holder renewing normally, and
# still listing the IP in db/show l2-announce, while it has quietly stopped
# answering ARP. Nothing reports unhealthy: the Gateway stays Programmed=True
# and Envoy stays ready on every node. Only the LAN notices.
#
# The discriminator is that in-cluster requests to the Gateway succeed while
# the LoadBalancer IP times out, so that is what this checks before touching
# anything. Deleting the lease hands the IP to another node within seconds.
gateway-repair-l2:
    #!/usr/bin/env bash
    set -euo pipefail

    svc=cilium-gateway-homelab-gateway
    ip=$(kubectl get svc -n kube-system "$svc" \
      -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
    [ -n "$ip" ] || { echo "no LoadBalancer IP assigned to $svc" >&2; exit 1; }

    if nc -z -w 5 "$ip" 80 2>/dev/null; then
      echo "$ip:80 reachable — nothing to do"
      exit 0
    fi

    echo "$ip:80 unreachable; re-electing L2 announcement"
    kubectl delete lease -n kube-system "cilium-l2announce-kube-system-$svc"

    for _ in $(seq 1 12); do
      sleep 5
      if nc -z -w 5 "$ip" 80 2>/dev/null; then
        echo "recovered, now held by $(kubectl get lease -n kube-system \
          "cilium-l2announce-kube-system-$svc" \
          -o jsonpath='{.spec.holderIdentity}')"
        exit 0
      fi
    done

    echo "still unreachable after re-election — look past L2" >&2
    exit 1

# Cluster health. Must be clean before upgrading the next node.
talos-health:
    {{TALOSCTL}} -n 192.168.1.211 health

# Talos version reported by every node.
talos-versions:
    {{TALOSCTL}} -n {{NODES}} version

# Upgrade Kubernetes. Separate from the OS; requires all nodes healthy.
#   just talos-upgrade-k8s 1.36.2
talos-upgrade-k8s version:
    {{TALOSCTL}} -n 192.168.1.211 upgrade-k8s --to {{version}}
