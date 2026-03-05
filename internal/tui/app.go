package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/viewport"
	"charm.land/lipgloss/v2"
	"github.com/AeternaLabsHQ/claude-code-stats/internal/extract"
	"github.com/AeternaLabsHQ/claude-code-stats/internal/tui/views"
)

var tabNames = []string{
	"Overview",
	"Tokens",
	"Activity",
	"Projects",
	"Sessions",
	"Plan & Billing",
	"Insights",
}

// dataReadyMsg is sent when async extraction completes.
type dataReadyMsg struct {
	data *extract.DashboardData
	err  error
}

// ExtractFunc is a function that performs data extraction.
type ExtractFunc func() (*extract.DashboardData, error)

type App struct {
	data       *extract.DashboardData
	activeTab  int
	viewport   viewport.Model
	spinner    spinner.Model
	width      int
	height     int
	limit      int
	ready      bool
	loading    bool
	extractFn  ExtractFunc
	loadingErr error
}

func NewApp(data *extract.DashboardData, limit int, extractFn ExtractFunc) App {
	s := spinner.New(spinner.WithSpinner(spinner.Dot))
	s.Style = lipgloss.NewStyle().Foreground(colorPrimary)

	app := App{
		data:    data,
		limit:   limit,
		spinner: s,
	}

	if data == nil && extractFn != nil {
		app.loading = true
		app.extractFn = extractFn
	}

	return app
}

func (a App) Init() tea.Cmd {
	if a.loading {
		return tea.Batch(a.spinner.Tick, a.startExtraction())
	}
	return nil
}

func (a App) startExtraction() tea.Cmd {
	fn := a.extractFn
	return func() tea.Msg {
		data, err := fn()
		return dataReadyMsg{data: data, err: err}
	}
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case dataReadyMsg:
		a.loading = false
		if msg.err != nil {
			a.loadingErr = msg.err
			return a, nil
		}
		a.data = msg.data
		if a.ready {
			a.viewport.SetContent(a.renderTab())
		}
		return a, nil

	case spinner.TickMsg:
		if a.loading {
			a.spinner, cmd = a.spinner.Update(msg)
			return a, cmd
		}

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height

		headerHeight := 5
		footerHeight := 2

		vpHeight := a.height - headerHeight - footerHeight
		if vpHeight < 1 {
			vpHeight = 1
		}

		if !a.ready {
			a.viewport = viewport.New(
				viewport.WithWidth(a.width),
				viewport.WithHeight(vpHeight),
			)
			a.ready = true
		} else {
			a.viewport.SetWidth(a.width)
			a.viewport.SetHeight(vpHeight)
		}
		if a.data != nil {
			a.viewport.SetContent(a.renderTab())
		}
		return a, nil

	case tea.KeyPressMsg:
		if a.loading {
			if msg.String() == "q" || msg.String() == "ctrl+c" {
				return a, tea.Quit
			}
			return a, nil
		}
		switch msg.String() {
		case "q", "ctrl+c":
			return a, tea.Quit
		case "1":
			a.activeTab = 0
		case "2":
			a.activeTab = 1
		case "3":
			a.activeTab = 2
		case "4":
			a.activeTab = 3
		case "5":
			a.activeTab = 4
		case "6":
			a.activeTab = 5
		case "7":
			a.activeTab = 6
		case "tab":
			a.activeTab = (a.activeTab + 1) % len(tabNames)
		case "shift+tab":
			a.activeTab = (a.activeTab - 1 + len(tabNames)) % len(tabNames)
		default:
			a.viewport, cmd = a.viewport.Update(msg)
			return a, cmd
		}
		a.viewport.SetContent(a.renderTab())
		a.viewport.GotoTop()
		return a, nil
	}

	a.viewport, cmd = a.viewport.Update(msg)
	return a, cmd
}

func (a App) View() tea.View {
	if a.loadingErr != nil {
		v := tea.NewView(fmt.Sprintf("\n  Error: %v\n\n  Press q to quit.", a.loadingErr))
		v.AltScreen = true
		return v
	}

	if a.loading || a.data == nil {
		content := fmt.Sprintf("\n  %s Extracting Claude Code usage data...\n\n  %s",
			a.spinner.View(),
			dimText.Render("Press q to quit"))
		v := tea.NewView(content)
		v.AltScreen = true
		return v
	}

	if !a.ready {
		v := tea.NewView("Initializing...")
		v.AltScreen = true
		return v
	}

	var b strings.Builder
	b.WriteString(a.renderKPI())
	b.WriteString("\n")
	b.WriteString(a.renderTabs())
	b.WriteString("\n")
	b.WriteString(a.viewport.View())
	b.WriteString("\n")
	b.WriteString(a.renderHelp())

	v := tea.NewView(b.String())
	v.AltScreen = true
	return v
}

func (a App) renderKPI() string {
	k := a.data.KPI

	cost := kpiValue.Foreground(colorOrange).Render(fmtCost(k.TotalCost))
	msgs := kpiValue.Foreground(colorGreen).Render(fmt.Sprintf("%d", k.TotalMessages))
	sess := kpiValue.Foreground(colorCyan).Render(fmt.Sprintf("%d", k.TotalSessions))
	output := kpiValue.Foreground(colorMagenta).Render(fmtTokens(k.TotalOutput))

	return fmt.Sprintf(" %s %s%s%s %s%s%s %s%s%s %s%s%s %s",
		lipgloss.NewStyle().Bold(true).Foreground(colorBright).Render("Claude Code Stats"),
		kpiSep.String(),
		kpiLabel.Render("API Value: "), cost,
		kpiSep.String(),
		kpiLabel.Render("Messages: "), msgs,
		kpiSep.String(),
		kpiLabel.Render("Sessions: "), sess,
		kpiSep.String(),
		kpiLabel.Render("Output: "), output,
		dimText.Render(fmt.Sprintf("  %s – %s", k.FirstSession, k.LastSession)),
	)
}

func (a App) renderTabs() string {
	var tabs []string
	for i, name := range tabNames {
		label := fmt.Sprintf(" %d %s ", i+1, name)
		if i == a.activeTab {
			tabs = append(tabs, activeTabStyle.Render(label))
		} else {
			tabs = append(tabs, tabStyle.Render(label))
		}
	}
	return " " + strings.Join(tabs, "")
}

func (a App) renderTab() string {
	switch a.activeTab {
	case 0:
		return views.RenderOverview(a.data)
	case 1:
		return views.RenderTokens(a.data, a.width)
	case 2:
		return views.RenderActivity(a.data, a.width)
	case 3:
		return views.RenderProjects(a.data, a.limit)
	case 4:
		return views.RenderSessions(a.data, a.limit)
	case 5:
		return views.RenderPlan(a.data)
	case 6:
		return views.RenderInsights(a.data)
	default:
		return ""
	}
}

func (a App) renderHelp() string {
	return helpStyle.Render(" 1-7 switch tabs • tab/shift+tab cycle • ↑/↓/pgup/pgdn scroll • q quit")
}

// Run starts the TUI application with pre-loaded data.
func Run(data *extract.DashboardData, limit int) error {
	p := tea.NewProgram(NewApp(data, limit, nil))
	_, err := p.Run()
	return err
}

// RunWithExtraction starts the TUI with a loading spinner while extraction runs async.
func RunWithExtraction(limit int, extractFn ExtractFunc) error {
	p := tea.NewProgram(NewApp(nil, limit, extractFn))
	_, err := p.Run()
	return err
}
