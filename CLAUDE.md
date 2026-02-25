# pom

A terminal Pomodoro timer built in Go using the Bubble Tea TUI framework.

## Project Structure

Single-file Go application (`main.go`) following the Elm architecture (Model/Update/View) via Bubble Tea.

## Key Dependencies

- `charmbracelet/bubbletea` — TUI framework (Elm architecture)
- `charmbracelet/bubbles` — Pre-built components (progress bar, text input)
- `charmbracelet/lipgloss` — Terminal styling and layout
- `gen2brain/beeep` — Desktop notifications

## Build & Run

```bash
go build -o pom .
./pom [flags]
```

**Flags:**
- `-name` — Optional session label
- `-session` — Focus duration (default: `25m`)
- `-sbreak` — Short break duration (default: `5m`)
- `-lbreak` — Long break duration (default: `15m`)
- `-nbreak` — Sessions before a long break (default: `4`)

## Architecture

- `Config` — Parsed CLI flags
- `model` — All TUI state (timer, pause, rename, progress)
- `SessionResult` — JSON output on exit (name, session count, start/end times)
- `SessionType` — Enum: `Focus Session`, `Short Break`, `Long Break`

## Key Behaviors

- Timer counts down; advances automatically through focus → short break → long break cycle
- Pause/resume freezes the target time; unpausing recalculates from remaining duration
- Session rename pauses the timer implicitly (target time recalculated on confirm/cancel)
- On quit (`q` / `ctrl+c`), outputs a JSON summary to stdout
- Progress bar fills left-to-right as each session/break elapses

## Keybindings

| Key        | Action           |
|------------|------------------|
| `space`/`p` | Pause / Resume  |
| `r`        | Rename session   |
| `?`        | Toggle help      |
| `q`/`ctrl+c` | Quit           |
