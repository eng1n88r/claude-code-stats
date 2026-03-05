package views

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/AeternaLabsHQ/claude-code-stats/internal/extract"
)

func RenderInsights(data *extract.DashboardData) string {
	var b strings.Builder

	title := lipgloss.NewStyle().Foreground(lipgloss.Color("99")).Bold(true)
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	green := lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
	cyan := lipgloss.NewStyle().Foreground(lipgloss.Color("87"))

	// Tool Usage
	b.WriteString(title.Render("Tool Usage (Top 20)"))
	b.WriteString("\n\n")

	tools := data.ToolSummary
	showing := len(tools)
	if showing > 20 {
		showing = 20
	}

	maxCount := 0
	if showing > 0 {
		maxCount = tools[0].Count
	}

	for i := 0; i < showing; i++ {
		t := tools[i]
		bar := makeBar(float64(t.Count), float64(maxCount), 25)
		b.WriteString(fmt.Sprintf("  %-20s %s %d\n",
			t.Name,
			cyan.Render(bar),
			t.Count))
	}

	// Storage
	b.WriteString("\n")
	b.WriteString(title.Render(fmt.Sprintf("Storage (~/.claude) — %.1f MB", data.Insights.Storage.TotalMB)))
	b.WriteString("\n\n")

	maxSize := 0.0
	for _, item := range data.Insights.Storage.Items {
		if item.SizeMB > maxSize {
			maxSize = item.SizeMB
		}
	}
	for _, item := range data.Insights.Storage.Items {
		bar := makeBar(item.SizeMB, maxSize, 25)
		b.WriteString(fmt.Sprintf("  %-25s %s %.2f MB\n",
			item.Name,
			lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Render(bar),
			item.SizeMB))
	}

	// Plugins
	installed := data.Insights.Plugins.Installed
	if len(installed) > 0 {
		b.WriteString("\n")
		b.WriteString(title.Render("Installed Plugins"))
		b.WriteString("\n\n")

		enabled := data.Insights.Plugins.Settings.EnabledPlugins

		rows := make([][]string, 0, len(installed))
		for _, p := range installed {
			status := "○ Off"
			if enabled[p.Name] {
				status = "● Active"
			}
			date := p.InstalledAt
			if len(date) > 10 {
				date = date[:10]
			}
			rows = append(rows, []string{
				p.ShortName,
				status,
				p.Version,
				date,
			})
		}

		t := table.New().
			Border(lipgloss.NormalBorder()).
			BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
			StyleFunc(func(row, col int) lipgloss.Style {
				s := lipgloss.NewStyle().Padding(0, 1)
				if row == table.HeaderRow {
					return s.Bold(true).Foreground(lipgloss.Color("99"))
				}
				// Color the status column
				if col == 1 && row >= 0 {
					if row < len(rows) && strings.HasPrefix(rows[row][1], "●") {
						return s.Foreground(lipgloss.Color("82"))
					}
					return s.Foreground(lipgloss.Color("243"))
				}
				return s
			}).
			Headers("Plugin", "Status", "Version", "Installed").
			Rows(rows...)

		b.WriteString(t.String())
	}

	// Configuration
	settings := data.Insights.Plugins.Settings
	b.WriteString("\n\n")
	b.WriteString(title.Render("Configuration"))
	b.WriteString("\n\n")

	activeCount := 0
	for _, v := range settings.EnabledPlugins {
		if v {
			activeCount++
		}
	}

	b.WriteString(fmt.Sprintf("  %-20s %s\n", muted.Render("Permission Mode:"), settings.PermissionMode))
	b.WriteString(fmt.Sprintf("  %-20s %s\n", muted.Render("Auto-Updates:"), settings.AutoUpdates))
	b.WriteString(fmt.Sprintf("  %-20s %d\n", muted.Render("Plugins Installed:"), len(installed)))
	b.WriteString(fmt.Sprintf("  %-20s %s\n", muted.Render("Plugins Active:"), green.Render(fmt.Sprintf("%d", activeCount))))

	// Todos & File History
	b.WriteString("\n")
	b.WriteString(title.Render("File Snapshots & Todos"))
	b.WriteString("\n\n")

	fh := data.Insights.FileHistory
	b.WriteString(fmt.Sprintf("  %-25s %d\n", muted.Render("File Snapshots:"), fh.TotalFiles))
	b.WriteString(fmt.Sprintf("  %-25s %d\n", muted.Render("Sessions w/ Snapshots:"), fh.TotalSessions))
	b.WriteString(fmt.Sprintf("  %-25s %.1f MB\n", muted.Render("Snapshot Size:"), fh.TotalSizeMB))

	td := data.Insights.Todos
	if td.Total > 0 {
		rate := 0.0
		if td.Total > 0 {
			rate = float64(td.Completed) / float64(td.Total) * 100
		}
		b.WriteString(fmt.Sprintf("  %-25s %d (%d completed, %.0f%%)\n",
			muted.Render("Todos:"), td.Total, td.Completed, rate))
	}

	// Plan-mode plans
	plans := data.Insights.Plans
	if len(plans) > 0 {
		b.WriteString("\n")
		b.WriteString(title.Render("Plan-Mode Plans"))
		b.WriteString("\n\n")

		rows := make([][]string, 0, len(plans))
		for _, p := range plans {
			rows = append(rows, []string{
				p.Title,
				p.Created[:10],
				fmt.Sprintf("%d", p.Lines),
				fmt.Sprintf("%.1f", p.SizeKB),
			})
		}

		t := table.New().
			Border(lipgloss.NormalBorder()).
			BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
			StyleFunc(func(row, col int) lipgloss.Style {
				s := lipgloss.NewStyle().Padding(0, 1)
				if row == table.HeaderRow {
					return s.Bold(true).Foreground(lipgloss.Color("99"))
				}
				return s
			}).
			Headers("Title", "Created", "Lines", "KB").
			Rows(rows...)

		b.WriteString(t.String())
	}

	_ = cyan

	return b.String()
}
