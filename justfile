# justfile

# Backtick values below (yq reads of config/) are evaluated only by recipes that use them.
set lazy

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

# Enable the gitleaks + trufflehog secret scan on commit and push (needs Docker).
hooks:
    git config core.hooksPath .githooks

# Scan the whole git history for secrets (both scanners, findings redacted).
secret-scan-history:
    docker run --rm -v "$PWD:/repo" ghcr.io/gitleaks/gitleaks:v8.30.1 git --no-banner --redact /repo
    docker run --rm -v "$PWD:/repo" trufflesecurity/trufflehog:3.99.2 git file:///repo --no-update --results=verified,unknown

# Docs: every version in the Software Inventory must still be in the file that
# pins it, and no other page may state a version. Also runs in CI.
docs-check:
    ./scripts/docs-version-check.sh

# Serve the docs locally at http://127.0.0.1:1111
[working-directory: 'docs']
docs-serve:
    zola serve

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

# kubectl exec into the OpenBao server; bao_root takes the token as the first stdin
# line so it never appears in a process list.
bao_exec := "kubectl exec -i -n openbao openbao-0 -c openbao --"
bao_root := bao_exec + " sh -c 'read -r BAO_TOKEN; export BAO_TOKEN; exec bao \"$@\"' bao"

# Saves only the unseal key (mode 600); the init root token is revoked at once.
# One-time OpenBao initialisation (run after first ArgoCD deploy of OpenBao)
openbao-init:
    #!/usr/bin/env bash
    set -euo pipefail
    umask 077
    out="$HOME/openbao-unseal-key"
    if [[ -e "$out" ]]; then echo "❌  $out already exists — move it into SOPS or delete it first." >&2; exit 1; fi
    init="$({{bao_exec}} bao operator init -key-shares=1 -key-threshold=1 -format=json </dev/null)"
    jq -r '.unseal_keys_b64[0]' <<<"$init" >"$out"
    jq -r '.unseal_keys_b64[0]' <<<"$init" | {{bao_exec}} sh -c 'read -r k; bao operator unseal "$k" >/dev/null'
    jq -r '.root_token' <<<"$init" | {{bao_root}} token revoke -self >/dev/null
    echo "✅  OpenBao initialised and unsealed; init root token revoked."
    echo "    Unseal key written to $out (mode 600). Next:"
    echo "      1. sops secrets/bootstrap.sops.yaml  → set OPENBAO_UNSEAL_KEY to the contents of $out"
    echo "      2. just create-secrets && kubectl delete pod -n openbao openbao-0   # OnDelete StatefulSet"
    echo "      3. rm $out"
    echo "      4. just openbao-setup"

# Configure OpenBao K8s auth, policies, roles and seed app secrets (idempotent)
openbao-setup:
    SOPS_AGE_KEY_FILE="$HOME/.config/sops/age/keys.txt" \
      sops exec-env secrets/bootstrap.sops.yaml 'bash scripts/openbao-setup.sh'

# Refuses to print to a terminal; capture it and revoke it when done:
#   export ROOT_TOKEN=$(just openbao-token) ... just openbao-revoke
# Generate a root token from the unseal key in the openbao-unseal-key Secret.
# Goes through the in-pod socket: the only listener allowing generate-root (openbao.go).
openbao-token:
    #!/usr/bin/env bash
    set -euo pipefail
    if [[ -t 1 ]]; then echo "Capture the token instead: export ROOT_TOKEN=\$(just openbao-token)" >&2; exit 1; fi
    gen="{{bao_exec}} env BAO_ADDR=unix:///tmp/openbao.sock bao"
    $gen delete sys/generate-root/attempt </dev/null >/dev/null
    init="$($gen write -format=json -f sys/generate-root/attempt </dev/null)"
    nonce="$(jq -r .data.nonce <<<"$init")"
    otp="$(jq -r .data.otp <<<"$init")"
    encoded="$(kubectl get secret openbao-unseal-key -n openbao -o jsonpath='{.data.unseal-key}' | base64 -d |
      $gen write -format=json sys/generate-root/update nonce="$nonce" key=- | jq -r .data.encoded_token)"
    # The bao CLI decodes via the authenticated endpoint; the token is the
    # unpadded-base64 encoded_token XORed with the OTP.
    printf '%s %s' "$encoded" "$otp" | python3 -c 'import base64, sys
    e, o = sys.stdin.read().split()
    print(bytes(b ^ ord(c) for b, c in zip(base64.b64decode(e + "=" * (-len(e) % 4)), o)).decode())'

