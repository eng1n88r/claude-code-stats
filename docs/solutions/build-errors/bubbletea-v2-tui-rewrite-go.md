---
title: "Bubble Tea v2 and Lipgloss v2 API Migration Gotchas"
date: 2026-03-05
category: build-errors
tags:
  - bubble-tea
  - bubble-tea-v2
  - lipgloss
  - lipgloss-v2
  - tui
  - go
  - cobra
  - windows
  - breaking-changes
severity: high
component: tui/rendering
symptom: >
  Compilation errors when using Bubble Tea v1 / Lipgloss v1 patterns with v2 modules —
  View() signature mismatch, KeyMsg type not found, viewport constructor arity errors,
  lipgloss.Color used as type instead of function. Plus runtime gotchas: stdout pollution,
  Cobra flag inheritance, and date overflow.
---

# Bubble Tea v2 and Lipgloss v2 API Migration Gotchas

Lessons learned from rewriting claude-dashboard from Python to Go with Bubble Tea v2, Lipgloss v2, and Bubbles v2.

## Problems and Solutions

### 1. View() Must Return `tea.View`, Not `string`

**Problem**: `View()` returning `string` causes a compile error. The v1 signature `View() string` no longer works in v2.

**Root Cause**: Bubble Tea v2 changed `View()` to return `tea.View`, a struct that carries content plus rendering directives (like alt-screen mode).

**Solution** (`internal/tui/app.go:171`):
```go
func (a App) View() tea.View {
    v := tea.NewView(b.String())
    v.AltScreen = true
    return v
}
```

**Key Insight**: Use `tea.NewView("content")` to construct the return value, and set `v.AltScreen = true` instead of calling `tea.WithAltScreen()` on the program.

---

### 2. `tea.KeyMsg` Replaced by `tea.KeyPressMsg`

**Problem**: Matching on `tea.KeyMsg` compiles but never fires — no key events are caught at runtime.

**Root Cause**: Bubble Tea v2 renamed `tea.KeyMsg` to `tea.KeyPressMsg`. The old type still exists but is no longer dispatched for ordinary key presses.

**Solution** (`internal/tui/app.go:130`):
```go
case tea.KeyPressMsg:
    switch msg.String() {
    case "q", "ctrl+c":
        return a, tea.Quit
    case "tab":
        // ...
    }
```

**Key Insight**: Switch on `tea.KeyPressMsg` (not `tea.KeyMsg`) to handle keyboard input in v2.

---

### 3. Viewport Constructor Uses Functional Options

**Problem**: `viewport.Model{Width: w, Height: h}` fails — fields are unexported in Bubbles v2.

**Root Cause**: Bubbles v2 viewport switched to a functional-options constructor pattern.

**Solution** (`internal/tui/app.go:116-124`):
```go
// Create
a.viewport = viewport.New(
    viewport.WithWidth(a.width),
    viewport.WithHeight(vpHeight),
)

// Resize
a.viewport.SetWidth(a.width)
a.viewport.SetHeight(vpHeight)
```

**Key Insight**: Use `viewport.New(viewport.WithWidth(w), viewport.WithHeight(h))` and `SetWidth()`/`SetHeight()` for updates.

---

### 4. `lipgloss.Color` Is a Function, Not a Type

**Problem**: Using `lipgloss.Color` as a struct field type fails to compile — it's a function in v2.

**Root Cause**: In Lipgloss v2, `lipgloss.Color()` is a function returning a `color.Color` value, not a type you can use for declarations.

**Solution** (`internal/tui/styles.go:8-22`):
```go
// Package-level color variables — function calls, not type constructors
var (
    colorPrimary = lipgloss.Color("99")
    colorAccent  = lipgloss.Color("212")
)
```

For data structures that need deferred color application (`internal/tui/views/tokens.go:61-71`):
```go
type tokenType struct {
    key   string
    label string
    color string  // store ANSI code as plain string
}
typeOrder := []tokenType{
    {"output", "Output", "82"},
    {"cache_write", "Cache Write", "208"},
}
// At render time:
style := lipgloss.NewStyle().Foreground(lipgloss.Color(tt.color))
```

**Key Insight**: Store ANSI color codes as `string` in data structures, call `lipgloss.Color(s)` at render time. Never use `lipgloss.Color` as a type.

---

### 5. Windows Console UTF-8 — `CP_UTF8` Not Exported

**Problem**: `windows.CP_UTF8` doesn't exist in `golang.org/x/sys/windows`. Box-drawing characters render as garbage on Windows.

**Root Cause**: The `x/sys/windows` package simply doesn't export this constant.

