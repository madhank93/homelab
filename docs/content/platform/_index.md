+++
title = "Platform"
description = "core/platform/ Pulumi stacks: Proxmox/Talos cluster, Cilium CNI, Argo CD GitOps, cert-manager, and secrets."
weight = 50
sort_by = "weight"
+++

Platform covers both cluster provisioning (`core/platform/`) and the GitOps delivery layer.

## Pulumi Stacks (`core/platform/`)

| Stack | Command | Manages |
|-------|---------|---------|
| `talos` | `just core talos up` | Proxmox VMs and Talos cluster bootstrap |
| `platform` | `just core platform up` | Cilium CNI, Gateway API + IP pool, cert-manager, Argo CD |

## GitOps Layer

- **No secrets in git** — CDK8s generates zero `Secret` resources. Bootstrap secrets are created by `just create-secrets` from a laptop.
- **Manifests are generated, not hand-written** — The `v0.1.7-manifests` branch is machine-generated YAML. Never edit it by hand.
- **One CDK8s app per Argo CD Application** — Each `main.go` entry writes to a separate `app/<name>/` directory, which becomes one Argo CD Application.
- **Argo CD is the single source of truth** — All drift is auto-corrected (`selfHeal: true`), all removed resources are pruned (`prune: true`).
- **The chart pins the version; never float a tag** — Every Helm release carries an explicit chart version, and any image tag we set is pinned. `latest` is never used.

### Running an image ahead of its chart

Charts sometimes lag an upstream release. Overriding the image tag to get there
early is allowed **only** when all of the following hold:

1. It is a **patch** bump (`x.y.0` → `x.y.1`). Never a minor or major — a newer
   binary may expect CRD fields or flags the older chart does not render.
2. The upstream diff between the two tags changes **no CRDs and no manifests**
   beyond the image tag itself. Verify it, do not assume:
   ```bash
   curl -sL "https://api.github.com/repos/<org>/<repo>/compare/<old>...<new>" \
     | python3 -c "import json,sys; [print(f['filename']) for f in json.load(sys.stdin)['files']]"
   ```
3. The override renders as intended — confirm with `helm template --set ...`
   rather than trusting the value key. A wrong key is silently ignored.
4. A comment records **why** and says to drop it once the chart catches up.

Current instance: `core/platform/argocd.go` runs an Argo CD image ahead of the one
its chart pins, for the server-side-diff Secret-masking fixes. Upstream's own diff
between the two is nothing but the image tag. Both numbers are in the
[Software Inventory](@/architecture/software-inventory.md).

See [GitOps Flow](@/architecture/gitops-flow/index.md) for the end-to-end diagram.
