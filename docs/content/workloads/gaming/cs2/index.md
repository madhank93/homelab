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

## Player setup

The server is LAN only: players must be on the home network (a `192.168.1.x`
address). Wi-Fi works; wired is steadier.

1. **Install Steam** from [store.steampowered.com](https://store.steampowered.com/about/)
   and sign in (a free account is enough).
2. **Install CS2.** Search *Counter-Strike 2* in the Steam store and click
   *Play Game*. It is free to play; Steam asks for ~85 GB of free disk space.
3. **Launch it once** and finish the intro so the game writes its settings.
4. **Enable the console:** gear icon (Settings) → Game →
   *Enable Developer Console (~)* → Yes. Press <kbd>~</kbd> (the key under
   <kbd>Esc</kbd>, left of <kbd>1</kbd>) to open it. If nothing opens — common
   on non-US keyboards — bind another key under Settings → Keyboard / Mouse →
   *Toggle Console*.
5. **Join from the console.** Type `connect 192.168.1.230; password ` (with
   the trailing space), paste the password the host shared with
   <kbd>Ctrl</kbd>+<kbd>V</kbd>, and press <kbd>Enter</kbd>:

   ```text
   connect 192.168.1.230; password <JOIN_PASSWORD>
   ```

   Paste rather than type: the password is long and random, and one wrong
   character is rejected.

### Server browser

The **LAN** tab stays empty: it finds servers by broadcast, which the
LoadBalancer IP does not answer. Add the server to **Favorites** instead —
Play → Community Server Browser → Favorites → *Add a server* →
`192.168.1.230:27015`.

The browser's password prompt closes as soon as it opens and joins with an
empty password (`NETWORK_DISCONNECT_REJECT_BADPASSWORD`). Set the password in
the console first with `password <JOIN_PASSWORD>`, then double-click the
favorite.

### Troubleshooting

| Symptom | Fix |
|---|---|
| Connect times out | The PC is not on the home network (guest Wi-Fi or a VPN). Check with `Test-NetConnection 192.168.1.230 -Port 27015` in PowerShell; it should print `TcpTestSucceeded : True`. |
| "Bad password" | The password was empty or mistyped. Paste it in the console as above. |
| Ping above ~40 ms or spiky | Move to a wired connection or closer to the router. 1–10 ms is normal on the LAN. |

### Low-spec PCs

CS2 needs 4 CPU threads, 8 GB RAM and a DirectX 11 GPU. Laptops on integrated
graphics (e.g. Intel UHD 620) run it at roughly 20–35 FPS; to get the most out
of one:

- Plug in the charger and set Windows power mode to *Best performance*.
- Run RAM in dual channel (two sticks) — the biggest gain for integrated graphics.
- Install the latest GPU driver from the vendor (Intel, AMD or NVIDIA).
- In Video settings: 1280×720 or 1024×768, every option on lowest, and
  FidelityFX Super Resolution on *Performance*.
- On laptops with a dedicated GPU, set `cs2.exe` to *High performance* under
  Windows Settings → System → Display → Graphics.
- Close the browser and overlays before playing.

## Changing the map

From any connected client's console, with `RCON_PASSWORD` from `secret/cs2`:

```text
rcon_address 192.168.1.230:27015
rcon_password <RCON_PASSWORD>
rcon changelevel de_mirage
```

## Updating the game

A CS2 patch only needs a restart; the container updates before it starts:

```bash
kubectl -n cs2 rollout restart deployment/cs2
```
