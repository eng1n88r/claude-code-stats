#!/usr/bin/env python3
"""Interactive CLI dashboard for Claude Code usage statistics.

Uses textual for a tabbed TUI interface, similar to Claude Code's plan mode.
"""

import argparse
import io
import json
import os
import subprocess
import sys
import time
from pathlib import Path

# Force UTF-8 output on Windows
if sys.platform == "win32":
    sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding="utf-8", errors="replace")
    sys.stderr = io.TextIOWrapper(sys.stderr.buffer, encoding="utf-8", errors="replace")
    os.environ.setdefault("PYTHONIOENCODING", "utf-8")

SCRIPT_DIR = Path(__file__).parent
DATA_FILE = SCRIPT_DIR / "public" / "dashboard_data.json"
EXTRACT_SCRIPT = SCRIPT_DIR / "extract_stats.py"


# ── Helpers ───────────────────────────────────────────────────────────────

def fmt_cost(val):
    if val is None:
        return "N/A"
    return f"${val:,.2f}"


def fmt_tokens(val):
    if val is None:
        return "0"
    if val >= 1_000_000:
        return f"{val / 1_000_000:.1f}M"
    if val >= 1_000:
        return f"{val / 1_000:.1f}K"
    return str(int(val))


def make_bar(val, max_val, width=30):
    if max_val <= 0:
        return ""
    filled = int(val / max_val * width)
    return "\u2588" * filled + "\u2591" * (width - filled)


def maybe_refresh(no_refresh):
    if no_refresh:
        return
    if not DATA_FILE.exists():
        print("Data file missing, running extract_stats.py...")
        subprocess.run([sys.executable, str(EXTRACT_SCRIPT)], check=True)
        return
    age = time.time() - os.path.getmtime(str(DATA_FILE))
    if age > 600:
        print("Data older than 10 min, refreshing...")
        subprocess.run([sys.executable, str(EXTRACT_SCRIPT)], check=True)


def load_data():
    with open(DATA_FILE, "r", encoding="utf-8") as f:
        return json.load(f)


# ── Textual TUI App ──────────────────────────────────────────────────────

from textual.app import App, ComposeResult
from textual.binding import Binding
from textual.containers import Horizontal, Vertical, VerticalScroll
from textual.widgets import DataTable, Footer, Header, Static, TabbedContent, TabPane

# Color mapping for rich markup inside Static widgets
C_MUTED = "[dim]"
C_GREEN = "[green]"
C_RED = "[red]"
C_CYAN = "[cyan]"
C_YELLOW = "[yellow]"
C_MAGENTA = "[magenta]"
C_BOLD = "[bold]"
C_ORANGE = "[dark_orange]"


def build_kpi_text(data):
    kpi = data["kpi"]
    lines = []
    lines.append("")
    lines.append(
        f"  {C_ORANGE}{C_BOLD}API Equivalent[/][/]  {C_BOLD}{fmt_cost(kpi['total_cost'])}[/]"
        f"  {C_MUTED}(Actually paid: {fmt_cost(kpi.get('actual_plan_cost'))})[/]"
    )
    lines.append(
        f"  {C_GREEN}{C_BOLD}Messages[/][/]        {C_BOLD}{kpi['total_messages']}[/]"
        f"  {C_MUTED}in {kpi['total_sessions']} sessions[/]"
    )
    lines.append(
        f"  {C_CYAN}{C_BOLD}Sessions[/][/]        {C_BOLD}{kpi['total_sessions']}[/]"
        f"  {C_MUTED}{kpi['total_projects']} projects[/]"
    )
    lines.append(
        f"  {C_MAGENTA}{C_BOLD}Output Tokens[/][/]   {C_BOLD}{fmt_tokens(kpi['total_output_tokens'])}[/]"
        f"  {C_MUTED}Input: {fmt_tokens(kpi['total_input_tokens'])}[/]"
    )
    lines.append(
        f"\n  {C_MUTED}Date range: {kpi.get('first_session', '?')} to {kpi.get('last_session', '?')}[/]"
    )
    return "\n".join(lines)


