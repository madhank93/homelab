+++
title = "Hetzner Bifrost"
description = "Hetzner VPS running NetBird VPN management plane and TURN server (Bifrost)."
weight = 30
+++

## Overview

A lightweight Hetzner Cloud VPS serves as **Bifrost** — the network bridge for the homelab. It runs:

- **NetBird** management plane — VPN mesh coordination for remote access to the cluster
- **TURN server** — WebRTC relay for NAT traversal

The VPS is provisioned by Pulumi from `infra/pulumi/hetzner_vps.go`.

## Pulumi Configuration

The Hetzner stack reads config from a `hetzner` key in `infra/pulumi/config.yaml`:

```yaml
server_name: bifrost
image: ubuntu-24.04
server_type: cx22
location: nbg1
ssh_key: <your-hetzner-ssh-key-name>
```

`HCLOUD_TOKEN` is injected at runtime via SOPS from `infra/secrets/bootstrap.sops.yaml`.

## Firewall Rules

The Bifrost firewall allows inbound:

| Protocol | Port | Purpose |
|----------|------|---------|
| TCP | 22 | SSH |
| TCP | 80 | HTTP |
| TCP | 443 | HTTPS / NetBird management |
| TCP+UDP | 3478 | STUN |
| TCP+UDP | 5349 | TURNS (TLS TURN) |
| UDP | 50000–50500 | TURN ephemeral relay range |

## Cloud-Init

Server configuration (NetBird + TURN setup) is applied via `infra/pulumi/cloud-init/cloud-init-hetzner.yml` at first boot.

Additional config files are copied from `infra/pulumi/bifrost/` to `/etc/` on the remote server via `pulumi-command` `CopyToRemote`.
