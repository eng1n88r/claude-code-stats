package tui

import (
	"fmt"

	"charm.land/lipgloss/v2"
	"github.com/AeternaLabsHQ/claude-code-stats/internal/extract"
	"github.com/AeternaLabsHQ/claude-code-stats/internal/tui/views"
)

// Non-interactive render functions for --all / --section output.

func RenderKPIText(data *extract.DashboardData) string {
	k := data.KPI
	return fmt.Sprintf(
		"%s  API Value: %s  Messages: %s  Sessions: %s  Output: %s  %s",
		lipgloss.NewStyle().Bold(true).Foreground(colorBright).Render("Claude Code Stats"),
		kpiValue.Foreground(colorOrange).Render(fmtCost(k.TotalCost)),
		kpiValue.Foreground(colorGreen).Render(fmt.Sprintf("%d", k.TotalMessages)),
		kpiValue.Foreground(colorCyan).Render(fmt.Sprintf("%d", k.TotalSessions)),
		kpiValue.Foreground(colorMagenta).Render(fmtTokens(k.TotalOutput)),
		dimText.Render(fmt.Sprintf("%s – %s", k.FirstSession, k.LastSession)),
	)
}

func RenderOverviewText(data *extract.DashboardData) string {
	return views.RenderOverview(data)
}

func RenderTokensText(data *extract.DashboardData, width int) string {
	return views.RenderTokens(data, width)
}

func RenderActivityText(data *extract.DashboardData, width int) string {
	return views.RenderActivity(data, width)
}

func RenderProjectsText(data *extract.DashboardData, limit int) string {
	return views.RenderProjects(data, limit)
}

func RenderSessionsText(data *extract.DashboardData, limit int) string {
	return views.RenderSessions(data, limit)
}

func RenderPlanText(data *extract.DashboardData) string {
	return views.RenderPlan(data)
}

func RenderInsightsText(data *extract.DashboardData) string {
	return views.RenderInsights(data)
}
