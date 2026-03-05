# Python: Claude Code Stats

Python implementation of the Claude Code usage statistics extractor and interactive CLI dashboard.

- **`extract_stats.py`** -- Parses `~/.claude/` session data, generates `public/dashboard_data.json` and `public/dashboard.html`
- **`cli.py`** -- Interactive TUI dashboard using [Textual](https://textual.textualize.io/)

## Development Setup

Requires Python 3.10+ and [uv](https://docs.astral.sh/uv/).

```bash
# From the repo root
cd python

# Create venv and install dependencies
uv sync

# Create your config (if not done yet)
cp ../config.example.json ../config.json
# Edit ../config.json with your plan details
```

## Usage

### HTML Dashboard

```bash
# Extract stats and generate HTML dashboard
uv run python extract_stats.py

# Open the dashboard
open ../public/dashboard.html        # macOS
xdg-open ../public/dashboard.html    # Linux
start ..\public\dashboard.html       # Windows
```

### Interactive CLI Dashboard

```bash
uv run python cli.py                           # TUI with auto-refresh
uv run python cli.py --all                     # Print all sections
uv run python cli.py --section tokens,plan     # Specific sections
uv run python cli.py --limit 10                # Limit table rows
uv run python cli.py --json                    # Raw JSON to stdout
uv run python cli.py --no-refresh              # Skip re-extraction
```

### Installed Commands

After `uv sync`, these commands are available in the venv:

```bash
uv run claude-extract              # Same as python extract_stats.py
uv run claude-dashboard            # Same as python cli.py
```

## Dependencies

- **extract_stats.py** -- Python stdlib only (no external deps)
- **cli.py** -- [Textual](https://textual.textualize.io/) for the TUI

## Automation

To auto-refresh the HTML dashboard:

```bash
*/10 * * * * cd /path/to/claude-code-stats/python && uv run python extract_stats.py 2>&1 >> ../update.log
```
