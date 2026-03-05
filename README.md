# Claude Code Usage Statistics

A comprehensive analytics dashboard for [Claude Code](https://docs.anthropic.com/en/docs/claude-code) usage data. Parses your local Claude Code session transcripts, calculates hypothetical API costs, and displays an interactive dashboard.

***Disclaimer:*** *This is an unofficial, community-built tool. Not affiliated with or endorsed by Anthropic.*

## Features

- **KPI Dashboard** -- Total API-equivalent cost, messages, sessions, output tokens
- **Token & API Value** -- Daily costs, cumulative costs, model distribution
- **Activity** -- Message patterns, hourly distribution, weekday distribution
- **Projects** -- Top projects by cost, detailed project metrics
- **Sessions** -- Filterable/searchable session details with expandable metadata
- **Plan & Billing** -- Cost savings analysis vs. your subscription plan
- **Insights** -- Tool usage, storage breakdown, plugins, todos, file snapshots

![Dashboard Screenshot](docs/images/claude-code-stats-01.png)

## Implementations

This repo contains two implementations sharing the same config and data format:

| | [**Go (TUI)**](go/) | [**Python**](python/) |
|---|---|---|
| Interface | Interactive terminal TUI (Bubble Tea v2) | HTML dashboard + Textual TUI |
| Install | Single binary, zero deps | Python 3.10+ with uv |
| Best for | Daily use, quick checks | HTML generation, automation |

Both read from `~/.claude/` and use the same `config.json`.

## Quick Start

### Option A: Go Binary (Recommended)

```bash
# Download from releases, or build from source:
cd go
go build -o claude-dashboard ./cmd/claude-dashboard
./claude-dashboard
```

See [go/README.md](go/README.md) for full details.

### Option B: Python

```bash
cd python
uv sync
uv run python extract_stats.py      # Generate HTML dashboard
uv run python cli.py                # Interactive TUI
```

See [python/README.md](python/README.md) for full details.

## Configuration

Create your config file in the repo root:

```bash
cp config.example.json config.json
```

See [`config.example.json`](config.example.json) for all options:

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `language` | `string` | `"en"` | UI language (`"en"` or `"de"`) |
| `plan_history` | `array` | `[]` | Your subscription plan history |
| `migration.enabled` | `bool` | `false` | Enable data from a migration backup |
| `migration.dir` | `string` | `null` | Path to migration backup directory |

### Plan History

Each entry in `plan_history` represents a subscription period:

```json
{
  "plan": "Max",
  "start": "2026-01-23",
  "end": null,
  "cost_eur": 87.61,
  "cost_usd": 93.00,
  "billing_day": 23
}
```

- `end: null` means the plan is currently active
- `billing_day` determines billing cycle boundaries for cost analysis

### Migration Support

If you migrated Claude Code data from another machine, you can include that historical data:

```json
{
  "migration": {
    "enabled": true,
    "dir": "~/backups/old-machine",
    "claude_dir_name": ".claude-windows",
    "dot_claude_json_name": ".claude-windows.json"
  }
}
```

Sessions are deduplicated across both sources automatically.

## Localization

Supports English and German. Set `"language": "en"` or `"language": "de"` in `config.json`.

To add a new language, create a file in `locales/` following the structure of [`locales/en.json`](locales/en.json).

## Repository Structure

```
claude-code-stats/
  config.example.json          # Shared configuration template
  locales/                     # Shared locale files (en, de)
  docs/                        # Documentation, screenshots, solutions
  python/                      # Python implementation
    extract_stats.py           #   HTML dashboard generator (stdlib only)
    cli.py                     #   Textual TUI dashboard
    pyproject.toml             #   uv/pip project file
  go/                          # Go implementation
    cmd/claude-dashboard/      #   CLI entry point
    internal/                  #   Extraction + TUI code
    .goreleaser.yml            #   Cross-platform build config
```

## License

MIT