# Read an OpenBao secret with a throwaway root token. Optionally pass a field name.
# Usage:
#   just openbao-get secret/grafana
#   just openbao-get secret/grafana ADMIN_PASSWORD
openbao-get path field='':
    #!/usr/bin/env bash
    set -euo pipefail
    token="$(just openbao-token)"
    bao() { printf '%s\n' "$token" | {{bao_root}} "$@"; }
    trap 'bao token revoke -self >/dev/null' EXIT
    if [[ -n "{{field}}" ]]; then bao kv get -field="{{field}}" "{{path}}"; else bao kv get "{{path}}"; fi

# Sync one Argo CD Application from git, without the argocd CLI.
# Used while argocd_manual_sync is true. Example: just argocd-sync openbao
argocd-sync app:
    kubectl patch application {{app}} -n argocd --type merge \
      -p '{"operation":{"initiatedBy":{"username":"just"},"sync":{"prune":true,"syncOptions":["CreateNamespace=true","ServerSideApply=true"]}}}'
    @echo "watch: kubectl get application {{app}} -n argocd -w"

# Revoke a root token from openbao-token (argument, or $ROOT_TOKEN if omitted)
openbao-revoke $token='':
    #!/usr/bin/env bash
    set -euo pipefail
    printf '%s\n' "${token:-${ROOT_TOKEN:?pass a token or export ROOT_TOKEN}}" | {{bao_root}} token revoke -self

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

# One-time: create Grafana's "Public" org (must get id 2, see workloads/monitoring/grafana.go).
# Grafana won't start while provisioning references a missing org, so run this
# against the running Grafana BEFORE syncing the two-org config. Idempotent.
grafana-init:
    #!/usr/bin/env bash
    set -euo pipefail
    pw=$(kubectl exec -n grafana deploy/grafana -c grafana -- cat /mnt/secrets/ADMIN_PASSWORD)
    auth="header = \"Authorization: Basic $(printf 'admin:%s' "$pw" | base64 | tr -d '\n')\""
    unset pw
    kubectl port-forward -n grafana svc/grafana 13030:3000 >/dev/null &
    pf=$!
    trap 'kill $pf' EXIT
    for _ in $(seq 20); do curl -sf -o /dev/null http://localhost:13030/api/health && break; sleep 0.5; done
    api() { printf '%s\n' "$auth" | curl -sf -K - -H 'Content-Type: application/json' "$@"; }
    id=$(api http://localhost:13030/api/orgs/name/Public | jq -r .id) \
      || id=$(api -X POST -d '{"name":"Public"}' http://localhost:13030/api/orgs | jq -r .orgId)
    [ "$id" = 2 ] || { echo "Grafana org Public has id $id, expected 2" >&2; exit 1; }
    echo "Grafana org Public = id 2"

# Short-lived Headlamp login token (cluster-admin); paste into the Headlamp UI
headlamp-token duration="24h":
    kubectl create token headlamp-admin -n headlamp --duration={{duration}}

# Decrypt and view a SOPS-encrypted file (opens in $EDITOR by default)
# Usage:
#   just sops-view secrets/bootstrap.sops.yaml
# To print to stdout instead:
#   just sops-decrypt secrets/bootstrap.sops.yaml
sops-view file:
    SOPS_AGE_KEY_FILE="$HOME/.config/sops/age/keys.txt" sops {{file}}

sops-decrypt file:
    SOPS_AGE_KEY_FILE="$HOME/.config/sops/age/keys.txt" sops -d {{file}}

SUBNET     := `yq -e '.cluster.subnet' config/config.yml`
VIP        := `yq -e '.cluster.vip' config/config.yml`
CONTROLLER := `yq -e '[.cluster.nodes[] | select(.role == "control")][0].ip' config/config.yml`
NODES      := `yq -e '[.cluster.nodes[].ip] | join(",")' config/config.yml`
TALOSCTL   := 'talosctl --talosconfig "$HOME/.talos/config" -e ' + VIP

ping_scan:
    nmap -sn {{SUBNET}}

# Upgrade one running Talos node.
# Upgrade controllers one at a time first, then workers, then the GPU worker
# last; verify cluster health between nodes.
#
# Never pass --preserve: it forces the deprecated MachineService.Upgrade path,
# whose installer runs in metal mode and rejects configs without a
# machine.install section. Talos 1.13+ upgrades via LifecycleService, which
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

    # Same version and schematics Pulumi builds new VMs from.
    case "{{schematic}}" in
      base|gpu) ;;
      *) echo "schematic must be 'base' or 'gpu'" >&2; exit 1 ;;
    esac
    # The wrong schematic silently drops (or adds) the NVIDIA extensions.
    case "$(yq ".cluster.nodes[] | select(.ip == \"{{node}}\") | has(\"pcie_ids\")" config/config.yml)" in
      true) want=gpu ;;
      false) want=base ;;
      *) echo "{{node}} is not in config/config.yml" >&2; exit 1 ;;
    esac
    [ "{{schematic}}" = "$want" ] || { echo "{{node}} needs schematic '$want'" >&2; exit 1; }
    VERSION=$(yq -e '.talos.version' config/versions.yaml)
    ID=$(yq -e '.talos.{{schematic}}_schematic' config/versions.yaml)

    IMAGE="factory.talos.dev/installer/$ID:$VERSION"
    echo "upgrading {{node}} to $VERSION"
    echo "  $IMAGE"
    read -rp "reboot this node? [y/N] " ok
    case "$ok" in y|Y|yes|YES) ;; *) echo "aborted"; exit 1 ;; esac

    {{TALOSCTL}} upgrade --nodes {{node}} --image "$IMAGE" --drain={{drain}}
    {{TALOSCTL}} -n {{CONTROLLER}} health --wait-timeout=10m
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

