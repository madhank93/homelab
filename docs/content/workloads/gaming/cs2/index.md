+++
title = "Counter-Strike 2"
description = "LAN-only Counter-Strike 2 dedicated server."
weight = 10
+++

## How It's Used Here

{{ src(path="workloads/gaming/cs2.go") }} runs the
[joedwards32/cs2](https://github.com/joedwards32/CS2) image as a single-replica
Deployment in the `cs2` namespace.

- **LAN only.** `CS2_LAN=1` starts the server with `sv_lan 1`: no Steam game
  server token is needed, the server is not listed in the public browser, and
  only clients with private addresses can join.
- **Fixed address.** A `LoadBalancer` Service pins `192.168.1.230` through the
  `lbipam.cilium.io/ips` annotation and exposes `27015/udp` (game) and
  `27015/tcp` (RCON).
- **Game files on Longhorn.** steamcmd downloads the ~60 GB game into an 80Gi
  PVC (`longhorn-cs2`, 1 replica) on every start and updates it in place.
  The first start takes as long as the download.
- **Defaults:** competitive, `de_inferno` from `mg_active`, 10 players, no bots.
  Change the `CS2_*` env in `cs2.go`.

## Secrets

`RCON_PASSWORD` and `JOIN_PASSWORD` live in OpenBao at `secret/cs2`, seeded by
`just openbao-setup` and mirrored into the `cs2-secrets` Secret by the CSI
driver. Read the join password to hand out:

```bash
just openbao-get secret/cs2 JOIN_PASSWORD
```

## Playing

From the CS2 console (enable it under Settings → Game → Enable Developer Console):

```text
connect 192.168.1.230; password <JOIN_PASSWORD>
```

## Updating the game

A CS2 patch only needs a restart; the container updates before it starts:

```bash
kubectl -n cs2 rollout restart deployment/cs2
```
