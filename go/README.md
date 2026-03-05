# Go: Claude Code Dashboard

Single-binary TUI dashboard for Claude Code usage statistics, built with [Bubble Tea v2](https://github.com/charmbracelet/bubbletea).

## Features

- Interactive tabbed TUI with 7 views (Overview, Tokens, Activity, Projects, Sessions, Plan & Billing, Insights)
- Async data extraction with loading spinner
- Auto-refresh when data is older than 10 minutes
- Non-interactive modes: `--json`, `--all`, `--section`
- Cross-platform (Windows/macOS/Linux), zero runtime dependencies

## Install

### Pre-built Binary

Download from [Releases](https://github.com/AeternaLabsHQ/claude-code-stats/releases) for your platform.

### From Source

```bash
go install github.com/AeternaLabsHQ/claude-code-stats/cmd/claude-dashboard@latest
```

## Development Setup

Requires Go 1.25+.

```bash
# From the repo root
cd go

# Build
go build -o claude-dashboard ./cmd/claude-dashboard

# Run
./claude-dashboard

# Create your config (if not done yet)
cp ../config.example.json ../config.json
```

## Usage

```bash
claude-dashboard                           # Extract + launch TUI
claude-dashboard --no-refresh              # TUI only, skip extraction
claude-dashboard --json                    # Extract + dump JSON to stdout
claude-dashboard --all                     # Extract + print all sections
claude-dashboard --section tokens,plan     # Specific sections
claude-dashboard --limit 10               # Limit table rows (default: 20)
claude-dashboard --quiet                   # Suppress progress output
claude-dashboard --config ./config.json    # Explicit config path
claude-dashboard --output ./out            # Output directory (default: ./public)
claude-dashboard extract                   # Extract only (no TUI), write JSON
claude-dashboard version                   # Print version
```

### TUI Keybindings

| Key | Action |
|-----|--------|
| `1`-`7` | Switch to tab |
| `Tab` / `Shift+Tab` | Cycle tabs |
| `Up` / `Down` / `PgUp` / `PgDn` | Scroll |
| `q` / `Ctrl+C` | Quit |

## Building Releases

Uses [GoReleaser](https://goreleaser.com/):

```bash
# Snapshot build (no publish)
goreleaser release --snapshot --clean

# Targets: linux/darwin/windows x amd64/arm64
```

## Project Structure

```
go/
  cmd/claude-dashboard/main.go     # CLI entry point (Cobra)
  internal/
    extract/                       # Data extraction from ~/.claude/
      types.go                     # Data structures
      sessions.go                  # JSONL transcript parser
      pricing.go                   # Model pricing + cost calculation
      billing.go                   # Plan/billing period analysis
      config.go                    # Config + locale loading
      extract.go                   # Main orchestrator
      history.go, plugins.go,      # Additional data sources
      plans.go, storage.go
      locales/                     # Embedded locale files (go:embed)
    tui/                           # Bubble Tea v2 TUI
      app.go                       # Root model + tab navigation
      styles.go                    # Lipgloss v2 styles
      helpers.go                   # Formatting utilities
      print.go                     # Non-interactive renderers
      windows.go                   # Windows UTF-8 console setup
      views/                       # Tab view renderers
    output/
      json.go                      # JSON output writer
  .goreleaser.yml
```