# Start or stop ComfyUI.
#
# Argo CD runs selfHeal, so `kubectl scale` is reverted within minutes — the
# replica count has to change in git. Scaling to 0 is also the only thing that
# reliably frees GPU memory: ComfyUI caches its model in VRAM after a run and
# does not release it, which would leave Ollama unable to load one. The card has
# 16 GB and no Kubernetes limit applies to it, so the two do not share well.
#   just comfyui on
#   just comfyui off
comfyui state:
    #!/usr/bin/env bash
    set -euo pipefail

    case "{{state}}" in
      on)  want=1 ;;
      off) want=0 ;;
      *)   echo "usage: just comfyui on|off" >&2; exit 1 ;;
    esac

    # The replica count lives in a literal tagged `// comfyui-replicas`.
    f=workloads/ai/comfyui.go
    re='(replicas := float64\()\d+(\) // comfyui-replicas)'

    have=$(perl -ne 'print $1 if m{replicas := float64\((\d+)\) // comfyui-replicas}' "$f")
    if [ -z "$have" ]; then
      echo "no comfyui-replicas marker in $f" >&2
      exit 1
    fi
    if [ "$have" = "$want" ]; then
      echo "ComfyUI is already {{state}}"
      exit 0
    fi

    perl -pi -e "s{$re}{\${1}$want\${2}}" "$f"
    just synth
    git add "$f"
    git commit -m "chore(comfyui): turn {{state}}"
    git push
    echo
    echo "pushed — Argo CD applies it once CI republishes the manifests branch"
    [ "$want" = "1" ] && echo "first start pulls a 12.3 GB image: kubectl get pods -n comfyui -w" || true

# Cluster health. Must be clean before upgrading the next node.
talos-health:
    {{TALOSCTL}} -n {{CONTROLLER}} health

# Talos version reported by every node.
talos-versions:
    {{TALOSCTL}} -n {{NODES}} version

# Upgrade Kubernetes to the version pinned in config/versions.yaml. Separate
# from the OS; requires all nodes healthy. Run it before `just core talos up`,
# which would otherwise push the new images to every node at once.
#   just talos-upgrade-k8s --dry-run
#   just talos-upgrade-k8s
talos-upgrade-k8s *flags:
    {{TALOSCTL}} -n {{CONTROLLER}} upgrade-k8s --to "$(yq -e '.kubernetes.version' config/versions.yaml)" {{flags}}
