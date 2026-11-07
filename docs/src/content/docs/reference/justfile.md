---
title: justfile
description: Every recipe in the repo's justfile, grouped by purpose, with its arguments and what it does.
sidebar:
  order: 4
---

All recipes run from the repo root. Values such as the VIP and node IPs are read from `config/config.yml` with `yq`. Source: [`justfile`](src:justfile).

## Provisioning

| Recipe | Args | What it does |
|---|---|---|
| `core` | `stack action` | Runs `pulumi <action>` on a stack (`talos`, `platform`, `hetzner`, `authentik`, `cloudflare`) inside `sops exec-env` of the bootstrap secrets. Adds `--yes` except for `preview` |
| `synth` | — | CDK8s synthesis (`go run .` in `workloads/`) into `app/` |
| `ping_scan` | — | `nmap -sn` over the cluster subnet |

## Talos

| Recipe | Args | What it does |
|---|---|---|
| `talos-upgrade` | `node [schematic=base] [drain=true]` | Upgrades one node to the pinned Talos with the Image Factory installer. Refuses the wrong schematic, asks before rebooting, waits for health, then runs `longhorn-repair-iscsi` |
| `talos-upgrade-k8s` | `*flags` | `talosctl upgrade-k8s` to the pinned Kubernetes; pass `--dry-run` first |
| `talos-health` | — | Cluster health; must be clean before the next node |
| `talos-versions` | — | Talos version on every node |

## Secrets

| Recipe | Args | What it does |
|---|---|---|
| `create-secrets` | — | Creates the bootstrap Kubernetes Secrets from SOPS (`scripts/create-bootstrap-secrets.sh`) |
| `openbao-init` | — | One-time: initialises and unseals OpenBao, writes the unseal key to `~/openbao-unseal-key` (mode 600), revokes the init root token, prints next steps |
| `openbao-setup` | — | Kubernetes auth, policies, roles and seeded app secrets (`scripts/openbao-setup.sh`). Idempotent; never overwrites values |
| `openbao-root` | `bao args…` | Runs one `bao` command with a throwaway root token (generate-root over the in-pod socket), revoked on exit. Piped stdin reaches `bao` |
| `openbao-token` | — | Root token kept for longer, e.g. for the UI. Refuses a terminal: `just openbao-token \| pbcopy` |
| `openbao-revoke` | stdin | Revokes the token piped in: `pbpaste \| just openbao-revoke` |
| `openbao-get` | `path [field]` | Reads a KV path with a throwaway root token |
| `openbao-sa-token` | `sa ns role` | Short-lived OpenBao token via Kubernetes auth for a service account |
| `openbao-sa-get` | `sa ns role path [field]` | Reads a KV path as that service account |
| `sops-view` | `file` | Opens a SOPS file decrypted in `$EDITOR` |
| `sops-decrypt` | `file` | Prints a SOPS file decrypted to stdout |
| `hooks` | — | Enables the gitleaks + trufflehog pre-commit and pre-push hooks (`core.hooksPath .githooks`) |
| `secret-scan-history` | — | Scans the whole git history with both scanners, findings redacted |

## Repairs and operations

| Recipe | Args | What it does |
|---|---|---|
| `gateway-repair-l2` | — | Re-elects the Gateway's L2 announcement when its LB IP is dark — see [Repairs](/operate/repairs/) |
| `longhorn-repair-iscsi` | — | Strips iSCSI node-record parameters the current open-iscsi rejects |
| `comfyui` | `on\|off` | Sets the ComfyUI replica literal, synths, commits and pushes |
| `argocd-sync` | `app` | Triggers a sync of one Argo CD Application |
| `grafana-init` | — | One-time: creates Grafana org "Public" (must be id 2) |
| `headlamp-token` | `[duration=24h]` | Cluster-admin login token for Headlamp |

## Images

| Recipe | Args | What it does |
|---|---|---|
| `build-push` | `image tag` | Builds `images/<image>` for `linux/amd64` and pushes to `harbor.madhan.app/library/<image>:<tag>` |

## Docs

| Recipe | Args | What it does |
|---|---|---|
| `docs-serve` | — | `npm run dev` in `docs/`, at `http://localhost:4321` |
| `docs-check` | — | Version and link guard (`scripts/docs-version-check.sh`); also runs in CI |
