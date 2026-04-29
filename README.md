# claude2-d2

A Go CLI that brings your Sphero R2-D2 to life alongside Claude Code. A persistent daemon holds the BLE connection and reacts to Claude's lifecycle — R2-D2 idles while Claude thinks, celebrates when it finishes, and throws a fit if it's left waiting.

## How it works

Claude Code hooks POST lifecycle events to a local HTTP server the daemon runs on `:2187`. The daemon drives an animation state machine over a persistent BLE connection.

```
Claude Code (any session)
    │  SessionStart / UserPromptSubmit / PreToolUse / Stop / StopFailure / PermissionRequest hooks
    ▼
curl -s -X POST http://localhost:2187/event -d "<event>"
    ▼
claude2-d2 daemon — animation state machine
    ▼
Sphero R2-D2 over BLE
```

| Event | What R2-D2 does |
|---|---|
| Session starts | Plays a random greeting animation |
| User sends a prompt | Starts a 60-second timer — if Claude hasn't used a tool by then, plays a random impatient animation |
| Claude uses a tool | Cancels the timer; cycles through idle animations every 30 seconds |
| Claude finishes responding | Plays a random celebratory animation for 10 seconds then goes quiet |
| Claude fails with an API error | Plays a random failure animation |
| Claude requests permission | After a 3-second debounce, plays an anxious animation (cancelled if Claude resumes quickly) |

## Requirements

- macOS (uses CoreBluetooth via `tinygo.org/x/bluetooth`)
- Sphero R2-D2 toy
- [Claude Code](https://claude.ai/code) CLI

## Installation

### Homebrew (recommended)

```bash
brew tap peterfox/claude2-d2 https://github.com/peterfox/claude2-d2
brew install claude2-d2
```

### Build from source

```bash
git clone https://github.com/peterfox/claude2-d2
cd claude2-d2
go build -o claude2-d2 .
mv claude2-d2 /usr/local/bin/claude2-d2
```

## Setup

**1. Power on your R2-D2 and scan for it:**

```bash
claude2-d2 setup
```

Scans over BLE and saves the droid's address to `~/.claude2-d2`. Only needed once.

**2. Start the daemon:**

Via Homebrew (auto-starts on login):
```bash
brew services start claude2-d2
```

Via source install (auto-starts on login):
```bash
claude2-d2 install
```

Manually (foreground, useful for development):
```bash
claude2-d2 daemon --debug
```

**3. Install the Claude Code plugin:**

Open Claude Code and run:
```
/plugin install peterfox/claude2-d2 https://github.com/peterfox/claude2-d2
```

This wires up all lifecycle hooks automatically. Toggle the plugin on/off anytime via `/plugin`.

That's it — open any Claude Code session and R2-D2 will start reacting.

## Claude Code hooks

The plugin in `hooks/hooks.json` registers six lifecycle hooks:

| Hook | Event sent | Trigger |
|---|---|---|
| `SessionStart` | `session_start` | A Claude Code session opens or resumes |
| `UserPromptSubmit` | `prompt` | User submits a message |
| `PreToolUse` | `thinking` | Claude is about to use a tool |
| `Stop` | `stop` | Claude finishes responding |
| `StopFailure` | `stop_failure` | Turn ends due to an API error |
| `PermissionRequest` | `permission_request` | Claude requests tool permission |

The `|| true` guard in every command means a stopped daemon is silently ignored.

## Testing

Use `claude2-d2 signal` to fire events without needing a Claude session:

```bash
claude2-d2 signal session_start      # greeting animation
claude2-d2 signal thinking           # starts idle loop
claude2-d2 signal prompt             # starts the 60s impatient timer
claude2-d2 signal stop               # celebratory animation
claude2-d2 signal stop_failure       # failure animation
claude2-d2 signal permission_request # anxious animation (after 3s debounce)
```

## Commands

| Command | Description |
|---|---|
| `claude2-d2 setup` | Scan for R2-D2 and save its BLE address to `~/.claude2-d2` |
| `claude2-d2 daemon [--debug]` | Connect and listen for HTTP events on `:2187` |
| `claude2-d2 signal <event>` | Send an event directly to the daemon |
| `claude2-d2 install` | Register as a launchd user agent (auto-starts on login) |
| `claude2-d2 uninstall` | Stop and remove the launchd agent |

## Daemon behaviour

- **Inactivity reconnect** — if no event is received for 5 minutes, the daemon sends a keepalive ping. On failure it reconnects automatically using the saved address.
- **Permission debounce** — `permission_request` waits 3 seconds before playing; if any other event arrives first, it is cancelled.
- **Animation safety** — starting a new animation always cancels any in-flight timed animation to prevent the droid wiggling after it should have stopped.

## Project structure

```
cmd/                 CLI commands (setup, daemon, signal, install, uninstall)
internal/r2/         BLE client, packet protocol, animation definitions, config
internal/daemon/     HTTP event server, animation state machine
internal/launchd/    macOS launchd plist management
hooks/               Claude Code plugin hooks
.claude-plugin/      Claude Code plugin manifest
Formula/             Homebrew formula
.github/workflows/   Release automation via GoReleaser
```

## Available animations

The droid has 56 built-in animations:

| Category | Animations |
|---|---|
| Charger | `charger_1` – `charger_7` |
| Emote | `emote_alarm`, `emote_angry`, `emote_annoyed`, `emote_chatty`, `emote_drive`, `emote_excited`, `emote_happy`, `emote_ion_blast`, `emote_laugh`, `emote_no`, `emote_sad`, `emote_sassy`, `emote_scared`, `emote_scan`, `emote_sleep`, `emote_spin`, `emote_surprised`, `emote_yes` |
| Idle | `idle_1`, `idle_2`, `idle_3` |
| Patrol | `patrol_alarm`, `patrol_hit`, `patrol_patrolling` |
| WWM | `wwm_angry`, `wwm_anxious`, `wwm_bow`, `wwm_concern`, `wwm_curious`, `wwm_double_take`, `wwm_excited`, `wwm_fiery`, `wwm_frustrated`, `wwm_happy`, `wwm_jittery`, `wwm_laugh`, `wwm_long_shake`, `wwm_no`, `wwm_ominous`, `wwm_relieved`, `wwm_sad`, `wwm_scared`, `wwm_shake`, `wwm_surprised`, `wwm_taunting`, `wwm_whisper`, `wwm_yelling`, `wwm_yoohoo` |
| Other | `motor` |

## Releasing

Tag a commit to trigger a release:

```bash
git tag v0.1.0
git push origin v0.1.0
```

GitHub Actions runs GoReleaser on a macOS runner, builds a universal binary (arm64 + amd64), creates a GitHub release, and updates `Formula/claude2-d2.rb` with the correct version and sha256 automatically.

## License

MIT