def build_overview_text(data):
    kpi = data["kpi"]
    models = data.get("model_summary", [])
    projects = data.get("projects", [])
    hourly = data.get("hourly_distribution", [])
    daily = data.get("daily_messages", [])
    plan_data = data.get("plan", {})
    cb = plan_data.get("current_billing", {})
    insights = data.get("insights", {})
    tools = data.get("tool_summary", [])
    plugins = insights.get("plugins", {})
    storage = insights.get("storage", {})

    top_model = max(models, key=lambda m: m["cost"])["model"] if models else "?"
    top_model_cost = max(models, key=lambda m: m["cost"])["cost"] if models else 0
    peak_hour = max(hourly, key=lambda h: h["messages"]) if hourly else {"hour": 0}
    active_days = len([d for d in daily if d["messages"] > 0])
    top_project = max(projects, key=lambda p: p["cost"])["name"] if projects else "?"
    top_project_cost = max(projects, key=lambda p: p["cost"])["cost"] if projects else 0
    avg_msgs = kpi["total_messages"] / kpi["total_sessions"] if kpi["total_sessions"] else 0
    plan_name = cb.get("plan", "?")
    plan_cost = cb.get("plan_cost_usd")
    roi = cb.get("roi_factor", 0)
    days_left = cb.get("days_remaining", 0)

    lines = [
        "",
        f"  {C_BOLD}Tokens:[/]   {len(models)} models, top: {top_model} ({fmt_cost(top_model_cost)})",
        f"  {C_BOLD}Activity:[/] {kpi['total_messages']} messages, {active_days} active days, peak: {peak_hour['hour']:02d}:00",
        f"  {C_BOLD}Projects:[/] {kpi['total_projects']} projects, top: {top_project} ({fmt_cost(top_project_cost)})",
        f"  {C_BOLD}Sessions:[/] {kpi['total_sessions']} total, avg {avg_msgs:.1f} msgs/session",
        f"  {C_BOLD}Plan:[/]     {plan_name} @ {fmt_cost(plan_cost)}/mo, ROI {roi}x, {days_left} days left",
        f"  {C_BOLD}Insights:[/] {len(tools)} tools, {len(plugins.get('installed', []))} plugins, {storage.get('total_mb', 0):.1f} MB storage",
    ]
    return "\n".join(lines)


def build_tokens_text(data):
    models = data.get("model_summary", [])
    lines = [
        "",
        f"  {C_BOLD}Model Detail[/]",
        "",
        f"  {'Model':<12} {'API Value':>10} {'Output':>8} {'Input':>8} {'Cache Read':>11} {'Calls':>6}",
        f"  {'─' * 12} {'─' * 10} {'─' * 8} {'─' * 8} {'─' * 11} {'─' * 6}",
    ]
    for m in models:
        lines.append(
            f"  {m['model']:<12} {C_GREEN}{fmt_cost(m['cost']):>10}[/] "
            f"{fmt_tokens(m['output_tokens']):>8} {fmt_tokens(m['input_tokens']):>8} "
            f"{fmt_tokens(m['cache_read_tokens']):>11} {m['calls']:>6}"
        )

    # Cost by token type
    cbt = data.get("cost_by_token_type", {})
    if cbt:
        total = sum(cbt.values())
        max_val = max(cbt.values()) if cbt else 1
        colors = {"input": C_CYAN, "output": C_GREEN, "cache_read": C_YELLOW, "cache_write": C_MAGENTA}
        lines.append(f"\n  {C_BOLD}Cost by Token Type[/]\n")
        for typ, val in cbt.items():
            pct = val / total * 100 if total else 0
            c = colors.get(typ, "")
            bar = make_bar(val, max_val, 30)
            lines.append(f"  {c}{typ:<14}[/] {c}{bar}[/] {fmt_cost(val)} ({pct:.0f}%)")

    # Daily costs
    daily = data.get("daily_costs", [])[-30:]
    if daily:
        max_total = max(d["total"] for d in daily) or 1
        lines.append(f"\n  {C_BOLD}Daily API Value (last {len(daily)} days)[/]\n")
        for d in daily:
            bar = make_bar(d["total"], max_total, 30)
            lines.append(f"  {d['date']}  {bar}  {fmt_cost(d['total'])}")

    return "\n".join(lines)


