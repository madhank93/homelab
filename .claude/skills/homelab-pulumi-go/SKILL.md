---
name: homelab-pulumi-go
description: "Homelab Pulumi (Golang) conventions for core/ — stack routing in core/main.go (talos, platform, hetzner, authentik, cloudflare), secrets arriving as env vars through sops exec-env and cfg.K, non-secret settings from the shared config module, resource options (Protect, IgnoreChanges, DeleteBeforeReplace, DependsOn, Provider), Output and ApplyT rules, the Terraform-bridged Authentik SDK generated into core/sdks from Pulumi.yaml, and reading just core <stack> preview output for unexpected replaces. Use whenever a task touches core/, Pulumi.yaml, Proxmox VMs, Talos machine config, Cilium, cert-manager or Argo CD installation, the Hetzner Bifrost VPS, Authentik applications, or Cloudflare DNS records. Do NOT use for CDK8s workloads in workloads/ — use homelab-cdk8s. Do NOT use for general Go style — use golang-code-style, golang-error-handling or golang-structs-interfaces."
---

# Homelab Pulumi

One Go program, five stacks. Changes here hit live infrastructure (etcd-member VMs, cluster PKI, public DNS), so the job is to keep `preview` boring.

## Hard rules

- Never read, print or decrypt anything under `secrets/`, and never print `core/kubeconfig`, `core/talosconfig` or `core/cloud/bifrost/env/`; they hold live credentials.
- Never run `pulumi` or `just core`. Pulumi runs only from the user's laptop, never CI; `just core` decrypts SOPS. Ask the user to run `just core <stack> preview` and paste the output.
- `just core <stack> <action>` adds `--yes` to every action except `preview`, so `up` applies without a confirmation prompt. Preview first, every time.
- Never create resources inside `ApplyT`; Pulumi cannot plan them and preview will not show them.
- Never rename a resource's logical name (the second argument of `New*`) or move it under a different parent without `pulumi.Aliases`; otherwise preview shows delete + create.
- Never edit `core/sdks/` (generated, gitignored).

## Stack map (`core/main.go`)

| Stack | Entry point | Depends on |
|---|---|---|
| `talos` | `platform.DeployTalosCluster` | Proxmox; writes `core/kubeconfig` and `core/talosconfig` |
| `platform` | `platform.DeployPlatform` (Gateway API CRDs, Cilium, Gateway, cert-manager, Argo CD) | `core/kubeconfig` from `talos` |
| `hetzner` | `cloud.DeployHetznerVPS` (Bifrost VPS, bootstrap via `remote.Command`) | — |
| `authentik` | `cloud.DeployAuthentik` | Authentik running on Bifrost |
| `cloudflare` | `cloud.ManageCloudflare` (wildcard + `publicServices` records) | Bifrost IP from `config/config.yml` |

Add a stack in three places: the `switch` in `main.go` (and its error message), an empty `Pulumi.<stack>.yaml`, and the stack list in the `justfile` comment.

## Configuration

| Kind | Source | Read with |
|---|---|---|
| Non-secret (IPs, domain, node specs, versions) | `config/config.yml`, `config/versions.yaml` | `config.C.*`, `config.HelmChart`, `config.ContainerImage` |
| Secret (API tokens, OAuth client secrets) | `secrets/bootstrap.sops.yaml` → `sops exec-env` → process env | `cfg.K.String("NAME")` |

- `Pulumi.<stack>.yaml` stays `config: {}`; never add Pulumi config or secrets there.
- Check each required secret is non-empty and return an error naming it (`cloud/hetzner.go`, `cloud/authentik.go`).
- Wrap a secret placed in a resource input with `pulumi.ToSecret` so state and preview mask it (`platform/argocd.go`, `platform/proxmox.go`).
- A new secret needs the user to add it to the SOPS file; ask, never edit it.

## Resource options in use

| Option | Use when | Example |
|---|---|---|
| `pulumi.Protect(true)` | Replacement destroys state that cannot be recreated | Talos `NewSecrets` (cluster PKI), Proxmox VMs |
| `pulumi.IgnoreChanges` | An input only seeds creation, or the provider reads it back differently | VM `disks[0].fileId`/`disks[0].speed`; image `fileName` |
| `pulumi.DeleteBeforeReplace(true)` | Old and new cannot coexist (same file name on disk) | `DownloadImage` in `platform/proxmox.go` |
| `pulumi.DependsOn` | Ordering not expressed by an Output (CRDs before CRs, chart before its Secret) | `platform/cilium.go`, `platform/cert_manager.go` |
| `pulumi.Provider(p)` | Every resource of a non-default provider | `k8sProvider`, Authentik and Proxmox providers |

Write a one-line comment beside each `Protect`/`IgnoreChanges` saying what it protects against; the existing ones are the model.

## Outputs

- Use `ApplyT` only to transform a value or write a local file (`platform/talos.go` writes kubeconfig/talosconfig).
- Pass Outputs straight into inputs instead of resolving them; use `.ApplyT` + a typed func for conversions (`idFloat` in `cloud/authentik.go`).
- Use data-source lookups (`Lookup*`, `Get*`) for objects Pulumi does not own (Authentik default flows, signing key).

## Bridged SDK (Authentik)

`Pulumi.yaml` `packages.authentik` pins the Terraform provider (`goauthentik/authentik`) and its version; Pulumi generates the Go SDK into `core/sdks/authentik/`, wired in by a `replace` in `core/go.mod`. Bump the version in `Pulumi.yaml`, regenerate on the laptop, then `cd core && go build ./...`. CI cannot build `core/` because the SDK is not committed.

## Verify

1. `just lint core` and `cd core && go build ./...` — fix until clean.
2. Ask the user for `just core <stack> preview`; read it with [references/preview.md](references/preview.md).
3. Only after a clean preview, the user runs `just core <stack> up`.

Before upgrading a provider or chart pinned in `config/versions.yaml`, check its GitHub issues for the target version. For Go style, load `golang-code-style`, `golang-error-handling` or `golang-structs-interfaces`.
