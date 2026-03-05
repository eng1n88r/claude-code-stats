package views

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/AeternaLabsHQ/claude-code-stats/internal/extract"
)

func RenderPlan(data *extract.DashboardData) string {
	var b strings.Builder

	title := lipgloss.NewStyle().Foreground(lipgloss.Color("99")).Bold(true)
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	green := lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	bright := lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Bold(true)

	cb := data.Plan.CurrentBilling
	if cb != nil {
		b.WriteString(title.Render("Current Billing Period"))
		b.WriteString("\n\n")

		// Progress bar
		pct := 0.0
		if cb.DaysTotal > 0 {
			pct = float64(cb.DaysElapsed) / float64(cb.DaysTotal) * 100
		}
		barWidth := 40
		filled := int(pct / 100 * float64(barWidth))
		if filled > barWidth {
			filled = barWidth
		}
		progressBar := ""
		for i := 0; i < filled; i++ {
			progressBar += "█"
		}
		for i := filled; i < barWidth; i++ {
			progressBar += "░"
		}

		b.WriteString(fmt.Sprintf("  %s %s  Day %d/%d (%d remaining)\n\n",
			bright.Render(cb.Plan),
			lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Render(progressBar),
			cb.DaysElapsed, cb.DaysTotal, cb.DaysRemaining))

		// Stats grid
		savingsStyle := green
		if cb.Savings < 0 {
			savingsStyle = red
		}

		b.WriteString(fmt.Sprintf("  %-20s %s\n", muted.Render("API Cost So Far:"), green.Render(fmtCost(cb.APICost))))
		b.WriteString(fmt.Sprintf("  %-20s %s\n", muted.Render("Projected:"), fmtCost(cb.ProjectedCost)))
		b.WriteString(fmt.Sprintf("  %-20s %s\n", muted.Render("Plan Cost:"), fmtCost(cb.PlanCostUSD)))
		b.WriteString(fmt.Sprintf("  %-20s %s\n", muted.Render("Savings:"), savingsStyle.Render(fmtCost(cb.Savings))))
		b.WriteString(fmt.Sprintf("  %-20s %.1fx\n", muted.Render("ROI:"), cb.ROIFactor))
		b.WriteString(fmt.Sprintf("  %-20s %d\n", muted.Render("Sessions:"), cb.Sessions))
		b.WriteString(fmt.Sprintf("  %-20s %d\n", muted.Render("Messages:"), cb.Messages))
		b.WriteString(fmt.Sprintf("  %-20s %s\n", muted.Render("Avg/Day:"), fmtCost(cb.CostPerDay)))
	}

	// Period Detail Table
	b.WriteString("\n")
	b.WriteString(title.Render("Period Detail"))
	b.WriteString("\n\n")

	periods := data.Plan.Periods
	if len(periods) > 0 {
		rows := make([][]string, 0, len(periods))
		for _, p := range periods {
			savings := fmtCost(p.Savings)
			if p.Savings >= 0 {
				savings = green.Render(savings)
			} else {
				savings = red.Render(savings)
			}

			rows = append(rows, []string{
				fmt.Sprintf("%s – %s", p.Start, p.End),
				p.Plan,
				fmt.Sprintf("%d (%d active)", p.TotalDays, p.DaysActive),
				fmtCost(p.APICost),
				fmtCost(p.PlanCostUSD),
				savings,
				fmt.Sprintf("%.1fx", p.ROIFactor),
				fmt.Sprintf("%d", p.Sessions),
				fmt.Sprintf("%d", p.Messages),
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
			Headers("Period", "Plan", "Days", "API Cost", "Plan Cost", "Savings", "ROI", "Sessions", "Msgs").
			Rows(rows...)

		b.WriteString(t.String())
	}

	// Totals
	b.WriteString("\n\n")
	savingsStyle := green
	if data.Plan.TotalSavings < 0 {
		savingsStyle = red
	}
	b.WriteString(fmt.Sprintf("  %s  API: %s  Plan: %s  Savings: %s  ROI: %.1fx\n",
		bright.Render("Total"),
		green.Render(fmtCost(data.Plan.TotalAPICost)),
		fmtCost(data.Plan.TotalPlanCost),
		savingsStyle.Render(fmtCost(data.Plan.TotalSavings)),
		data.Plan.OverallROI))

	return b.String()
}
