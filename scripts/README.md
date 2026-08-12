# scripts

Bootstrap helpers invoked from the `justfile`. Run them through `just`, not
directly — they expect environment variables that `sops exec-env` provides.

| Script | Recipe | Purpose |
|---|---|---|
| `create-bootstrap-secrets.sh` | `just create-secrets` | Creates the two bootstrap Secrets (OpenBao unseal key, Cloudflare API token) from `secrets/bootstrap.sops.yaml`. |
| `openbao-setup.sh` | `just openbao-setup` | One-time OpenBao configuration: Kubernetes auth, policies, roles. |

## Talos upgrades

Talos node upgrades live in the `justfile`, not here:

```bash
just talos-versions                    # what each node runs today
just talos-upgrade 192.168.1.223       # one worker
just talos-upgrade 192.168.1.224 gpu   # worker4, nvidia schematic
just talos-health                      # must be clean before the next node
just talos-upgrade-k8s 1.36.2          # Kubernetes, after all nodes are done
```

Upgrade workers first, then the GPU worker, then controllers one at a time.
The target version and both schematic IDs are read from
`core/platform/talos.go`, so they cannot drift from what Pulumi provisions.

`just core talos up` does not upgrade running nodes — it only stages the image
on the Proxmox datastore for VMs created afterwards.