def build_activity_text(data):
    lines = [""]

    # Hourly
    hourly = data.get("hourly_distribution", [])
    if hourly:
        max_msgs = max(h["messages"] for h in hourly) or 1
        peak = max(hourly, key=lambda h: h["messages"])
        lines.append(f"  {C_BOLD}Hourly Distribution[/]\n")
        for h in hourly:
            style = C_YELLOW + C_BOLD if (h["hour"] == peak["hour"] and h["messages"] > 0) else C_CYAN
            end = "[/][/]" if C_BOLD in style else "[/]"
            bar = make_bar(h["messages"], max_msgs, 30)
            lines.append(f"  {style}{h['hour']:02d}:00{end}  {bar}  {h['messages']}")

    # Weekday
    weekday = data.get("weekday_distribution", [])
    if weekday:
        max_msgs = max(w["messages"] for w in weekday) or 1
        weekend = {"Sat", "Sun"}
        lines.append(f"\n  {C_BOLD}Weekday Distribution[/]\n")
        for w in weekday:
            style = C_MAGENTA if w["day"] in weekend else C_CYAN
            bar = make_bar(w["messages"], max_msgs, 30)
            lines.append(f"  {style}{w['day']:<3}[/]  {bar}  {w['messages']}")

    # Daily messages table
    daily = data.get("daily_messages", [])[-30:]
    if daily:
        lines.append(f"\n  {C_BOLD}Daily Messages (last {len(daily)} days)[/]\n")
        lines.append(f"  {'Date':<12} {'Messages':>8} {'Sessions':>8}")
        lines.append(f"  {'─' * 12} {'─' * 8} {'─' * 8}")
        for d in daily:
            lines.append(f"  {d['date']:<12} {d['messages']:>8} {d['sessions']:>8}")

    return "\n".join(lines)


def build_projects_text(data, limit=20):
    projects = sorted(data.get("projects", []), key=lambda p: p["cost"], reverse=True)
    lines = [
        "",
        f"  {C_BOLD}Projects by API Value[/]\n",
        f"  {'Project':<25} {'Sessions':>8} {'Messages':>8} {'API Value':>10} {'Output':>10} {'Size':>8}",
        f"  {'─' * 25} {'─' * 8} {'─' * 8} {'─' * 10} {'─' * 10} {'─' * 8}",
    ]
    for p in projects[:limit]:
        lines.append(
            f"  {p['name']:<25} {p['sessions']:>8} {p['messages']:>8} "
            f"{C_GREEN}{fmt_cost(p['cost']):>10}[/] "
            f"{fmt_tokens(p['output_tokens']):>10} {p.get('file_size_mb', 0):>7.1f}M"
        )
    # Totals
    all_p = data.get("projects", [])
    lines.append(f"  {'─' * 25} {'─' * 8} {'─' * 8} {'─' * 10} {'─' * 10} {'─' * 8}")
    lines.append(
        f"  {C_BOLD}{'Total':<25}[/] {sum(p['sessions'] for p in all_p):>8} "
        f"{sum(p['messages'] for p in all_p):>8} "
        f"{C_GREEN}{C_BOLD}{fmt_cost(sum(p['cost'] for p in all_p)):>10}[/][/] "
        f"{fmt_tokens(sum(p['output_tokens'] for p in all_p)):>10} "
        f"{sum(p.get('file_size_mb', 0) for p in all_p):>7.1f}M"
    )
    return "\n".join(lines)


def build_sessions_text(data, limit=20):
    sessions = sorted(data.get("sessions", []), key=lambda s: s["date"], reverse=True)
    lines = [
        "",
        f"  {C_BOLD}Sessions[/]\n",
        f"  {'Date':<12} {'Project':<22} {'Dur':>5} {'Cost':>8} {'Msgs':>5} {'Model':<10} First Prompt",
        f"  {'─' * 12} {'─' * 22} {'─' * 5} {'─' * 8} {'─' * 5} {'─' * 10} {'─' * 40}",
    ]
    for s in sessions[:limit]:
        dur = s.get("duration_min", 0)
        prompt = (s.get("first_prompt") or "")[:50]
        lines.append(
            f"  {s['date']:<12} {s.get('project', ''):<22} {dur:>4.0f}m "
            f"{C_GREEN}{fmt_cost(s['cost']):>8}[/] {s['messages']:>5} "
            f"{s.get('primary_model', ''):<10} {C_MUTED}{prompt}[/]"
        )
    remaining = len(sessions) - limit
    if remaining > 0:
        lines.append(f"\n  {C_MUTED}... and {remaining} more sessions[/]")
    return "\n".join(lines)


