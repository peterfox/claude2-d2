# claude2-d2

Go CLI that integrates a Sphero R2-D2 BLE toy with Claude Code. A persistent daemon maintains the BLE connection and reacts to Claude Code lifecycle events by playing animations on the droid.

## Quick start

```bash
go build -o r2 .
./r2 setup        # scan and save the droid's address to ~/.r2d2
./r2 daemon       # connect and start listening (or use install for auto-start)
```

## Commands

| Command | Purpose |
|---|---|
| `r2 setup` | Scan for R2-D2, save address to `~/.r2d2` |
| `r2 daemon [--debug]` | Connect by stored address, listen for HTTP events on `:2187` |
| `r2 signal <event>` | POST an event to the daemon HTTP server (for testing) |
| `r2 install` | Register daemon as a launchd user agent (auto-starts on login) |
| `r2 uninstall` | Stop and remove the launchd agent |

## Config file — `~/.r2d2`

Written by `r2 setup`. Contains the BLE address so the daemon connects directly without scanning.

```json
{
  "device_address": "6560a97d-7978-69d2-8e16-478fe653f5b8",
  "device_name": "D2-22BF"
}
```

## HTTP server — `:2187`

The daemon listens on `http://localhost:2187`. Claude Code hooks POST event names to `/event`.

Valid events: `prompt`, `thinking`, `stop`, `session_start`, `stop_failure`, `permission_request`

```bash
curl -s -X POST http://localhost:2187/event -d session_start
curl -s -X POST http://localhost:2187/event -d thinking
curl -s -X POST http://localhost:2187/event -d prompt
curl -s -X POST http://localhost:2187/event -d stop
curl -s -X POST http://localhost:2187/event -d stop_failure
curl -s -X POST http://localhost:2187/event -d permission_request
```

## Event → animation mapping

| Event | Trigger | R2-D2 behaviour |
|---|---|---|
| `session_start` | Claude Code session opens | Play `wwm_yoohoo` for 10s |
| `prompt` | User submits a message | Start 60s timer; if no `thinking` arrives → random impatient animation |
| `thinking` | Claude uses a tool | Cancel timer; cycle `idle_1/2/3` every 30s |
| `stop` | Claude finishes responding | Stop idle cycle; play random celebratory animation for 10s |
| `stop_failure` | API error ends the turn | Play random failure animation for 10s |
| `permission_request` | Claude requests tool permission | After 3s debounce, play `wwm_anxious` for 10s |

## Animation pools

| Pool | Used by | Animations |
|---|---|---|
| `stopAnimations` | `stop` | `emote_excited`, `emote_laugh`, `emote_yes`, `emote_spin`, `wwm_bow`, `wwm_happy`, `wwm_excited`, `wwm_relieved` |
| `impatientAnimations` | angry timeout | `emote_angry`, `emote_annoyed`, `wwm_angry`, `wwm_frustrated`, `wwm_fiery`, `wwm_jittery` |
| `failureAnimations` | `stop_failure` | `emote_angry`, `emote_alarm`, `emote_annoyed`, `emote_no`, `emote_sad`, `emote_ion_blast`, `wwm_angry`, `wwm_frustrated`, `wwm_fiery`, `wwm_ominous`, `wwm_no`, `wwm_yelling` |

## Claude Code plugin

Hooks are packaged as a Claude Code plugin in `hooks/hooks.json` and `.claude-plugin/plugin.json`. Install with:

```
/plugin install peterfox/claude2-d2 https://github.com/peterfox/claude2-d2
```

Six hooks are registered: `SessionStart`, `UserPromptSubmit`, `PreToolUse`, `Stop`, `StopFailure`, `PermissionRequest`. All use `|| true` so a stopped daemon is silently ignored.

## Architecture

```
internal/r2/         BLE client, packet framing, animation commands, config file
internal/daemon/     HTTP event server, animation state machine
internal/launchd/    macOS launchd plist management
cmd/                 Cobra CLI commands
hooks/               Claude Code plugin hooks
.claude-plugin/      Claude Code plugin manifest
Formula/             Homebrew formula (updated automatically by GoReleaser)
.github/workflows/   Release workflow
```

## Daemon behaviour

- **Inactivity reconnect** — checks every 1 minute; if no event for 5 minutes, sends a `StopAnimation` ping. On failure, reconnects via stored address.
- **Permission debounce** — `permission_request` starts a 3-second timer; any subsequent event cancels it.
- **Animation cancellation** — `playTimedAnimLocked` closes the previous `animCancel` channel before starting a new animation, preventing stale `StopAnimation` calls from interrupting the next state.

## BLE protocol notes

- Device advertises as `D2-XXXX` (name prefix match, not UUID match)
- Connection: handshake bytes to connect characteristic → 500ms wait → init packet → 5s wait
- Checksum: `0xFF - (sum(body bytes) & 0xFF)` — NOT XOR
- Packet frame: `0x8D | body... | checksum | 0xD8` with byte escaping via `0xAB`

## Animations

56 animations available in `r2.Animations` map. Categories:
- `charger_1` – `charger_7`
- `emote_alarm`, `emote_angry`, `emote_annoyed`, `emote_chatty`, `emote_drive`, `emote_excited`, `emote_happy`, `emote_ion_blast`, `emote_laugh`, `emote_no`, `emote_sad`, `emote_sassy`, `emote_scared`, `emote_scan`, `emote_sleep`, `emote_spin`, `emote_surprised`, `emote_yes`
- `idle_1`, `idle_2`, `idle_3`
- `patrol_alarm`, `patrol_hit`, `patrol_patrolling`
- `wwm_angry`, `wwm_anxious`, `wwm_bow`, `wwm_concern`, `wwm_curious`, `wwm_double_take`, `wwm_excited`, `wwm_fiery`, `wwm_frustrated`, `wwm_happy`, `wwm_jittery`, `wwm_laugh`, `wwm_long_shake`, `wwm_no`, `wwm_ominous`, `wwm_relieved`, `wwm_sad`, `wwm_scared`, `wwm_shake`, `wwm_surprised`, `wwm_taunting`, `wwm_whisper`, `wwm_yelling`, `wwm_yoohoo`
- `motor`

## Build

```bash
go build -o r2 .
```

Requires macOS with Bluetooth. The `tinygo.org/x/bluetooth` package uses CoreBluetooth via CGo — no external BLE library needed.

## Releasing

```bash
git tag v0.1.0
git push origin v0.1.0
```

GoReleaser builds a universal binary, creates the GitHub release, and updates `Formula/claude2-d2.rb` automatically.