**Solution** (`internal/tui/windows.go`):
```go
//go:build windows

package tui

import "golang.org/x/sys/windows"

const cpUTF8 = 65001

func init() {
    _ = windows.SetConsoleOutputCP(cpUTF8)
    _ = windows.SetConsoleCP(cpUTF8)
}
```

**Key Insight**: Define `const cpUTF8 = 65001` locally in a `//go:build windows` file. The `init()` runs before `main()` and fixes all UTF-8 rendering.

---

### 6. Cobra `PersistentFlags` vs `Flags`

**Problem**: `--quiet` defined with `rootCmd.Flags()` is invisible to the `extract` subcommand. Running `claude-dashboard extract --quiet` errors with "unknown flag: --quiet".

**Root Cause**: Cobra's `Flags()` registers flags only on that specific command. Subcommands don't inherit them.

**Solution** (`cmd/claude-dashboard/main.go:42-45`):
```go
// Local to root only:
rootCmd.Flags().BoolVar(&flagJSON, "json", false, "Dump JSON to stdout")

// Shared with all subcommands:
rootCmd.PersistentFlags().BoolVar(&flagQuiet, "quiet", false, "Suppress progress output")
rootCmd.PersistentFlags().StringVar(&flagConfig, "config", "", "Path to config.json")
rootCmd.PersistentFlags().StringVar(&flagOutput, "output", "", "Output directory")
```

**Key Insight**: Use `PersistentFlags()` for any flag that subcommands need to inherit.

---

### 7. Date Overflow — `billing_day=31` in February

**Problem**: `time.Date(2026, 2, 31, ...)` silently normalizes to March 3rd. Billing period calculations span into the wrong month.

**Root Cause**: Go's `time.Date` normalizes day overflow instead of panicking. The Python version had the same bug.

**Solution** (`internal/extract/billing.go:182-189`):
```go
func clampedDate(year int, month time.Month, day int) time.Time {
    lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
    if day > lastDay {
        day = lastDay
    }
    return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}
```

**Key Insight**: `time.Date(year, month+1, 0, ...).Day()` gives the last day of `month` — use this to clamp before constructing dates.

---

### 8. Progress Output Leaking into `--json` Stdout

**Problem**: `fmt.Println("Processing sessions...")` in the extract package polluted stdout, making `--json | jq` fail.

**Root Cause**: All progress output defaulted to stdout. With `--json`, both JSON and progress text went to the same stream.

**Solution**: Change all progress output to stderr:
```go
// WRONG
fmt.Printf("Processing %d files...\n", n)

// RIGHT
fmt.Fprintf(os.Stderr, "Processing %d files...\n", n)
```

**Key Insight**: Reserve stdout exclusively for machine-readable output. All progress, logs, and TUI output go to stderr.

## Prevention Checklist

| Issue | Grep / Vet Check |
|---|---|
| BubbleTea v1 signatures | `grep 'tea\.KeyMsg\|View() string'` |
| Lipgloss Color as type | `grep 'lipgloss\.Color[^(]'` |
| Windows CP_UTF8 | `grep 'windows\.CP_UTF8'` |
| Stdout pollution | `cmd --json 2>/dev/null \| jq .` |
| Cobra local vs persistent | `grep '\.Flags()\..*Var'` — audit each |
| Unclamped billing day | `grep 'time\.Date('` — variable day without clamp |

## Test Cases

```go
// Date clamping
func TestClampedDate(t *testing.T) {
    cases := []struct {
        year, day int
        month     time.Month
        wantDay   int
    }{
        {2026, 31, time.January, 31},
        {2026, 31, time.February, 28},
        {2024, 31, time.February, 29},  // leap year
        {2026, 31, time.April, 30},
        {2026, 29, time.February, 28},
    }
    for _, tc := range cases {
        got := clampedDate(tc.year, tc.month, tc.day)
        if got.Day() != tc.wantDay {
            t.Errorf("clampedDate(%d, %v, %d).Day() = %d, want %d",
                tc.year, tc.month, tc.day, got.Day(), tc.wantDay)
        }
    }
}
```

## Related Documentation

- [Python TUI encoding and alignment fixes](../ui-bugs/claude-code-tui-dashboard-encoding-and-alignment-fixes.md) — the Windows encoding bugs that the Go rewrite eliminates
- [Go rewrite plan](../../plans/2026-03-05-refactor-go-bubbletea-rewrite-plan.md) — full implementation plan
- [Bubble Tea v2 upgrade guide](https://github.com/charmbracelet/bubbletea/blob/main/UPGRADE_GUIDE_V2.md)