def build_plan_text(data):
    plan = data.get("plan", {})
    cb = plan.get("current_billing", {})
    lines = [""]

    if cb:
        days_elapsed = cb.get("days_elapsed", 0)
        days_total = cb.get("days_total", 1)
        pct = days_elapsed / days_total if days_total else 0
        bar_width = 40
        filled = int(pct * bar_width)
        bar = C_GREEN + "\u2588" * filled + "[/]" + C_MUTED + "\u2591" * (bar_width - filled) + "[/]"

        lines.append(f"  {C_BOLD}Current Billing Period[/]  ({cb.get('period_start')} to {cb.get('period_end')})")
        lines.append(f"  Day {days_elapsed} of {days_total}  [{pct * 100:.0f}%]")
        lines.append(f"  {bar}  {cb.get('days_remaining', 0)} days remaining")
        lines.append("")

        stats = [
            ("Plan", f"{cb.get('plan', '?')} @ {fmt_cost(cb.get('plan_cost_usd'))}/mo"),
            ("API Cost", fmt_cost(cb.get("api_cost"))),
            ("Projected", fmt_cost(cb.get("projected_cost"))),
            ("Savings", None),
            ("ROI", f"{cb.get('roi_factor', 0)}x"),
            ("Sessions", str(cb.get("sessions", 0))),
            ("Messages", str(cb.get("messages", 0))),
            ("Avg/Day", fmt_cost(cb.get("cost_per_day"))),
        ]
        for label, val in stats:
            if label == "Savings":
                s = cb.get("savings", 0)
                color = C_GREEN if s and s >= 0 else C_RED
                lines.append(f"  {label:<12} {color}{fmt_cost(s)}[/]")
            else:
                lines.append(f"  {label:<12} {val}")

    # Period detail
    periods = plan.get("periods", [])
    if periods:
        lines.append(f"\n  {C_BOLD}Period Detail[/]\n")
        lines.append(
            f"  {'Period':<25} {'Plan':<5} {'Days':>15} {'API Cost':>10} "
            f"{'Plan Cost':>10} {'Savings':>10} {'ROI':>5} {'Sess':>5} {'Msgs':>5}"
        )
        lines.append(f"  {'─' * 25} {'─' * 5} {'─' * 15} {'─' * 10} {'─' * 10} {'─' * 10} {'─' * 5} {'─' * 5} {'─' * 5}")
        for p in periods:
            plan_cost_str = fmt_cost(p.get("plan_cost_usd")) if p.get("plan_cost_usd") is not None else (
                fmt_cost(p.get("plan_cost_eur")) if p.get("plan_cost_eur") is not None else "N/A"
            )
            savings = p.get("savings", 0)
            sc = C_GREEN if savings and savings >= 0 else C_RED
            days_str = f"{p['total_days']} ({p['days_active']} active)"
            lines.append(
                f"  {p.get('start', '?')+' - '+p.get('end', '?'):<25} {p.get('plan', ''):<5} "
                f"{days_str:>15} {fmt_cost(p.get('api_cost')):>10} "
                f"{plan_cost_str:>10} {sc}{fmt_cost(savings):>10}[/] "
                f"{p.get('roi_factor', 0):>4.1f}x {p.get('sessions', 0):>5} {p.get('messages', 0):>5}"
            )

    return "\n".join(lines)


