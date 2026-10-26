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

## Server settings

Edit `workloads/gaming/cs2_config.yml` and push. Reloader restarts the server
(about 2 minutes) and the new values apply from the first map load.

`convars` are written to `gamemode_<mode>_server.cfg`, which CS2 runs after its
own `gamemode_<mode>.cfg` on every map load, so they win over Valve's mode
defaults. Any other server convar can be added there too.

### `server`

| Key | Image env var | Meaning |
|---|---|---|
| `name` | `CS2_SERVERNAME` | Name shown in the server browser and scoreboard. |
| `game_alias` | `CS2_GAMEALIAS` | Game mode: `competitive`, `casual`, `wingman`, `deathmatch`, `armsrace`. |
| `map_group` | `CS2_MAPGROUP` | Map pool for the end-of-match vote. `mg_active` is the Active Duty pool. |
| `start_map` | `CS2_STARTMAP` | Map loaded at startup. |
| `max_players` | `CS2_MAXPLAYERS` | Player slots. |

Source: [joedwards32/CS2](https://github.com/joedwards32/CS2) (the server image).

### `convars`

*Engine* is the built-in default; *competitive* is what
`gamemode_competitive.cfg` sets, and what `cs2_config.yml` ships with.

| Convar | Engine | Competitive | Meaning |
|---|---|---|---|
| **Economy** | | | |
| `mp_startmoney` | 800 | 800 | Money each player starts with. |
| `mp_maxmoney` | 16000 | 16000 | Most money a player can hold. |
| `mp_buytime` | 90 | 20 | Seconds after round start that buying stays open. |
| `mp_buy_anywhere` | 0 | 0 | Buy outside buy zones: 0 off, 1 both teams, 2 T, 3 CT. |
| `mp_afterroundmoney` | 0 | 0 | Money given to every player after each round. |
| `cash_team_loser_bonus` | 1400 | 1400 | Base money for the losing team. |
| `cash_team_loser_bonus_consecutive_rounds` | 500 | 500 | Extra loss bonus per consecutive loss. |
| `cash_player_killed_enemy_default` | 300 | 300 | Kill reward (weapons with their own reward override it). |
| **Round and match timing** | | | |
| `mp_roundtime` | 5 | 1.92 | Round length in minutes (1.92 = 1:55). |
| `mp_roundtime_defuse` | 0 | 1.92 | Round length on bomb maps; 0 uses `mp_roundtime`. |
| `mp_roundtime_hostage` | 0 | 1.92 | Round length on hostage maps; 0 uses `mp_roundtime`. |
| `mp_freezetime` | 6 | 15 | Seconds players are frozen at round start (0–60). |
| `mp_maxrounds` | 0 | 24 | Rounds per match before the map changes. |
| `mp_halftime` | 0 | 1 | Swap sides at half time. |
| `mp_match_can_clinch` | 1 | 1 | End early once a team can no longer be caught. |
| `mp_warmuptime` | 30 | 120 | Warmup length in seconds. |
| `mp_warmuptime_all_players_connected` | 0 | 15 | Warmup once everyone is connected; 0 disables. |
| **Gameplay** | | | |
| `mp_friendlyfire` | 0 | 1 | Teammates can damage each other. |
| `mp_solid_teammates` | 1 | 1 | 0 pass through, 1 solid, 2 can stand on heads. |
| `bot_quota` | 10 | 0 | Number of bots. |
| `bot_difficulty` | 1 | 2 | 0 easy, 1 normal, 2 hard, 3 expert. |

Sources:

- Engine defaults, ranges and descriptions:
  [SteamDatabase/GameTracking-CS2 `DumpSource2/convars.txt`](https://github.com/SteamDatabase/GameTracking-CS2/blob/master/DumpSource2/convars.txt),
  dumped from the CS2 binaries on every update. Search it for any other convar.
- Competitive defaults:
  [`game/csgo/cfg/gamemode_competitive.cfg`](https://github.com/SteamDatabase/GameTracking-CS2/blob/master/game/csgo/cfg/gamemode_competitive.cfg),
  also on the server at `/home/steam/cs2-dedicated/game/csgo/cfg/`.
- [Valve Developer Community: CS2 console commands](https://developer.valvesoftware.com/wiki/List_of_Counter-Strike_2_console_commands).

A value set live with `rcon <convar> <value>` lasts until the next map load,
when `cs2_config.yml` applies again.

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
