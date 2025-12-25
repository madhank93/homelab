+++
title = "Hetzner Bifrost"
weight = 30
+++

# Hetzner Bifrost VPS

The "Bifrost" server serves as a gateway/jump-host hosted on Hetzner Cloud. It is defined in `infra/pulumi/hetzner_vps.go`.

## Server Specifications

- **Name**: `bifrost-public-vps1`
- **Type**: `cpx21`
- **Image**: `ubuntu-24.04`
- **Location**: `ash` (Ashburn, VA)

## Firewall Configuration

A strict firewall (`bifrost-fw`) is applied to the server using `pulumi-hcloud`.

### Inbound Rules (Ingress)

| Port | Protocol | Description | Source |
|------|----------|-------------|--------|
| 22 | TCP | SSH Access | Any (`0.0.0.0/0`, `::/0`) |
| 80 | TCP | HTTP | Any |
| 443 | TCP | HTTPS | Any |
| 3478 | TCP/UDP | STUN/TURN | Any |
| 5349 | TCP/UDP | STUN/TURN TLS | Any |
| 50000-50500 | UDP | TURN Ephemeral Range | Any |

## Bootstrapping

After the VPS is provisioned:
1.  **File Copy**: The contents of the `./bifrost` directory are copied to `/etc` on the remote server.
2.  **Setup Script**: A bootstrap script (`/etc/bifrost/bootstrap.sh`) is executed via SSH to finalize the configuration.