def build_insights_text(data):
    lines = [""]

    # Tool usage
    tools = data.get("tool_summary", [])[:20]
    if tools:
        max_count = tools[0]["count"] if tools else 1
        lines.append(f"  {C_BOLD}Tool Usage (Top 20)[/]\n")
        lines.append(f"  {'Tool':<20} {'Calls':>6}  Bar")
        lines.append(f"  {'─' * 20} {'─' * 6}  {'─' * 20}")
        for tool in tools:
            bar = make_bar(tool["count"], max_count, 20)
            lines.append(f"  {tool['name']:<20} {tool['count']:>6}  {C_CYAN}{bar}[/]")

    # Storage
    insights = data.get("insights", {})
    storage = insights.get("storage", {})
    items = storage.get("items", [])
    if items:
        max_size = max(i["size_mb"] for i in items) or 1
        lines.append(f"\n  {C_BOLD}Storage ({storage.get('total_mb', 0):.1f} MB total)[/]\n")
        lines.append(f"  {'Item':<20} {'Size MB':>8}  Bar")
        lines.append(f"  {'─' * 20} {'─' * 8}  {'─' * 20}")
        for item in items:
            if item["size_mb"] > 0:
                bar = make_bar(item["size_mb"], max_size, 20)
                lines.append(f"  {item['name']:<20} {item['size_mb']:>8.2f}  {C_YELLOW}{bar}[/]")

    # Plugins
    plugins = insights.get("plugins", {})
    installed = plugins.get("installed", [])
    enabled = plugins.get("settings", {}).get("enabled_plugins", {})
    if installed:
        lines.append(f"\n  {C_BOLD}Installed Plugins ({len(installed)})[/]\n")
        lines.append(f"  {'Plugin':<25} {'Status':<8} {'Version':<14} {'Installed':<10}")
        lines.append(f"  {'─' * 25} {'─' * 8} {'─' * 14} {'─' * 10}")
        for p in installed:
            is_active = enabled.get(p["name"], False)
            dot = f"{C_GREEN}\u25cf Active[/] " if is_active else f"{C_MUTED}\u25cb Off[/]    "
            date = p.get("installed_at", "")[:10]
            lines.append(f"  {p['short_name']:<25} {dot} {p.get('version', ''):<14} {date:<10}")

    # Config
    settings = plugins.get("settings", {})
    active_count = sum(1 for v in enabled.values() if v)
    lines.append(f"\n  {C_BOLD}Configuration[/]\n")
    lines.append(f"  Permission Mode: {settings.get('permission_mode') or 'default'}")
    lines.append(f"  Auto-Updates:    {settings.get('auto_updates', 'unknown')}")
    lines.append(f"  Plugins:         {len(installed)} installed, {active_count} active")
    lines.append(f"  Total Storage:   {storage.get('total_mb', 0):.1f} MB")

    # Todos & file history
    todos = insights.get("todos", {})
    fh = insights.get("file_history", {})
    lines.append(f"\n  {C_BOLD}Todos & File History[/]\n")
    if todos.get("total", 0) > 0:
        rate = f"{todos['completed'] / todos['total'] * 100:.0f}%"
        lines.append(f"  Todos: {todos['total']} total, {todos['completed']} completed ({rate})")
    else:
        lines.append(f"  Todos: {todos.get('total', 0)} total across {todos.get('files', 0)} files")
    lines.append(
        f"  File History: {fh.get('total_files', 0)} files, "
        f"{fh.get('total_sessions', 0)} sessions, {fh.get('total_size_mb', 0):.1f} MB"
    )

    return "\n".join(lines)


# ── Textual App ───────────────────────────────────────────────────────────

DASHBOARD_CSS = """
Screen {
    background: $surface;
}
#kpi-bar {
    height: 3;
    padding: 0 1;
    background: $boost;
    color: $text;
}
TabbedContent {
    height: 1fr;
}
TabPane {
    padding: 0;
}
.tab-content {
    padding: 0 0;
}
"""


