---
title: Bifrost
description: Bump Traefik, NetBird, Authentik and Gatus on the Hetzner VPS, independent of the cluster.
sidebar:
  order: 16
---

Phase 7. Risk: low, independent of the cluster, but every public URL goes through it. Image tags live in [`docker-compose.yml`](src:core/cloud/bifrost/docker-compose.yml).

| Image | Note |
|---|---|
| `traefik` | |
| `netbirdio/netbird-server`, `netbirdio/reverse-proxy`, `netbirdio/netbird` (agent) | Same version, always |
| `netbirdio/dashboard` | Separate `v2.x` line; take the release current when the server version shipped |
| `ghcr.io/goauthentik/server` | Two services: server and worker |
| `postgres` | Minor bumps only |
| `ghcr.io/twin/gatus` | |

:::caution[NetBird versions move together]
Server, reverse-proxy, agent and the in-cluster peer must run the same version. Bump `images.netbird` in [`config/versions.yaml`](src:config/versions.yaml) in the same commit.
:::

```bash frame="terminal"
just core hetzner preview   # no server replace: a user_data change replaces the VPS and its IP
just core hetzner up
ssh root@178.156.199.250 "ip rule show | grep 7120"   # NetBird policy route present
```

Gate: Gatus all green at `uptime.madhan.app`.

## Authentik migrations

[`bootstrap.sh`](src:core/cloud/bifrost/bootstrap.sh) runs `migrate` in a one-off server container before starting `authentik-server` and `authentik-worker`, so the server never queries columns that do not exist yet. If `just core hetzner up` fails at the Authentik health check:

```bash frame="terminal"
docker exec authentik-server ak migrate
docker restart authentik-server authentik-worker
```

Authentik refuses to skip a minor: the newest row of `authentik_version_history` must be the previous minor.

## NetBird ip rules

After restarting `netbird-agent`, `ip rule show` must list table `7120`, which carries the `192.168.1.0/24` route Traefik needs. If it is missing, restart the container again: `docker restart netbird-agent`.
