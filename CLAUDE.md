# pom

A terminal Pomodoro timer built in Go using the Bubble Tea TUI framework.

## Project Structure

```
cmd/pom/main.go              # Entry point
internal/
  config/config.go            # Config struct and defaults
  pomodoro/
    session.go                # SessionType enum, SessionResult type
    timer.go                  # Pure timer state-machine logic
  storage/
    store.go                  # Store interface
    sqlite.go                 # SQLite implementation
    queries.go                # QueryFilter, Statistics types
  tui/
    model.go                  # Bubble Tea model + Init
    update.go                 # Update() — input handling, state transitions
    view.go                   # View() — rendering
    commands.go               # tickCmd, notifyCmd
  commands/
    root.go                   # Cobra root command
    start.go                  # `pom start` — launches TUI timer
    summary.go                # `pom summary` — aggregated stats
    history.go                # `pom history` — list past sessions
```

## Key Dependencies

- `charmbracelet/bubbletea` — TUI framework (Elm architecture)
- `charmbracelet/bubbles` — Pre-built components (progress bar, text input)
- `charmbracelet/lipgloss` — Terminal styling and layout
- `charmbracelet/lipgloss/table` — Table rendering for summary/history
- `gen2brain/beeep` — Desktop notifications
- `spf13/cobra` — CLI subcommand framework
- `modernc.org/sqlite` — Pure-Go SQLite driver

## Build & Run

```bash
go build -o pom ./cmd/pom
```

### Commands

```bash
pom start [flags]       # Start a pomodoro session
pom summary [flags]     # Show aggregated session statistics
pom history [flags]     # List completed sessions
```

### Start Flags

- `--name, -n` — Optional session label
- `--session, -s` — Focus duration (default: `25m`)
- `--sbreak` — Short break duration (default: `5m`)
- `--lbreak` — Long break duration (default: `15m`)
- `--nbreak` — Sessions before a long break (default: `4`)

### Summary / History Flags

- `--name` — Filter by session name
- `--from` — Start date (`YYYY-MM-DD`)
- `--to` — End date (`YYYY-MM-DD`)
- `--type` — Filter by type: `focus`, `short-break`, `long-break`
- `--json` — Output as JSON
- `--limit` — Max sessions to show (history only)

## Architecture

- **`internal/config`** — `Config` struct with session/break durations
- **`internal/pomodoro`** — Domain types (`SessionType`, `SessionResult`) and pure timer logic (`NextSession`)
- **`internal/storage`** — `Store` interface with SQLite backend; persists sessions to `~/.local/share/pom/history.db`
- **`internal/tui`** — Bubble Tea model/update/view following Elm architecture
- **`internal/commands`** — Cobra command definitions wiring everything together

## Key Behaviors

- Timer counts down; advances automatically through focus → short break → long break cycle
- Each completed session/break is persisted to SQLite automatically
- Pause/resume freezes the target time; unpausing recalculates from remaining duration
- Session rename pauses the timer implicitly (target time recalculated on confirm/cancel)
- On quit (`q` / `ctrl+c`), outputs a JSON summary to stdout
- `pom summary` and `pom history` query the SQLite database with optional filters

## Keybindings (during timer)

| Key        | Action           |
|------------|------------------|
| `space`/`p` | Pause / Resume  |
| `r`        | Rename session   |
| `?`        | Toggle help      |
| `q`/`ctrl+c` | Quit           |