class DashboardApp(App):
    CSS = DASHBOARD_CSS
    TITLE = "Claude Code Stats"
    BINDINGS = [
        Binding("q", "quit", "Quit"),
        Binding("1", "tab_1", "Overview", show=True),
        Binding("2", "tab_2", "Tokens", show=True),
        Binding("3", "tab_3", "Activity", show=True),
        Binding("4", "tab_4", "Projects", show=True),
        Binding("5", "tab_5", "Sessions", show=True),
        Binding("6", "tab_6", "Plan", show=True),
        Binding("7", "tab_7", "Insights", show=True),
    ]

    def __init__(self, data: dict, limit: int = 20):
        super().__init__()
        self.data = data
        self.limit = limit

    def compose(self) -> ComposeResult:
        yield Header()
        yield Static(build_kpi_text(self.data), id="kpi-bar")
        with TabbedContent(
            "Overview",
            "Tokens",
            "Activity",
            "Projects",
            "Sessions",
            "Plan & Billing",
            "Insights",
        ):
            with TabPane("Overview"):
                yield VerticalScroll(Static(build_overview_text(self.data), classes="tab-content"))
            with TabPane("Tokens"):
                yield VerticalScroll(Static(build_tokens_text(self.data), classes="tab-content"))
            with TabPane("Activity"):
                yield VerticalScroll(Static(build_activity_text(self.data), classes="tab-content"))
            with TabPane("Projects"):
                yield VerticalScroll(Static(build_projects_text(self.data, self.limit), classes="tab-content"))
            with TabPane("Sessions"):
                yield VerticalScroll(Static(build_sessions_text(self.data, self.limit), classes="tab-content"))
            with TabPane("Plan & Billing"):
                yield VerticalScroll(Static(build_plan_text(self.data), classes="tab-content"))
            with TabPane("Insights"):
                yield VerticalScroll(Static(build_insights_text(self.data), classes="tab-content"))
        yield Footer()

    def _switch_tab(self, index: int) -> None:
        tabs = self.query_one(TabbedContent)
        tab_ids = list(tabs.query("TabPane"))
        if 0 <= index < len(tab_ids):
            tabs.active = tab_ids[index].id

    def action_tab_1(self) -> None:
        self._switch_tab(0)

    def action_tab_2(self) -> None:
        self._switch_tab(1)

    def action_tab_3(self) -> None:
        self._switch_tab(2)

    def action_tab_4(self) -> None:
        self._switch_tab(3)

    def action_tab_5(self) -> None:
        self._switch_tab(4)

    def action_tab_6(self) -> None:
        self._switch_tab(5)

    def action_tab_7(self) -> None:
        self._switch_tab(6)


# ── Non-interactive fallback (--print mode) ───────────────────────────────

def print_static(data, sections, limit):
    """Non-interactive output using rich console (for piping / --print)."""
    from rich.console import Console
    from rich.panel import Panel
    console = Console(force_terminal=True)

    console.print(build_kpi_text(data))

    if not sections:
        console.print(build_overview_text(data))
        console.print(f"\n  {C_MUTED}Use --all for full output, or run without --print for interactive tabs[/]\n")
        return

    dispatch = {
        "tokens": lambda: console.print(build_tokens_text(data)),
        "activity": lambda: console.print(build_activity_text(data)),
        "projects": lambda: console.print(build_projects_text(data, limit)),
        "sessions": lambda: console.print(build_sessions_text(data, limit)),
        "plan": lambda: console.print(build_plan_text(data)),
        "insights": lambda: console.print(build_insights_text(data)),
    }
    for s in sections:
        console.print(f"\n{'─' * 40} {s.upper()} {'─' * 40}")
        dispatch[s]()
    console.print()


# ── Main ──────────────────────────────────────────────────────────────────

VALID_SECTIONS = ["tokens", "activity", "projects", "sessions", "plan", "insights"]


def main():
    parser = argparse.ArgumentParser(description="CLI dashboard for Claude Code stats")
    parser.add_argument("--all", action="store_true", help="Show all sections (non-interactive)")
    parser.add_argument("--section", type=str, help=f"Comma-separated sections: {','.join(VALID_SECTIONS)}")
    parser.add_argument("--limit", type=int, default=20, help="Limit rows for sessions/projects (default: 20)")
    parser.add_argument("--json", action="store_true", help="Dump raw JSON to stdout")
    parser.add_argument("--no-refresh", action="store_true", help="Skip auto-running extract_stats.py")
    parser.add_argument("--print", action="store_true", dest="print_mode", help="Non-interactive output (no TUI)")
    args = parser.parse_args()

    maybe_refresh(args.no_refresh)

    if not DATA_FILE.exists():
        print("Error: dashboard_data.json not found. Run extract_stats.py first.")
        sys.exit(1)

    data = load_data()

    if args.json:
        print(json.dumps(data, indent=2, ensure_ascii=False))
        return

    # Non-interactive modes: --print, --all, --section
    if args.print_mode or args.all or args.section:
        sections = []
        if args.all:
            sections = VALID_SECTIONS
        elif args.section:
            sections = [s.strip() for s in args.section.split(",")]
            invalid = [s for s in sections if s not in VALID_SECTIONS]
            if invalid:
                print(f"Invalid sections: {', '.join(invalid)}")
                print(f"Valid: {', '.join(VALID_SECTIONS)}")
                sys.exit(1)
        print_static(data, sections, args.limit)
        return

    # Interactive TUI (default)
    app = DashboardApp(data, limit=args.limit)
    app.run()


if __name__ == "__main__":
    main()
