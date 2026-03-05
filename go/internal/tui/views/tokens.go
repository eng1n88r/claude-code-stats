package views

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/AeternaLabsHQ/claude-code-stats/internal/extract"
)

func RenderTokens(data *extract.DashboardData, width int) string {
	var b strings.Builder

	title := lipgloss.NewStyle().Foreground(lipgloss.Color("99")).Bold(true)
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	green := lipgloss.NewStyle().Foreground(lipgloss.Color("82"))

	// Model Detail Table
	b.WriteString(title.Render("Model Detail"))
	b.WriteString("\n\n")

	if len(data.ModelSummary) > 0 {
		rows := make([][]string, 0, len(data.ModelSummary))
		for _, m := range data.ModelSummary {
			rows = append(rows, []string{
				m.Model,
				fmtCost(m.Cost),
				fmtTokens(m.OutputTokens),
				fmtTokens(m.InputTokens),
				fmtTokens(m.CacheReadTokens),
				fmt.Sprintf("%d", m.Calls),
			})
		}

		t := table.New().
			Border(lipgloss.NormalBorder()).
			BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
			StyleFunc(func(row, col int) lipgloss.Style {
				if row == table.HeaderRow {
					return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99")).Padding(0, 1)
				}
				return lipgloss.NewStyle().Padding(0, 1)
			}).
			Headers("Model", "API Value", "Output", "Input", "Cache Read", "Calls").
			Rows(rows...)

		b.WriteString(t.String())
		b.WriteString("\n\n")
	}

	// Cost by Token Type
	b.WriteString(title.Render("API Value by Token Type"))
	b.WriteString("\n\n")

	totalTypeCost := 0.0
	for _, v := range data.CostByType {
		totalTypeCost += v
	}

	type tokenType struct {
		key   string
		label string
		color string
	}
	typeOrder := []tokenType{
		{"output", "Output", "82"},
		{"cache_write", "Cache Write", "208"},
		{"input", "Input", "87"},
		{"cache_read", "Cache Read", "243"},
	}

	barWidth := 30
	if width > 80 {
		barWidth = 40
	}

	for _, tt := range typeOrder {
		val := data.CostByType[tt.key]
		bar := makeBar(val, totalTypeCost, barWidth)
		style := lipgloss.NewStyle().Foreground(lipgloss.Color(tt.color))
		pct := ""
		if totalTypeCost > 0 {
			pct = fmt.Sprintf("%.0f%%", val/totalTypeCost*100)
		}
		b.WriteString(fmt.Sprintf("  %-12s %s %s %s\n",
			tt.label,
			style.Render(bar),
			green.Render(fmtCost(val)),
			muted.Render(pct)))
	}

	// Daily Costs (last 30 days)
	b.WriteString("\n")
	b.WriteString(title.Render("Daily API Value (last 30 days)"))
	b.WriteString("\n\n")

	costs := data.DailyCosts
	start := 0
	if len(costs) > 30 {
		start = len(costs) - 30
	}

	maxDayCost := 0.0
	for i := start; i < len(costs); i++ {
		if t, ok := costs[i]["total"].(float64); ok && t > maxDayCost {
			maxDayCost = t
		}
	}

	for i := start; i < len(costs); i++ {
		d, _ := costs[i]["date"].(string)
		t, _ := costs[i]["total"].(float64)
		bar := makeBar(t, maxDayCost, barWidth)
		b.WriteString(fmt.Sprintf("  %s %s %s\n",
			muted.Render(d),
			lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Render(bar),
			green.Render(fmtCost(t))))
	}

	return b.String()
}

func makeBar(val, maxVal float64, width int) string {
	if maxVal <= 0 {
		return ""
	}
	filled := int(val / maxVal * float64(width))
	if filled > width {
		filled = width
	}
	s := ""
	for i := 0; i < filled; i++ {
		s += "█"
	}
	for i := filled; i < width; i++ {
		s += "░"
	}
	return s
}
